package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"engram/internal/proto"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Ollama struct {
	client *http.Client
	host   string
}

func NewOllama(host string) *Ollama {
	return &Ollama{
		client: http.DefaultClient,
		host:   host,
	}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (o *Ollama) Generate(ctx context.Context, model string, events []*proto.Event) (string, error) {
	prompt := o.buildPrompt(events)

	reqBody, err := json.Marshal(ollamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	url := o.host + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call ollama api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama api returned non-200 status: %d %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}

func (o *Ollama) buildPrompt(events []*proto.Event) string {
	var sb strings.Builder
	sb.WriteString("You are a helpful AI assistant. This is the conversation history:\n")

	for _, event := range events {
		ts := event.Timestamp.AsTime().Format("2006-01-02 15:04:05")
		switch p := event.Payload.(type) {
		case *proto.Event_Created:
			sb.WriteString(fmt.Sprintf("[%s] SYSTEM: Goal: %s\n", ts, p.Created.GetGoal()))
		case *proto.Event_UserMessage:
			sb.WriteString(fmt.Sprintf("[%s] User: %s\n", ts, p.UserMessage.GetContent()))
		case *proto.Event_LlmResponse:
			sb.WriteString(fmt.Sprintf("[%s] Assistant: %s\n", ts, p.LlmResponse.GetContent()))
		case *proto.Event_Edited:
			sb.WriteString(fmt.Sprintf("[%s] SYSTEM INTERVENTION: %s\n", ts, p.Edited.Description))
		case *proto.Event_Paused:
			sb.WriteString(fmt.Sprintf("[%s] SYSTEM: Paused. Reason: %s\n", ts, p.Paused.Reason))
		case *proto.Event_Resumed:
			sb.WriteString(fmt.Sprintf("[%s] SYSTEM: Resumed.\n", ts))
		}
	}
	sb.WriteString(fmt.Sprintf("\nCurrent Time: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("Assistant: ")
	return sb.String()
}
