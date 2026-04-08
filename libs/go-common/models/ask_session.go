package models

import "time"

// AskSession represents a multi-turn conversation in the "Ask Insights" feature.
// Stored in the "ask_sessions" collection.
type AskSession struct {
	ID        string              `bson:"_id" json:"id"`
	ProjectID string              `bson:"project_id" json:"project_id"`
	UserID    string              `bson:"user_id" json:"user_id"`
	Title     string              `bson:"title" json:"title"` // first question, used as display title
	Messages     []AskSessionMessage `bson:"messages" json:"messages"`
	MessageCount int                 `bson:"message_count" json:"message_count"`
	CreatedAt    time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time           `bson:"updated_at" json:"updated_at"`
}

// AskSessionMessage is a single Q&A turn within a conversation.
type AskSessionMessage struct {
	Question   string             `bson:"question" json:"question"`
	Answer     string             `bson:"answer" json:"answer"`
	Sources    []AskSessionSource `bson:"sources" json:"sources"`
	Model      string             `bson:"model" json:"model"`
	TokensUsed int                `bson:"tokens_used" json:"tokens_used"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// AskSessionSource is a reference to an insight or recommendation used as context.
type AskSessionSource struct {
	ID           string  `bson:"id" json:"id"`
	Type         string  `bson:"type" json:"type"` // "insight" or "recommendation"
	Name         string  `bson:"name" json:"name"`
	Score        float64 `bson:"score" json:"score"`
	Severity     string  `bson:"severity,omitempty" json:"severity,omitempty"`
	AnalysisArea string  `bson:"analysis_area,omitempty" json:"analysis_area,omitempty"`
	Description  string  `bson:"description,omitempty" json:"description,omitempty"`
	DiscoveryID  string  `bson:"discovery_id" json:"discovery_id"`
}
