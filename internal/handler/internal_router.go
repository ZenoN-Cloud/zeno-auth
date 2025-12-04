package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/ZenoN-Cloud/zeno-auth/internal/repository"
)

func SetupInternalRouter(orgRepo repository.OrganizationRepository, logger zerolog.Logger) *chi.Mux {
	if orgRepo == nil {
		logger.Error().Msg("Organization repository is nil, cannot setup internal router")
		return nil
	}

	r := chi.NewRouter()
	if r == nil {
		logger.Error().Msg("Failed to create chi router")
		return nil
	}

	orgHandler := NewOrganizationHandlerWithRepo(orgRepo, logger)
	if orgHandler == nil {
		logger.Error().Msg("Failed to create organization handler")
		return nil
	}

	r.Route("/internal/v1", func(r chi.Router) {
		r.Put("/organizations/{org_id}/status", orgHandler.UpdateOrganizationStatus)
	})

	return r
}
