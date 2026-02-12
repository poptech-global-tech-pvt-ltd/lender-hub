package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"lending-hub-service/internal/infrastructure/http/circuitbreaker"
	"lending-hub-service/internal/infrastructure/http/retry"
)

// paymentExecutor implements HttpExecutor for Payment operations
type paymentExecutor struct {
	client   *http.Client
	breaker  *circuitbreaker.CircuitBreaker
	retryCfg retry.Config
}

// NewPaymentExecutor creates a new payment HTTP executor
func NewPaymentExecutor() HttpExecutor {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
	breaker := circuitbreaker.New(circuitbreaker.DefaultPaymentConfig())
	retryCfg := retry.Config{
		MaxAttempts:  2,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		JitterFactor:    0.1,
	}
	return &paymentExecutor{
		client:   client,
		breaker:  breaker,
		retryCfg: retryCfg,
	}
}

// Do executes an HTTP request with circuit breaker and retry
func (e *paymentExecutor) Do(ctx context.Context, req Request) (*Response, error) {
	// Check circuit breaker
	allowed, err := e.breaker.AllowRequest()
	if err != nil {
		return nil, fmt.Errorf("circuit breaker: %w", err)
	}
	if !allowed {
		return nil, circuitbreaker.ErrCircuitOpen
	}

	start := time.Now()

	// Execute with retry
	httpResp, err := retry.Do(ctx, e.retryCfg, func(ctx context.Context, attempt int) (*http.Response, error) {
		// Build HTTP request
		httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, req.Body)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		// Set headers
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}

		// Execute request
		resp, err := e.client.Do(httpReq)
		if err != nil {
			return nil, err
		}

		return resp, nil
	})

	if err != nil {
		e.breaker.RecordFailure()
		return nil, err
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		e.breaker.RecordFailure()
		return nil, fmt.Errorf("read response body: %w", err)
	}

	// Check if response is retryable (shouldn't happen here, but check anyway)
	if retry.IsRetryable(httpResp.StatusCode) {
		e.breaker.RecordFailure()
		return nil, fmt.Errorf("non-retryable error: status %d", httpResp.StatusCode)
	}

	// Success
	e.breaker.RecordSuccess()

	duration := time.Since(start)
	return &Response{
		StatusCode: httpResp.StatusCode,
		Body:       body,
		Headers:    httpResp.Header,
		Duration:   duration,
	}, nil
}

// CircuitState returns the current circuit breaker state as a string
func (e *paymentExecutor) CircuitState() string {
	state := e.breaker.GetState()
	switch state {
	case circuitbreaker.StateClosed:
		return "CLOSED"
	case circuitbreaker.StateOpen:
		return "OPEN"
	case circuitbreaker.StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}
