package models

import "time"

// Project represents a DecisionBox project. Same schema as agent's model
// (both read/write the same MongoDB "projects" collection).
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

type WarehouseConfig struct {
	Provider    string `bson:"provider" json:"provider"`
	ProjectID   string `bson:"project_id,omitempty" json:"project_id,omitempty"`
	Dataset     string `bson:"dataset" json:"dataset"`
	Location    string `bson:"location,omitempty" json:"location,omitempty"`
	FilterField string `bson:"filter_field,omitempty" json:"filter_field,omitempty"`
	FilterValue string `bson:"filter_value,omitempty" json:"filter_value,omitempty"`
}

type LLMConfig struct {
	Provider string `bson:"provider" json:"provider"`
	Model    string `bson:"model" json:"model"`
}

type ScheduleConfig struct {
	Enabled  bool   `bson:"enabled" json:"enabled"`
	CronExpr string `bson:"cron_expr" json:"cron_expr"`
	MaxSteps int    `bson:"max_steps" json:"max_steps"`
}
