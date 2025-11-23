package config

import (
	"fmt"
	"strings"
	"time"
)

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	var errs []string

	// Database
	if c.Database.URL == "" {
		errs = append(errs, "DATABASE_URL is required")
	}

	// JWT
	if c.JWT.PrivateKey == "" && c.Env != "development" {
		errs = append(errs, "JWT_PRIVATE_KEY is required in production")
	}
	// JWT_PUBLIC_KEY is optional - we have embedded public key in internal/token/jwt_public.pem
	if c.JWT.AccessTokenTTL <= 0 {
		errs = append(errs, "JWT_ACCESS_TOKEN_TTL must be > 0")
	}
	if c.JWT.AccessTokenTTL > 86400 { // 24 hours
		errs = append(errs, "JWT_ACCESS_TOKEN_TTL too large (max 24h)")
	}
	if c.JWT.RefreshTokenTTL <= 0 {
		errs = append(errs, "JWT_REFRESH_TOKEN_TTL must be > 0")
	}
	if c.JWT.RefreshTokenTTL > 2592000 { // 30 days
		errs = append(errs, "JWT_REFRESH_TOKEN_TTL too large (max 30 days)")
	}

	// Server
	if c.Server.Port == "" {
		errs = append(errs, "SERVER_PORT is required")
	}

	// CORS - warn if wildcard in production
	if c.Env == "production" {
		for _, origin := range c.Server.CORSAllowedOrigins {
			if origin == "*" {
				errs = append(errs, "CORS wildcard (*) not allowed in production")
			}
		}
	}

	// Timezone
	if c.Timezone != "" {
		if _, err := time.LoadLocation(c.Timezone); err != nil {
			errs = append(errs, fmt.Sprintf("invalid timezone: %s", c.Timezone))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

// Validate проверяет корректность конфигурации базы данных
func (d *Database) Validate() error {
	if d.URL == "" {
		return fmt.Errorf("database URL is required")
	}
	if !strings.HasPrefix(d.URL, "postgres://") && !strings.HasPrefix(d.URL, "postgresql://") {
		return fmt.Errorf("database URL must start with postgres:// or postgresql://")
	}
	return nil
}
