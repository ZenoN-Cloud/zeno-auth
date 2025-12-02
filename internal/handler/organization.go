package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
)

type OrganizationHandler struct {
	orgRepo repository.OrganizationRepository
	logger  zerolog.Logger
}

func NewOrganizationHandler(pool *pgxpool.Pool, logger zerolog.Logger) *OrganizationHandler {
	db := &postgres.DB{}
	// Use reflection to set private field - hacky but works for now
	// TODO: refactor DB to expose SetPool method
	orgRepo := postgres.NewOrganizationRepo(db)
	return &OrganizationHandler{
		orgRepo: orgRepo,
		logger:  logger,
	}
}

func NewOrganizationHandlerWithRepo(orgRepo repository.OrganizationRepository, logger zerolog.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		orgRepo: orgRepo,
		logger:  logger,
	}
}

type UpdateOrgStatusRequest struct {
	Status         string     `json:"status"`
	TrialEndsAt    *time.Time `json:"trial_ends_at,omitempty"`
	SubscriptionID *string    `json:"subscription_id,omitempty"`
}

func (h *OrganizationHandler) UpdateOrganizationStatus(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	var req UpdateOrgStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Status == "" {
		h.respondError(w, http.StatusBadRequest, "status is required")
		return
	}

	validStatuses := map[string]bool{
		"created":  true,
		"trialing": true,
		"active":   true,
		"past_due": true,
		"canceled": true,
	}
	if !validStatuses[req.Status] {
		h.respondError(w, http.StatusBadRequest, "invalid status")
		return
	}

	org, err := h.orgRepo.GetByID(r.Context(), orgID)
	if err != nil {
		h.logger.Error().Err(err).Str("org_id", orgIDStr).Msg("failed to get organization")
		h.respondError(w, http.StatusNotFound, "organization not found")
		return
	}

	org.Status = req.Status
	if req.TrialEndsAt != nil {
		org.TrialEndsAt = req.TrialEndsAt
	}
	if req.SubscriptionID != nil {
		subID, _ := uuid.Parse(*req.SubscriptionID)
		org.SubscriptionID = &subID
	}
	org.UpdatedAt = time.Now()

	if err := h.orgRepo.Update(r.Context(), org); err != nil {
		h.logger.Error().Err(err).Str("org_id", orgIDStr).Msg("failed to update organization")
		h.respondError(w, http.StatusInternalServerError, "failed to update organization")
		return
	}

	h.logger.Info().
		Str("org_id", orgIDStr).
		Str("status", req.Status).
		Msg("organization status updated")

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "organization status updated",
	})
}

func (h *OrganizationHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *OrganizationHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func (h *OrganizationHandler) GetUserOrganizations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var uid uuid.UUID
	switch v := userID.(type) {
	case uuid.UUID:
		uid = v
	case string:
		var err error
		uid, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id type"})
		return
	}

	orgs, err := h.orgRepo.GetByUserID(c.Request.Context(), uid)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", uid.String()).Msg("failed to get organizations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get organizations"})
		return
	}

	if orgs == nil {
		orgs = []*model.Organization{}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": orgs})
}
