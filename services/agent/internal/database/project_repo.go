package database

import (
	"context"
	"fmt"

	"github.com/decisionbox-io/decisionbox/services/agent/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionProjects = "projects"

// ProjectRepository provides access to project documents.
type ProjectRepository struct {
	db *DB
}

// NewProjectRepository creates a new project repository.
func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// GetByID returns a project by its ID.
func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*models.Project, error) {
	col := r.db.Collection(CollectionProjects)

	// Try as ObjectID first, then as string
	filter := bson.M{}
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		filter["_id"] = oid
	} else {
		filter["_id"] = id
	}

	var project models.Project
	if err := col.FindOne(ctx, filter).Decode(&project); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return &project, nil
}
