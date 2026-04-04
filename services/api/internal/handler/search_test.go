package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	goembedding "github.com/decisionbox-io/decisionbox/libs/go-common/embedding"
	commonmodels "github.com/decisionbox-io/decisionbox/libs/go-common/models"
	gosecrets "github.com/decisionbox-io/decisionbox/libs/go-common/secrets"
	"github.com/decisionbox-io/decisionbox/libs/go-common/vectorstore"
	"github.com/decisionbox-io/decisionbox/services/api/internal/models"
)

// mockProjectRepoForSearch implements ProjectRepo with embedding config.
type mockProjectRepoForSearch struct {
	project *models.Project
}

func (m *mockProjectRepoForSearch) Create(_ context.Context, _ *models.Project) error { return nil }
func (m *mockProjectRepoForSearch) GetByID(_ context.Context, id string) (*models.Project, error) {
	if m.project != nil && m.project.ID == id {
		return m.project, nil
	}
	return nil, context.DeadlineExceeded
}
func (m *mockProjectRepoForSearch) List(_ context.Context, _, _ int) ([]*models.Project, error) {
	return nil, nil
}
func (m *mockProjectRepoForSearch) Update(_ context.Context, _ string, _ *models.Project) error {
	return nil
}
func (m *mockProjectRepoForSearch) Delete(_ context.Context, _ string) error { return nil }

// mockVectorStoreForSearch returns pre-set search results.
type mockVectorStoreForSearch struct {
	results []vectorstore.SearchResult
}

func (m *mockVectorStoreForSearch) Upsert(_ context.Context, _ []vectorstore.Point) error {
	return nil
}
func (m *mockVectorStoreForSearch) Search(_ context.Context, _ []float64, _ vectorstore.SearchOpts) ([]vectorstore.SearchResult, error) {
	return m.results, nil
}
func (m *mockVectorStoreForSearch) FindDuplicates(_ context.Context, _ []float64, _, _, _ string, _ float64) ([]vectorstore.SearchResult, error) {
	return nil, nil
}
func (m *mockVectorStoreForSearch) Delete(_ context.Context, _ []string) error      { return nil }
func (m *mockVectorStoreForSearch) HealthCheck(_ context.Context) error              { return nil }
func (m *mockVectorStoreForSearch) EnsureCollection(_ context.Context, _ int) error  { return nil }

// mockSearchHistoryRepo discards all saves.
type mockSearchHistoryRepo struct{}

func (m *mockSearchHistoryRepo) Save(_ context.Context, _ *commonmodels.SearchHistory) error {
	return nil
}
func (m *mockSearchHistoryRepo) ListByUser(_ context.Context, _ string, _ int) ([]*commonmodels.SearchHistory, error) {
	return nil, nil
}
func (m *mockSearchHistoryRepo) ListByProject(_ context.Context, _ string, _ int) ([]*commonmodels.SearchHistory, error) {
	return nil, nil
}

// mockAskSessionRepo discards all operations.
type mockAskSessionRepo struct{}

func (m *mockAskSessionRepo) Create(_ context.Context, _ *commonmodels.AskSession) error {
	return nil
}
func (m *mockAskSessionRepo) AppendMessage(_ context.Context, _ string, _ commonmodels.AskSessionMessage) error {
	return nil
}
func (m *mockAskSessionRepo) GetByID(_ context.Context, _ string) (*commonmodels.AskSession, error) {
	return nil, nil
}
func (m *mockAskSessionRepo) ListByProject(_ context.Context, _ string, _ int) ([]*commonmodels.AskSession, error) {
	return nil, nil
}
func (m *mockAskSessionRepo) Delete(_ context.Context, _ string) error { return nil }

// mockSecretProviderForSearch returns a pre-set API key.
type mockSecretProviderForSearch struct{}

func (m *mockSecretProviderForSearch) Get(_ context.Context, _, _ string) (string, error) {
	return "test-key", nil
}
func (m *mockSecretProviderForSearch) Set(_ context.Context, _, _, _ string) error   { return nil }
func (m *mockSecretProviderForSearch) Delete(_ context.Context, _, _ string) error   { return nil }
func (m *mockSecretProviderForSearch) List(_ context.Context, _ string) ([]gosecrets.SecretEntry, error) {
	return nil, nil
}

func init() {
	// Register a mock embedding provider for tests
	goembedding.RegisterWithMeta("test-embedding", func(cfg goembedding.ProviderConfig) (goembedding.Provider, error) {
		return &testEmbeddingProvider{}, nil
	}, goembedding.ProviderMeta{
		ID:   "test-embedding",
		Name: "Test Embedding",
		Models: []goembedding.ModelInfo{
			{ID: "test-model", Dimensions: 3},
		},
	})
}

type testEmbeddingProvider struct{}

func (t *testEmbeddingProvider) Embed(_ context.Context, texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i := range texts {
		result[i] = []float64{0.1, 0.2, 0.3}
	}
	return result, nil
}
func (t *testEmbeddingProvider) Dimensions() int        { return 3 }
func (t *testEmbeddingProvider) ModelName() string       { return "test-model" }
func (t *testEmbeddingProvider) Validate(_ context.Context) error { return nil }

func TestSearchHandler_Search(t *testing.T) {
	insightID := "11111111-1111-4111-8111-111111111111"

	projectRepo := &mockProjectRepoForSearch{
		project: &models.Project{
			ID:   "proj-1",
			Name: "Test Project",
			Embedding: goembedding.ProjectConfig{
				Provider: "test-embedding",
				Model:    "test-model",
			},
		},
	}

	insightRepo := &mockInsightRepo{
		insights: []*commonmodels.StandaloneInsight{
			{
				ID:           insightID,
				ProjectID:    "proj-1",
				DiscoveryID:  "disc-1",
				Name:         "High churn",
				Description:  "Players leaving",
				Severity:     "high",
				AnalysisArea: "churn",
				DiscoveredAt: time.Now(),
			},
		},
	}

	vs := &mockVectorStoreForSearch{
		results: []vectorstore.SearchResult{
			{
				ID:    insightID,
				Score: 0.89,
				Payload: map[string]interface{}{
					"type": "insight",
				},
			},
		},
	}

	h := NewSearchHandler(
		projectRepo,
		insightRepo,
		&mockRecommendationRepo{},
		&mockSearchHistoryRepo{},
		&mockAskSessionRepo{},
		&mockSecretProviderForSearch{},
		vs,
	)

	body, _ := json.Marshal(searchRequest{
		Query: "why are players leaving?",
		Limit: 10,
	})

	req := httptest.NewRequest("POST", "/api/v1/projects/proj-1/search", bytes.NewReader(body))
	req.SetPathValue("id", "proj-1")
	w := httptest.NewRecorder()

	h.Search(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	results := data["results"].([]interface{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0].(map[string]interface{})
	if result["id"] != insightID {
		t.Errorf("expected ID %s, got %v", insightID, result["id"])
	}
	if result["score"].(float64) < 0.8 {
		t.Errorf("expected score >= 0.8, got %v", result["score"])
	}
	if result["name"] != "High churn" {
		t.Errorf("expected name=High churn, got %v", result["name"])
	}
}

func TestSearchHandler_NoVectorStore(t *testing.T) {
	h := NewSearchHandler(nil, nil, nil, nil, nil, nil, nil) // no Qdrant

	body, _ := json.Marshal(searchRequest{Query: "test"})
	req := httptest.NewRequest("POST", "/api/v1/projects/proj-1/search", bytes.NewReader(body))
	req.SetPathValue("id", "proj-1")
	w := httptest.NewRecorder()

	h.Search(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestSearchHandler_EmptyQuery(t *testing.T) {
	h := NewSearchHandler(nil, nil, nil, nil, nil, nil, &mockVectorStoreForSearch{})

	body, _ := json.Marshal(searchRequest{Query: ""})
	req := httptest.NewRequest("POST", "/api/v1/projects/proj-1/search", bytes.NewReader(body))
	req.SetPathValue("id", "proj-1")
	w := httptest.NewRecorder()

	h.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSearchHandler_NoEmbeddingConfig(t *testing.T) {
	projectRepo := &mockProjectRepoForSearch{
		project: &models.Project{
			ID:   "proj-1",
			Name: "No Embedding",
		},
	}

	h := NewSearchHandler(projectRepo, nil, nil, nil, nil, nil, &mockVectorStoreForSearch{})

	body, _ := json.Marshal(searchRequest{Query: "test"})
	req := httptest.NewRequest("POST", "/api/v1/projects/proj-1/search", bytes.NewReader(body))
	req.SetPathValue("id", "proj-1")
	w := httptest.NewRecorder()

	h.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
