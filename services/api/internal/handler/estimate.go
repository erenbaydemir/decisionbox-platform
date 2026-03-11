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

	// Extract JSON from output (agent logs may be mixed in on stdout)
	jsonBytes := extractJSONObject(output)
	if jsonBytes == nil {
		writeError(w, http.StatusInternalServerError, "no JSON found in agent output")
		return
	}

	var estimate models.CostEstimate
	if err := json.Unmarshal(jsonBytes, &estimate); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to parse estimate: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, estimate)
}

// extractJSONObject finds the first top-level JSON object in mixed output.
func extractJSONObject(data []byte) []byte {
	s := string(data)
	start := strings.Index(s, "{")
	if start == -1 {
		return nil
	}
	// Find matching closing brace
	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == '{' {
			depth++
		} else if s[i] == '}' {
			depth--
			if depth == 0 {
				return []byte(s[start : i+1])
			}
		}
	}
	return nil
}
