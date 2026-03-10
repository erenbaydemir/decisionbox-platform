package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockChecker struct {
	name string
	err  error
}

func (m mockChecker) Name() string                    { return m.name }
func (m mockChecker) Check(ctx context.Context) error { return m.err }

func TestLivenessHandler(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.LivenessHandler()(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var resp response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Status, "ok")
	}
}

func TestReadinessAllHealthy(t *testing.T) {
	h := NewHandler(
		mockChecker{name: "mongodb", err: nil},
		mockChecker{name: "redis", err: nil},
	)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()

	h.ReadinessHandler()(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var resp response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Status, "ok")
	}
	if len(resp.Services) != 2 {
		t.Errorf("services count = %d, want 2", len(resp.Services))
	}
}

func TestReadinessDegraded(t *testing.T) {
	h := NewHandler(
		mockChecker{name: "mongodb", err: nil},
		mockChecker{name: "redis", err: fmt.Errorf("connection refused")},
	)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()

	h.ReadinessHandler()(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}

	var resp response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "degraded" {
		t.Errorf("status = %q, want %q", resp.Status, "degraded")
	}
	if resp.Services["redis"].Status != "error" {
		t.Errorf("redis status = %q, want %q", resp.Services["redis"].Status, "error")
	}
	if resp.Services["mongodb"].Status != "ok" {
		t.Errorf("mongodb status = %q, want %q", resp.Services["mongodb"].Status, "ok")
	}
}

func TestReadinessNoCheckers(t *testing.T) {
	h := NewHandler()

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()

	h.ReadinessHandler()(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
