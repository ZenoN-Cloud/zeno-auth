package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		existingHeader string
		expectNew      bool
	}{
		{
			name:           "generates new request ID when none provided",
			existingHeader: "",
			expectNew:      true,
		},
		{
			name:           "uses existing request ID when provided",
			existingHeader: "existing-request-id-123",
			expectNew:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingHeader != "" {
				req.Header.Set("X-Request-ID", tt.existingHeader)
			}
			c.Request = req

			// Test handler that checks request ID
			var capturedRequestID string
			handler := RequestID()
			testHandler := gin.HandlerFunc(func(c *gin.Context) {
				capturedRequestID = c.GetString("request_id")
				c.Status(http.StatusOK)
			})

			// Execute
			handler(c)
			testHandler(c)

			// Assertions
			if tt.expectNew {
				assert.NotEmpty(t, capturedRequestID)
				assert.NotEqual(t, tt.existingHeader, capturedRequestID)
				assert.Len(t, capturedRequestID, 36) // UUID length
			} else {
				assert.Equal(t, tt.existingHeader, capturedRequestID)
			}

			// Check response header
			assert.Equal(t, capturedRequestID, w.Header().Get("X-Request-ID"))
		})
	}
}

func TestRequestIDFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	var requestID string
	handler := RequestID()
	testHandler := gin.HandlerFunc(func(c *gin.Context) {
		requestID = c.GetString("request_id")
		c.Status(http.StatusOK)
	})

	handler(c)
	testHandler(c)

	// Check UUID format (8-4-4-4-12 characters)
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, requestID)
}
