package store

import (
	"context"
	ipb "engram/internal/proto"
	pb "engram/proto/v1"
)

type EventStore interface {
	AppendEvent(ctx context.Context, agentID string, event *ipb.Event) error
	LoadEvents(ctx context.Context, agentID string, fromSeq uint64) ([]*ipb.Event, error)
	GetNextSequence(ctx context.Context, agentID string) (uint64, error)
	SubscribeEvents(ctx context.Context, agentID string, handler func(*ipb.Event)) error
}

type MetadataStore interface {
	GetMetadata(ctx context.Context, agentID string) (*ipb.AgentMetadata, error)
	SetMetadata(ctx context.Context, agentID string, meta *ipb.AgentMetadata) error
	GetAgent(ctx context.Context, agentID string) (*pb.Agent, error)
	SetAgent(ctx context.Context, agentID string, agent *pb.Agent) error
	DeleteAgent(ctx context.Context, agentID string) error
	AtomicClaim(ctx context.Context, agentID, workerID string) (*ipb.AgentMetadata, error)
	UpdateStatus(ctx context.Context, agentID, status string) error
}

type IdempotencyStore interface {
	GetToolResult(ctx context.Context, key string) (string, bool, error)
	SetToolResult(ctx context.Context, key, result string) error
}

type WorkerRegistry interface {
	RegisterWorker(ctx context.Context, workerID string) error
	Heartbeat(ctx context.Context, workerID string) error
	GetDeadWorkers(ctx context.Context, timeout int64) ([]string, error)
	UnregisterWorker(ctx context.Context, workerID string) error
}

type DurableStore interface {
	EventStore
	MetadataStore
	IdempotencyStore
	WorkerRegistry
	ListAgentIDs(ctx context.Context) ([]string, error)
	Close() error
}
