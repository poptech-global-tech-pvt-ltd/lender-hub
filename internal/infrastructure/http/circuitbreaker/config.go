package circuitbreaker

import "time"

// Config holds circuit breaker configuration
type Config struct {
	// FailureThreshold: number of consecutive failures to trip to OPEN
	FailureThreshold int
	// Timeout: duration to stay OPEN before transitioning to HALF_OPEN
	Timeout time.Duration
	// HalfOpenRequests: number of probe requests allowed in HALF_OPEN
	HalfOpenRequests int
}

// DefaultProfileConfig returns default config for profile operations
func DefaultProfileConfig() Config {
	return Config{
		FailureThreshold: 5,
		Timeout:          30 * time.Second,
		HalfOpenRequests: 1,
	}
}

// DefaultPaymentConfig returns default config for payment operations
func DefaultPaymentConfig() Config {
	return Config{
		FailureThreshold: 10,
		Timeout:          15 * time.Second,
		HalfOpenRequests: 1,
	}
}
