package vertexai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
)

// modelGardenChat sends a request to a Model Garden deployed endpoint using
// the OpenAI-compatible Chat Completions API exposed by Google's vLLM/SGLang
// serving containers.
//
// For dedicated endpoints, the model name is omitted from the request body —
// the endpoint itself serves a single deployed model. The configured model
// string is still echoed back on the response for DecisionBox logging and
// debug-log correlation.
func (p *VertexAIProvider) modelGardenChat(ctx context.Context, req gollm.ChatRequest) (*gollm.ChatResponse, error) {
	messages := make([]openaiMessage, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, openaiMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, msg := range req.Messages {
		messages = append(messages, openaiMessage{Role: msg.Role, Content: msg.Content})
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	body := openaiRequest{
		Messages:  messages,
		MaxTokens: maxTokens,
	}
	if req.Temperature > 0 {
		temp := req.Temperature
		body.Temperature = &temp
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("vertex-ai/modelgarden: failed to marshal request: %w", err)
	}

	url := p.endpointURL + "/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("vertex-ai/modelgarden: failed to create request: %w", err)
	}

	token, err := p.auth.token(ctx)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("vertex-ai/modelgarden: request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("vertex-ai/modelgarden: failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vertex-ai/modelgarden: API error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	var resp openaiResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("vertex-ai/modelgarden: failed to parse response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("vertex-ai/modelgarden: empty response (no choices)")
	}

	return &gollm.ChatResponse{
		Content:    resp.Choices[0].Message.Content,
		Model:      req.Model,
		StopReason: resp.Choices[0].FinishReason,
		Usage: gollm.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}, nil
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiRequest struct {
	Messages    []openaiMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
