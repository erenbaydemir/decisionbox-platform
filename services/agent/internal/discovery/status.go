package discovery

import (
	"context"
	"fmt"

	"github.com/decisionbox-io/decisionbox/services/agent/internal/database"
	"github.com/decisionbox-io/decisionbox/services/agent/internal/models"
)

// StatusReporter writes live status updates to MongoDB during a discovery run.
// If runID is empty, status reporting is disabled (agent run without API).
type StatusReporter struct {
	repo     *database.RunRepository
	runID    string
	maxSteps int
}

// NewStatusReporter creates a status reporter. Pass empty runID to disable.
func NewStatusReporter(repo *database.RunRepository, runID string, maxSteps int) *StatusReporter {
	if maxSteps <= 0 {
		maxSteps = 100
	}
	return &StatusReporter{repo: repo, runID: runID, maxSteps: maxSteps}
}

func (s *StatusReporter) enabled() bool {
	return s.runID != "" && s.repo != nil
}

// SetPhase updates the current phase and progress.
func (s *StatusReporter) SetPhase(ctx context.Context, phase, detail string, progress int) {
	if !s.enabled() {
		return
	}
	s.repo.UpdateStatus(ctx, s.runID, models.RunStatusRunning, phase, detail, progress)
}

// AddStep appends a step to the live log.
func (s *StatusReporter) AddStep(ctx context.Context, step models.RunStep) {
	if !s.enabled() {
		return
	}
	s.repo.AddStep(ctx, s.runID, step)
}

// AddExplorationStep logs an exploration step with LLM thinking and query.
func (s *StatusReporter) AddExplorationStep(ctx context.Context, stepNum int, thinking, query string, rowCount int, queryTimeMs int64, queryFixed bool, err string) {
	if !s.enabled() {
		return
	}

	msg := fmt.Sprintf("Step %d", stepNum)
	if thinking != "" {
		// Truncate thinking for display
		t := thinking
		if len(t) > 200 {
			t = t[:200] + "..."
		}
		msg = fmt.Sprintf("Step %d: %s", stepNum, t)
	}

	resultSummary := ""
	if rowCount > 0 {
		resultSummary = fmt.Sprintf("%d rows returned", rowCount)
	}

	step := models.RunStep{
		Phase:       models.PhaseExploration,
		StepNum:     stepNum,
		Type:        "query",
		Message:     msg,
		LLMThinking: thinking,
		Query:       query,
		QueryResult: resultSummary,
		RowCount:    rowCount,
		QueryTimeMs: queryTimeMs,
		QueryFixed:  queryFixed,
		Error:       err,
	}

	s.repo.AddStep(ctx, s.runID, step)

	// Update progress: exploration is 10-60% of total
	progress := 10 + (stepNum * 50 / s.maxSteps)
	if progress > 60 {
		progress = 60
	}
	detail := fmt.Sprintf("Step %d/%d: exploring data...", stepNum, s.maxSteps)
	s.repo.UpdateStatus(ctx, s.runID, models.RunStatusRunning, models.PhaseExploration, detail, progress)

	// Update query count
	s.repo.IncrementQueryCount(ctx, s.runID, err == "")
}

// AddAnalysisStep logs an analysis area completion.
func (s *StatusReporter) AddAnalysisStep(ctx context.Context, areaID, areaName string, insightCount int, err string) {
	if !s.enabled() {
		return
	}

	msg := fmt.Sprintf("Analyzed %s: %d insights found", areaName, insightCount)
	stepType := "analysis"
	if err != "" {
		msg = fmt.Sprintf("Analysis of %s failed: %s", areaName, err)
		stepType = "error"
	}

	step := models.RunStep{
		Phase:   models.PhaseAnalysis,
		Type:    stepType,
		Message: msg,
		Error:   err,
	}

	s.repo.AddStep(ctx, s.runID, step)
}

// AddInsightStep logs a discovered insight.
func (s *StatusReporter) AddInsightStep(ctx context.Context, name, severity, area string) {
	if !s.enabled() {
		return
	}

	step := models.RunStep{
		Phase:           models.PhaseAnalysis,
		Type:            "insight",
		Message:         fmt.Sprintf("Found: %s (%s)", name, severity),
		InsightName:     name,
		InsightSeverity: severity,
	}

	s.repo.AddStep(ctx, s.runID, step)
}

// AddValidationStep logs a validation check result.
func (s *StatusReporter) AddValidationStep(ctx context.Context, insightName, status string, claimed, verified int) {
	if !s.enabled() {
		return
	}

	msg := fmt.Sprintf("Validated \"%s\": %s", insightName, status)
	if claimed > 0 {
		msg = fmt.Sprintf("Validated \"%s\": %s (claimed: %d, verified: %d)", insightName, status, claimed, verified)
	}

	step := models.RunStep{
		Phase:   models.PhaseValidation,
		Type:    "validation",
		Message: msg,
	}

	s.repo.AddStep(ctx, s.runID, step)
}

// Complete marks the run as completed.
func (s *StatusReporter) Complete(ctx context.Context, insightsFound int) {
	if !s.enabled() {
		return
	}
	s.repo.Complete(ctx, s.runID, insightsFound)
}

// Fail marks the run as failed.
func (s *StatusReporter) Fail(ctx context.Context, errMsg string) {
	if !s.enabled() {
		return
	}
	s.repo.Fail(ctx, s.runID, errMsg)
}
