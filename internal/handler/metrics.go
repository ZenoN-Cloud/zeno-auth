package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MetricsGetter interface for getting metrics
type MetricsGetter interface {
	GetMetrics() interface{}
}

// MetricsHandler handles metrics endpoint
type MetricsHandler struct {
	metrics MetricsGetter
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(m interface{}) *MetricsHandler {
	if mg, ok := m.(MetricsGetter); ok {
		return &MetricsHandler{
			metrics: mg,
		}
	}
	return &MetricsHandler{metrics: nil}
}

// GetMetrics returns current metrics
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	if h.metrics == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Metrics not available"})
		return
	}
	snapshot := h.metrics.GetMetrics()
	c.JSON(http.StatusOK, snapshot)
}
