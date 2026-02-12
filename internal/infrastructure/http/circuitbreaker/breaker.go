package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed is the normal operating state
	StateClosed State = iota
	// StateOpen means the circuit is failing and rejecting requests
	StateOpen
	// StateHalfOpen is the probing state to test recovery
	StateHalfOpen
)

var (
	// ErrCircuitOpen is returned when circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker implements a circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.Mutex
	config          Config
	state           State
	failureCount    int
	successCount    int // for half-open tracking
	lastFailureTime time.Time
	halfOpenAttempts int
}

// New creates a new circuit breaker with the given config
func New(cfg Config) *CircuitBreaker {
	return &CircuitBreaker{
		config:          cfg,
		state:           StateClosed,
		halfOpenAttempts: 0,
	}
}

// AllowRequest checks if a request should be allowed
func (cb *CircuitBreaker) AllowRequest() (bool, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true, nil

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
			// Transition to HALF_OPEN
			cb.state = StateHalfOpen
			cb.halfOpenAttempts = 0
			cb.successCount = 0
			return true, nil
		}
		return false, ErrCircuitOpen

	case StateHalfOpen:
		if cb.halfOpenAttempts < cb.config.HalfOpenRequests {
			cb.halfOpenAttempts++
			return true, nil
		}
		return false, ErrCircuitOpen

	default:
		return false, ErrCircuitOpen
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		// Transition back to CLOSED
		cb.state = StateClosed
		cb.failureCount = 0
		cb.halfOpenAttempts = 0
		cb.successCount = 0

	case StateClosed:
		// Reset failure count on success
		cb.failureCount = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
		}

	case StateHalfOpen:
		// Any failure in HALF_OPEN goes back to OPEN
		cb.state = StateOpen
		cb.lastFailureTime = time.Now()
		cb.halfOpenAttempts = 0
		cb.failureCount = cb.config.FailureThreshold
	}
}

// GetState returns the current state (for metrics)
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset forces a reset to CLOSED state (for testing)
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.halfOpenAttempts = 0
	cb.lastFailureTime = time.Time{}
}
