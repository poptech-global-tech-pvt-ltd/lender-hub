package metrics

import "time"

// NoopClient implements MetricsClient but does nothing
// Used when Datadog is not configured (local dev)
type NoopClient struct{}

// NewNoopClient creates a new noop metrics client
func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

// Count does nothing
func (c *NoopClient) Count(name string, value int64, tags []string) {}

// Gauge does nothing
func (c *NoopClient) Gauge(name string, value float64, tags []string) {}

// Histogram does nothing
func (c *NoopClient) Histogram(name string, value float64, tags []string) {}

// Timing does nothing
func (c *NoopClient) Timing(name string, duration time.Duration, tags []string) {}

// Close does nothing
func (c *NoopClient) Close() error {
	return nil
}

// Verify interface compliance
var _ MetricsClient = (*NoopClient)(nil)
