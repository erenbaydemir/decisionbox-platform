package logger

import (
	"fmt"
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	l := New("test-service", "info")
	if l == nil {
		t.Fatal("New() returned nil")
	}
	// Should not panic
	l.Info("test message", zap.String("key", "value"))
}

func TestNewWithLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "warning", "error", "unknown"}
	for _, level := range levels {
		l := New("test-service", level)
		if l == nil {
			t.Errorf("New() with level %q returned nil", level)
		}
	}
}

func TestNewProductionJSON(t *testing.T) {
	os.Setenv("ENV", "prod")
	defer os.Unsetenv("ENV")

	l := New("test-service", "info")
	if l == nil {
		t.Fatal("New() returned nil in prod mode")
	}
	// Should not panic — outputs JSON
	l.Info("production log")
}

func TestWith(t *testing.T) {
	l := New("test-service", "info")
	child := l.With(zap.String("app_id", "app123"))
	if child == nil {
		t.Fatal("With() returned nil")
	}
	// Should include app_id in output
	child.Info("child logger message")
}

func TestFieldConstructors(t *testing.T) {
	l := New("test-service", "debug")

	// All field constructors should not panic
	l.Info("test fields",
		AppID("app123"),
		OrgID("org456"),
		UserID("user789"),
		SessionID("sess000"),
		CorrelationID("corr111"),
	)

	err := fmt.Errorf("test error")
	l.Error("error test", Err(err))
}
