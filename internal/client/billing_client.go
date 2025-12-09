package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type BillingClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewBillingClient(baseURL string) *BillingClient {
	return &BillingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CreateTrialRequest struct {
	OrgID string `json:"org_id"`
}

type CreateTrialResponse struct {
	SubscriptionID string `json:"subscription_id"`
	TrialEndsAt    string `json:"trial_ends_at"`
}

// CreateTrialSubscription creates a trial subscription for a new organization
// Note: The billing service automatically creates a trial subscription when
// it's first accessed, so we just make a dummy request to trigger that logic.
func (c *BillingClient) CreateTrialSubscription(ctx context.Context, orgID uuid.UUID) error {
	if c.baseURL == "" {
		log.Warn().Msg("Billing service URL not configured, skipping trial creation")
		return nil
	}

	// Call the legacy org subscription endpoint to trigger trial creation
	url := fmt.Sprintf("%s/v1/billing/org/%s", c.baseURL, orgID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str("org_id", orgID.String()).Msg("Failed to call billing service for trial creation")
		// Don't fail registration if billing service is down
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Warn().
			Int("status", resp.StatusCode).
			Str("body", string(bodyBytes)).
			Str("org_id", orgID.String()).
			Msg("Billing service returned non-OK status (trial may be created on first real request)")
		// Don't fail registration - trial will be created later
		return nil
	}

	log.Info().Str("org_id", orgID.String()).Msg("Triggered trial subscription creation in billing service")
	return nil
}
