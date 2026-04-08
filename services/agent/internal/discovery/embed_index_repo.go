package discovery

import (
	"context"

	commonmodels "github.com/decisionbox-io/decisionbox/libs/go-common/models"
)

// EmbedIndexStore handles Phase 9 database operations.
// This interface allows unit testing without a real MongoDB connection.
type EmbedIndexStore interface {
	InsertInsights(ctx context.Context, insights []*commonmodels.StandaloneInsight) error
	InsertRecommendations(ctx context.Context, recs []*commonmodels.StandaloneRecommendation) error
	UpdateEmbedding(ctx context.Context, collection, id, embeddingText, embeddingModel string) error
	UpdateDuplicate(ctx context.Context, collection, id, duplicateOf string, score float64) error
}
