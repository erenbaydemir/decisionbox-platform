package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestStandaloneRecommendationBuildEmbeddingText(t *testing.T) {
	rec := &StandaloneRecommendation{
		Title:       "Add retry mechanics at Level 45",
		Description: "Implement a lives/retry system to reduce churn",
		ExpectedImpact: Impact{
			Metric:               "D7 retention",
			EstimatedImprovement: "15-20%",
		},
		TargetSegment: "players at level 45",
	}

	text := rec.BuildEmbeddingText()

	expected := []string{
		"Add retry mechanics at Level 45",
		"Implement a lives/retry system",
		"Impact: D7 retention 15-20%",
		"Segment: players at level 45",
	}
	for _, s := range expected {
		if !strings.Contains(text, s) {
			t.Errorf("expected embedding text to contain %q, got: %s", s, text)
		}
	}
}

func TestStandaloneRecommendationJSONRoundTrip(t *testing.T) {
	rec := &StandaloneRecommendation{
		ID:                     "660e8400-e29b-41d4-a716-446655440000",
		ProjectID:              "proj-123",
		DiscoveryID:            "disc-456",
		Domain:                 "gaming",
		Category:               "match3",
		RecommendationCategory: "engagement",
		Title:                  "Add retry mechanics",
		Description:            "Implement retries",
		Priority:               1,
		TargetSegment:          "players at level 45",
		SegmentSize:            12450,
		ExpectedImpact: Impact{
			Metric:               "D7 retention",
			EstimatedImprovement: "15-20%",
		},
		Actions:           []string{"Add 3 free retries per day"},
		RelatedInsightIDs: []string{"insight-abc"},
		Confidence:        0.78,
		EmbeddingText:     "Add retry mechanics...",
		EmbeddingModel:    "text-embedding-3-small",
		CreatedAt:         time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded StandaloneRecommendation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ID != rec.ID {
		t.Errorf("ID mismatch: %s != %s", decoded.ID, rec.ID)
	}
	if decoded.RecommendationCategory != "engagement" {
		t.Errorf("Category mismatch: %s", decoded.RecommendationCategory)
	}
	if decoded.EmbeddingModel != "text-embedding-3-small" {
		t.Errorf("EmbeddingModel mismatch: %s", decoded.EmbeddingModel)
	}
}
