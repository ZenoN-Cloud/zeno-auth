package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/ZenoN-Cloud/zeno-auth/internal/repository"
)

func SetupInternalRouter(orgRepo repository.OrganizationRepository, logger zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	orgHandler := NewOrganizationHandlerWithRepo(orgRepo, logger)

	r.Route("/internal/v1", func(r chi.Router) {
		r.Put("/organizations/{org_id}/status", orgHandler.UpdateOrganizationStatus)
	})

	return r
}
