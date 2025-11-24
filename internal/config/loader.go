package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	env := os.Getenv("ENV")
	if env != "prod" && env != "production" {
		_ = godotenv.Load()
	}

	cfg := &Config{
		Env:             getEnv("ENV", "dev"),
		AppName:         getEnv("APP_NAME", "zeno-auth"),
		Timezone:        getEnv("TIMEZONE", "UTC"),
		FrontendBaseURL: getEnv("FRONTEND_BASE_URL", "http://localhost:5173"),
		Server: Server{
			Port:               getEnv("PORT", "8080"),
			CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
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
	// Cleanup jobs can skip DB + JWT validation
	skipJWT := os.Getenv("SKIP_JWT_VALIDATION") == "1"

	if !skipJWT && cfg.JWT.PrivateKey == "" {
		return fmt.Errorf("JWT_PRIVATE_KEY is required")
	}

	if port, err := strconv.Atoi(cfg.Server.Port); err != nil || port <= 0 {
		return fmt.Errorf("PORT must be a valid positive integer")
	}

	if !skipJWT && cfg.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("ACCESS_TOKEN_TTL must be positive")
	}

	if cfg.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("REFRESH_TOKEN_TTL must be positive")
	}

	validEnvs := map[string]bool{
		"dev":         true,
		"development": true,
		"staging":     true,
		"prod":        true,
		"production":  true,
		"test":        true,
	}
	if !validEnvs[cfg.Env] {
		return fmt.Errorf("ENV must be one of: dev, development, staging, prod, production, test")
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

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		var result []string
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}
