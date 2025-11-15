package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	env := os.Getenv("ENV")
	if env != "prod" {
		_ = godotenv.Load()
	}

	cfg := &Config{
		Database: Database{
			URL: getEnv("DATABASE_URL", ""),
		},
		Server: Server{
			Port: getEnv("PORT", "8080"),
		},
		JWT: JWT{
			PrivateKey: getEnv("JWT_PRIVATE_KEY", ""),
		},
		Log: Log{
			Level:    getEnv("LOG_LEVEL", "info"),
			FilePath: getEnv("LOG_FILE", "logs/app.log"),
		},
		Env: getEnv("ENV", "dev"),
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWT.PrivateKey == "" {
		return fmt.Errorf("JWT_PRIVATE_KEY is required")
	}
	if port, err := strconv.Atoi(cfg.Server.Port); err != nil || port <= 0 {
		return fmt.Errorf("PORT must be a valid positive integer")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}