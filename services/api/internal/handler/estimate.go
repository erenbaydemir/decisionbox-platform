package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/decisionbox-io/decisionbox/services/api/internal/database"
	"github.com/decisionbox-io/decisionbox/services/api/internal/models"
)

// EstimateHandler handles cost estimation endpoints.
type EstimateHandler struct {
	projectRepo *database.ProjectRepository
}

func NewEstimateHandler(projectRepo *database.ProjectRepository) *EstimateHandler {
	return &EstimateHandler{projectRepo: projectRepo}
}

// Estimate runs a cost estimation for a project.
// POST /api/v1/projects/{id}/discover/estimate
func (h *EstimateHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	p, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil || p == nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	// Parse optional request body
	var body struct {
		Areas    []string `json:"areas"`
		MaxSteps int      `json:"max_steps"`
	}
	decodeJSON(r, &body)

	if body.MaxSteps <= 0 {
		body.MaxSteps = 100
	}

	// Spawn agent with --estimate flag (synchronous — captures stdout)
	args := []string{
		"--project-id", projectID,
		"--estimate",
		"--max-steps", strconv.Itoa(body.MaxSteps),
	}
	if len(body.Areas) > 0 {
		args = append(args, "--areas", strings.Join(body.Areas, ","))
	}

	cmd := exec.Command("decisionbox-agent", args...)
	cmd.Env = append(os.Environ(),
		"MONGODB_URI="+getEnvOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
		"MONGODB_DB="+getEnvOrDefault("MONGODB_DB", "decisionbox"),
	)

	output, err := cmd.Output()
	if err != nil {
		errMsg := fmt.Sprintf("estimation failed: %s", err.Error())
		if exitErr, ok := err.(*exec.ExitError); ok {
			errMsg += "\n" + string(exitErr.Stderr)
		}
		writeError(w, http.StatusInternalServerError, errMsg)
		return
	}

	// Parse the JSON output from the agent
	var estimate models.CostEstimate
	if err := json.Unmarshal(output, &estimate); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to parse estimate: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, estimate)
}
