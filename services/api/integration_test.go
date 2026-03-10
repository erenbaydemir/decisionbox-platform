//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	gomongo "github.com/decisionbox-io/decisionbox/libs/go-common/mongodb"
	"github.com/decisionbox-io/decisionbox/services/api/internal/database"
	"github.com/decisionbox-io/decisionbox/services/api/internal/handler"
	"github.com/decisionbox-io/decisionbox/services/api/internal/models"
	"github.com/decisionbox-io/decisionbox/services/api/internal/server"
	tcmongo "github.com/testcontainers/testcontainers-go/modules/mongodb"

	_ "github.com/decisionbox-io/decisionbox/domain-packs/gaming/go"
)

var testServer *httptest.Server
var testDB *database.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcmongo.Run(ctx, "mongo:7.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "MongoDB start failed: %v\n", err)
		os.Exit(1)
	}
	defer container.Terminate(ctx)

	uri, _ := container.ConnectionString(ctx)
	mongoCfg := gomongo.DefaultConfig()
	mongoCfg.URI = uri
	mongoCfg.Database = "api_integration_test"

	client, err := gomongo.NewClient(ctx, mongoCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MongoDB connect failed: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect(ctx)

	testDB = database.New(client)
	testServer = httptest.NewServer(server.New(testDB))
	defer testServer.Close()

	os.Exit(m.Run())
}

func doRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, testServer.URL+path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func decodeResponse(t *testing.T, resp *http.Response) handler.APIResponse {
	t.Helper()
	var r handler.APIResponse
	json.NewDecoder(resp.Body).Decode(&r)
	resp.Body.Close()
	return r
}

// --- Health ---

func TestInteg_Health(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/health", nil)
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

// --- Domains ---

func TestInteg_ListDomains(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/domains", nil)
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
	r := decodeResponse(t, resp)
	domains := r.Data.([]interface{})
	if len(domains) == 0 {
		t.Error("should return gaming domain")
	}
}

func TestInteg_ListCategories(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/domains/gaming/categories", nil)
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

func TestInteg_GetProfileSchema(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/domains/gaming/categories/match3/schema", nil)
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

func TestInteg_GetAnalysisAreas(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/domains/gaming/categories/match3/areas", nil)
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

func TestInteg_DomainNotFound(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/domains/nonexistent/categories", nil)
	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// --- Projects CRUD ---

func TestInteg_ProjectCRUD(t *testing.T) {
	// Create
	project := models.Project{
		Name:     "Test Game",
		Domain:   "gaming",
		Category: "match3",
		Warehouse: models.WarehouseConfig{Provider: "bigquery", Dataset: "test"},
		LLM:       models.LLMConfig{Provider: "claude", Model: "claude-sonnet-4-20250514"},
	}

	resp := doRequest(t, "POST", "/api/v1/projects", project)
	if resp.StatusCode != 201 {
		t.Fatalf("create status = %d", resp.StatusCode)
	}
	r := decodeResponse(t, resp)
	created := r.Data.(map[string]interface{})
	id := created["id"].(string)
	if id == "" {
		t.Fatal("should return ID")
	}

	// Get
	resp = doRequest(t, "GET", "/api/v1/projects/"+id, nil)
	if resp.StatusCode != 200 {
		t.Errorf("get status = %d", resp.StatusCode)
	}
	r = decodeResponse(t, resp)
	got := r.Data.(map[string]interface{})
	if got["name"] != "Test Game" {
		t.Errorf("name = %v", got["name"])
	}

	// List
	resp = doRequest(t, "GET", "/api/v1/projects", nil)
	if resp.StatusCode != 200 {
		t.Errorf("list status = %d", resp.StatusCode)
	}
	r = decodeResponse(t, resp)
	list := r.Data.([]interface{})
	if len(list) == 0 {
		t.Error("should have projects")
	}

	// Update
	project.Name = "Updated Game"
	resp = doRequest(t, "PUT", "/api/v1/projects/"+id, project)
	if resp.StatusCode != 200 {
		t.Errorf("update status = %d", resp.StatusCode)
	}

	// Delete
	resp = doRequest(t, "DELETE", "/api/v1/projects/"+id, nil)
	if resp.StatusCode != 200 {
		t.Errorf("delete status = %d", resp.StatusCode)
	}

	// Get after delete — should be gone
	resp = doRequest(t, "GET", "/api/v1/projects/"+id, nil)
	if resp.StatusCode != 404 {
		t.Errorf("after delete status = %d, want 404", resp.StatusCode)
	}
}

func TestInteg_ProjectCreate_Validation(t *testing.T) {
	// Missing name
	resp := doRequest(t, "POST", "/api/v1/projects", models.Project{Domain: "gaming", Category: "match3"})
	if resp.StatusCode != 400 {
		t.Errorf("missing name: status = %d, want 400", resp.StatusCode)
	}

	// Missing domain
	resp = doRequest(t, "POST", "/api/v1/projects", models.Project{Name: "Test"})
	if resp.StatusCode != 400 {
		t.Errorf("missing domain: status = %d, want 400", resp.StatusCode)
	}
}

func TestInteg_ProjectNotFound(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/projects/000000000000000000000000", nil)
	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// --- Discoveries ---

func TestInteg_DiscoveryEndpoints(t *testing.T) {
	// Create a project first
	project := models.Project{
		Name: "Disc Test", Domain: "gaming", Category: "match3",
		Warehouse: models.WarehouseConfig{Provider: "bigquery", Dataset: "test"},
		LLM: models.LLMConfig{Provider: "claude", Model: "test"},
	}
	resp := doRequest(t, "POST", "/api/v1/projects", project)
	r := decodeResponse(t, resp)
	id := r.Data.(map[string]interface{})["id"].(string)

	// No discoveries yet
	resp = doRequest(t, "GET", "/api/v1/projects/"+id+"/discoveries/latest", nil)
	if resp.StatusCode != 404 {
		t.Errorf("latest before discovery: status = %d, want 404", resp.StatusCode)
	}

	// List empty
	resp = doRequest(t, "GET", "/api/v1/projects/"+id+"/discoveries", nil)
	if resp.StatusCode != 200 {
		t.Errorf("list status = %d", resp.StatusCode)
	}

	// Trigger discovery (returns accepted)
	resp = doRequest(t, "POST", "/api/v1/projects/"+id+"/discover", nil)
	if resp.StatusCode != 202 {
		t.Errorf("trigger status = %d, want 202", resp.StatusCode)
	}

	// Status
	resp = doRequest(t, "GET", "/api/v1/projects/"+id+"/status", nil)
	if resp.StatusCode != 200 {
		t.Errorf("status endpoint = %d", resp.StatusCode)
	}

	// Invalid date
	resp = doRequest(t, "GET", "/api/v1/projects/"+id+"/discoveries/not-a-date", nil)
	if resp.StatusCode != 400 {
		t.Errorf("invalid date: status = %d, want 400", resp.StatusCode)
	}

	// Insert a discovery directly to test retrieval
	discCol := testDB.Collection("discoveries")
	disc := models.DiscoveryResult{
		ProjectID: id, Domain: "gaming", Category: "match3",
		DiscoveryDate: time.Now(), TotalSteps: 50,
		Insights: []models.Insight{{ID: "i1", AnalysisArea: "churn", Name: "Test"}},
		Summary:  models.Summary{TotalInsights: 1},
		CreatedAt: time.Now(),
	}
	discCol.InsertOne(context.Background(), disc)

	// Now latest should work
	resp = doRequest(t, "GET", "/api/v1/projects/"+id+"/discoveries/latest", nil)
	if resp.StatusCode != 200 {
		t.Errorf("latest after insert: status = %d", resp.StatusCode)
	}

	// By date
	today := time.Now().Format("2006-01-02")
	resp = doRequest(t, "GET", "/api/v1/projects/"+id+"/discoveries/"+today, nil)
	if resp.StatusCode != 200 {
		t.Errorf("by date: status = %d", resp.StatusCode)
	}

	// Cleanup
	doRequest(t, "DELETE", "/api/v1/projects/"+id, nil)
}

// --- CORS ---

func TestInteg_CORS(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", testServer.URL+"/api/v1/health", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing CORS header")
	}
}
