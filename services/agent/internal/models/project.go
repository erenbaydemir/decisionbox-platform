package models

import "time"

// Project represents a DecisionBox project configuration.
// Stored in MongoDB "projects" collection.
type Project struct {
	ID          string `bson:"_id,omitempty" json:"id"`
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	Domain      string `bson:"domain" json:"domain"`
	Category    string `bson:"category" json:"category"`

	Warehouse WarehouseConfig `bson:"warehouse" json:"warehouse"`
	LLM       LLMConfig       `bson:"llm" json:"llm"`
	Schedule  ScheduleConfig  `bson:"schedule" json:"schedule"`

	Profile map[string]interface{} `bson:"profile,omitempty" json:"profile,omitempty"`

	Status        string     `bson:"status" json:"status"`
	LastRunAt     *time.Time `bson:"last_run_at,omitempty" json:"last_run_at,omitempty"`
	LastRunStatus string     `bson:"last_run_status,omitempty" json:"last_run_status,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// WarehouseConfig holds data warehouse connection settings.
type WarehouseConfig struct {
	Provider  string `bson:"provider" json:"provider"`
	ProjectID string `bson:"project_id,omitempty" json:"project_id,omitempty"`
	Location  string `bson:"location,omitempty" json:"location,omitempty"`

	// Datasets to explore. Supports multiple datasets in the same warehouse.
	// BigQuery: ["events_prod", "decisionbox_features_prod"]
	// The agent discovers schemas from ALL listed datasets.
	Datasets []string `bson:"datasets" json:"datasets"`

	// Deprecated: single dataset field. Use Datasets instead.
	// Kept for backward compatibility — if set and Datasets is empty,
	// treated as Datasets: [Dataset].
	Dataset string `bson:"dataset,omitempty" json:"dataset,omitempty"`

	// Optional: filter for multi-tenant warehouses.
	FilterField string `bson:"filter_field,omitempty" json:"filter_field,omitempty"`
	FilterValue string `bson:"filter_value,omitempty" json:"filter_value,omitempty"`
}

// GetDatasets returns the list of datasets to explore.
// Handles backward compatibility with the single Dataset field.
func (w *WarehouseConfig) GetDatasets() []string {
	if len(w.Datasets) > 0 {
		return w.Datasets
	}
	if w.Dataset != "" {
		return []string{w.Dataset}
	}
	return nil
}

// LLMConfig holds LLM provider settings.
type LLMConfig struct {
	Provider string `bson:"provider" json:"provider"`
	Model    string `bson:"model" json:"model"`
}

// ScheduleConfig holds discovery scheduling settings.
type ScheduleConfig struct {
	Enabled  bool   `bson:"enabled" json:"enabled"`
	CronExpr string `bson:"cron_expr" json:"cron_expr"`
	MaxSteps int    `bson:"max_steps" json:"max_steps"`
}
