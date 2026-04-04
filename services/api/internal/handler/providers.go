package handler

import (
	"net/http"

	goembedding "github.com/decisionbox-io/decisionbox/libs/go-common/embedding"
	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
	gowarehouse "github.com/decisionbox-io/decisionbox/libs/go-common/warehouse"
)

// ProvidersHandler handles provider listing endpoints.
type ProvidersHandler struct{}

func NewProvidersHandler() *ProvidersHandler {
	return &ProvidersHandler{}
}

// ListLLMProviders returns registered LLM providers with config metadata.
// GET /api/v1/providers/llm
func (h *ProvidersHandler) ListLLMProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, gollm.RegisteredProvidersMeta())
}

// ListWarehouseProviders returns registered warehouse providers with config metadata.
// GET /api/v1/providers/warehouse
func (h *ProvidersHandler) ListWarehouseProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, gowarehouse.RegisteredProvidersMeta())
}

// ListEmbeddingProviders returns registered embedding providers with config metadata.
// GET /api/v1/providers/embedding
func (h *ProvidersHandler) ListEmbeddingProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, goembedding.RegisteredProvidersMeta())
}
