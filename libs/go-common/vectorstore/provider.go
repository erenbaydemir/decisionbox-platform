package vectorstore

import "context"

// Provider abstracts vector store operations for semantic search.
// The primary implementation is Qdrant. The interface exists for testability
// (mock in unit tests), not for swapping backends.
type Provider interface {
	// Upsert stores vectors with metadata. Idempotent by ID.
	Upsert(ctx context.Context, points []Point) error

	// Search finds vectors similar to the query vector, with optional filters.
	Search(ctx context.Context, vector []float64, opts SearchOpts) ([]SearchResult, error)

	// FindDuplicates searches for existing vectors above the similarity threshold.
	// Used during indexing to flag near-duplicate insights/recommendations across discovery runs.
	// Filters by projectID and docType ("insight"/"recommendation"), excludes current discoveryID.
	FindDuplicates(ctx context.Context, vector []float64, projectID string, docType string, excludeDiscoveryID string, threshold float64) ([]SearchResult, error)

	// Delete removes vectors by ID.
	Delete(ctx context.Context, ids []string) error

	// HealthCheck verifies the vector store is reachable.
	HealthCheck(ctx context.Context) error

	// EnsureCollection creates the collection if it doesn't exist.
	// The collection is named decisionbox_{dimensions} and configured with cosine distance.
	EnsureCollection(ctx context.Context, dimensions int) error
}

// Point represents a vector with metadata to store in the vector database.
type Point struct {
	ID      string
	Vector  []float64
	Payload map[string]interface{}
}

// SearchOpts configures a vector search query.
type SearchOpts struct {
	ProjectIDs     []string // required: scope to one or more projects
	Types          []string // optional: "insight", "recommendation"
	EmbeddingModel string   // optional: filter by model (for cross-project search)
	Severity       string   // optional: "critical", "high", "medium", "low"
	AnalysisArea   string   // optional: filter by area
	Limit          int      // max results to return
	MinScore       float64  // optional: minimum similarity threshold
}

// SearchResult represents a single vector search match.
type SearchResult struct {
	ID      string
	Score   float64
	Payload map[string]interface{}
}
