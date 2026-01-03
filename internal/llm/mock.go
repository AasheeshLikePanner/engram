package llm

import (
	"context"
	ipb "engram/internal/proto"
	"strings"
)

type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, model string, events []*ipb.Event) (string, error) {
	// Simple deterministic logic for verification
	var hasIntervention bool
	var lastUserMsg string

	for _, e := range events {
		switch p := e.Payload.(type) {
		case *ipb.Event_UserMessage:
			lastUserMsg = p.UserMessage.Content
		case *ipb.Event_Edited:
			hasIntervention = true
		}
	}

	if strings.Contains(lastUserMsg, "[TOOL_CALL]") && !hasIntervention {
		return "[TOOL_CALL] I need to use the calculator.", nil
	}

	if hasIntervention {
		return "I see the system intervention. Based on the tool result and the edit, the answer is 4.", nil
	}

	return "Hello! I am a durable AI agent. How can I help you?", nil
}
