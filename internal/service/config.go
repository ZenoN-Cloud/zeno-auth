package service

import "github.com/ZenoN-Cloud/zeno-auth/internal/config"

type Config struct {
	AccessTokenTTL  int
	RefreshTokenTTL int
}

func NewConfig(cfg *config.Config) *Config {
	return &Config{
		AccessTokenTTL:  cfg.JWT.AccessTokenTTL,
		RefreshTokenTTL: cfg.JWT.RefreshTokenTTL,
	}
}
