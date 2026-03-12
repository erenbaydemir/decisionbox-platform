package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/decisionbox-io/decisionbox/services/api/internal/database"
	apilog "github.com/decisionbox-io/decisionbox/services/api/internal/log"
)

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// DiscoveriesHandler handles discovery result endpoints.
type DiscoveriesHandler struct {
	repo        *database.DiscoveryRepository
	projectRepo *database.ProjectRepository
	runRepo     *database.RunRepository
	tracker     *ProcessTracker
}

func NewDiscoveriesHandler(repo *database.DiscoveryRepository, projectRepo *database.ProjectRepository, runRepo *database.RunRepository, tracker *ProcessTracker) *DiscoveriesHandler {
	return &DiscoveriesHandler{repo: repo, projectRepo: projectRepo, runRepo: runRepo, tracker: tracker}
}

// List returns discovery results for a project.
// GET /api/v1/projects/{id}/discoveries
func (h *DiscoveriesHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	p, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil || p == nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	results, err := h.repo.List(r.Context(), projectID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list discoveries: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// GetDiscoveryByID returns a specific discovery by its ID.
// GET /api/v1/discoveries/{id}
func (h *DiscoveriesHandler) GetDiscoveryByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	result, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get discovery: "+err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "discovery not found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetLatest returns the most recent discovery for a project.
// GET /api/v1/projects/{id}/discoveries/latest
func (h *DiscoveriesHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	result, err := h.repo.GetLatest(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get discovery: "+err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "no discoveries found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetByDate returns a discovery for a specific date.
// GET /api/v1/projects/{id}/discoveries/{date}
func (h *DiscoveriesHandler) GetByDate(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	dateStr := r.PathValue("date")

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	result, err := h.repo.GetByDate(r.Context(), projectID, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get discovery: "+err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "no discovery found for date "+dateStr)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// TriggerDiscovery triggers a discovery run for a project.
// POST /api/v1/projects/{id}/discover
func (h *DiscoveriesHandler) TriggerDiscovery(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	p, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil || p == nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	// Check if there's already a running discovery
	running, _ := h.runRepo.GetRunningByProject(r.Context(), projectID)
	if running != nil {
		writeJSON(w, http.StatusConflict, map[string]string{
			"status":  "already_running",
			"run_id":  running.ID,
			"message": "A discovery is already running for this project",
		})
		return
	}

	// Parse optional request body
	var body struct {
		Areas    []string `json:"areas"`     // optional: run only these areas
		MaxSteps int      `json:"max_steps"` // optional: override exploration steps (default 100)
	}
	decodeJSON(r, &body) // ignore error — body is optional

	// Create a run record
	runID, err := h.runRepo.Create(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create run: "+err.Error())
		return
	}

	// Spawn the agent as a background subprocess
	args := []string{
		"--project-id", projectID,
		"--run-id", runID,
	}
	if len(body.Areas) > 0 {
		args = append(args, "--areas", strings.Join(body.Areas, ","))
	}
	if body.MaxSteps > 0 {
		args = append(args, "--max-steps", strconv.Itoa(body.MaxSteps))
	}
	cmd := exec.Command("decisionbox-agent", args...)

	// Inherit parent environment so agent gets LLM_API_KEY, DOMAIN_PACK_PATH, etc.
	cmd.Env = append(os.Environ(),
		"MONGODB_URI="+getEnvOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
		"MONGODB_DB="+getEnvOrDefault("MONGODB_DB", "decisionbox"),
	)

	if err := cmd.Start(); err != nil {
		apilog.WithFields(apilog.Fields{
			"project_id": projectID, "run_id": runID, "error": err.Error(),
		}).Error("Failed to start agent subprocess")
		h.runRepo.Fail(r.Context(), runID, "failed to start: "+err.Error())
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to start agent: %s", err.Error()))
		return
	}

	apilog.WithFields(apilog.Fields{
		"project_id": projectID,
		"run_id":     runID,
		"pid":        cmd.Process.Pid,
		"areas":      body.Areas,
		"max_steps":  body.MaxSteps,
	}).Info("Discovery agent spawned")

	// Track the process so we can cancel it
	h.tracker.Track(runID, cmd.Process)

	// Wait in background, clean up when done
	go func() {
		err := cmd.Wait()
		h.tracker.Remove(runID)
		if err != nil {
			apilog.WithFields(apilog.Fields{
				"run_id": runID, "error": err.Error(),
			}).Warn("Agent process exited with error")
			h.runRepo.Fail(context.Background(), runID, "agent exited with error: "+err.Error())
		} else {
			apilog.WithField("run_id", runID).Info("Agent process completed")
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status": "started",
		"run_id": runID,
		"message": "Discovery agent started",
	})
}

// GetStatus returns the live discovery status for a project.
// GET /api/v1/projects/{id}/status
func (h *DiscoveriesHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	p, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil || p == nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	// Get the latest run (for live status)
	latestRun, _ := h.runRepo.GetLatestByProject(r.Context(), projectID)

	status := map[string]interface{}{
		"project_id": projectID,
	}

	if latestRun != nil {
		status["run"] = latestRun
	}

	// Also include latest completed discovery stats
	latest, _ := h.repo.GetLatest(r.Context(), projectID)
	if latest != nil {
		status["last_discovery"] = map[string]interface{}{
			"date":            latest.DiscoveryDate,
			"insights_count":  len(latest.Insights),
			"total_steps":     latest.TotalSteps,
		}
	}

	writeJSON(w, http.StatusOK, status)
}

// GetRun returns a specific discovery run by ID.
// GET /api/v1/runs/{runId}
func (h *DiscoveriesHandler) GetRun(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("runId")

	run, err := h.runRepo.GetByID(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get run: "+err.Error())
		return
	}
	if run == nil {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}

	writeJSON(w, http.StatusOK, run)
}

// CancelRun cancels a running discovery.
// DELETE /api/v1/runs/{runId}
func (h *DiscoveriesHandler) CancelRun(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("runId")

	run, err := h.runRepo.GetByID(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get run: "+err.Error())
		return
	}
	if run == nil {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}

	if run.Status != "running" && run.Status != "pending" {
		writeError(w, http.StatusBadRequest, "run is not active (status: "+run.Status+")")
		return
	}

	// Kill the subprocess if it's tracked
	killed := h.tracker.Kill(runID)

	// Mark as cancelled in MongoDB
	h.runRepo.Cancel(r.Context(), runID)

	msg := "Run cancelled"
	if killed {
		msg = "Run cancelled and agent process killed"
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "cancelled",
		"message": msg,
	})
}
