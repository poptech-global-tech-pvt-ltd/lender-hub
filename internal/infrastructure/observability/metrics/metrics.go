package metrics

import "time"

// MetricsClient is the shared interface for all metrics emission
type MetricsClient interface {
	Count(name string, value int64, tags []string)
	Gauge(name string, value float64, tags []string)
	Histogram(name string, value float64, tags []string)
	Timing(name string, duration time.Duration, tags []string)
	Close() error
}

// ═══════════════════════════════════════════
// Business metrics — track product outcomes
// ═══════════════════════════════════════════
const (
	MetricOrdersCreated         = "orders.created"
	MetricOrdersSuccess         = "orders.success"
	MetricOrdersFailed          = "orders.failed"
	MetricEligibilityChecks     = "eligibility.checks"
	MetricEligibilityEligible   = "eligibility.eligible"
	MetricEligibilityIneligible = "eligibility.ineligible"
	MetricOnboardingStarted     = "onboarding.started"
	MetricOnboardingCompleted   = "onboarding.completed"
	MetricOnboardingFailed      = "onboarding.failed"
	MetricRefundsInitiated      = "refunds.initiated"
	MetricRefundsCompleted      = "refunds.completed"
	MetricIdempotencyDuplicate  = "idempotency.duplicates"
)

// ═══════════════════════════════════════════
// Provider metrics — track Lazypay call health
// ═══════════════════════════════════════════
const (
	MetricLazypayRequests     = "lazypay.requests"
	MetricLazypayErrors       = "lazypay.errors"
	MetricLazypayLatency      = "lazypay.latency"
	MetricLazypayCircuitState = "lazypay.circuit_breaker.state"
	MetricLazypayTimeout      = "lazypay.timeout"
)

// ═══════════════════════════════════════════
// System metrics — track infra health
// ═══════════════════════════════════════════
const (
	MetricAPILatency    = "api.latency"
	MetricAPIRequests   = "api.requests"
	MetricAPIErrors     = "api.errors"
	MetricDBQueryTime   = "db.query.duration"
	MetricDBConnections = "db.connections.active"
	MetricCacheHit      = "cache.hit"
	MetricCacheMiss     = "cache.miss"
	MetricPanics        = "panics"
)
