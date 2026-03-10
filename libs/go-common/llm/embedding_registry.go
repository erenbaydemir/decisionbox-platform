package llm

import (
	"fmt"
	"sync"
)

// EmbeddingProviderFactory creates an EmbeddingProvider from configuration.
type EmbeddingProviderFactory func(cfg ProviderConfig) (EmbeddingProvider, error)

var (
	embeddingProvidersMu sync.RWMutex
	embeddingProviders   = make(map[string]EmbeddingProviderFactory)
)

// RegisterEmbedding makes an embedding provider available by name.
// Provider packages call this in their init() function:
//
//	func init() {
//	    llm.RegisterEmbedding("openai", func(cfg llm.ProviderConfig) (llm.EmbeddingProvider, error) {
//	        return NewOpenAIEmbeddingProvider(cfg["api_key"], cfg["model"], cfg["dimensions"])
//	    })
//	}
//
// Services then select the provider via EMBEDDING_PROVIDER env var.
func RegisterEmbedding(name string, factory EmbeddingProviderFactory) {
	embeddingProvidersMu.Lock()
	defer embeddingProvidersMu.Unlock()
	if factory == nil {
		panic("llm: RegisterEmbedding factory is nil for " + name)
	}
	if _, exists := embeddingProviders[name]; exists {
		panic("llm: RegisterEmbedding called twice for " + name)
	}
	embeddingProviders[name] = factory
}

// NewEmbeddingProvider creates an embedding provider by name.
//
// Usage in services:
//
//	provider, err := llm.NewEmbeddingProvider("openai", llm.ProviderConfig{
//	    "api_key":    os.Getenv("EMBEDDING_API_KEY"),
//	    "model":      "text-embedding-3-large",
//	    "dimensions": "3072",
//	})
func NewEmbeddingProvider(name string, cfg ProviderConfig) (EmbeddingProvider, error) {
	embeddingProvidersMu.RLock()
	factory, exists := embeddingProviders[name]
	embeddingProvidersMu.RUnlock()

	if !exists {
		registered := make([]string, 0, len(embeddingProviders))
		embeddingProvidersMu.RLock()
		for k := range embeddingProviders {
			registered = append(registered, k)
		}
		embeddingProvidersMu.RUnlock()
		return nil, fmt.Errorf("llm: unknown embedding provider %q (registered: %v)", name, registered)
	}

	return factory(cfg)
}

// RegisteredEmbeddingProviders returns the names of all registered embedding providers.
func RegisteredEmbeddingProviders() []string {
	embeddingProvidersMu.RLock()
	defer embeddingProvidersMu.RUnlock()
	names := make([]string, 0, len(embeddingProviders))
	for k := range embeddingProviders {
		names = append(names, k)
	}
	return names
}
