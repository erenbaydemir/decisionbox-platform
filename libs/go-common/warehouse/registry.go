package warehouse

import (
	"fmt"
	"sync"
)

// ProviderConfig is a generic key-value configuration passed to provider factories.
// Each provider defines which keys it expects (e.g., "project_id", "dataset", "location").
type ProviderConfig map[string]string

// ProviderFactory creates a Provider from configuration.
// Provider packages implement this and register it via Register().
type ProviderFactory func(cfg ProviderConfig) (Provider, error)

var (
	providersMu sync.RWMutex
	providers   = make(map[string]ProviderFactory)
)

// Register makes a provider available by name.
// Provider packages call this in their init() function:
//
//	func init() {
//	    warehouse.Register("clickhouse", func(cfg warehouse.ProviderConfig) (warehouse.Provider, error) {
//	        return NewClickHouseProvider(cfg["dsn"], cfg["database"])
//	    })
//	}
//
// Services then select the provider via WAREHOUSE_PROVIDER env var.
func Register(name string, factory ProviderFactory) {
	providersMu.Lock()
	defer providersMu.Unlock()
	if factory == nil {
		panic("warehouse: Register factory is nil for " + name)
	}
	if _, exists := providers[name]; exists {
		panic("warehouse: Register called twice for " + name)
	}
	providers[name] = factory
}

// NewProvider creates a provider by name using the registered factory.
// Returns an error if the provider name is not registered.
//
// Usage in services:
//
//	provider, err := warehouse.NewProvider("bigquery", warehouse.ProviderConfig{
//	    "project_id": os.Getenv("WAREHOUSE_PROJECT_ID"),
//	    "dataset":    os.Getenv("WAREHOUSE_DATASET"),
//	})
func NewProvider(name string, cfg ProviderConfig) (Provider, error) {
	providersMu.RLock()
	factory, exists := providers[name]
	providersMu.RUnlock()

	if !exists {
		registered := make([]string, 0, len(providers))
		providersMu.RLock()
		for k := range providers {
			registered = append(registered, k)
		}
		providersMu.RUnlock()
		return nil, fmt.Errorf("warehouse: unknown provider %q (registered: %v)", name, registered)
	}

	return factory(cfg)
}

// RegisteredProviders returns the names of all registered providers.
func RegisteredProviders() []string {
	providersMu.RLock()
	defer providersMu.RUnlock()
	names := make([]string, 0, len(providers))
	for k := range providers {
		names = append(names, k)
	}
	return names
}
