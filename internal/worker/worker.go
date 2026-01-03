package worker

import (
	"context"
	"engram/internal/llm"
	ipb "engram/internal/proto"
	"engram/internal/sandbox"
	"engram/internal/store"
	pb "engram/proto/v1"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WorkerConfig struct {
	HeartbeatInterval time.Duration
	StepTimeout       time.Duration
	MaxRetries        int
	RetryBaseDelay    time.Duration
}

func DefaultConfig() WorkerConfig {
	return WorkerConfig{
		HeartbeatInterval: 5 * time.Second,
		StepTimeout:       60 * time.Second,
		MaxRetries:        3,
		RetryBaseDelay:    1 * time.Second,
	}
}

type Worker struct {
	store     store.DurableStore
	llm       llm.LLM
	sandbox   sandbox.Sandbox
	workerID  string
	toolsPath string
	config    WorkerConfig
	stopCh    chan struct{}
}

func NewWorker(s store.DurableStore, l llm.LLM, id string, toolsPath string, cfg ...WorkerConfig) *Worker {
	config := DefaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	sb, _ := sandbox.NewWasmSandbox(context.Background())
	return &Worker{
		store:     s,
		llm:       l,
		sandbox:   sb,
		workerID:  id,
		toolsPath: toolsPath,
		config:    config,
		stopCh:    make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	// Register worker
	if err := w.store.RegisterWorker(ctx, w.workerID); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	// Start heartbeat goroutine
	go w.runHeartbeat(ctx)

	// Start scanner
	go w.runScanner(ctx)

	// Start dead worker detector
	go w.runDeadWorkerDetector(ctx)

	<-ctx.Done()
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
				fmt.Printf("Worker [%s]: Heartbeat failed: %v\n", w.workerID, err)
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
	agentIDs, err := w.store.ListAgentIDs(ctx)
	if err != nil {
		return
	}

	for _, agentID := range agentIDs {
		meta, err := w.store.GetMetadata(ctx, agentID)
		if err != nil {
			continue
		}

		if meta.Status == "waiting_for_step" && meta.RunningOnWorker == "" {
			go w.ClaimAndProcess(ctx, agentID)
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

	fmt.Printf("Worker [%s]: Found %d dead workers, recovering tasks...\n", w.workerID, len(deadWorkers))

	agentIDs, _ := w.store.ListAgentIDs(ctx)
	for _, agentID := range agentIDs {
		meta, err := w.store.GetMetadata(ctx, agentID)
		if err != nil {
			continue
		}

		for _, deadWorker := range deadWorkers {
			if meta.RunningOnWorker == deadWorker && meta.Status == "processing" {
				fmt.Printf("Worker [%s]: Recovering task %s from dead worker %s\n", w.workerID, agentID, deadWorker)
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

	fmt.Printf("Worker [%s]: Claimed agent %s for processing\n", w.workerID, agentID)
	w.ExecuteStepWithRetry(ctx, meta)
}

func (w *Worker) ExecuteStepWithRetry(ctx context.Context, meta *ipb.AgentMetadata) {
	var lastErr error

	for attempt := 0; attempt <= w.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * w.config.RetryBaseDelay
			fmt.Printf("Worker [%s]: Retry %d for agent %s after %v\n", w.workerID, attempt, meta.AgentId, delay)
			time.Sleep(delay)
		}

		stepCtx, cancel := context.WithTimeout(ctx, w.config.StepTimeout)
		err := w.ExecuteStep(stepCtx, meta)
		cancel()

		if err == nil {
			return
		}

		lastErr = err
		if ctx.Err() != nil {
			break
		}

		fmt.Printf("Worker [%s]: Step failed for %s: %v\n", w.workerID, meta.AgentId, err)
	}

	fmt.Printf("Worker [%s]: All retries exhausted for %s: %v\n", w.workerID, meta.AgentId, lastErr)
	w.recordError(ctx, meta.AgentId, fmt.Sprintf("Max retries exceeded: %v", lastErr))
}

func (w *Worker) ExecuteStep(ctx context.Context, meta *ipb.AgentMetadata) error {
	agentID := meta.AgentId

	events, err := w.store.LoadEvents(ctx, agentID, 0)
	if err != nil {
		return fmt.Errorf("failed to load events: %w", err)
	}

	fmt.Printf("Worker [%s]: Replaying %d events for agent %s\n", w.workerID, len(events), agentID)

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

	fmt.Printf("Worker [%s]: Got LLM response for %s\n", w.workerID, agentID)

	nextSeq, _ := w.store.GetNextSequence(ctx, agentID)
	var responseEvent *ipb.Event
	newStatus := "waiting_for_input"

	if strings.HasPrefix(llmResponse, "[TOOL_CALL]") {
		fmt.Printf("Worker [%s]: Tool call detected for %s. Pausing for HITL.\n", w.workerID, agentID)
		responseEvent = &ipb.Event{
			Sequence:  nextSeq,
			Timestamp: timestamppb.Now(),
			Payload: &ipb.Event_ToolCall{
				ToolCall: &ipb.ToolCall{
					Name:   "calculator",
					Input:  `{"expression": "2+2"}`,
					CallId: uuid.New().String(),
				},
			},
		}
		newStatus = "paused"
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

	fmt.Printf("Worker [%s]: Step completed for agent %s\n", w.workerID, agentID)
	return nil
}

func (w *Worker) executeToolCall(ctx context.Context, meta *ipb.AgentMetadata, tc *ipb.ToolCall, events []*ipb.Event) error {
	agentID := meta.AgentId

	// Idempotency check
	existingResult, found, _ := w.store.GetToolResult(ctx, tc.CallId)
	if found {
		fmt.Printf("Worker [%s]: Using cached result for tool %s\n", w.workerID, tc.CallId)
		return w.appendToolResult(ctx, agentID, tc.CallId, existingResult, false)
	}

	fmt.Printf("Worker [%s]: Executing tool %s for %s\n", w.workerID, tc.Name, agentID)

	// Load WASM binary from toolsPath
	wasmPath := fmt.Sprintf("%s/%s.wasm", w.toolsPath, tc.Name)
	wasmBinary, err := os.ReadFile(wasmPath)
	if err != nil {
		fmt.Printf("Worker [%s]: Error reading tool binary %s: %v\n", w.workerID, wasmPath, err)
		return w.appendToolResult(ctx, agentID, tc.CallId, fmt.Sprintf("Error: Tool binary not found for %s", tc.Name), true)
	}

	result, err := w.sandbox.Execute(ctx, wasmBinary, []string{tc.Name, tc.Input})
	isError := err != nil
	if isError {
		result = err.Error()
	}

	fmt.Printf("Worker [%s]: Sandboxed execution result for %s: %q (err=%v)\n", w.workerID, tc.Name, result, err)

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
