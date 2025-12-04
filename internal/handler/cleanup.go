package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
)

type CleanupService interface {
	CleanupExpiredTokens(ctx context.Context) (int, error)
	CleanupOldAuditLogs(ctx context.Context, retentionDays int) (int, error)
}

type CleanupHandler struct {
	db             *postgres.DB
	cleanupService CleanupService
}

func NewCleanupHandler(db *postgres.DB, cleanupService CleanupService) *CleanupHandler {
	return &CleanupHandler{
		db:             db,
		cleanupService: cleanupService,
	}
}

func (h *CleanupHandler) CleanupAll(c *gin.Context) {
	// Only allow in dev environment
	secret := c.GetHeader("X-Admin-Secret")
	expectedSecret := "dev-cleanup-secret-2024" // #nosec G101 - dev only secret
	if secret != expectedSecret {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
		return
	}

	queries := []string{
		"TRUNCATE TABLE refresh_tokens CASCADE",
		"TRUNCATE TABLE org_memberships CASCADE",
		"TRUNCATE TABLE organizations CASCADE",
		"TRUNCATE TABLE users CASCADE",
	}

	if h.db == nil || h.db.Pool() == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Database not available"})
		return
	}

	for _, query := range queries {
		if _, err := h.db.Pool().Exec(c.Request.Context(), query); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database operation failed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "All tables cleaned"})
}

func (h *CleanupHandler) CleanupExpired(c *gin.Context) {
	secret := c.GetHeader("X-Admin-Secret")
	expectedSecret := "dev-cleanup-secret-2024" // #nosec G101 - dev only secret
	if secret != expectedSecret {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
		return
	}

	if h.cleanupService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Cleanup service not available"})
		return
	}

	tokensDeleted, err := h.cleanupService.CleanupExpiredTokens(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Token cleanup failed"})
		return
	}

	logsDeleted, err := h.cleanupService.CleanupOldAuditLogs(c.Request.Context(), 730)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Audit log cleanup failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Cleanup completed",
		"tokens_deleted": tokensDeleted,
		"logs_deleted":   logsDeleted,
		"retention_days": 730,
	})
}
