package vertexai

import (
	"testing"

	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
)

func TestIsClaude(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"claude-sonnet-4-20250514", true},
		{"claude-opus-4-20250514", true},
		{"claude-haiku-4-5-20251001", true},
		{"gemini-2.5-pro", false},
		{"gpt-4o", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isClaude(tt.model); got != tt.want {
			t.Errorf("isClaude(%q) = %v, want %v", tt.model, got, tt.want)
		}
	}
}

func TestIsGemini(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"gemini-2.5-pro", true},
		{"gemini-2.5-flash", true},
		{"gemini-1.5-pro", true},
		{"claude-sonnet-4-20250514", false},
		{"gpt-4o", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isGemini(tt.model); got != tt.want {
			t.Errorf("isGemini(%q) = %v, want %v", tt.model, got, tt.want)
		}
	}
}

func TestVertexAIProvider_ChatUnsupportedModel(t *testing.T) {
	p := &VertexAIProvider{
		projectID: "test-project",
		location:  "us-central1",
		model:     "llama-3",
	}

	_, err := p.Chat(nil, gollm.ChatRequest{
		Model: "llama-3",
	})

	if err == nil {
		t.Fatal("expected error for unsupported model without endpoint_url")
	}
	if !contains(err.Error(), "endpoint_url") {
		t.Errorf("error = %q, should suggest setting endpoint_url", err.Error())
	}
}

func TestVertexAIProvider_ChatDefaultModel(t *testing.T) {
	p := &VertexAIProvider{
		projectID: "test-project",
		location:  "us-central1",
		model:     "claude-sonnet-4-20250514",
	}

	// Empty model in request should fall back to provider default
	// Will fail on auth (no credentials in test), but verifies routing
	_, err := p.Chat(nil, gollm.ChatRequest{
		Messages: []gollm.Message{{Role: "user", Content: "test"}},
	})

	// Should fail on auth, not on model routing
	if err != nil && contains(err.Error(), "unsupported model") {
		t.Error("should route to claude, not fail on model check")
	}
}

func TestVertexAIProvider_Registered(t *testing.T) {
	meta, ok := gollm.GetProviderMeta("vertex-ai")
	if !ok {
		t.Fatal("vertex-ai not registered")
	}
	if meta.Name == "" {
		t.Error("missing provider name")
	}
	if meta.Description == "" || contains(meta.Description, "coming soon") {
		t.Error("description still says 'coming soon' — should be updated")
	}
	if len(meta.DefaultPricing) == 0 {
		t.Error("no default pricing")
	}
	if _, ok := meta.DefaultPricing["claude-sonnet-4"]; !ok {
		t.Error("missing claude-sonnet-4 pricing")
	}
	if _, ok := meta.DefaultPricing["gemini-2.5-pro"]; !ok {
		t.Error("missing gemini-2.5-pro pricing")
	}

	// MaxOutputTokens
	if meta.MaxOutputTokens == nil {
		t.Fatal("MaxOutputTokens should not be nil")
	}
	if len(meta.MaxOutputTokens) != 11 {
		t.Errorf("MaxOutputTokens has %d entries, want 11", len(meta.MaxOutputTokens))
	}
	if meta.MaxOutputTokens["claude-opus-4-6"] != 128000 {
		t.Errorf("MaxOutputTokens[claude-opus-4-6] = %d, want 128000", meta.MaxOutputTokens["claude-opus-4-6"])
	}
	if meta.MaxOutputTokens["gemini-2.5-pro"] != 65536 {
		t.Errorf("MaxOutputTokens[gemini-2.5-pro] = %d, want 65536", meta.MaxOutputTokens["gemini-2.5-pro"])
	}

	// Verify GetMaxOutputTokens helper
	if got := gollm.GetMaxOutputTokens("vertex-ai", "gemini-2.5-flash"); got != 65536 {
		t.Errorf("GetMaxOutputTokens(vertex-ai, gemini-2.5-flash) = %d, want 65536", got)
	}
	// Verify _default fallback
	if got := gollm.GetMaxOutputTokens("vertex-ai", "some-new-model"); got != 16384 {
		t.Errorf("GetMaxOutputTokens(vertex-ai, some-new-model) = %d, want 16384 (_default)", got)
	}
}

func TestVertexAIProvider_Validate_UnsupportedModel(t *testing.T) {
	p := &VertexAIProvider{
		projectID: "test-project",
		location:  "us-east5",
		model:     "llama-3",
	}
	err := p.Validate(nil)
	if err == nil {
		t.Error("Validate should fail with unsupported model")
	}
}

func TestVertexAIProvider_ConfigFields(t *testing.T) {
	meta, ok := gollm.GetProviderMeta("vertex-ai")
	if !ok {
		t.Fatal("vertex-ai not registered")
	}

	fieldKeys := make(map[string]bool)
	for _, f := range meta.ConfigFields {
		fieldKeys[f.Key] = true
	}

	if !fieldKeys["project_id"] {
		t.Error("missing project_id config field")
	}
	if !fieldKeys["location"] {
		t.Error("missing location config field")
	}
	if !fieldKeys["model"] {
		t.Error("missing model config field")
	}
	if !fieldKeys["endpoint_url"] {
		t.Error("missing endpoint_url config field (needed for Model Garden deployed models)")
	}
	// Should NOT have api_key — uses GCP ADC
	if fieldKeys["api_key"] {
		t.Error("vertex-ai should not have api_key field — uses GCP ADC")
	}
}

func TestBuildEndpointURL(t *testing.T) {
	projectID := "decisionbox"
	location := "us-central1"
	full := "https://mg-endpoint-abc.us-central1-12345.prediction.vertexai.goog/v1beta1/projects/decisionbox/locations/us-central1/endpoints/mg-endpoint-abc"

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "empty input returns empty",
			in:   "",
			want: "",
		},
		{
			name: "whitespace returns empty",
			in:   "   ",
			want: "",
		},
		{
			name: "full URL passed through",
			in:   full,
			want: full,
		},
		{
			name: "full URL with trailing slash is trimmed",
			in:   full + "/",
			want: full,
		},
		{
			name: "full URL with /chat/completions is stripped",
			in:   full + "/chat/completions",
			want: full,
		},
		{
			name: "DNS only, no scheme → full URL constructed",
			in:   "mg-endpoint-abc.us-central1-12345.prediction.vertexai.goog",
			want: full,
		},
		{
			name: "DNS only with https:// scheme → full URL constructed",
			in:   "https://mg-endpoint-abc.us-central1-12345.prediction.vertexai.goog",
			want: full,
		},
		{
			name: "DNS with trailing slash → still works",
			in:   "https://mg-endpoint-abc.us-central1-12345.prediction.vertexai.goog/",
			want: full,
		},
		{
			name:    "URL with non-empty path but no /endpoints/ is rejected",
			in:      "https://example.com/some/wrong/path",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildEndpointURL(tt.in, projectID, location)
			if tt.wantErr {
				if err == nil {
					t.Errorf("buildEndpointURL(%q) = %q, want error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Errorf("buildEndpointURL(%q): unexpected error: %v", tt.in, err)
				return
			}
			if got != tt.want {
				t.Errorf("buildEndpointURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestVertexAI_Factory_InvalidEndpointURL(t *testing.T) {
	_, err := gollm.NewProvider("vertex-ai", gollm.ProviderConfig{
		"project_id":   "my-project",
		"location":     "us-central1",
		"model":        "gemma-4-31b-it",
		"endpoint_url": "https://example.com/some/wrong/path",
	})
	if err == nil {
		t.Fatal("expected error for endpoint_url with wrong path")
	}
	if !contains(err.Error(), "/endpoints/") && !contains(err.Error(), "path") {
		t.Errorf("error = %q, should mention /endpoints/ or path", err.Error())
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
