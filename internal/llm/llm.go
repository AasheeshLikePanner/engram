package llm

import (
	"context"
	ipb "engram/internal/proto"
)

type LLM interface {
	Generate(ctx context.Context, model string, events []*ipb.Event) (string, error)
}
