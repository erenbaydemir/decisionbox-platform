package mongodb

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxPoolSize != 100 {
		t.Errorf("MaxPoolSize = %d, want 100", cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize != 10 {
		t.Errorf("MinPoolSize = %d, want 10", cfg.MinPoolSize)
	}
	if cfg.MaxConnIdleTime != 60*time.Second {
		t.Errorf("MaxConnIdleTime = %v, want 60s", cfg.MaxConnIdleTime)
	}
	if cfg.ConnectTimeout != 10*time.Second {
		t.Errorf("ConnectTimeout = %v, want 10s", cfg.ConnectTimeout)
	}
	if cfg.ServerSelectionTimeout != 10*time.Second {
		t.Errorf("ServerSelectionTimeout = %v, want 10s", cfg.ServerSelectionTimeout)
	}
}
