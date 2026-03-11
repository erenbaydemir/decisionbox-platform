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

func TestLLMAPIKeyOptional(t *testing.T) {
	// API key is not required at config level — it's optional
	// (vertex-ai, ollama don't need it)
	cfg := &Config{}
	cfg.MongoDB.URI = "mongodb://localhost"

	err := cfg.Validate()
	if err != nil {
		t.Errorf("should pass without LLM_API_KEY: %v", err)
	}
}

func TestLoadWithAPIKey(t *testing.T) {
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("LLM_API_KEY", "test-key")
	defer func() {
		os.Unsetenv("MONGODB_URI")
		os.Unsetenv("LLM_API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.LLM.APIKey != "test-key" {
		t.Errorf("APIKey = %q", cfg.LLM.APIKey)
	}
}
