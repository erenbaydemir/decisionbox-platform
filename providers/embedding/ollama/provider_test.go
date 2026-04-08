package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	goembedding "github.com/decisionbox-io/decisionbox/libs/go-common/embedding"
)

func TestRegistration(t *testing.T) {
	names := goembedding.RegisteredProviders()
	found := false
	for _, n := range names {
		if n == "ollama" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected ollama to be registered")
	}
}

func TestRegistrationMeta(t *testing.T) {
	meta, ok := goembedding.GetProviderMeta("ollama")
	if !ok {
		t.Fatal("expected ollama metadata to exist")
	}
	if meta.Name != "Ollama (Local)" {
		t.Errorf("expected Name='Ollama (Local)', got %s", meta.Name)
	}
	if len(meta.Models) != 3 {
		t.Errorf("expected 3 models, got %d", len(meta.Models))
	}
}

func TestFactoryUnsupportedModel(t *testing.T) {
	_, err := goembedding.NewProvider("ollama", goembedding.ProviderConfig{
		"model": "nonexistent-model",
	})
	if err == nil {
		t.Fatal("expected error for unsupported model")
	}
	if !strings.Contains(err.Error(), "unsupported model") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFactoryDefaultModel(t *testing.T) {
	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(embedResponse{
			Embeddings: [][]float64{make([]float64, 768)},
		})
	})
	defer server.Close()

	p, err := goembedding.NewProvider("ollama", goembedding.ProviderConfig{
		"host": server.URL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ModelName() != "nomic-embed-text" {
		t.Errorf("expected default model nomic-embed-text, got %s", p.ModelName())
	}
	if p.Dimensions() != 768 {
		t.Errorf("expected 768 dims, got %d", p.Dimensions())
	}
}

func TestFactoryCustomHost(t *testing.T) {
	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(embedResponse{
			Embeddings: [][]float64{make([]float64, 768)},
		})
	})
	defer server.Close()

	p, err := goembedding.NewProvider("ollama", goembedding.ProviderConfig{
		"host":  server.URL,
		"model": "nomic-embed-text",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it uses the custom host by making a request
	_, err = p.Embed(context.Background(), []string{"test"})
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
}

func TestEmbedSingleText(t *testing.T) {
	expectedVec := []float64{0.1, 0.2, 0.3}

	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/embed" {
			t.Errorf("expected /api/embed path, got %s", r.URL.Path)
		}

		var req embedRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "nomic-embed-text" {
			t.Errorf("expected model nomic-embed-text, got %s", req.Model)
		}
		if len(req.Input) != 1 || req.Input[0] != "hello world" {
			t.Errorf("unexpected input: %v", req.Input)
		}

		json.NewEncoder(w).Encode(embedResponse{
			Model:      "nomic-embed-text",
			Embeddings: [][]float64{expectedVec},
		})
	})
	defer server.Close()

	p := newProvider(server.URL, "nomic-embed-text", 768)
	result, err := p.Embed(context.Background(), []string{"hello world"})
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0][0] != 0.1 {
		t.Errorf("expected first value 0.1, got %f", result[0][0])
	}
}

func TestEmbedBatch(t *testing.T) {
	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req embedRequest
		json.NewDecoder(r.Body).Decode(&req)

		embeddings := make([][]float64, len(req.Input))
		for i := range req.Input {
			embeddings[i] = make([]float64, 768)
		}
		json.NewEncoder(w).Encode(embedResponse{Embeddings: embeddings})
	})
	defer server.Close()

	p := newProvider(server.URL, "nomic-embed-text", 768)
	result, err := p.Embed(context.Background(), []string{"text1", "text2", "text3"})
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
}

func TestEmbedEmpty(t *testing.T) {
	p := newProvider("http://unused", "nomic-embed-text", 768)
	result, err := p.Embed(context.Background(), []string{})
	if err != nil {
		t.Fatalf("Embed empty failed: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result for empty input, got %v", result)
	}
}

func TestEmbedServerError(t *testing.T) {
	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "model not found"}`))
	})
	defer server.Close()

	p := newProvider(server.URL, "nomic-embed-text", 768)
	_, err := p.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected HTTP 500 in error, got: %v", err)
	}
}

func TestEmbedConnectionError(t *testing.T) {
	p := newProvider("http://localhost:1", "nomic-embed-text", 768)
	_, err := p.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Fatal("expected error for connection failure")
	}
	if !strings.Contains(err.Error(), "is Ollama running") {
		t.Errorf("expected helpful error message, got: %v", err)
	}
}

func TestEmbedMismatchedCount(t *testing.T) {
	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(embedResponse{
			Embeddings: [][]float64{make([]float64, 768)},
		})
	})
	defer server.Close()

	p := newProvider(server.URL, "nomic-embed-text", 768)
	_, err := p.Embed(context.Background(), []string{"text1", "text2"})
	if err == nil {
		t.Fatal("expected error for mismatched count")
	}
	if !strings.Contains(err.Error(), "expected 2 embeddings") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate(t *testing.T) {
	server := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(embedResponse{
			Embeddings: [][]float64{make([]float64, 768)},
		})
	})
	defer server.Close()

	p := newProvider(server.URL, "nomic-embed-text", 768)
	err := p.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestModelName(t *testing.T) {
	p := newProvider("http://unused", "mxbai-embed-large", 1024)
	if p.ModelName() != "mxbai-embed-large" {
		t.Errorf("expected mxbai-embed-large, got %s", p.ModelName())
	}
}

func TestDimensions(t *testing.T) {
	tests := []struct {
		model string
		dims  int
	}{
		{"nomic-embed-text", 768},
		{"mxbai-embed-large", 1024},
		{"all-minilm", 384},
	}
	for _, tt := range tests {
		p := newProvider("http://unused", tt.model, tt.dims)
		if p.Dimensions() != tt.dims {
			t.Errorf("model %s: expected %d dims, got %d", tt.model, tt.dims, p.Dimensions())
		}
	}
}

// Verify provider implements the interface at compile time.
var _ goembedding.Provider = (*provider)(nil)

func newMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}
