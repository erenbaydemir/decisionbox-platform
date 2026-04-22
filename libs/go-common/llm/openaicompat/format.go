// Package openaicompat provides shared types and helpers for LLM providers
// that speak the OpenAI /chat/completions request/response schema.
//
// Consumers:
//   - providers/llm/openai (OpenAI direct and OpenAI-compatible APIs)
//   - providers/llm/azure-foundry (Azure AI Foundry for OpenAI-family deployments)
//   - providers/llm/bedrock (Qwen models on AWS Bedrock)
//
// Only the request and response bodies are shared here. Transport (HTTP POST vs
// AWS Bedrock InvokeModel), authentication (Bearer token vs api-key header vs
// AWS SigV4), and endpoint URLs stay in each provider.
package openaicompat

import (
	"encoding/json"
	"fmt"

	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
)

// Message is a single chat message in the OpenAI schema.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request is the OpenAI chat completions request body.
// Model is omitempty because some transports (e.g. AWS Bedrock InvokeModel)
// carry the model identifier outside the body.
type Request struct {
	Model       string    `json:"model,omitempty"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Choice is a single completion choice in the response.
type Choice struct {
	Index        int     `json:"index,omitempty"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage reports token consumption for a single call.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

// Response is the OpenAI chat completions response body.
type Response struct {
	ID      string    `json:"id,omitempty"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	Error   *APIError `json:"error,omitempty"`
}

// APIError is a structured error object returned inside the response body.
// Providers that receive HTTP-level errors (non-200 responses) typically surface
// those separately; this type represents body-level errors that may appear on
// 200 or on error responses where the body is still a Response shape.
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// Error implements the error interface. The rendered form includes the type and
// message so callers that wrap it with a provider prefix still produce a useful
// string.
func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("%s: %s", e.Type, e.Message)
	}
	return e.Message
}

// BuildRequestBody marshals a gollm.ChatRequest into an OpenAI-compatible JSON
// body. Pass model to embed it in the body; pass "" to omit it (useful when the
// transport supplies the model identifier elsewhere, e.g., Bedrock's ModelId).
// A non-empty SystemPrompt is prepended as the first message with role="system".
func BuildRequestBody(req gollm.ChatRequest, model string) ([]byte, error) {
	messages := make([]Message, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, Message{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.Messages {
		messages = append(messages, Message{Role: m.Role, Content: m.Content})
	}

	body := Request{
		Model:    model,
		Messages: messages,
	}
	if req.MaxTokens > 0 {
		body.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		body.Temperature = req.Temperature
	}

	return json.Marshal(body)
}

// ParseResponseBody decodes an OpenAI-compatible response body into a
// gollm.ChatResponse.
//
// Returns an error that is a *APIError when the body contains a structured
// error object; callers can errors.As to inspect the error code if needed.
// Returns a plain error if the body is malformed or contains no choices.
func ParseResponseBody(data []byte) (*gollm.ChatResponse, error) {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("response contained no choices")
	}

	choice := resp.Choices[0]
	return &gollm.ChatResponse{
		Content:    choice.Message.Content,
		Model:      resp.Model,
		StopReason: choice.FinishReason,
		Usage: gollm.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}, nil
}

// ExtractAPIError returns the embedded *APIError if the body matches the
// OpenAI-compatible error shape; otherwise returns nil. Intended for providers
// that receive an HTTP-level error and want to produce a structured message
// when the body is JSON.
func ExtractAPIError(data []byte) *APIError {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil
	}
	return resp.Error
}
