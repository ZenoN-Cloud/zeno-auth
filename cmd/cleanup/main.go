package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ZenoN-Cloud/zeno-auth/internal/config"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
)

func main() {
	retentionDays := flag.Int("retention-days", 730, "Audit log retention in days (default: 730 = 2 years)")
	flag.Parse()

	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting cleanup job")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Connect to database
	db, err := postgres.New(cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize repositories
	refreshTokenRepo := postgres.NewRefreshTokenRepo(db)
	auditLogRepo := postgres.NewAuditLogRepository(db.Pool())

	// Initialize cleanup service
	cleanupService := service.NewCleanupService(refreshTokenRepo, auditLogRepo)

	ctx := context.Background()

	// Cleanup expired refresh tokens
	log.Info().Msg("Cleaning up expired refresh tokens")
	if _, err := cleanupService.CleanupExpiredTokens(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to cleanup expired tokens")
	} else {
		log.Info().Msg("Expired tokens cleaned up successfully")
	}

	// Cleanup old audit logs
	log.Info().Int("retention_days", *retentionDays).Msg("Cleaning up old audit logs")
	if _, err := cleanupService.CleanupOldAuditLogs(ctx, *retentionDays); err != nil {
		log.Error().Err(err).Msg("Failed to cleanup old audit logs")
	} else {
		log.Info().Msg("Old audit logs cleaned up successfully")
	}

	// Cleanup expired email verifications
	log.Info().Msg("Cleaning up expired email verifications")
	emailVerificationRepo := postgres.NewEmailVerificationRepository(db.Pool())
	if err := emailVerificationRepo.DeleteExpired(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to cleanup expired email verifications")
	} else {
		log.Info().Msg("Expired email verifications cleaned up successfully")
	}

	// Cleanup expired password reset tokens
	log.Info().Msg("Cleaning up expired password reset tokens")
	passwordResetRepo := postgres.NewPasswordResetRepository(db.Pool())
	if err := passwordResetRepo.DeleteExpired(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to cleanup expired password reset tokens")
	} else {
		log.Info().Msg("Expired password reset tokens cleaned up successfully")
	}

	log.Info().Msg("Cleanup job completed successfully")
}
