package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apperrors "github.com/ZenoN-Cloud/zeno-auth/internal/errors"
	"github.com/ZenoN-Cloud/zeno-auth/internal/response"
)

type GDPRService interface {
	ExportUserData(ctx context.Context, userID uuid.UUID) (interface{}, error)
	DeleteUserAccount(ctx context.Context, userID uuid.UUID) error
}

type GDPRHandler struct {
	gdprService  GDPRService
	auditService AuditService
	emailService EmailNotifier
}

type EmailNotifier interface {
	SendAccountDeletionNotification(ctx context.Context, userID uuid.UUID) error
	SendDataExportNotification(ctx context.Context, userID uuid.UUID) error
}

func NewGDPRHandler(gdprService GDPRService, auditService AuditService, emailService EmailNotifier) *GDPRHandler {
	return &GDPRHandler{
		gdprService:  gdprService,
		auditService: auditService,
		emailService: emailService,
	}
}

func (h *GDPRHandler) ExportData(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, err := uuid.Parse(userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	data, err := h.gdprService.ExportUserData(c.Request.Context(), uid)
	if err != nil {
		httpErr := apperrors.MapErrorToHTTP(err)
		response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(c.Request.Context(), &uid, "data_exported", nil, c.ClientIP(), c.Request.UserAgent())
	}

	if h.emailService != nil {
		go func() { _ = h.emailService.SendDataExportNotification(c.Request.Context(), uid) }()
	}

	response.Success(c, http.StatusOK, gin.H{"data": data})
}

func (h *GDPRHandler) DeleteAccount(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, err := uuid.Parse(userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	if err := h.gdprService.DeleteUserAccount(c.Request.Context(), uid); err != nil {
		httpErr := apperrors.MapErrorToHTTP(err)
		response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
		return
	}

	if h.auditService != nil {
		_ = h.auditService.Log(c.Request.Context(), &uid, "account_deleted", nil, c.ClientIP(), c.Request.UserAgent())
	}

	if h.emailService != nil {
		go func() { _ = h.emailService.SendAccountDeletionNotification(c.Request.Context(), uid) }()
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
