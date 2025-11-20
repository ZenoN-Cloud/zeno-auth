package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
)

type ConsentHandler struct {
	consentService ConsentService
}

func NewConsentHandler(consentService ConsentService) *ConsentHandler {
	return &ConsentHandler{
		consentService: consentService,
	}
}

type GrantConsentRequest struct {
	ConsentType string `json:"consent_type" binding:"required,oneof=terms privacy marketing analytics"`
	Version     string `json:"version" binding:"required"`
}

func (h *ConsentHandler) GetConsents(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	consents, err := h.consentService.GetUserConsents(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get consents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"consents": consents})
}

func (h *ConsentHandler) GrantConsent(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var req GrantConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.consentService.GrantConsent(c.Request.Context(), uid, model.ConsentType(req.ConsentType), req.Version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grant consent"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Consent granted successfully"})
}

func (h *ConsentHandler) RevokeConsent(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	consentType := c.Param("type")

	if err := h.consentService.RevokeConsent(c.Request.Context(), uid, model.ConsentType(consentType)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consent revoked successfully"})
}
