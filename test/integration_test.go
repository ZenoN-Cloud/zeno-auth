package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_HealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	baseURL := getBaseURL()
	if !isServerAvailable(t, baseURL) {
		t.Skip("Server not available at " + baseURL)
	}

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(baseURL + "/health")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "zeno-auth", response["service"])
}

func TestIntegration_JWKSEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	baseURL := getBaseURL()
	if !isServerAvailable(t, baseURL) {
		t.Skip("Server not available at " + baseURL)
	}

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(baseURL + "/jwks")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var jwks map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)
	require.NoError(t, err)
	assert.Contains(t, jwks, "keys")
}

func TestIntegration_AuthFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	baseURL := getBaseURL()
	if !isServerAvailable(t, baseURL) {
		t.Skip("Server not available at " + baseURL)
	}

	client := &http.Client{Timeout: 5 * time.Second}

	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())

	t.Run("Register", func(t *testing.T) {
		registerReq := map[string]string{
			"email":     email,
			"password":  "testpassword123",
			"full_name": "Integration Test User",
		}

		body, _ := json.Marshal(registerReq)
		resp, err := client.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var userResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&userResp)
		require.NoError(t, err)
		assert.Equal(t, email, userResp["email"])
	})

	t.Run("Login", func(t *testing.T) {
		loginReq := map[string]string{
			"email":    email,
			"password": "testpassword123",
		}

		body, _ := json.Marshal(loginReq)
		resp, err := client.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var authResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&authResp)
		require.NoError(t, err)
		assert.Contains(t, authResp, "access_token")
		assert.Contains(t, authResp, "refresh_token")
	})
}

func getBaseURL() string {
	if url := os.Getenv("TEST_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

func isServerAvailable(_ *testing.T, baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}
