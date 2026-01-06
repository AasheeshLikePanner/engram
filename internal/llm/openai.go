package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"engram/internal/proto"
	"fmt"
	"io"
	"net/http"
)

type OpenAI struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

func NewOpenAI(apiKey, baseURL string) *OpenAI {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAI{
		client:  &http.Client{},
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (o *OpenAI) Generate(ctx context.Context, model string, events []*proto.Event) (string, error) {
	messages := o.buildMessages(events)

	reqBody, err := json.Marshal(openaiChatRequest{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal openai request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create openai request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call openai api: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai api returned non-200 status: %d %s", resp.StatusCode, string(bodyBytes))
	}

	var openaiResp openaiChatResponse
	if err := json.Unmarshal(bodyBytes, &openaiResp); err != nil {
		return "", fmt.Errorf("failed to decode openai response: %w", err)
	}

	if openaiResp.Error != nil {
		return "", fmt.Errorf("openai api error: %s", openaiResp.Error.Message)
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("openai api returned no choices")
	}

	return openaiResp.Choices[0].Message.Content, nil
}

func (o *OpenAI) buildMessages(events []*proto.Event) []chatMessage {
	var messages []chatMessage

	for _, event := range events {
		switch p := event.Payload.(type) {
		case *proto.Event_Created:
			messages = append(messages, chatMessage{
				Role:    "system",
				Content: fmt.Sprintf("You are a helpful AI assistant. Your goal is: %s. Use [TOOL_CALL:tool_name:input_json] format to call tools.", p.Created.GetGoal()),
			})
		case *proto.Event_UserMessage:
			messages = append(messages, chatMessage{
				Role:    "user",
				Content: p.UserMessage.GetContent(),
			})
		case *proto.Event_LlmResponse:
			messages = append(messages, chatMessage{
				Role:    "assistant",
				Content: p.LlmResponse.GetContent(),
			})
		case *proto.Event_ToolCall:
			messages = append(messages, chatMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("[TOOL_CALL:%s:%s]", p.ToolCall.Name, p.ToolCall.Input),
			})
		case *proto.Event_ToolResult:
			messages = append(messages, chatMessage{
				Role:    "user",
				Content: fmt.Sprintf("Tool Result (%s): %s", p.ToolResult.CallId, p.ToolResult.Output),
			})
		case *proto.Event_Error:
			messages = append(messages, chatMessage{
				Role:    "system",
				Content: fmt.Sprintf("Error occurred: %s", p.Error.Message),
			})
		}
	}

	return messages
}
