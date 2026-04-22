package openaicompat

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
)

func TestBuildRequestBody_BasicShape(t *testing.T) {
	got, err := BuildRequestBody(gollm.ChatRequest{
		Messages:    []gollm.Message{{Role: "user", Content: "Hello"}},
		MaxTokens:   100,
		Temperature: 0.5,
	}, "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed Request
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}

	if parsed.Model != "gpt-4o" {
		t.Errorf("model = %q, want gpt-4o", parsed.Model)
	}
	if len(parsed.Messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(parsed.Messages))
	}
	if parsed.Messages[0].Role != "user" || parsed.Messages[0].Content != "Hello" {
		t.Errorf("message[0] = %+v", parsed.Messages[0])
	}
	if parsed.MaxTokens != 100 {
		t.Errorf("max_tokens = %d", parsed.MaxTokens)
	}
	if parsed.Temperature != 0.5 {
		t.Errorf("temperature = %v", parsed.Temperature)
	}
}

func TestBuildRequestBody_SystemPromptPrepended(t *testing.T) {
	got, err := BuildRequestBody(gollm.ChatRequest{
		SystemPrompt: "You are a calculator.",
		Messages:     []gollm.Message{{Role: "user", Content: "2+2"}},
	}, "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed Request
	_ = json.Unmarshal(got, &parsed)

	if len(parsed.Messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(parsed.Messages))
	}
	if parsed.Messages[0].Role != "system" {
		t.Errorf("messages[0].role = %q, want system", parsed.Messages[0].Role)
	}
	if parsed.Messages[0].Content != "You are a calculator." {
		t.Errorf("messages[0].content = %q", parsed.Messages[0].Content)
	}
	if parsed.Messages[1].Role != "user" {
		t.Errorf("messages[1].role = %q, want user", parsed.Messages[1].Role)
	}
}

func TestBuildRequestBody_EmptyModelOmitted(t *testing.T) {
	got, err := BuildRequestBody(gollm.ChatRequest{
		Messages: []gollm.Message{{Role: "user", Content: "hi"}},
	}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify the model field is absent, not present with empty string
	if strings.Contains(string(got), `"model"`) {
		t.Errorf("empty model should be omitted, got: %s", string(got))
	}
}

func TestBuildRequestBody_ZeroTemperatureOmitted(t *testing.T) {
	got, err := BuildRequestBody(gollm.ChatRequest{
		Messages: []gollm.Message{{Role: "user", Content: "hi"}},
	}, "m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(string(got), `"temperature"`) {
		t.Errorf("zero temperature should be omitted, got: %s", string(got))
	}
}

func TestParseResponseBody_Success(t *testing.T) {
	body, _ := json.Marshal(Response{
		Model: "gpt-4o",
		Choices: []Choice{
			{Message: Message{Role: "assistant", Content: "Hello!"}, FinishReason: "stop"},
		},
		Usage: Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	})

	resp, err := ParseResponseBody(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello!" {
		t.Errorf("content = %q", resp.Content)
	}
	if resp.Model != "gpt-4o" {
		t.Errorf("model = %q", resp.Model)
	}
	if resp.StopReason != "stop" {
		t.Errorf("stop_reason = %q", resp.StopReason)
	}
	if resp.Usage.InputTokens != 10 {
		t.Errorf("input tokens = %d", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 5 {
		t.Errorf("output tokens = %d", resp.Usage.OutputTokens)
	}
}

func TestParseResponseBody_APIError(t *testing.T) {
	body, _ := json.Marshal(Response{
		Error: &APIError{Message: "invalid key", Type: "authentication_error", Code: "401"},
	})

	_, err := ParseResponseBody(body)
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error should be *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "401" {
		t.Errorf("code = %q, want 401", apiErr.Code)
	}
	if !strings.Contains(err.Error(), "authentication_error") {
		t.Errorf("error should render type, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "invalid key") {
		t.Errorf("error should render message, got %q", err.Error())
	}
}

func TestParseResponseBody_NoChoices(t *testing.T) {
	body, _ := json.Marshal(Response{Model: "gpt-4o", Choices: []Choice{}})
	_, err := ParseResponseBody(body)
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if !strings.Contains(err.Error(), "no choices") {
		t.Errorf("error = %q, should mention no choices", err.Error())
	}
}

func TestParseResponseBody_Malformed(t *testing.T) {
	_, err := ParseResponseBody([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestExtractAPIError(t *testing.T) {
	body, _ := json.Marshal(Response{
		Error: &APIError{Message: "rate limited", Type: "rate_limit_error"},
	})
	apiErr := ExtractAPIError(body)
	if apiErr == nil {
		t.Fatal("expected APIError")
	}
	if apiErr.Type != "rate_limit_error" {
		t.Errorf("type = %q", apiErr.Type)
	}

	// Non-error body should return nil
	okBody, _ := json.Marshal(Response{Model: "gpt-4o", Choices: []Choice{{}}})
	if apiErr := ExtractAPIError(okBody); apiErr != nil {
		t.Errorf("should return nil for non-error body, got %v", apiErr)
	}

	// Malformed JSON should return nil, not panic
	if apiErr := ExtractAPIError([]byte("not json")); apiErr != nil {
		t.Errorf("should return nil for malformed, got %v", apiErr)
	}
}

func TestAPIError_ErrorString(t *testing.T) {
	e := &APIError{Message: "oops"}
	if got := e.Error(); got != "oops" {
		t.Errorf("no-type: got %q", got)
	}
	e2 := &APIError{Type: "x_err", Message: "oops"}
	if got := e2.Error(); got != "x_err: oops" {
		t.Errorf("with-type: got %q", got)
	}
}
