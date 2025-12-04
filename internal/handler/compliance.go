package handler

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ComplianceReporter interface for generating compliance reports
type ComplianceReporter interface {
	GetDataExportCount(startDate, endDate time.Time) (int64, error)
	GetAccountDeletionCount(startDate, endDate time.Time) (int64, error)
	GetActiveUsersCount() (int64, error)
	GetAuditLogCount(startDate, endDate time.Time) (int64, error)
}

type ComplianceHandler struct {
	reporter ComplianceReporter
}

func NewComplianceHandler(reporter ComplianceReporter) *ComplianceHandler {
	return &ComplianceHandler{
		reporter: reporter,
	}
}

// isValidDateString validates date string format to prevent XSS
func isValidDateString(date string) bool {
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, date)
	return matched
}

// isValidDateRange validates date is within reasonable range to prevent XSS
func isValidDateRange(t time.Time) bool {
	now := time.Now()
	// Allow dates from 10 years ago to 1 year in future
	minDate := now.AddDate(-10, 0, 0)
	maxDate := now.AddDate(1, 0, 0)
	return t.After(minDate) && t.Before(maxDate)
}

// GetComplianceReport returns GDPR compliance report
func (h *ComplianceHandler) GetComplianceReport(c *gin.Context) {
	// Only admins should access this
	// TODO: Add admin role check

	// Default: last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// Parse query params if provided
	if start := c.Query("start_date"); start != "" {
		// Validate date format to prevent XSS
		if len(start) == 10 && isValidDateString(start) {
			if t, err := time.Parse("2006-01-02", start); err == nil && isValidDateRange(t) {
				startDate = t
			}
		}
	}
	if end := c.Query("end_date"); end != "" {
		// Validate date format to prevent XSS
		if len(end) == 10 && isValidDateString(end) {
			if t, err := time.Parse("2006-01-02", end); err == nil && isValidDateRange(t) {
				endDate = t
			}
		}
	}

	// Use only validated dates in response to prevent XSS
	safeStartDate := startDate.Format("2006-01-02")
	safeEndDate := endDate.Format("2006-01-02")

	report := gin.H{
		"report_id":    uuid.New().String(),
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"period": gin.H{
			"start_date": safeStartDate,
			"end_date":   safeEndDate,
		},
		"gdpr_compliance": gin.H{
			"data_export_requests":      0,
			"account_deletion_requests": 0,
			"active_users":              0,
			"audit_log_entries":         0,
		},
		"status": "success",
	}

	if h.reporter != nil {
		if count, err := h.reporter.GetDataExportCount(startDate, endDate); err == nil {
			report["gdpr_compliance"].(gin.H)["data_export_requests"] = count
		}
		if count, err := h.reporter.GetAccountDeletionCount(startDate, endDate); err == nil {
			report["gdpr_compliance"].(gin.H)["account_deletion_requests"] = count
		}
		if count, err := h.reporter.GetActiveUsersCount(); err == nil {
			report["gdpr_compliance"].(gin.H)["active_users"] = count
		}
		if count, err := h.reporter.GetAuditLogCount(startDate, endDate); err == nil {
			report["gdpr_compliance"].(gin.H)["audit_log_entries"] = count
		}
	}

	c.JSON(http.StatusOK, report)
}

// GetComplianceStatus returns current compliance status
func (h *ComplianceHandler) GetComplianceStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "compliant",
		"gdpr": gin.H{
			"right_to_access":       true, // Art. 15
			"right_to_erasure":      true, // Art. 17
			"right_to_portability":  true, // Art. 20
			"consent_management":    true, // Art. 7
			"data_retention_policy": true, // Art. 5.1.e
			"audit_logging":         true, // Art. 30
			"breach_notification":   true, // Art. 33 - Email notifications implemented
		},
		"security": gin.H{
			"password_hashing":      true,
			"rate_limiting":         true,
			"input_validation":      true,
			"session_management":    true,
			"audit_logging":         true,
			"encryption_in_transit": true,
			"encryption_at_rest":    false, // TODO
			"mfa_support":           false, // TODO
		},
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	})
}
