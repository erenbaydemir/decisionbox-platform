package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFeedbackHandler_Submit_MissingFields(t *testing.T) {
	h := NewFeedbackHandler(nil)

	// Empty body
	req := httptest.NewRequest("POST", "/api/v1/discoveries/run1/feedback",
		strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("runId", "run1")
	w := httptest.NewRecorder()

	h.Submit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestFeedbackHandler_Submit_InvalidRating(t *testing.T) {
	h := NewFeedbackHandler(nil)

	req := httptest.NewRequest("POST", "/api/v1/discoveries/run1/feedback",
		strings.NewReader(`{"target_type":"insight","target_id":"0","rating":"meh"}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("runId", "run1")
	w := httptest.NewRecorder()

	h.Submit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error == "" {
		t.Error("should have error message")
	}
}

func TestFeedbackHandler_Submit_InvalidTargetType(t *testing.T) {
	h := NewFeedbackHandler(nil)

	req := httptest.NewRequest("POST", "/api/v1/discoveries/run1/feedback",
		strings.NewReader(`{"target_type":"sql_query","target_id":"0","rating":"like"}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("runId", "run1")
	w := httptest.NewRecorder()

	h.Submit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestFeedbackHandler_Submit_InvalidBody(t *testing.T) {
	h := NewFeedbackHandler(nil)

	req := httptest.NewRequest("POST", "/api/v1/discoveries/run1/feedback",
		strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("runId", "run1")
	w := httptest.NewRecorder()

	h.Submit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for invalid body", w.Code)
	}
}

func TestFeedbackHandler_Submit_ExplorationStepValid(t *testing.T) {
	h := NewFeedbackHandler(nil)

	// exploration_step should pass validation (will panic at DB, not 400)
	req := httptest.NewRequest("POST", "/api/v1/discoveries/run1/feedback",
		strings.NewReader(`{"target_type":"exploration_step","target_id":"3","rating":"like"}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("runId", "run1")
	w := httptest.NewRecorder()

	defer func() { recover() }() // nil repo will panic
	h.Submit(w, req)

	if w.Code == http.StatusBadRequest {
		t.Error("exploration_step should pass validation")
	}
}

func TestFeedbackHandler_Submit_EmptyRating(t *testing.T) {
	h := NewFeedbackHandler(nil)

	req := httptest.NewRequest("POST", "/api/v1/discoveries/run1/feedback",
		strings.NewReader(`{"target_type":"insight","target_id":"0","rating":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("runId", "run1")
	w := httptest.NewRecorder()

	h.Submit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("empty rating: status = %d, want 400", w.Code)
	}
}

func TestFeedbackHandler_Submit_EmptyTargetID(t *testing.T) {
	h := NewFeedbackHandler(nil)

	req := httptest.NewRequest("POST", "/api/v1/discoveries/run1/feedback",
		strings.NewReader(`{"target_type":"insight","target_id":"","rating":"like"}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("runId", "run1")
	w := httptest.NewRecorder()

	h.Submit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("empty target_id: status = %d, want 400", w.Code)
	}
}

func TestFeedbackHandler_List_NilRepo(t *testing.T) {
	h := NewFeedbackHandler(nil)

	req := httptest.NewRequest("GET", "/api/v1/discoveries/run1/feedback", nil)
	req.SetPathValue("runId", "run1")
	w := httptest.NewRecorder()

	// Will panic on nil repo — that's expected in unit tests without DB
	defer func() { recover() }()
	h.List(w, req)
}
