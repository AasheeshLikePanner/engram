package worker

import (
	"context"
	"encoding/json"
	"engram/internal/llm"
	ipb "engram/internal/proto"
	"engram/internal/sandbox"
	"engram/internal/store"
	pb "engram/proto/v1"
	"fmt"
	"log/slog"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WorkerConfig struct {
	HeartbeatInterval  time.Duration
	StepTimeout        time.Duration
	MaxRetries         int
	RetryBaseDelay     time.Duration
	MaxConcurrentTasks int
}

func DefaultConfig() WorkerConfig {
	return WorkerConfig{
		HeartbeatInterval:  5 * time.Second,
		StepTimeout:        60 * time.Second,
		MaxRetries:         3,
		RetryBaseDelay:     1 * time.Second,
		MaxConcurrentTasks: 10,
	}
}

type Worker struct {
	store     store.DurableStore
	llm       llm.LLM
	sandbox   sandbox.Sandbox
	workerID  string
	toolsPath string
	config    WorkerConfig
	wg        sync.WaitGroup
	semaphore chan struct{}
	logger    *slog.Logger
}

func NewWorker(s store.DurableStore, l llm.LLM, id string, toolsPath string, cfg ...WorkerConfig) *Worker {
	config := DefaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	sb, _ := sandbox.NewWasmSandbox(context.Background())
	logger := slog.Default().With("worker_id", id)

	return &Worker{
		store:     s,
		llm:       l,
		sandbox:   sb,
		workerID:  id,
		toolsPath: toolsPath,
		config:    config,
		semaphore: make(chan struct{}, config.MaxConcurrentTasks),
		logger:    logger,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	w.logger.Info("Starting worker")

	// Register worker
	if err := w.store.RegisterWorker(ctx, w.workerID); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	// Start heartbeat goroutine
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.runHeartbeat(ctx)
	}()

	// Start scanner
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.runScanner(ctx)
	}()

	// Start dead worker detector
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.runDeadWorkerDetector(ctx)
	}()

	<-ctx.Done()
	w.logger.Info("Shutting down worker, waiting for active tasks...")
	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Wait for goroutines to finish
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("Graceful shutdown complete")
	case <-shutdownCtx.Done():
		w.logger.Warn("Shutdown timed out, some tasks may have been interrupted")
	}

	return w.store.UnregisterWorker(context.Background(), w.workerID)
}

func (w *Worker) runHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(w.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.store.Heartbeat(ctx, w.workerID); err != nil {
				w.logger.Error("Heartbeat failed", "error", err)
			}
		}
	}
}

func (w *Worker) runScanner(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.Scan(ctx)
		}
	}
}

func (w *Worker) Scan(ctx context.Context) {
	agentIDs, err := w.store.ListPendingAgents(ctx)
	if err != nil {
		w.logger.Error("Failed to list pending agents", "error", err)
		return
	}

	for _, agentID := range agentIDs {
		select {
		case <-ctx.Done():
			return
		case w.semaphore <- struct{}{}:
			w.wg.Add(1)
			go func(aid string) {
				defer w.wg.Done()
				defer func() { <-w.semaphore }()
				w.ClaimAndProcess(ctx, aid)
			}(agentID)
		default:
			// Semaphore full, skip for now
			return
		}
	}
}

func (w *Worker) runDeadWorkerDetector(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.recoverDeadWorkerTasks(ctx)
		}
	}
}

func (w *Worker) recoverDeadWorkerTasks(ctx context.Context) {
	deadWorkers, err := w.store.GetDeadWorkers(ctx, 30)
	if err != nil {
		return
	}

	if len(deadWorkers) == 0 {
		return
	}

	w.logger.Info("Found dead workers, recovering tasks", "count", len(deadWorkers))

	agentIDs, _ := w.store.ListAgentIDs(ctx)
	for _, agentID := range agentIDs {
		meta, err := w.store.GetMetadata(ctx, agentID)
		if err != nil {
			continue
		}

		for _, deadWorker := range deadWorkers {
			if meta.RunningOnWorker == deadWorker && meta.Status == "processing" {
				w.logger.Info("Recovering task from dead worker", "agent_id", agentID, "dead_worker", deadWorker)
				w.store.UpdateStatus(ctx, agentID, "waiting_for_step")
			}
		}
	}

	for _, deadWorker := range deadWorkers {
		w.store.UnregisterWorker(ctx, deadWorker)
	}
}

func (w *Worker) ClaimAndProcess(ctx context.Context, agentID string) {
	meta, err := w.store.AtomicClaim(ctx, agentID, w.workerID)
	if err != nil {
		return
	}

	if meta == nil {
		return
	}

	w.logger.Info("Claimed agent for processing", "agent_id", agentID)
	w.ExecuteStepWithRetry(ctx, meta)
}

func (w *Worker) ExecuteStepWithRetry(ctx context.Context, meta *ipb.AgentMetadata) {
	var lastErr error

	for attempt := 0; attempt <= w.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * w.config.RetryBaseDelay
			w.logger.Info("Retrying step", "agent_id", meta.AgentId, "attempt", attempt, "delay", delay)
			
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		}

		stepCtx, cancel := context.WithTimeout(ctx, w.config.StepTimeout)
		err := w.ExecuteStep(stepCtx, meta)
		cancel()

		if err == nil {
			return
		}

		lastErr = err
		if ctx.Err() != nil {
			return
		}

		w.logger.Error("Step failed", "agent_id", meta.AgentId, "error", err)
	}

	w.logger.Error("All retries exhausted", "agent_id", meta.AgentId, "error", lastErr)
	w.recordError(ctx, meta.AgentId, fmt.Sprintf("Max retries exceeded: %v", lastErr))
}

var toolCallRegex = regexp.MustCompile(`\[TOOL_CALL:([^:]+):?([\s\S]*?)\]`)

func (w *Worker) ExecuteStep(ctx context.Context, meta *ipb.AgentMetadata) error {
	agentID := meta.AgentId

	// Load only the last 50 events for context to prevent overflow
	// In a real production system, you might want to load more or use a summarization strategy
	allEvents, err := w.store.LoadEvents(ctx, agentID, 0)
	if err != nil {
		return fmt.Errorf("failed to load events: %w", err)
	}

	windowSize := 50
	events := allEvents
	if len(allEvents) > windowSize {
		events = allEvents[len(allEvents)-windowSize:]
		w.logger.Debug("Truncated events for context", "agent_id", agentID, "original", len(allEvents), "window", windowSize)
	}

	w.logger.Debug("Replaying events", "agent_id", agentID, "count", len(events))

	// Check for pending tool call that needs execution
	if len(events) > 0 {
		lastEvent := events[len(events)-1]
		if tc := lastEvent.GetToolCall(); tc != nil {
			return w.executeToolCall(ctx, meta, tc, events)
		}
	}

	// Generate LLM response
	llmResponse, err := w.llm.Generate(ctx, meta.CurrentModel, events)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	w.logger.Info("Got LLM response", "agent_id", agentID)

	nextSeq, _ := w.store.GetNextSequence(ctx, agentID)
	var responseEvent *ipb.Event
	newStatus := "waiting_for_input"

	// Robust tool call parsing using Regex
	match := toolCallRegex.FindStringSubmatch(llmResponse)
	if len(match) > 1 {
		toolName := strings.TrimSpace(match[1])
		toolInput := strings.TrimSpace(match[2])
		if toolInput == "" {
			toolInput = "{}"
		}

		// Validate JSON if present
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(toolInput), &js); err != nil {
			w.logger.Warn("LLM provided invalid JSON for tool", "input", toolInput)
		}

		w.logger.Info("Tool call detected", "agent_id", agentID, "tool", toolName)
		responseEvent = &ipb.Event{
			Sequence:  nextSeq,
			Timestamp: timestamppb.Now(),
			Payload: &ipb.Event_ToolCall{
				ToolCall: &ipb.ToolCall{
					Name:   toolName,
					Input:  toolInput,
					CallId: uuid.New().String(),
				},
			},
		}
		newStatus = "waiting_for_step"
	} else {
		responseEvent = &ipb.Event{
			Sequence:  nextSeq,
			Timestamp: timestamppb.Now(),
			Payload: &ipb.Event_LlmResponse{
				LlmResponse: &ipb.LLMResponse{Content: llmResponse},
			},
		}
	}

	if err := w.store.AppendEvent(ctx, agentID, responseEvent); err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}

	if err := w.store.UpdateStatus(ctx, agentID, newStatus); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	w.logger.Info("Step completed", "agent_id", agentID)
	return nil
}

func (w *Worker) executeToolCall(ctx context.Context, meta *ipb.AgentMetadata, tc *ipb.ToolCall, events []*ipb.Event) error {
	agentID := meta.AgentId

	// Idempotency check
	existingResult, found, _ := w.store.GetToolResult(ctx, tc.CallId)
	if found {
		w.logger.Info("Using cached tool result", "call_id", tc.CallId)
		return w.appendToolResult(ctx, agentID, tc.CallId, existingResult, false)
	}

	w.logger.Info("Executing tool", "agent_id", agentID, "tool", tc.Name)

	// Load WASM binary from toolsPath
	wasmPath := fmt.Sprintf("%s/%s.wasm", w.toolsPath, tc.Name)
	wasmBinary, err := os.ReadFile(wasmPath)
	if err != nil {
		w.logger.Error("Tool binary not found", "path", wasmPath)
		return w.appendToolResult(ctx, agentID, tc.CallId, fmt.Sprintf("Error: Tool binary not found for %s", tc.Name), true)
	}

	result, err := w.sandbox.Execute(ctx, wasmBinary, []string{tc.Name, tc.Input})
	isError := err != nil
	if isError {
		result = err.Error()
	}

	w.logger.Info("Tool execution complete", "agent_id", agentID, "tool", tc.Name, "is_error", isError)

	// Cache result for idempotency
	w.store.SetToolResult(ctx, tc.CallId, result)

	return w.appendToolResult(ctx, agentID, tc.CallId, result, isError)
}

func (w *Worker) appendToolResult(ctx context.Context, agentID, callID, result string, isError bool) error {
	nextSeq, _ := w.store.GetNextSequence(ctx, agentID)

	resultEvent := &ipb.Event{
		Sequence:  nextSeq,
		Timestamp: timestamppb.Now(),
		Payload: &ipb.Event_ToolResult{
			ToolResult: &ipb.ToolResult{
				CallId:  callID,
				Output:  result,
				IsError: isError,
			},
		},
	}

	if err := w.store.AppendEvent(ctx, agentID, resultEvent); err != nil {
		return err
	}

	return w.store.UpdateStatus(ctx, agentID, "waiting_for_step")
}

func (w *Worker) recordError(ctx context.Context, agentID, message string) {
	nextSeq, _ := w.store.GetNextSequence(ctx, agentID)

	errorEvent := &ipb.Event{
		Sequence:  nextSeq,
		Timestamp: timestamppb.Now(),
		Payload: &ipb.Event_Error{
			Error: &ipb.Error{Message: message},
		},
	}

	w.store.AppendEvent(ctx, agentID, errorEvent)
	w.store.UpdateStatus(ctx, agentID, "failed")
}

// Unused imports placeholder
var _ = proto.Marshal
var _ = pb.Agent{}
