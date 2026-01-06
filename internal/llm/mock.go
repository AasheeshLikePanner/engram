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
		case *ipb.Event_ToolResult:
			hasIntervention = true
		case *ipb.Event_Edited:
			hasIntervention = true
		}
	}

	if strings.Contains(lastUserMsg, "What time is it") && !hasIntervention {
		return "[TOOL_CALL:clock:{}] Checking the time now.", nil
	}

	if strings.Contains(lastUserMsg, "What time did I say") {
		return fmt.Sprintf("You said: '%s'.", saidTime), nil
	}

	if hasIntervention {
		return "The current time is 12:00 PM according to the clock tool.", nil
	}

	return "Hello! I am a durable AI agent. How can I help you?", nil
}
