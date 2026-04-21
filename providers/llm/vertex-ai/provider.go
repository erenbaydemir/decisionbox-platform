// Package vertexai provides an llm.Provider for Google Vertex AI.
// Supports Claude (via Anthropic partnership) and Gemini (Google's own models)
// with GCP-native authentication (Application Default Credentials).
//
// Configuration:
//
//	LLM_PROVIDER=vertex-ai
//	LLM_MODEL=claude-sonnet-4-20250514  (or gemini-2.5-pro, gemini-2.5-flash)
//	VERTEX_PROJECT_ID=my-gcp-project
//	VERTEX_LOCATION=us-east5
//
// Authentication:
//
//	Uses Application Default Credentials (ADC). On GKE this works via
//	Workload Identity. Locally, run: gcloud auth application-default login
package vertexai

import (
	"context"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strconv"
	"strings"
	"time"

	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
)

func init() {
	gollm.RegisterWithMeta("vertex-ai", func(cfg gollm.ProviderConfig) (gollm.Provider, error) {
		projectID := cfg["project_id"]
		if projectID == "" {
			return nil, fmt.Errorf("vertex-ai: project_id is required")
		}
		location := cfg["location"]
		if location == "" {
			location = "us-east5"
		}
		model := cfg["model"]
		if model == "" {
			return nil, fmt.Errorf("vertex-ai: model is required")
		}

		endpointURL, err := buildEndpointURL(cfg["endpoint_url"], projectID, location)
		if err != nil {
			return nil, err
		}

		timeoutSec, _ := strconv.Atoi(cfg["timeout_seconds"])
		if timeoutSec == 0 {
			timeoutSec = 300 // 5 minutes default — Opus needs more time for large prompts
		}
		ctx := context.Background()

		// Initialize GCP auth
		auth, err := newGCPAuth(ctx)
		if err != nil {
			return nil, err
		}

		return &VertexAIProvider{
			projectID:   projectID,
			location:    location,
			model:       model,
			endpointURL: endpointURL,
			auth:        auth,
			httpClient:  &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
		}, nil
	}, gollm.ProviderMeta{
		Name:        "Google Vertex AI",
		Description: "GCP-managed AI platform — Claude, Gemini, and Model Garden deployed models with GCP auth",
		ConfigFields: []gollm.ConfigField{
			{Key: "project_id", Label: "GCP Project ID", Required: true, Type: "string", Placeholder: "my-gcp-project"},
			{Key: "location", Label: "Region", Type: "string", Default: "us-east5", Description: "GCP region (us-east5 for Claude, us-central1 for Gemini)"},
			{Key: "model", Label: "Model", Required: true, Type: "string", Default: "claude-sonnet-4-20250514", Placeholder: "claude-sonnet-4-20250514, gemini-2.5-pro, or gemma-4-31b-it"},
			{Key: "endpoint_url", Label: "Model Garden Endpoint URL", Type: "string", Description: "Only for deployed Model Garden models (e.g. Gemma, Mistral). Paste either the full base URL ending in /endpoints/{ID}, or just the Dedicated DNS from the Vertex console (e.g. mg-endpoint-xxx.us-central1-123.prediction.vertexai.goog). Leave empty for Claude/Gemini serverless models.", Placeholder: "mg-endpoint-xxx.us-central1-123.prediction.vertexai.goog"},
		},
		DefaultPricing: map[string]gollm.TokenPricing{
			"claude-opus-4-6":   {InputPerMillion: 15.0, OutputPerMillion: 75.0},
			"claude-sonnet-4-6": {InputPerMillion: 3.0, OutputPerMillion: 15.0},
			"claude-sonnet-4-5": {InputPerMillion: 3.0, OutputPerMillion: 15.0},
			"claude-opus-4-5":   {InputPerMillion: 15.0, OutputPerMillion: 75.0},
			"claude-opus-4-1":   {InputPerMillion: 15.0, OutputPerMillion: 75.0},
			"claude-sonnet-4":   {InputPerMillion: 3.0, OutputPerMillion: 15.0},
			"claude-opus-4":     {InputPerMillion: 15.0, OutputPerMillion: 75.0},
			"claude-haiku-4-5":  {InputPerMillion: 0.80, OutputPerMillion: 4.0},
			"gemini-2.5-pro":    {InputPerMillion: 1.25, OutputPerMillion: 10.0},
			"gemini-2.5-flash":  {InputPerMillion: 0.15, OutputPerMillion: 0.60},
		},
		MaxOutputTokens: map[string]int{
			"claude-opus-4-6":   128000,
			"claude-sonnet-4-6": 64000,
			"claude-sonnet-4-5": 64000,
			"claude-opus-4-5":   64000,
			"claude-opus-4-1":   32000,
			"claude-sonnet-4":   64000,
			"claude-opus-4":     32000,
			"claude-haiku-4-5":  64000,
			"gemini-2.5-pro":    65536,
			"gemini-2.5-flash":  65536,
			"_default":          16384,
		},
	})
}

// VertexAIProvider implements llm.Provider for Google Vertex AI.
// Routes to Claude, Gemini, or a Model Garden deployed endpoint based on
// configuration: endpoint_url takes priority, then model name prefix.
type VertexAIProvider struct {
	projectID   string
	location    string
	model       string
	endpointURL string // if set, routes through the Model Garden OpenAI-compatible path
	auth        *gcpAuth
	httpClient  *http.Client
}

// buildEndpointURL returns a fully-qualified Model Garden endpoint base URL
// (ending in /endpoints/{ID}, with no trailing slash or /chat/completions
// suffix) given whatever the user pasted into the config.
//
// Accepted inputs:
//   - "" → returns "" (Model Garden path disabled, falls back to Claude/Gemini routing)
//   - Full base URL, e.g. "https://<dns>/v1beta1/projects/<p>/locations/<r>/endpoints/<id>"
//   - Full base URL with trailing "/chat/completions" or "/" (normalized away)
//   - Dedicated endpoint DNS only, e.g. "mg-endpoint-<uuid>.<region>-<project>.prediction.vertexai.goog"
//     (with or without "https://" prefix) — the endpoint ID is derived from
//     the first DNS subdomain, then the full path is constructed using the
//     configured project_id and location.
func buildEndpointURL(raw, projectID, location string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	raw = strings.TrimRight(raw, "/")
	raw = strings.TrimSuffix(raw, "/chat/completions")
	raw = strings.TrimRight(raw, "/")

	if !strings.HasPrefix(raw, "https://") && !strings.HasPrefix(raw, "http://") {
		raw = "https://" + raw
	}

	if strings.Contains(raw, "/endpoints/") {
		return raw, nil
	}

	parsed, err := urlpkg.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("vertex-ai: invalid endpoint_url %q: %w", raw, err)
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf("vertex-ai: endpoint_url %q has a path but does not contain /endpoints/{ID}", raw)
	}

	idx := strings.Index(parsed.Host, ".")
	if idx <= 0 {
		return "", fmt.Errorf("vertex-ai: cannot derive endpoint ID from %q — paste either the full base URL including /endpoints/{ID}, or the dedicated endpoint DNS ({ENDPOINT_ID}.<region>-<project>.prediction.vertexai.goog)", raw)
	}
	endpointID := parsed.Host[:idx]

	return fmt.Sprintf("%s/v1beta1/projects/%s/locations/%s/endpoints/%s", raw, projectID, location, endpointID), nil
}

// Validate checks that GCP credentials are valid and the model endpoint is accessible.
// Makes a minimal request (max_tokens=1) to verify auth and model access.
func (p *VertexAIProvider) Validate(ctx context.Context) error {
	_, err := p.Chat(ctx, gollm.ChatRequest{
		Model:     p.model,
		Messages:  []gollm.Message{{Role: "user", Content: "hi"}},
		MaxTokens: 1,
	})
	if err != nil {
		return fmt.Errorf("vertex-ai: validation failed: %w", err)
	}
	return nil
}

// Chat sends a conversation to Vertex AI.
// If endpoint_url is configured, routes through the Model Garden OpenAI-compatible
// path regardless of model name. Otherwise routes to Claude or Gemini by model
// name prefix.
func (p *VertexAIProvider) Chat(ctx context.Context, req gollm.ChatRequest) (*gollm.ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	if p.endpointURL != "" {
		return p.modelGardenChat(ctx, req)
	}
	if isClaude(req.Model) {
		return p.claudeChat(ctx, req)
	}
	if isGemini(req.Model) {
		return p.geminiChat(ctx, req)
	}

	return nil, fmt.Errorf("vertex-ai: model %q is not claude-* or gemini-*; set endpoint_url to use a Model Garden deployed endpoint", req.Model)
}

// isClaude returns true if the model name is a Claude model.
func isClaude(model string) bool {
	return strings.HasPrefix(model, "claude-") || strings.Contains(model, "claude")
}

// isGemini returns true if the model name is a Gemini model.
func isGemini(model string) bool {
	return strings.HasPrefix(model, "gemini-") || strings.Contains(model, "gemini")
}
