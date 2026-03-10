package models

import "time"

// DiscoveryResult — read-only view of agent's discovery output.
// Same BSON schema as agent's model.
type DiscoveryResult struct {
	ID            string    `bson:"_id,omitempty" json:"id"`
	ProjectID     string    `bson:"project_id" json:"project_id"`
	Domain        string    `bson:"domain" json:"domain"`
	Category      string    `bson:"category" json:"category"`
	DiscoveryDate time.Time `bson:"discovery_date" json:"discovery_date"`

	TotalSteps int   `bson:"total_steps" json:"total_steps"`
	Duration   int64 `bson:"duration" json:"duration"`

	Insights        []Insight        `bson:"insights" json:"insights"`
	Recommendations []Recommendation `bson:"recommendations" json:"recommendations"`
	Summary         Summary          `bson:"summary" json:"summary"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type Insight struct {
	ID           string                 `bson:"id" json:"id"`
	AnalysisArea string                 `bson:"analysis_area" json:"analysis_area"`
	Name         string                 `bson:"name" json:"name"`
	Description  string                 `bson:"description" json:"description"`
	Severity     string                 `bson:"severity" json:"severity"`
	AffectedCount int                   `bson:"affected_count" json:"affected_count"`
	RiskScore     float64               `bson:"risk_score" json:"risk_score"`
	Confidence    float64               `bson:"confidence" json:"confidence"`
	Metrics       map[string]interface{} `bson:"metrics,omitempty" json:"metrics,omitempty"`
	Indicators    []string               `bson:"indicators,omitempty" json:"indicators,omitempty"`
	TargetSegment string                 `bson:"target_segment,omitempty" json:"target_segment,omitempty"`
	Validation    *InsightValidation     `bson:"validation,omitempty" json:"validation,omitempty"`
	DiscoveredAt  time.Time              `bson:"discovered_at" json:"discovered_at"`
}

type InsightValidation struct {
	Status        string    `bson:"status" json:"status"`
	VerifiedCount int       `bson:"verified_count,omitempty" json:"verified_count,omitempty"`
	OriginalCount int       `bson:"original_count,omitempty" json:"original_count,omitempty"`
	Reasoning     string    `bson:"reasoning,omitempty" json:"reasoning,omitempty"`
	ValidatedAt   time.Time `bson:"validated_at" json:"validated_at"`
}

type Recommendation struct {
	ID          string `bson:"id" json:"id"`
	Category    string `bson:"category" json:"category"`
	Title       string `bson:"title" json:"title"`
	Description string `bson:"description" json:"description"`
	Priority    int    `bson:"priority" json:"priority"`
	TargetSegment string `bson:"target_segment" json:"target_segment"`
	SegmentSize   int    `bson:"segment_size" json:"segment_size"`
	ExpectedImpact Impact `bson:"expected_impact" json:"expected_impact"`
	Actions     []string `bson:"actions" json:"actions"`
	Confidence  float64  `bson:"confidence" json:"confidence"`
}

type Impact struct {
	Metric               string  `bson:"metric" json:"metric"`
	EstimatedImprovement string  `bson:"estimated_improvement" json:"estimated_improvement"`
	Reasoning            string  `bson:"reasoning" json:"reasoning"`
}

type Summary struct {
	Date                 time.Time `bson:"date" json:"date"`
	Text                 string    `bson:"text" json:"text"`
	KeyFindings          []string  `bson:"key_findings" json:"key_findings"`
	TopRecommendations   []string  `bson:"top_recommendations" json:"top_recommendations"`
	TotalInsights        int       `bson:"total_insights" json:"total_insights"`
	TotalRecommendations int       `bson:"total_recommendations" json:"total_recommendations"`
	QueriesExecuted      int       `bson:"queries_executed" json:"queries_executed"`
}
