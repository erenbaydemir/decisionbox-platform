// Package domainpack defines interfaces for domain-specific functionality.
//
// DecisionBox's core platform is domain-agnostic. Domain packs layer
// industry-specific logic on top: semantic text generation, warehouse
// queries, field mappings, and entity types.
//
// Each domain pack (gaming, e-commerce, social, etc.) implements these
// interfaces and registers itself via Register(). Services select the
// pack based on the app's "domain" field in MongoDB.
//
// Example — adding an e-commerce domain pack:
//
//  1. Create domain-packs/ecommerce/go/ with implementations
//  2. Call domainpack.Register("ecommerce", NewEcommercePack())
//  3. Service: import _ ".../domain-packs/ecommerce/go"
//  4. Set domain: "ecommerce" on the app in MongoDB
package domainpack

// Pack represents a complete domain pack with all domain-specific logic.
// Each domain (gaming, e-commerce, social) implements this interface.
type Pack interface {
	// Name returns the domain name (e.g., "gaming", "ecommerce").
	Name() string

	// SemanticGenerator returns the semantic text generator for this domain.
	SemanticGenerator() SemanticGenerator

	// DataFetcher returns the data fetcher for this domain.
	DataFetcher() DataFetcher

	// EntityTypes returns the entity types this domain supports
	// (e.g., gaming: ["users", "entities", "sessions", "cohorts"]).
	EntityTypes() []string
}

// SemanticGenerator converts raw warehouse data into human-readable text
// suitable for embedding. Each domain generates different text:
//
//   - Gaming: "Active gamer, level 45, quit rate 15%, uses boosters..."
//   - E-commerce: "Customer with $500 lifetime spend, 3 orders, last purchase 5 days ago..."
//   - Social: "Creator with 10K followers, posts 3x daily, 8% engagement rate..."
type SemanticGenerator interface {
	// GenerateText converts raw data into human-readable text for embedding.
	// entityType: domain-specific entity type (e.g., "users", "entities").
	// data: key-value pairs from warehouse query results.
	GenerateText(entityType string, data map[string]interface{}) string

	// SupportedEntityTypes returns entity types this generator handles.
	SupportedEntityTypes() []string
}

// DataFetcher provides domain-specific warehouse queries and identity mapping.
// Each domain knows which tables to query and how to identify entities.
type DataFetcher interface {
	// FetchQuery returns the SQL query for fetching data for a given entity type.
	// dataset: the warehouse dataset name.
	// appID: the application ID for filtering.
	// lookbackDays: how far back to look (0 = no limit).
	FetchQuery(entityType, dataset, appID string, lookbackDays int) string

	// GenerateDatapointID creates a unique datapoint identifier from raw data.
	// Each domain knows which fields uniquely identify an entity:
	//   - Gaming users: app_id + user_id
	//   - Gaming entities: app_id + level_number
	//   - E-commerce: app_id + product_id
	GenerateDatapointID(entityType, appID string, data map[string]interface{}) string

	// CollectionName returns the MongoDB collection name for an entity type.
	// Each domain can name its collections:
	//   - Gaming: "embedding_user_profiles", "embedding_entity_profiles"
	//   - E-commerce: "embedding_customer_profiles", "embedding_product_profiles"
	CollectionName(entityType string) string

	// SupportedEntityTypes returns entity types this fetcher handles.
	SupportedEntityTypes() []string
}
