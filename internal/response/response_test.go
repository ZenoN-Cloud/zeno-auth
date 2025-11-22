package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		data       interface{}
	}{
		{
			name:       "success with data",
			statusCode: http.StatusOK,
			data:       map[string]interface{}{"message": "test"},
		},
		{
			name:       "success with nil data",
			statusCode: http.StatusCreated,
			data:       nil,
		},
		{
			name:       "success with string data",
			statusCode: http.StatusAccepted,
			data:       "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Success(c, tt.statusCode, tt.data)

			assert.Equal(t, tt.statusCode, w.Code)

			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "ok", response.Status)
			assert.Equal(t, tt.data, response.Data)
			assert.Empty(t, response.Code)
			assert.Empty(t, response.Message)
		})
	}
}

func TestError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		code       string
		message    string
	}{
		{
			name:       "bad request error",
			statusCode: http.StatusBadRequest,
			code:       "bad_request",
			message:    "Invalid input",
		},
		{
			name:       "unauthorized error",
			statusCode: http.StatusUnauthorized,
			code:       "unauthorized",
			message:    "Access denied",
		},
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			code:       "internal_error",
			message:    "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Error(c, tt.statusCode, tt.code, tt.message)

			assert.Equal(t, tt.statusCode, w.Code)

			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response.Status)
			assert.Equal(t, tt.code, response.Code)
			assert.Equal(t, tt.message, response.Message)
			assert.Nil(t, response.Data)
		})
	}
}

func TestConvenienceMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         func(*gin.Context, string)
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "BadRequest",
			method:         BadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "bad_request",
		},
		{
			name:           "Unauthorized",
			method:         Unauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "unauthorized",
		},
		{
			name:           "Forbidden",
			method:         Forbidden,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "forbidden",
		},
		{
			name:           "NotFound",
			method:         NotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "not_found",
		},
		{
			name:           "Conflict",
			method:         Conflict,
			expectedStatus: http.StatusConflict,
			expectedCode:   "conflict",
		},
		{
			name:           "InternalError",
			method:         InternalError,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			message := "test message"
			tt.method(c, message)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response.Status)
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Equal(t, message, response.Message)
		})
	}
}
