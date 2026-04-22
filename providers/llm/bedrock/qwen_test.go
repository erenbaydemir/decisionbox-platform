package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
	"github.com/decisionbox-io/decisionbox/libs/go-common/llm/openaicompat"
)

func TestIsQwen(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"qwen.qwen3-next-80b-a3b-v1:0", true},
		{"us.qwen.qwen3-next-80b-a3b-v1:0", true},
		{"eu.qwen.qwen3-32b-v1:0", true},
		{"qwen.qwen3-coder-30b-a3b-instruct-v1:0", true},
		{"anthropic.claude-sonnet-4-6-v1:0", false},
		{"us.anthropic.claude-haiku-4-5-v1:0", false},
		{"meta.llama3-70b-instruct-v1:0", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isQwen(tt.model); got != tt.want {
			t.Errorf("isQwen(%q) = %v, want %v", tt.model, got, tt.want)
		}
	}
}

// buildQwenResponse builds a JSON response body matching the OpenAI-compatible
// chat completions schema that Qwen on Bedrock returns.
func buildQwenResponse(content, model, finishReason string, promptTokens, completionTokens int) []byte {
	resp := openaicompat.Response{
		ID:    "chatcmpl-test",
		Model: model,
		Choices: []openaicompat.Choice{
			{
				Message:      openaicompat.Message{Role: "assistant", Content: content},
				FinishReason: finishReason,
			},
		},
		Usage: openaicompat.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}
	body, _ := json.Marshal(resp)
	return body
}

func newQwenProvider(client bedrockClient) *BedrockProvider {
	return &BedrockProvider{
		client:     client,
		region:     "us-east-1",
		model:      "qwen.qwen3-next-80b-a3b-v1:0",
		httpClient: &http.Client{},
	}
}

func TestBedrockProvider_Qwen_Chat_Success(t *testing.T) {
	mock := &mockBedrockClient{
		responseBody: buildQwenResponse(
			"Hello from Qwen",
			"qwen.qwen3-next-80b-a3b-v1:0",
			"stop",
			12, 4,
		),
	}
	p := newQwenProvider(mock)

	resp, err := p.Chat(context.Background(), gollm.ChatRequest{
		Messages: []gollm.Message{{Role: "user", Content: "Hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello from Qwen" {
		t.Errorf("content = %q", resp.Content)
	}
	if resp.Model != "qwen.qwen3-next-80b-a3b-v1:0" {
		t.Errorf("model = %q", resp.Model)
	}
	if resp.StopReason != "stop" {
		t.Errorf("stop_reason = %q", resp.StopReason)
	}
	if resp.Usage.InputTokens != 12 {
		t.Errorf("input_tokens = %d", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 4 {
		t.Errorf("output_tokens = %d", resp.Usage.OutputTokens)
	}
}

func TestBedrockProvider_Qwen_Chat_APIError(t *testing.T) {
	mock := &mockBedrockClient{
		invokeErr: fmt.Errorf("AccessDeniedException: model not subscribed"),
	}
	p := newQwenProvider(mock)

	_, err := p.Chat(context.Background(), gollm.ChatRequest{
		Messages: []gollm.Message{{Role: "user", Content: "Hello"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "InvokeModel failed") {
		t.Errorf("error = %q, should mention InvokeModel failed", err.Error())
	}
	if !contains(err.Error(), "bedrock/qwen") {
		t.Errorf("error should carry bedrock/qwen prefix, got %q", err.Error())
	}
}

func TestBedrockProvider_Qwen_Chat_SystemPromptInMessages(t *testing.T) {
	wrapped := &capturingMockBedrockClient{
		delegate: &mockBedrockClient{
			responseBody: buildQwenResponse("4", "qwen.qwen3-32b-v1:0", "stop", 20, 1),
		},
	}
	p := &BedrockProvider{
		client:     wrapped,
		region:     "us-east-1",
		model:      "qwen.qwen3-32b-v1:0",
		httpClient: &http.Client{},
	}

	_, err := p.Chat(context.Background(), gollm.ChatRequest{
		SystemPrompt: "You are a calculator. Only respond with numbers.",
		Messages:     []gollm.Message{{Role: "user", Content: "2+2?"}},
		MaxTokens:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Qwen (OpenAI-style) carries the system prompt as the first message with
	// role="system", not as a top-level "system" field like Anthropic.
	var reqBody openaicompat.Request
	if err := json.Unmarshal(wrapped.lastBody, &reqBody); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	if _, hasSystemField := findKey(wrapped.lastBody, "system"); hasSystemField {
		t.Error("Qwen request body should NOT have a top-level 'system' field")
	}
	if len(reqBody.Messages) != 2 {
		t.Fatalf("messages = %d, want 2 (system + user)", len(reqBody.Messages))
	}
	if reqBody.Messages[0].Role != "system" {
		t.Errorf("messages[0].role = %q, want system", reqBody.Messages[0].Role)
	}
	if reqBody.Messages[0].Content != "You are a calculator. Only respond with numbers." {
		t.Errorf("messages[0].content = %q", reqBody.Messages[0].Content)
	}
}

func TestBedrockProvider_Qwen_Chat_TokenCounting(t *testing.T) {
	mock := &mockBedrockClient{
		responseBody: buildQwenResponse(
			"Response text",
			"qwen.qwen3-coder-next-v1:0",
			"length",
			100, 50,
		),
	}
	p := &BedrockProvider{
		client:     mock,
		region:     "us-east-1",
		model:      "qwen.qwen3-coder-next-v1:0",
		httpClient: &http.Client{},
	}

	resp, err := p.Chat(context.Background(), gollm.ChatRequest{
		Messages:  []gollm.Message{{Role: "user", Content: "Tell me a long story"}},
		MaxTokens: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Usage.InputTokens != 100 {
		t.Errorf("input_tokens = %d, want 100", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 50 {
		t.Errorf("output_tokens = %d, want 50", resp.Usage.OutputTokens)
	}
	if resp.StopReason != "length" {
		t.Errorf("stop_reason = %q, want length", resp.StopReason)
	}
}

func TestBedrockProvider_Qwen_Chat_ModelIDInEnvelopeNotBody(t *testing.T) {
	// Bedrock's InvokeModel carries the model identifier via ModelId. The body
	// must not include the model field — that's wasted bytes on a hot path.
	wrapped := &capturingMockBedrockClient{
		delegate: &mockBedrockClient{
			responseBody: buildQwenResponse("ok", "qwen.qwen3-next-80b-a3b-v1:0", "stop", 3, 1),
		},
	}
	p := newQwenProvider(wrapped)

	_, err := p.Chat(context.Background(), gollm.ChatRequest{
		Messages: []gollm.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, present := findKey(wrapped.lastBody, "model"); present {
		t.Errorf("body should omit 'model' field for Bedrock transport, got %s", string(wrapped.lastBody))
	}
}

func TestBedrockProvider_Qwen_Chat_NoChoices(t *testing.T) {
	body, _ := json.Marshal(openaicompat.Response{
		Model:   "qwen.qwen3-next-80b-a3b-v1:0",
		Choices: []openaicompat.Choice{},
	})
	mock := &mockBedrockClient{responseBody: body}
	p := newQwenProvider(mock)

	_, err := p.Chat(context.Background(), gollm.ChatRequest{
		Messages: []gollm.Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if !contains(err.Error(), "no choices") {
		t.Errorf("error should mention no choices, got %q", err.Error())
	}
}

func TestBedrockProvider_Qwen_Chat_ModelFallbackToRequest(t *testing.T) {
	// If the response omits the model field, the provider should fall back to
	// the requested model id so telemetry still has something to attribute.
	body, _ := json.Marshal(openaicompat.Response{
		Choices: []openaicompat.Choice{
			{Message: openaicompat.Message{Role: "assistant", Content: "ok"}, FinishReason: "stop"},
		},
	})
	mock := &mockBedrockClient{responseBody: body}
	p := newQwenProvider(mock)

	resp, err := p.Chat(context.Background(), gollm.ChatRequest{
		Model:    "qwen.qwen3-32b-v1:0",
		Messages: []gollm.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Model != "qwen.qwen3-32b-v1:0" {
		t.Errorf("model = %q, want fallback to request model", resp.Model)
	}
}

func TestBedrockProvider_Qwen_Registered(t *testing.T) {
	meta, ok := gollm.GetProviderMeta("bedrock")
	if !ok {
		t.Fatal("bedrock not registered")
	}
	// Qwen should appear in pricing and max-output-tokens maps
	if _, ok := meta.DefaultPricing["qwen3-next-80b-a3b"]; !ok {
		t.Error("missing pricing entry for qwen3-next-80b-a3b")
	}
	if got := gollm.GetMaxOutputTokens("bedrock", "qwen3-next-80b-a3b"); got != 32768 {
		t.Errorf("GetMaxOutputTokens(bedrock, qwen3-next-80b-a3b) = %d, want 32768", got)
	}
}

// findKey checks whether the given JSON object body contains a top-level key.
// Used in tests to verify request-body shape without being sensitive to field
// ordering.
func findKey(body []byte, key string) (json.RawMessage, bool) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, false
	}
	v, ok := obj[key]
	return v, ok
}

