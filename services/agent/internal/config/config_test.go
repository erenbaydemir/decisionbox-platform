package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	defer os.Unsetenv("MONGODB_URI")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Service.Name != "decisionbox-agent" {
		t.Errorf("Name = %q", cfg.Service.Name)
	}
	if cfg.MongoDB.Database != "decisionbox" {
		t.Errorf("Database = %q", cfg.MongoDB.Database)
	}
	if cfg.LLM.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d", cfg.LLM.MaxRetries)
	}
}

func TestValidateMissingMongoDB(t *testing.T) {
	cfg := &Config{}
	err := cfg.Validate()
	if err == nil {
		t.Error("should fail without MONGODB_URI")
	}
}

func TestValidateSuccess(t *testing.T) {
	cfg := &Config{}
	cfg.MongoDB.URI = "mongodb://localhost"

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() error: %v", err)
	}
}

func TestIsProduction(t *testing.T) {
	cfg := &Config{}

	cfg.Service.Environment = "dev"
	if cfg.IsProduction() {
		t.Error("dev should not be production")
	}

	cfg.Service.Environment = "prod"
	if !cfg.IsProduction() {
		t.Error("prod should be production")
	}
}

func TestLLMConfigDefaults(t *testing.T) {
	// LLM config has no API key — secrets come from secret provider
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	defer os.Unsetenv("MONGODB_URI")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.LLM.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", cfg.LLM.MaxRetries)
	}
}
