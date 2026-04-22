//go:build integration

package bedrock

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
)

func bedrockModel() string {
	if m := os.Getenv("INTEGRATION_TEST_BEDROCK_MODEL"); m != "" {
		return m
	}
	return "us.anthropic.claude-haiku-4-5-20251001-v1:0"
}

func TestIntegration_BasicChat(t *testing.T) {
	region := os.Getenv("INTEGRATION_TEST_BEDROCK_REGION")
	if region == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_REGION not set (also requires AWS credentials)")
	}

	model := bedrockModel()
	provider, err := gollm.NewProvider("bedrock", gollm.ProviderConfig{
		"region": region,
		"model":  model,
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.Chat(ctx, gollm.ChatRequest{
		Messages:  []gollm.Message{{Role: "user", Content: "Say hello in one word."}},
		MaxTokens: 10,
	})
	if err != nil {
		if strings.Contains(err.Error(), "ThrottlingException") || strings.Contains(err.Error(), "429") {
			t.Skipf("Rate limited (auth works, quota exceeded): %v", err)
		}
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content == "" {
		t.Error("response content should not be empty")
	}
	if resp.Model == "" {
		t.Error("response model should not be empty")
	}
	if resp.Usage.InputTokens == 0 {
		t.Error("should report input tokens")
	}
	if resp.Usage.OutputTokens == 0 {
		t.Error("should report output tokens")
	}
	t.Logf("Bedrock Claude: %q (model=%s, tokens: in=%d out=%d)",
		resp.Content, resp.Model, resp.Usage.InputTokens, resp.Usage.OutputTokens)
}

func TestIntegration_SystemPrompt(t *testing.T) {
	region := os.Getenv("INTEGRATION_TEST_BEDROCK_REGION")
	if region == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_REGION not set")
	}

	model := bedrockModel()
	provider, err := gollm.NewProvider("bedrock", gollm.ProviderConfig{
		"region": region,
		"model":  model,
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.Chat(ctx, gollm.ChatRequest{
		SystemPrompt: "You are a calculator. Only respond with numbers.",
		Messages:     []gollm.Message{{Role: "user", Content: "What is 2+2?"}},
		MaxTokens:    10,
	})
	if err != nil {
		if strings.Contains(err.Error(), "ThrottlingException") || strings.Contains(err.Error(), "429") {
			t.Skipf("Rate limited (auth works, quota exceeded): %v", err)
		}
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content == "" {
		t.Error("response should not be empty")
	}
	t.Logf("Bedrock system prompt: %q", resp.Content)
}

func TestIntegration_ModelOverride(t *testing.T) {
	region := os.Getenv("INTEGRATION_TEST_BEDROCK_REGION")
	if region == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_REGION not set")
	}

	model := bedrockModel()
	provider, err := gollm.NewProvider("bedrock", gollm.ProviderConfig{
		"region": region,
		"model":  model,
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.Chat(ctx, gollm.ChatRequest{
		Model:     model,
		Messages:  []gollm.Message{{Role: "user", Content: "Say yes."}},
		MaxTokens: 5,
	})
	if err != nil {
		if strings.Contains(err.Error(), "ThrottlingException") || strings.Contains(err.Error(), "429") {
			t.Skipf("Rate limited (auth works, quota exceeded): %v", err)
		}
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content == "" {
		t.Error("response should not be empty")
	}
	if resp.StopReason == "" {
		t.Error("stop_reason should not be empty")
	}
	t.Logf("Bedrock model override: %q (stop=%s)", resp.Content, resp.StopReason)
}

// --- Error path tests ---

func TestIntegration_InvalidModel(t *testing.T) {
	region := os.Getenv("INTEGRATION_TEST_BEDROCK_REGION")
	if region == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_REGION not set")
	}

	provider, err := gollm.NewProvider("bedrock", gollm.ProviderConfig{
		"region": region,
		"model":  "anthropic.nonexistent-model-v1:0",
	})
	if err != nil {
		t.Fatalf("Provider creation should succeed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = provider.Chat(ctx, gollm.ChatRequest{
		Messages:  []gollm.Message{{Role: "user", Content: "hello"}},
		MaxTokens: 5,
	})
	if err == nil {
		t.Fatal("should return error for invalid model")
	}
	t.Logf("Invalid model error: %v", err)
}

func TestIntegration_UnsupportedModelPrefix(t *testing.T) {
	region := os.Getenv("INTEGRATION_TEST_BEDROCK_REGION")
	if region == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_REGION not set")
	}

	provider, err := gollm.NewProvider("bedrock", gollm.ProviderConfig{
		"region": region,
		"model":  "meta.llama3-70b-instruct-v1:0",
	})
	if err != nil {
		t.Fatalf("Provider creation should succeed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = provider.Chat(ctx, gollm.ChatRequest{
		Messages:  []gollm.Message{{Role: "user", Content: "hello"}},
		MaxTokens: 5,
	})
	if err == nil {
		t.Fatal("should return error for unsupported model")
	}
	if !strings.Contains(err.Error(), "unsupported model") {
		t.Errorf("error should mention unsupported model, got: %v", err)
	}
	t.Logf("Unsupported model error: %v", err)
}

// --- Qwen tests ---
//
// These tests exercise the OpenAI-compatible route on a real Qwen model. They
// are gated by INTEGRATION_TEST_BEDROCK_QWEN_MODEL so that the default Claude
// integration tests do not require a Qwen-enabled account. Example:
//
//	INTEGRATION_TEST_BEDROCK_REGION=us-east-1 \
//	INTEGRATION_TEST_BEDROCK_QWEN_MODEL=qwen.qwen3-next-80b-a3b \
//	  go test -tags=integration ./providers/llm/bedrock/...

func TestIntegration_Qwen_BasicChat(t *testing.T) {
	region := os.Getenv("INTEGRATION_TEST_BEDROCK_REGION")
	if region == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_REGION not set (also requires AWS credentials)")
	}
	model := os.Getenv("INTEGRATION_TEST_BEDROCK_QWEN_MODEL")
	if model == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_QWEN_MODEL not set")
	}

	provider, err := gollm.NewProvider("bedrock", gollm.ProviderConfig{
		"region": region,
		"model":  model,
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := provider.Chat(ctx, gollm.ChatRequest{
		Messages:  []gollm.Message{{Role: "user", Content: "Say hello in one word."}},
		MaxTokens: 20,
	})
	if err != nil {
		if strings.Contains(err.Error(), "ThrottlingException") || strings.Contains(err.Error(), "429") {
			t.Skipf("Rate limited (auth works, quota exceeded): %v", err)
		}
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content == "" {
		t.Error("response content should not be empty")
	}
	if resp.Usage.InputTokens == 0 {
		t.Error("should report input tokens")
	}
	if resp.Usage.OutputTokens == 0 {
		t.Error("should report output tokens")
	}
	t.Logf("Bedrock Qwen: %q (model=%s, tokens: in=%d out=%d)",
		resp.Content, resp.Model, resp.Usage.InputTokens, resp.Usage.OutputTokens)
}

func TestIntegration_Qwen_SystemPrompt(t *testing.T) {
	region := os.Getenv("INTEGRATION_TEST_BEDROCK_REGION")
	if region == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_REGION not set")
	}
	model := os.Getenv("INTEGRATION_TEST_BEDROCK_QWEN_MODEL")
	if model == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_QWEN_MODEL not set")
	}

	provider, err := gollm.NewProvider("bedrock", gollm.ProviderConfig{
		"region": region,
		"model":  model,
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := provider.Chat(ctx, gollm.ChatRequest{
		SystemPrompt: "You are a calculator. Respond only with the numeric answer.",
		Messages:     []gollm.Message{{Role: "user", Content: "What is 2+2?"}},
		MaxTokens:    10,
	})
	if err != nil {
		if strings.Contains(err.Error(), "ThrottlingException") || strings.Contains(err.Error(), "429") {
			t.Skipf("Rate limited (auth works, quota exceeded): %v", err)
		}
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content == "" {
		t.Error("response should not be empty")
	}
	t.Logf("Bedrock Qwen system prompt: %q", resp.Content)
}

func TestIntegration_ContextCancellation(t *testing.T) {
	region := os.Getenv("INTEGRATION_TEST_BEDROCK_REGION")
	if region == "" {
		t.Skip("INTEGRATION_TEST_BEDROCK_REGION not set")
	}

	provider, err := gollm.NewProvider("bedrock", gollm.ProviderConfig{
		"region": region,
		"model":  bedrockModel(),
	})
	if err != nil {
		t.Fatalf("Provider creation failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = provider.Chat(ctx, gollm.ChatRequest{
		Messages:  []gollm.Message{{Role: "user", Content: "hello"}},
		MaxTokens: 5,
	})
	if err == nil {
		t.Fatal("should return error for cancelled context")
	}
	t.Logf("Cancelled context error: %v", err)
}
