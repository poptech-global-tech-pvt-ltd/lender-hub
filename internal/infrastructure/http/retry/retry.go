package retry

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxAttempts  int           // Total attempts (including first)
	InitialDelay time.Duration
	MaxDelay     time.Duration
	JitterFactor float64 // 0.0-1.0
}

// DefaultConfig returns default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		JitterFactor: 0.1,
	}
}

// IsRetryable checks if the HTTP status code warrants a retry
// Only retry on: 502, 503, 504, 429, or 0 (network error)
func IsRetryable(statusCode int) bool {
	return statusCode == 0 || statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout ||
		statusCode == http.StatusTooManyRequests
}

// CalculateDelay returns the backoff delay for attempt n (0-indexed)
func CalculateDelay(cfg Config, attempt int) time.Duration {
	delay := float64(cfg.InitialDelay) * math.Pow(2, float64(attempt))
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}
	jitter := delay * cfg.JitterFactor * rand.Float64()
	return time.Duration(delay + jitter)
}

// Do executes fn with retries. fn receives attempt number (0-based).
// Returns the last result and error.
func Do(ctx context.Context, cfg Config, fn func(ctx context.Context, attempt int) (*http.Response, error)) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		lastResp, lastErr = fn(ctx, attempt)

		// Success or non-retryable
		if lastErr == nil && lastResp != nil && !IsRetryable(lastResp.StatusCode) {
			return lastResp, nil
		}

		// Don't sleep after last attempt
		if attempt < cfg.MaxAttempts-1 {
			delay := CalculateDelay(cfg, attempt)
			select {
			case <-ctx.Done():
				return lastResp, ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return lastResp, lastErr
}
