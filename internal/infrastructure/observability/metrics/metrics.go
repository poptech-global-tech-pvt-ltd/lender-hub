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

// ═══════════════════════════════════════════════════════════
// Business metrics ONLY
// System metrics (API latency, DB, cache, panics) → Datadog Agent
// Provider metrics (Lazypay calls, errors, latency) → dd-trace-go
// ═══════════════════════════════════════════════════════════

// Eligibility
const (
	MetricEligibilityChecked    = "eligibility.checked"
	MetricEligibilityEligible   = "eligibility.eligible"
	MetricEligibilityIneligible = "eligibility.ineligible"
)

// Onboarding
const (
	MetricOnboardingStarted   = "onboarding.started"
	MetricOnboardingCompleted = "onboarding.completed"
	MetricOnboardingFailed    = "onboarding.failed"
)

// Order
const (
	MetricOrderCreated = "order.created"
	MetricOrderSuccess = "order.success"
	MetricOrderFailed  = "order.failed"
)

// Refund
const (
	MetricRefundInitiated = "refund.initiated"
	MetricRefundCompleted = "refund.completed"
)

// Idempotency
const (
	MetricIdempotencyDuplicate = "idempotency.duplicate"
)
