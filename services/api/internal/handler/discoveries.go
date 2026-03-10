package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/decisionbox-io/decisionbox/services/api/internal/database"
)

// DiscoveriesHandler handles discovery result endpoints.
type DiscoveriesHandler struct {
	repo        *database.DiscoveryRepository
	projectRepo *database.ProjectRepository
}

func NewDiscoveriesHandler(repo *database.DiscoveryRepository, projectRepo *database.ProjectRepository) *DiscoveriesHandler {
	return &DiscoveriesHandler{repo: repo, projectRepo: projectRepo}
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

	// TODO: spawn agent process or add to job queue
	// For now, return instructions to run manually
	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":  "accepted",
		"message": "Discovery run queued. Run the agent manually: ./bin/decisionbox-agent --project-id=" + projectID,
	})
}

// GetStatus returns the discovery status for a project.
// GET /api/v1/projects/{id}/status
func (h *DiscoveriesHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	p, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil || p == nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	status := map[string]interface{}{
		"project_id":     projectID,
		"status":         p.Status,
		"last_run_at":    p.LastRunAt,
		"last_run_status": p.LastRunStatus,
	}

	latest, _ := h.repo.GetLatest(r.Context(), projectID)
	if latest != nil {
		status["last_discovery_date"] = latest.DiscoveryDate
		status["last_insights_count"] = len(latest.Insights)
	}

	writeJSON(w, http.StatusOK, status)
}
