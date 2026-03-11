package executor

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Request represents an outbound HTTP request to a lender
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    io.Reader
}

// Response wraps the HTTP response with timing metadata
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
}

// HttpExecutor abstracts HTTP execution with circuit breaker + retry
type HttpExecutor interface {
	Do(ctx context.Context, req Request) (*Response, error)
	CircuitState() string // "CLOSED", "OPEN", "HALF_OPEN"
}
