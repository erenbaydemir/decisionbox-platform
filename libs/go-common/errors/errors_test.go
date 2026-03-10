package errors

import (
	"fmt"
	"testing"
)

func TestAppError(t *testing.T) {
	tests := []struct {
		name      string
		err       *AppError
		wantCode  int
		wantRetry bool
	}{
		{"NotFound", NotFound("not found"), 404, false},
		{"Unauthorized", Unauthorized("unauthorized"), 401, false},
		{"Forbidden", Forbidden("forbidden"), 403, false},
		{"BadRequest", BadRequest("bad request"), 400, false},
		{"Internal", Internal("server error", fmt.Errorf("cause")), 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", tt.err.Code, tt.wantCode)
			}
			if tt.err.IsRetryable() != tt.wantRetry {
				t.Errorf("IsRetryable() = %v, want %v", tt.err.IsRetryable(), tt.wantRetry)
			}
		})
	}
}

func TestAppErrorMessage(t *testing.T) {
	err := NotFound("user not found")
	if err.Error() != "user not found" {
		t.Errorf("Error() = %q, want %q", err.Error(), "user not found")
	}

	wrapped := Internal("failed", fmt.Errorf("db timeout"))
	if wrapped.Error() != "failed: db timeout" {
		t.Errorf("Error() = %q, want %q", wrapped.Error(), "failed: db timeout")
	}
}

func TestAppErrorUnwrap(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := Internal("wrapper", cause)

	if err.Unwrap() != cause {
		t.Error("Unwrap() did not return the wrapped error")
	}

	simple := NotFound("missing")
	if simple.Unwrap() != nil {
		t.Error("Unwrap() should return nil for errors without a cause")
	}
}
