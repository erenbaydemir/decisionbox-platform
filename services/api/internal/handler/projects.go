package handler

import (
	"net/http"
	"strconv"

	"github.com/decisionbox-io/decisionbox/services/api/internal/database"
	"github.com/decisionbox-io/decisionbox/services/api/internal/models"
)

// ProjectsHandler handles project CRUD endpoints.
type ProjectsHandler struct {
	repo *database.ProjectRepository
}

func NewProjectsHandler(repo *database.ProjectRepository) *ProjectsHandler {
	return &ProjectsHandler{repo: repo}
}

// Create creates a new project.
// POST /api/v1/projects
func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p models.Project
	if err := decodeJSON(r, &p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if p.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if p.Domain == "" {
		writeError(w, http.StatusBadRequest, "domain is required")
		return
	}
	if p.Category == "" {
		writeError(w, http.StatusBadRequest, "category is required")
		return
	}

	if err := h.repo.Create(r.Context(), &p); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create project: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

// List returns all projects.
// GET /api/v1/projects
func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	projects, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list projects: "+err.Error())
		return
	}

	if projects == nil {
		projects = make([]*models.Project, 0)
	}

	writeJSON(w, http.StatusOK, projects)
}

// Get returns a project by ID.
// GET /api/v1/projects/{id}
func (h *ProjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get project: "+err.Error())
		return
	}
	if p == nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

// Update updates a project.
// PUT /api/v1/projects/{id}
func (h *ProjectsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var p models.Project
	if err := decodeJSON(r, &p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if err := h.repo.Update(r.Context(), id, &p); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update project: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, p)
}

// Delete deletes a project.
// DELETE /api/v1/projects/{id}
func (h *ProjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete project: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
}
