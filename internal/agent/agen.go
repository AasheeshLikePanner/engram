package agent

import (
	"context"
	ipb "engram/internal/proto"
	"engram/internal/store"
	pb "engram/proto/v1"
	"io"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Agent struct {
	pb.UnimplementedAgentServiceServer
	store  store.DurableStore
	logger *slog.Logger
}

func NewAgent(store store.DurableStore) *Agent {
	return &Agent{
		store:  store,
		logger: slog.Default().With("service", "agent"),
	}
}

func (a *Agent) CreateAgent(ctx context.Context, req *pb.CreateAgentRequest) (*pb.CreateAgentResponse, error) {
	if req.Goal == "" {
		return nil, status.Error(codes.InvalidArgument, "goal is required")
	}
	if req.Model == "" {
		return nil, status.Error(codes.InvalidArgument, "model is required")
	}

	agentID := uuid.New().String()
	now := timestamppb.Now()

	a.logger.Info("Creating agent", "agent_id", agentID, "model", req.Model)

	createdEvent := &ipb.Event{
		Sequence:  1,
		Timestamp: now,
		Payload: &ipb.Event_Created{
			Created: &ipb.AgentCreated{
				Goal:  req.Goal,
				Model: req.Model,
			},
		},
	}

	if err := a.store.AppendEvent(ctx, agentID, createdEvent); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create initial event: %v", err)
	}

	agent := &pb.Agent{
		AgentId:   agentID,
		Goal:      req.Goal,
		Status:    "waiting_for_step",
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  req.Metadata,
	}

	if err := a.store.SetAgent(ctx, agentID, agent); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save agent: %v", err)
	}

	meta := &ipb.AgentMetadata{
		AgentId:      agentID,
		Status:       "waiting_for_step",
		UpdatedAt:    now,
		HeadSequence: 1,
		CurrentModel: req.Model,
	}
	if err := a.store.SetMetadata(ctx, agentID, meta); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save metadata: %v", err)
	}

	return &pb.CreateAgentResponse{
		AgentId: agentID,
		Agent:   agent,
	}, nil
}

func (a *Agent) GetAgent(ctx context.Context, req *pb.GetAgentRequest) (*pb.Agent, error) {
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	agent, err := a.store.GetAgent(ctx, req.AgentId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "agent %s not found: %v", req.AgentId, err)
	}

	return agent, nil
}

func (a *Agent) ListAgents(ctx context.Context, req *pb.ListAgentsRequest) (*pb.ListAgentsResponse, error) {
	ids, err := a.store.ListAgentIDs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list agent ids: %v", err)
	}

	var agents []*pb.Agent
	for _, id := range ids {
		agent, err := a.store.GetAgent(ctx, id)
		if err != nil {
			continue
		}
		if req.StatusFilter != "" && agent.Status != req.StatusFilter {
			continue
		}
		agents = append(agents, agent)
	}

	return &pb.ListAgentsResponse{Agents: agents}, nil
}

func (a *Agent) DeleteAgent(ctx context.Context, req *pb.DeleteAgentRequest) (*emptypb.Empty, error) {
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	if err := a.store.DeleteAgent(ctx, req.AgentId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete agent: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func eventToServiceMessage(aid string, event *ipb.Event) *pb.ServerMessage {
	serverMsg := &pb.ServerMessage{
		AgentId:    aid,
		SequenceId: int64(event.Sequence),
		Timestamp:  event.Timestamp,
	}

	switch p := event.Payload.(type) {
	case *ipb.Event_UserMessage:
		serverMsg.Payload = &pb.ServerMessage_Status{
			Status: &pb.StatusUpdate{
				Status:  "user_input",
				Message: p.UserMessage.Content,
			},
		}
	case *ipb.Event_Thought:
		serverMsg.Payload = &pb.ServerMessage_Thought{Thought: &pb.Thought{Content: p.Thought.Content}}
	case *ipb.Event_LlmResponse:
		serverMsg.Payload = &pb.ServerMessage_Text{Text: &pb.TextOutput{Content: p.LlmResponse.Content}}
	case *ipb.Event_Error:
		serverMsg.Payload = &pb.ServerMessage_Error{Error: &pb.Error{Code: "INTERNAL", Message: p.Error.Message}}
	case *ipb.Event_FinalAnswer:
		serverMsg.Payload = &pb.ServerMessage_FinalAnswer{FinalAnswer: &pb.FinalAnswer{Content: p.FinalAnswer.Content}}
	case *ipb.Event_ToolCall:
		serverMsg.Payload = &pb.ServerMessage_ToolCall{ToolCall: &pb.ToolCall{
			ToolName:  p.ToolCall.Name,
			InputJson: p.ToolCall.Input,
			CallId:    p.ToolCall.CallId,
		}}
	case *ipb.Event_ToolResult:
		serverMsg.Payload = &pb.ServerMessage_ToolResult{ToolResult: &pb.ToolResult{
			CallId:  p.ToolResult.CallId,
			Output:  p.ToolResult.Output,
			IsError: p.ToolResult.IsError,
		}}
	case *ipb.Event_Observation:
		serverMsg.Payload = &pb.ServerMessage_Observation{Observation: &pb.Observation{Content: p.Observation.Content}}
	}
	return serverMsg
}

func (a *Agent) Chat(stream pb.AgentService_ChatServer) error {
	var agentID string
	var subStarted bool
	var wg sync.WaitGroup
	ctx := stream.Context()

	errChan := make(chan error, 1)
	sendChan := make(chan *pb.ServerMessage, 10)

	// Single sender goroutine to ensure thread-safety
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-sendChan:
				if !ok {
					return
				}
				if err := stream.Send(msg); err != nil {
					select {
					case errChan <- err:
					default:
					}
					return
				}
			}
		}
	}()

	// Function to start subscription
	startSub := func(aid string) {
		if subStarted {
			return
		}
		subStarted = true
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.store.SubscribeEvents(ctx, aid, func(event *ipb.Event) {
				serverMsg := eventToServiceMessage(aid, event)

				select {
				case sendChan <- serverMsg:
				case <-ctx.Done():
				}
			})
			if err != nil && err != context.Canceled {
				a.logger.Error("Subscription error", "agent_id", aid, "error", err)
			}
		}()
	}

	// Receiver loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				select {
				case errChan <- err:
				default:
				}
				return
			}

			if msg.AgentId == "" {
				continue
			}
			agentID = msg.AgentId
			startSub(agentID)

			seq, err := a.store.GetNextSequence(ctx, agentID)
			if err != nil {
				select {
				case errChan <- status.Errorf(codes.Internal, "failed to get next sequence: %v", err):
				default:
				}
				return
			}

			var event *ipb.Event
			switch p := msg.Payload.(type) {
			case *pb.ClientMessage_UserInput:
				event = &ipb.Event{
					Sequence:  seq,
					Timestamp: timestamppb.Now(),
					Payload: &ipb.Event_UserMessage{
						UserMessage: &ipb.UserMessage{
							Content: p.UserInput.Content,
						},
					},
				}
			case *pb.ClientMessage_Pause:
				a.store.UpdateStatus(ctx, agentID, "paused")
				event = &ipb.Event{
					Sequence:  seq,
					Timestamp: timestamppb.Now(),
					Payload: &ipb.Event_Paused{
						Paused: &ipb.AgentPaused{Reason: p.Pause.Reason},
					},
				}
			case *pb.ClientMessage_Resume:
				a.store.UpdateStatus(ctx, agentID, "waiting_for_step")
				event = &ipb.Event{
					Sequence:  seq,
					Timestamp: timestamppb.Now(),
					Payload: &ipb.Event_Resumed{
						Resumed: &ipb.AgentResumed{},
					},
				}
			case *pb.ClientMessage_Edit:
				event = &ipb.Event{
					Sequence:  seq,
					Timestamp: timestamppb.Now(),
					Payload: &ipb.Event_Edited{
						Edited: &ipb.AgentEdited{
							Patch:       p.Edit.Patch,
							Description: p.Edit.Description,
						},
					},
				}
			case *pb.ClientMessage_Cancel:
				a.store.UpdateStatus(ctx, agentID, "canceled")
				return
			default:
				continue
			}

			if err := a.store.AppendEvent(ctx, agentID, event); err != nil {
				select {
				case errChan <- status.Errorf(codes.Internal, "failed to append event: %v", err):
				default:
				}
				return
			}

			if meta, err := a.store.GetMetadata(ctx, agentID); err == nil {
				meta.HeadSequence = seq
				meta.UpdatedAt = timestamppb.Now()
				a.store.SetMetadata(ctx, agentID, meta)
			}

			if err := a.store.UpdateStatus(ctx, agentID, "waiting_for_step"); err != nil {
				select {
				case errChan <- status.Errorf(codes.Internal, "failed to update status: %v", err):
				default:
				}
				return
			}

			select {
			case sendChan <- &pb.ServerMessage{
				AgentId:   agentID,
				Timestamp: timestamppb.Now(),
				Payload: &pb.ServerMessage_Status{
					Status: &pb.StatusUpdate{
						Status:  "waiting",
						Message: "Message received and queued for processing",
					},
				},
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	var returnErr error
	select {
	case <-ctx.Done():
		returnErr = ctx.Err()
	case err := <-errChan:
		returnErr = err
	}

	close(sendChan)
	wg.Wait()
	return returnErr
}

func (a *Agent) SetAgentStatus(ctx context.Context, req *pb.SetAgentStatusRequest) (*pb.SetAgentStatusResponse, error) {
	if err := a.store.UpdateStatus(ctx, req.AgentId, req.Status); err != nil {
		return nil, err
	}
	return &pb.SetAgentStatusResponse{Status: req.Status}, nil
}

func (a *Agent) ControlAgent(ctx context.Context, req *pb.ControlAgentRequest) (*pb.ControlAgentResponse, error) {
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	agent, err := a.GetAgent(ctx, &pb.GetAgentRequest{AgentId: req.AgentId})
	if err != nil {
		return nil, err
	}

	var newStatus string
	switch req.Command.(type) {
	case *pb.ControlAgentRequest_Pause:
		newStatus = "paused"
	case *pb.ControlAgentRequest_Resume:
		newStatus = "running"
	case *pb.ControlAgentRequest_Cancel:
		newStatus = "canceled"
	case *pb.ControlAgentRequest_Retry:
		if agent.Status != "failed" {
			return nil, status.Errorf(codes.FailedPrecondition, "agent must be in 'failed' state to be retried")
		}
		newStatus = "waiting_for_step"
	default:
		return nil, status.Error(codes.InvalidArgument, "unknown control command")
	}

	if err := a.store.UpdateStatus(ctx, req.AgentId, newStatus); err != nil {
		return nil, err
	}

	return &pb.ControlAgentResponse{Status: newStatus}, nil
}

func (a *Agent) Replay(req *pb.ReplayRequest, stream pb.AgentService_ReplayServer) error {
	if req.AgentId == "" {
		return status.Error(codes.InvalidArgument, "agent_id is required")
	}

	events, err := a.store.LoadEvents(stream.Context(), req.AgentId, uint64(req.FromSequence))
	if err != nil {
		return status.Errorf(codes.Internal, "failed to load events: %v", err)
	}

	for _, event := range events {
		serverMsg := eventToServiceMessage(req.AgentId, event)
		if err := stream.Send(serverMsg); err != nil {
			return err
		}
	}

	return nil
}
