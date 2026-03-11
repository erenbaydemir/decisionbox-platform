package models

import (
	"fmt"
	"testing"
	"time"
)

func TestNewProjectContext(t *testing.T) {
	ctx := NewProjectContext("proj-123")

	if ctx.ProjectID != "proj-123" {
		t.Errorf("ProjectID = %q, want %q", ctx.ProjectID, "proj-123")
	}
	if ctx.TotalDiscoveries != 0 {
		t.Errorf("TotalDiscoveries = %d, want 0", ctx.TotalDiscoveries)
	}
	if len(ctx.KnownSchemas) != 0 {
		t.Errorf("KnownSchemas should be empty")
	}
	if len(ctx.SuccessfulQueries) != 0 {
		t.Errorf("SuccessfulQueries should be empty")
	}
	if len(ctx.FailedQueries) != 0 {
		t.Errorf("FailedQueries should be empty")
	}
	if len(ctx.HistoricalPatterns) != 0 {
		t.Errorf("HistoricalPatterns should be empty")
	}
	if len(ctx.Notes) != 0 {
		t.Errorf("Notes should be empty")
	}
	if ctx.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestAddSuccessfulQuery(t *testing.T) {
	ctx := NewProjectContext("proj-123")

	ctx.AddSuccessfulQuery(QueryHistory{
		Query:   "SELECT * FROM test",
		Purpose: "test query",
	})

	if len(ctx.SuccessfulQueries) != 1 {
		t.Fatalf("len = %d, want 1", len(ctx.SuccessfulQueries))
	}
	if !ctx.SuccessfulQueries[0].Success {
		t.Error("query should be marked as success")
	}
}

func TestAddSuccessfulQueryLimit(t *testing.T) {
	ctx := NewProjectContext("proj-123")

	for i := 0; i < 150; i++ {
		ctx.AddSuccessfulQuery(QueryHistory{Query: "SELECT 1"})
	}

	if len(ctx.SuccessfulQueries) != 100 {
		t.Errorf("len = %d, want 100 (should trim to last 100)", len(ctx.SuccessfulQueries))
	}
}

func TestAddFailedQuery(t *testing.T) {
	ctx := NewProjectContext("proj-123")

	ctx.AddFailedQuery(QueryHistory{
		Query: "SELECT invalid",
		Error: "syntax error",
	})

	if len(ctx.FailedQueries) != 1 {
		t.Fatalf("len = %d, want 1", len(ctx.FailedQueries))
	}
	if ctx.FailedQueries[0].Success {
		t.Error("query should be marked as failure")
	}
}

func TestAddFailedQueryLimit(t *testing.T) {
	ctx := NewProjectContext("proj-123")

	for i := 0; i < 100; i++ {
		ctx.AddFailedQuery(QueryHistory{Query: "SELECT invalid"})
	}

	if len(ctx.FailedQueries) != 50 {
		t.Errorf("len = %d, want 50 (should trim to last 50)", len(ctx.FailedQueries))
	}
}

func TestAddNote(t *testing.T) {
	ctx := NewProjectContext("proj-123")

	ctx.AddNote("schema", "sessions table has 1M rows", 0.8)

	if len(ctx.Notes) != 1 {
		t.Fatalf("len = %d, want 1", len(ctx.Notes))
	}
	if ctx.Notes[0].Category != "schema" {
		t.Errorf("category = %q, want %q", ctx.Notes[0].Category, "schema")
	}
	if ctx.Notes[0].Relevance != 0.8 {
		t.Errorf("relevance = %f, want 0.8", ctx.Notes[0].Relevance)
	}
}

func TestAddNoteLimit(t *testing.T) {
	ctx := NewProjectContext("proj-123")

	for i := 0; i < 250; i++ {
		ctx.AddNote("test", "note", 0.5)
	}

	if len(ctx.Notes) != 200 {
		t.Errorf("len = %d, want 200 (should trim to last 200)", len(ctx.Notes))
	}
}

func TestRecordDiscoverySuccess(t *testing.T) {
	ctx := NewProjectContext("proj-123")
	ctx.ConsecutiveFailures = 3

	ctx.RecordDiscovery(true)

	if ctx.TotalDiscoveries != 1 {
		t.Errorf("TotalDiscoveries = %d, want 1", ctx.TotalDiscoveries)
	}
	if ctx.ConsecutiveFailures != 0 {
		t.Errorf("ConsecutiveFailures = %d, want 0 (should reset on success)", ctx.ConsecutiveFailures)
	}
	if ctx.LastDiscoveryDate.IsZero() {
		t.Error("LastDiscoveryDate should be set")
	}
}

func TestRecordDiscoveryFailure(t *testing.T) {
	ctx := NewProjectContext("proj-123")

	ctx.RecordDiscovery(false)
	ctx.RecordDiscovery(false)

	if ctx.TotalDiscoveries != 2 {
		t.Errorf("TotalDiscoveries = %d, want 2", ctx.TotalDiscoveries)
	}
	if ctx.ConsecutiveFailures != 2 {
		t.Errorf("ConsecutiveFailures = %d, want 2", ctx.ConsecutiveFailures)
	}
}

func TestUpdatedAtChanges(t *testing.T) {
	ctx := NewProjectContext("proj-123")
	initial := ctx.UpdatedAt

	time.Sleep(time.Millisecond)
	ctx.AddNote("test", "note", 0.5)

	if !ctx.UpdatedAt.After(initial) {
		t.Error("UpdatedAt should be updated after AddNote")
	}
}

// --- UpdatePatterns ---

func TestUpdatePatterns_NewInsight(t *testing.T) {
	ctx := NewProjectContext("proj-123")
	insights := []Insight{
		{Name: "High Churn", AnalysisArea: "churn", Description: "Players leaving"},
	}

	ctx.UpdatePatterns(insights)

	if len(ctx.HistoricalPatterns) != 1 {
		t.Fatalf("patterns = %d, want 1", len(ctx.HistoricalPatterns))
	}
	p := ctx.HistoricalPatterns[0]
	if p.PatternID != "churn:High Churn" {
		t.Errorf("patternID = %q", p.PatternID)
	}
	if p.SeenCount != 1 {
		t.Errorf("seenCount = %d, want 1", p.SeenCount)
	}
	if p.Status != "active" {
		t.Errorf("status = %q, want active", p.Status)
	}
}

func TestUpdatePatterns_RecurringInsight(t *testing.T) {
	ctx := NewProjectContext("proj-123")
	insights := []Insight{
		{Name: "High Churn", AnalysisArea: "churn"},
	}

	ctx.UpdatePatterns(insights)
	ctx.UpdatePatterns(insights) // seen again

	if len(ctx.HistoricalPatterns) != 1 {
		t.Fatalf("patterns = %d, want 1 (deduped)", len(ctx.HistoricalPatterns))
	}
	if ctx.HistoricalPatterns[0].SeenCount != 2 {
		t.Errorf("seenCount = %d, want 2", ctx.HistoricalPatterns[0].SeenCount)
	}
	if ctx.HistoricalPatterns[0].Status != "recurring" {
		t.Errorf("status = %q, want recurring", ctx.HistoricalPatterns[0].Status)
	}
}

func TestUpdatePatterns_MultipleInsights(t *testing.T) {
	ctx := NewProjectContext("proj-123")
	insights := []Insight{
		{Name: "High Churn", AnalysisArea: "churn"},
		{Name: "Low Revenue", AnalysisArea: "monetization"},
		{Name: "Level 45 Spike", AnalysisArea: "levels"},
	}

	ctx.UpdatePatterns(insights)

	if len(ctx.HistoricalPatterns) != 3 {
		t.Errorf("patterns = %d, want 3", len(ctx.HistoricalPatterns))
	}
}

func TestUpdatePatterns_Limit(t *testing.T) {
	ctx := NewProjectContext("proj-123")

	// Add 201 unique insights — should cap at 200
	for i := 0; i < 201; i++ {
		ctx.UpdatePatterns([]Insight{
			{Name: fmt.Sprintf("Insight %d", i), AnalysisArea: "test"},
		})
	}

	if len(ctx.HistoricalPatterns) > 200 {
		t.Errorf("patterns = %d, should be capped at 200", len(ctx.HistoricalPatterns))
	}
}

// --- Summary types ---

func TestInsightSummary_Fields(t *testing.T) {
	s := InsightSummary{
		Name: "High Churn", AnalysisArea: "churn", Severity: "critical",
		AffectedCount: 500, Date: "2026-03-10",
	}
	if s.Name != "High Churn" || s.Severity != "critical" || s.AffectedCount != 500 {
		t.Error("fields not set correctly")
	}
}

func TestFeedbackSummary_Fields(t *testing.T) {
	s := FeedbackSummary{InsightName: "Bad", Rating: "dislike", Comment: "wrong metric"}
	if s.Rating != "dislike" || s.Comment != "wrong metric" {
		t.Error("fields not set correctly")
	}
}

func TestRecommendationSummary_Fields(t *testing.T) {
	s := RecommendationSummary{Title: "Send Lives", Category: "churn", Priority: 1}
	if s.Priority != 1 || s.Category != "churn" {
		t.Error("fields not set correctly")
	}
}
