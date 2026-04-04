// Package ollama provides an embedding.Provider backed by a local Ollama instance.
// Ollama runs open-source embedding models locally (nomic-embed-text, mxbai-embed-large, etc.).
//
// Register via init():
//
//	import _ "github.com/decisionbox-io/decisionbox/providers/embedding/ollama"
//
// Supported models:
//   - nomic-embed-text (768 dims)
//   - mxbai-embed-large (1024 dims)
//   - all-minilm (384 dims)
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	goembedding "github.com/decisionbox-io/decisionbox/libs/go-common/embedding"
)

const defaultHost = "http://localhost:11434"

var modelDimensions = map[string]int{
	"nomic-embed-text": 768,
	"mxbai-embed-large": 1024,
	"all-minilm":        384,
}

func init() {
	goembedding.RegisterWithMeta("ollama", func(cfg goembedding.ProviderConfig) (goembedding.Provider, error) {
		host := cfg["host"]
		if host == "" {
			host = defaultHost
		}

		model := cfg["model"]
		if model == "" {
			model = "nomic-embed-text"
		}

		dims, ok := modelDimensions[model]
		if !ok {
			return nil, fmt.Errorf("ollama embedding: unsupported model %q (supported: nomic-embed-text, mxbai-embed-large, all-minilm)", model)
		}

		return newProvider(host, model, dims), nil
	}, goembedding.ProviderMeta{
		Name:        "Ollama (Local)",
		Description: "Run open-source embedding models locally — free, air-gapped",
		ConfigFields: []goembedding.ConfigField{
			{Key: "host", Label: "Ollama Host", Type: "string", Default: defaultHost, Placeholder: defaultHost},
			{Key: "model", Label: "Model", Required: true, Type: "string", Default: "nomic-embed-text"},
		},
		Models: []goembedding.ModelInfo{
			{ID: "nomic-embed-text", Name: "Nomic Embed Text", Dimensions: 768},
			{ID: "mxbai-embed-large", Name: "MxBai Embed Large", Dimensions: 1024},
			{ID: "all-minilm", Name: "All-MiniLM", Dimensions: 384},
		},
	})
}

// provider implements embedding.Provider using a local Ollama instance.
type provider struct {
	host   string
	model  string
	dims   int
	client *http.Client
}

func newProvider(host, model string, dims int) *provider {
	return &provider{
		host:  host,
		model: model,
		dims:  dims,
		client: &http.Client{
			Timeout: 120 * time.Second, // longer timeout for local models with cold start
		},
	}
}

// Embed generates vector embeddings for the given texts.
// Uses the Ollama /api/embed endpoint which supports batch inputs.
func (p *provider) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := embedRequest{
		Model: p.model,
		Input: texts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ollama embedding: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.host+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama embedding: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embedding: request failed (is Ollama running at %s?): %w", p.host, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama embedding: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama embedding: API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var embResp embedResponse
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("ollama embedding: failed to unmarshal response: %w", err)
	}

	if len(embResp.Embeddings) != len(texts) {
		return nil, fmt.Errorf("ollama embedding: expected %d embeddings, got %d", len(texts), len(embResp.Embeddings))
	}

	return embResp.Embeddings, nil
}

// Dimensions returns the vector dimensionality for this model.
func (p *provider) Dimensions() int {
	return p.dims
}

// ModelName returns the model identifier.
func (p *provider) ModelName() string {
	return p.model
}

// Validate checks that Ollama is reachable and the model is available.
func (p *provider) Validate(ctx context.Context) error {
	_, err := p.Embed(ctx, []string{"test"})
	return err
}

// embedRequest is the Ollama /api/embed request body.
type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// embedResponse is the Ollama /api/embed response body.
type embedResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
}
