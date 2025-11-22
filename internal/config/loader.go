package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	env := os.Getenv("ENV")
	if env != "prod" && env != "production" {
		_ = godotenv.Load()
	}

	cfg := &Config{
		Env:      getEnv("ENV", "dev"),
		AppName:  getEnv("APP_NAME", "zeno-auth"),
		Timezone: getEnv("TIMEZONE", "UTC"),
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
	// Skip JWT validation for cleanup jobs
	skipJWT := os.Getenv("SKIP_JWT_VALIDATION") == "1"
	if !skipJWT && cfg.JWT.PrivateKey == "" {
		return fmt.Errorf("JWT_PRIVATE_KEY is required")
	}

	// Validate port
	if port, err := strconv.Atoi(cfg.Server.Port); err != nil || port <= 0 {
		return fmt.Errorf("PORT must be a valid positive integer")
	}

	// Validate database URL (skip for cleanup jobs)
	if !skipJWT && cfg.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	// Validate token TTLs
	if cfg.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("ACCESS_TOKEN_TTL must be positive")
	}
	if cfg.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("REFRESH_TOKEN_TTL must be positive")
	}

	// Validate environment
	validEnvs := map[string]bool{"dev": true, "development": true, "staging": true, "prod": true, "production": true, "test": true}
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
		var result []string
		for _, v := range splitAndTrim(value, ",") {
			if v != "" {
				result = append(result, v)
			}
		}
		return result
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, v := range splitString(s, sep) {
		if trimmed := trimSpace(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
