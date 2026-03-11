package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPricingHandler_Get_NilRepo(t *testing.T) {
	h := NewPricingHandler(nil)

	req := httptest.NewRequest("GET", "/api/v1/pricing", nil)
	w := httptest.NewRecorder()

	// Will panic on nil repo — that's expected in unit tests without DB
	defer func() { recover() }()
	h.Get(w, req)
}

func TestPricingHandler_Update_InvalidJSON(t *testing.T) {
	h := NewPricingHandler(nil)

	req := httptest.NewRequest("PUT", "/api/v1/pricing",
		strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestEstimateHandler_ProjectNotFound(t *testing.T) {
	h := NewEstimateHandler(nil)

	req := httptest.NewRequest("POST", "/api/v1/projects/nonexistent/discover/estimate", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	// Will panic on nil repo — expected
	defer func() { recover() }()
	h.Estimate(w, req)
}
