package handler

import (
	"context"
	"net/http"
	"time"

	goembedding "github.com/decisionbox-io/decisionbox/libs/go-common/embedding"
	commonmodels "github.com/decisionbox-io/decisionbox/libs/go-common/models"
	gosecrets "github.com/decisionbox-io/decisionbox/libs/go-common/secrets"
	"github.com/decisionbox-io/decisionbox/libs/go-common/vectorstore"
	"github.com/decisionbox-io/decisionbox/services/api/internal/database"
	"github.com/google/uuid"
)

// SearchHandler handles semantic search endpoints.
type SearchHandler struct {
	projectRepo    database.ProjectRepo
	insightRepo    database.InsightRepo
	recRepo        database.RecommendationRepo
	historyRepo    database.SearchHistoryRepo
	secretProvider gosecrets.Provider
	vectorStore    vectorstore.Provider // nil if Qdrant not configured
}

func NewSearchHandler(
	projectRepo database.ProjectRepo,
	insightRepo database.InsightRepo,
	recRepo database.RecommendationRepo,
	historyRepo database.SearchHistoryRepo,
	secretProvider gosecrets.Provider,
	vectorStore vectorstore.Provider,
) *SearchHandler {
	return &SearchHandler{
		projectRepo:    projectRepo,
		insightRepo:    insightRepo,
		recRepo:        recRepo,
		historyRepo:    historyRepo,
		secretProvider: secretProvider,
		vectorStore:    vectorStore,
	}
}

type searchRequest struct {
	Query    string        `json:"query"`
	Types    []string      `json:"types,omitempty"`
	Limit    int           `json:"limit,omitempty"`
	MinScore float64       `json:"min_score,omitempty"`
	Filters  searchFilters `json:"filters,omitempty"`
}

type searchFilters struct {
	Severity     string `json:"severity,omitempty"`
	AnalysisArea string `json:"analysis_area,omitempty"`
}

type searchResponse struct {
	Results        []searchResultItem `json:"results"`
	EmbeddingModel string             `json:"embedding_model"`
}

type searchResultItem struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Score         float64 `json:"score"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Severity      string  `json:"severity,omitempty"`
	AnalysisArea  string  `json:"analysis_area,omitempty"`
	DiscoveryID   string  `json:"discovery_id"`
	DiscoveredAt  string  `json:"discovered_at,omitempty"`
	ProjectID     string  `json:"project_id,omitempty"`
	ProjectName   string  `json:"project_name,omitempty"`
}

// Search performs project-scoped semantic search.
// POST /api/v1/projects/{id}/search
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "project ID is required")
		return
	}

	if h.vectorStore == nil {
		writeError(w, http.StatusServiceUnavailable, "vector search is not configured (QDRANT_URL not set)")
		return
	}

	var req searchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	ctx := r.Context()

	// Load project to get embedding config
	project, err := h.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	if project.Embedding.Provider == "" {
		writeError(w, http.StatusBadRequest, "embedding provider not configured for this project")
		return
	}

	// Create embedding provider for this project
	embProvider, err := h.createEmbeddingProvider(ctx, project.Embedding.Provider, project.Embedding.Model, projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create embedding provider")
		return
	}

	// Embed the query
	vectors, err := embProvider.Embed(ctx, []string{req.Query})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to embed query")
		return
	}

	// Search Qdrant
	searchResults, err := h.vectorStore.Search(ctx, vectors[0], vectorstore.SearchOpts{
		ProjectIDs:     []string{projectID},
		Types:          req.Types,
		EmbeddingModel: embProvider.ModelName(),
		Severity:       req.Filters.Severity,
		AnalysisArea:   req.Filters.AnalysisArea,
		Limit:          req.Limit,
		MinScore:       req.MinScore,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	// Fetch full documents from MongoDB and build response
	items := h.enrichResults(ctx, searchResults)

	// Save search history (fire and forget)
	go h.saveSearchHistory(context.Background(), projectID, req, items)

	writeJSON(w, http.StatusOK, searchResponse{
		Results:        items,
		EmbeddingModel: embProvider.ModelName(),
	})
}

// createEmbeddingProvider creates an embedding provider for a project.
func (h *SearchHandler) createEmbeddingProvider(ctx context.Context, providerName, model, projectID string) (goembedding.Provider, error) {
	apiKey := ""
	if h.secretProvider != nil {
		key, err := h.secretProvider.Get(ctx, projectID, "embedding-api-key")
		if err == nil {
			apiKey = key
		}
	}

	return goembedding.NewProvider(providerName, goembedding.ProviderConfig{
		"api_key": apiKey,
		"model":   model,
	})
}

// enrichResults fetches full documents from MongoDB for each search result.
func (h *SearchHandler) enrichResults(ctx context.Context, results []vectorstore.SearchResult) []searchResultItem {
	items := make([]searchResultItem, 0, len(results))

	for _, sr := range results {
		docType, _ := sr.Payload["type"].(string)

		item := searchResultItem{
			ID:    sr.ID,
			Type:  docType,
			Score: sr.Score,
		}

		switch docType {
		case "insight":
			if ins, err := h.insightRepo.GetByID(ctx, sr.ID); err == nil {
				item.Name = ins.Name
				item.Description = ins.Description
				item.Severity = ins.Severity
				item.AnalysisArea = ins.AnalysisArea
				item.DiscoveryID = ins.DiscoveryID
				item.DiscoveredAt = ins.DiscoveredAt.Format(time.RFC3339)
				item.ProjectID = ins.ProjectID
			}
		case "recommendation":
			if rec, err := h.recRepo.GetByID(ctx, sr.ID); err == nil {
				item.Name = rec.Title
				item.Description = rec.Description
				item.DiscoveryID = rec.DiscoveryID
				item.ProjectID = rec.ProjectID
			}
		}

		items = append(items, item)
	}

	return items
}

// saveSearchHistory records the search for analytics.
func (h *SearchHandler) saveSearchHistory(ctx context.Context, projectID string, req searchRequest, items []searchResultItem) {
	topIDs := make([]string, 0, len(items))
	for i, item := range items {
		if i >= 5 {
			break
		}
		topIDs = append(topIDs, item.ID)
	}

	var topScore float64
	if len(items) > 0 {
		topScore = items[0].Score
	}

	entry := &commonmodels.SearchHistory{
		ID:             uuid.New().String(),
		UserID:         "anonymous", // NoAuth default — enterprise overrides
		ProjectID:      projectID,
		Query:          req.Query,
		Type:           "search",
		ResultsCount:   len(items),
		TopResultIDs:   topIDs,
		TopResultScore: topScore,
		CreatedAt:      time.Now(),
	}

	h.historyRepo.Save(ctx, entry)
}
