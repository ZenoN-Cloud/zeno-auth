package metrics

import (
	"sync"
	"time"
)

// Metrics holds all application metrics
type Metrics struct {
	mu sync.RWMutex

	// Counters
	registrationsTotal  int64
	loginsTotal         int64
	loginFailuresTotal  int64
	tokenRefreshesTotal int64

	// Gauges
	activeSessions int64

	// Histograms (simplified - store last N values)
	requestDurations []time.Duration
	maxDurations     int
}

// New creates a new Metrics instance
func New() *Metrics {
	return &Metrics{
		maxDurations: 1000, // Keep last 1000 request durations
	}
}

// IncrementRegistrations increments the registrations counter
func (m *Metrics) IncrementRegistrations() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registrationsTotal++
}

// IncrementLogins increments the successful logins counter
func (m *Metrics) IncrementLogins() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loginsTotal++
}

// IncrementLoginFailures increments the failed logins counter
func (m *Metrics) IncrementLoginFailures() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loginFailuresTotal++
}

// IncrementTokenRefreshes increments the token refreshes counter
func (m *Metrics) IncrementTokenRefreshes() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokenRefreshesTotal++
}

// SetActiveSessions sets the current number of active sessions
func (m *Metrics) SetActiveSessions(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeSessions = count
}

// RecordRequestDuration records a request duration
func (m *Metrics) RecordRequestDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestDurations = append(m.requestDurations, duration)
	if len(m.requestDurations) > m.maxDurations {
		m.requestDurations = m.requestDurations[1:]
	}
}

// GetMetrics returns current metrics snapshot
func (m *Metrics) GetMetrics() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MetricsSnapshot{
		RegistrationsTotal:  m.registrationsTotal,
		LoginsTotal:         m.loginsTotal,
		LoginFailuresTotal:  m.loginFailuresTotal,
		TokenRefreshesTotal: m.tokenRefreshesTotal,
		ActiveSessions:      m.activeSessions,
		RequestDurations:    m.calculateDurationStats(),
	}
}

// GetMetricsInterface returns metrics as interface{} for generic handlers
func (m *Metrics) GetMetricsInterface() interface{} {
	return m.GetMetrics()
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	RegistrationsTotal  int64         `json:"registrations_total"`
	LoginsTotal         int64         `json:"logins_total"`
	LoginFailuresTotal  int64         `json:"login_failures_total"`
	TokenRefreshesTotal int64         `json:"token_refreshes_total"`
	ActiveSessions      int64         `json:"active_sessions"`
	RequestDurations    DurationStats `json:"request_durations"`
}

// DurationStats holds statistics about request durations
type DurationStats struct {
	Count   int     `json:"count"`
	Average float64 `json:"average_ms"`
	Min     float64 `json:"min_ms"`
	Max     float64 `json:"max_ms"`
	P50     float64 `json:"p50_ms"`
	P95     float64 `json:"p95_ms"`
	P99     float64 `json:"p99_ms"`
}

func (m *Metrics) calculateDurationStats() DurationStats {
	if len(m.requestDurations) == 0 {
		return DurationStats{}
	}

	var sum time.Duration
	min := m.requestDurations[0]
	max := m.requestDurations[0]

	for _, d := range m.requestDurations {
		sum += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	count := len(m.requestDurations)
	avg := sum / time.Duration(count)

	p50 := m.requestDurations[count*50/100]
	p95 := m.requestDurations[count*95/100]
	p99 := m.requestDurations[count*99/100]

	return DurationStats{
		Count:   count,
		Average: float64(avg.Milliseconds()),
		Min:     float64(min.Milliseconds()),
		Max:     float64(max.Milliseconds()),
		P50:     float64(p50.Milliseconds()),
		P95:     float64(p95.Milliseconds()),
		P99:     float64(p99.Milliseconds()),
	}
}
