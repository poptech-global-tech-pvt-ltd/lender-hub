package circuitbreaker

import "time"

type CircuitBreakerConfig struct {
	// FailureThreshold: number of consecutive failures to trip to OPEN
	FailureThreshold int
	// Timeout: duration to stay OPEN before transitioning to HALF_OPEN
	Timeout time.Duration
	// HalfOpenRequests: number of probe requests allowed in HALF_OPEN
	HalfOpenRequests int
}

// DefaultProfileConfig returns default config for profile operations
func DefaultProfileConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		Timeout:          30 * time.Second,
		HalfOpenRequests: 1,
	}
}

// DefaultPaymentConfig returns default config for payment operations
func DefaultPaymentConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 10,
		Timeout:          15 * time.Second,
		HalfOpenRequests: 1,
	}
}
