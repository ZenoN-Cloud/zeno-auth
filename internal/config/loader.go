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
		Env:      getEnv("ENV", "dev"),
		AppName:  getEnv("APP_NAME", "zeno-auth"),
		Timezone: getEnv("TIMEZONE", "UTC"),
		Server: Server{
			Port: getEnv("PORT", "8080"),
		},
		Database: Database{
			URL: getEnv("DATABASE_URL", ""),
		},
		JWT: JWT{
			PrivateKey:      getEnv("JWT_PRIVATE_KEY", ""),
			PublicKey:       getEnv("JWT_PUBLIC_KEY", ""),
			AccessTokenTTL:  getEnvInt("ACCESS_TOKEN_TTL", 1800),
			RefreshTokenTTL: getEnvInt("REFRESH_TOKEN_TTL", 1209600),
		},
		Log: Log{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			File:   getEnv("LOG_FILE", "logs/app.log"),
		},
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func validate(cfg *Config) error {
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

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
