package discovery

import (
	"context"
	"fmt"

	commonmodels "github.com/decisionbox-io/decisionbox/libs/go-common/models"
	"github.com/decisionbox-io/decisionbox/services/agent/internal/database"
)

// MongoEmbedIndexStore implements EmbedIndexStore using MongoDB.
type MongoEmbedIndexStore struct {
	db *database.DB
}

// NewMongoEmbedIndexStore creates a new MongoDB-backed EmbedIndexStore.
func NewMongoEmbedIndexStore(db *database.DB) *MongoEmbedIndexStore {
	return &MongoEmbedIndexStore{db: db}
}

func (s *MongoEmbedIndexStore) InsertInsights(ctx context.Context, insights []*commonmodels.StandaloneInsight) error {
	if len(insights) == 0 {
		return nil
	}
	docs := make([]interface{}, len(insights))
	for i, ins := range insights {
		docs[i] = ins
	}
	_, err := s.db.Collection("insights").InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("insert insights: %w", err)
	}
	return nil
}

func (s *MongoEmbedIndexStore) InsertRecommendations(ctx context.Context, recs []*commonmodels.StandaloneRecommendation) error {
	if len(recs) == 0 {
		return nil
	}
	docs := make([]interface{}, len(recs))
	for i, rec := range recs {
		docs[i] = rec
	}
	_, err := s.db.Collection("recommendations").InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("insert recommendations: %w", err)
	}
	return nil
}

func (s *MongoEmbedIndexStore) UpdateEmbedding(ctx context.Context, collection, id, embeddingText, embeddingModel string) error {
	_, err := s.db.Collection(collection).UpdateByID(ctx, id, map[string]interface{}{
		"$set": map[string]interface{}{
			"embedding_text":  embeddingText,
			"embedding_model": embeddingModel,
		},
	})
	if err != nil {
		return fmt.Errorf("update embedding for %s/%s: %w", collection, id, err)
	}
	return nil
}

func (s *MongoEmbedIndexStore) UpdateDuplicate(ctx context.Context, collection, id, duplicateOf string, score float64) error {
	_, err := s.db.Collection(collection).UpdateByID(ctx, id, map[string]interface{}{
		"$set": map[string]interface{}{
			"duplicate_of":     duplicateOf,
			"similarity_score": score,
		},
	})
	if err != nil {
		return fmt.Errorf("update duplicate for %s/%s: %w", collection, id, err)
	}
	return nil
}
