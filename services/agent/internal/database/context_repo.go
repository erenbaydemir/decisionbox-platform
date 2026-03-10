package database

import (
	"context"
	"fmt"
	"time"

	applog "github.com/decisionbox-io/decisionbox/services/agent/internal/log"
	"github.com/decisionbox-io/decisionbox/services/agent/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionProjectContext = "project_context"

// ContextRepository manages ProjectContext persistence.
type ContextRepository struct {
	collection *mongo.Collection
}

// NewContextRepository creates a new context repository.
func NewContextRepository(client *DB) *ContextRepository {
	return &ContextRepository{
		collection: client.Collection(CollectionProjectContext),
	}
}

// GetByProjectID retrieves the project context.
func (r *ContextRepository) GetByProjectID(ctx context.Context, projectID string) (*models.ProjectContext, error) {
	filter := bson.M{"project_id": projectID}

	var pctx models.ProjectContext
	err := r.collection.FindOne(ctx, filter).Decode(&pctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.NewProjectContext(projectID), nil
		}
		return nil, fmt.Errorf("failed to get project context: %w", err)
	}

	return &pctx, nil
}

// Save saves or updates the project context.
func (r *ContextRepository) Save(ctx context.Context, pctx *models.ProjectContext) error {
	pctx.UpdatedAt = time.Now()

	filter := bson.M{"project_id": pctx.ProjectID}
	update := bson.M{"$set": pctx}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save project context: %w", err)
	}

	return nil
}

// EnsureIndexes creates necessary indexes.
func (r *ContextRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "project_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		applog.WithError(err).Warn("Failed to create context indexes")
		return err
	}
	return nil
}
