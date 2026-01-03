package llm

import (
	"context"
	ipb "engram/internal/proto"
	"fmt"
	"strings"
)

type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, model string, events []*ipb.Event) (string, error) {
	// Simple deterministic logic for verification
	var hasIntervention bool
	var lastUserMsg string
	var saidTime string

	for _, e := range events {
		switch p := e.Payload.(type) {
		case *ipb.Event_UserMessage:
			lastUserMsg = p.UserMessage.Content
			if strings.Contains(lastUserMsg, "It is exactly") {
				saidTime = lastUserMsg
			}
		case *ipb.Event_Edited:
			hasIntervention = true
		}
	}

	if strings.Contains(lastUserMsg, "What time did I say") {
		return fmt.Sprintf("You said: '%s'. The engine shows events happened then, but replayed now.", saidTime), nil
	}

	if strings.Contains(lastUserMsg, "[CLOCK_TOOL]") && !hasIntervention {
		return "[TOOL_CALL] I need to check the time.", nil
	}

	if hasIntervention {
		return "I see the system intervention. Based on the tool result and the edit, the answer is 4.", nil
	}

	return "Hello! I am a durable AI agent. How can I help you?", nil
}
