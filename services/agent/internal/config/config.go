package config

import (
	"fmt"
	"time"

	"github.com/decisionbox-io/decisionbox/libs/go-common/config"
)

// Config holds agent configuration.
// Most settings come from the project document in MongoDB.
// Only infrastructure secrets and operational config are env vars.
type Config struct {
	Service ServiceConfig
	MongoDB MongoDBConfig
	LLM     LLMConfig
}

type ServiceConfig struct {
	Name        string
	Environment string
	LogLevel    string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

// LLMConfig holds secrets that are NOT stored in the database.
// Provider and model come from the project config in MongoDB.
type LLMConfig struct {
	APIKey         string
	MaxRetries     int
	Timeout        time.Duration
	RequestDelayMs int
}

// Load reads configuration from environment variables.
// Only reads infrastructure config — warehouse/LLM provider settings
// come from the project document in MongoDB.
func Load() (*Config, error) {
	cfg := &Config{
		Service: ServiceConfig{
			Name:        config.GetEnvOrDefault("SERVICE_NAME", "decisionbox-agent"),
			Environment: config.GetEnvOrDefault("ENV", "dev"),
			LogLevel:    config.GetEnvOrDefault("LOG_LEVEL", "info"),
		},
		MongoDB: MongoDBConfig{
			URI:      config.GetEnv("MONGODB_URI"),
			Database: config.GetEnvOrDefault("MONGODB_DB", "decisionbox"),
		},
		LLM: LLMConfig{
			APIKey:         config.GetEnv("LLM_API_KEY"),
			MaxRetries:     config.GetEnvAsInt("LLM_MAX_RETRIES", 3),
			Timeout:        parseDuration("LLM_TIMEOUT", "120s"),
			RequestDelayMs: config.GetEnvAsInt("LLM_REQUEST_DELAY_MS", 1000),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.MongoDB.URI == "" {
		return fmt.Errorf("MONGODB_URI is required")
	}
	return nil
}

func (c *Config) IsProduction() bool {
	return c.Service.Environment == "prod" || c.Service.Environment == "production"
}

func parseDuration(key, defaultVal string) time.Duration {
	val := config.GetEnvOrDefault(key, defaultVal)
	d, err := time.ParseDuration(val)
	if err != nil {
		d, _ = time.ParseDuration(defaultVal)
	}
	return d
}
