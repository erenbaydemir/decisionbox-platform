package config

import (
	"fmt"

	goconfig "github.com/decisionbox-io/decisionbox/libs/go-common/config"
)

type Config struct {
	Service  ServiceConfig
	MongoDB  MongoDBConfig
	Server   ServerConfig
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

type ServerConfig struct {
	Port string
}

func Load() (*Config, error) {
	cfg := &Config{
		Service: ServiceConfig{
			Name:        goconfig.GetEnvOrDefault("SERVICE_NAME", "decisionbox-api"),
			Environment: goconfig.GetEnvOrDefault("ENV", "dev"),
			LogLevel:    goconfig.GetEnvOrDefault("LOG_LEVEL", "info"),
		},
		MongoDB: MongoDBConfig{
			URI:      goconfig.GetEnv("MONGODB_URI"),
			Database: goconfig.GetEnvOrDefault("MONGODB_DB", "decisionbox"),
		},
		Server: ServerConfig{
			Port: goconfig.GetEnvOrDefault("PORT", "8080"),
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
