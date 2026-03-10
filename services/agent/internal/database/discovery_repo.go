package database

import (
	"context"
	"fmt"
	"time"

	"github.com/decisionbox-io/decisionbox/services/agent/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DiscoveryRepository manages DiscoveryResult persistence.
type DiscoveryRepository struct {
	collection *mongo.Collection
}

// NewDiscoveryRepository creates a new discovery repository.
func NewDiscoveryRepository(client *DB) *DiscoveryRepository {
	return &DiscoveryRepository{
		collection: client.Collection(CollectionDiscoveries),
	}
}

// Save inserts a discovery result.
func (r *DiscoveryRepository) Save(ctx context.Context, result *models.DiscoveryResult) error {
	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, result)
	if err != nil {
		return fmt.Errorf("failed to save discovery result: %w", err)
	}
	return nil
}

// GetLatest retrieves the most recent discovery for a project.
func (r *DiscoveryRepository) GetLatest(ctx context.Context, projectID string) (*models.DiscoveryResult, error) {
	filter := bson.M{"project_id": projectID}
	opts := options.FindOne().SetSort(bson.D{{Key: "discovery_date", Value: -1}})

	var result models.DiscoveryResult
	err := r.collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest discovery: %w", err)
	}
	return &result, nil
}

// EnsureIndexes creates necessary indexes.
func (r *DiscoveryRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "project_id", Value: 1},
				{Key: "discovery_date", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
