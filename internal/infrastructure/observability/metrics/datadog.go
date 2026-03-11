package metrics

import (
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

// DatadogClient wraps the statsd client for Datadog
type DatadogClient struct {
	client *statsd.Client
}

// NewDatadogClient creates a new Datadog metrics client
func NewDatadogClient(cfg DatadogConfig) (*DatadogClient, error) {
	client, err := statsd.New(cfg.Address,
		statsd.WithNamespace(cfg.Namespace+"."),
		statsd.WithTags(cfg.Tags),
	)
	if err != nil {
		return nil, err
	}
	return &DatadogClient{client: client}, nil
}

// Count increments a counter metric
func (d *DatadogClient) Count(name string, value int64, tags []string) {
	_ = d.client.Count(name, value, tags, 1)
}

// Gauge sets a gauge metric
func (d *DatadogClient) Gauge(name string, value float64, tags []string) {
	_ = d.client.Gauge(name, value, tags, 1)
}

// Histogram records a histogram metric
func (d *DatadogClient) Histogram(name string, value float64, tags []string) {
	_ = d.client.Histogram(name, value, tags, 1)
}

// Timing records a timing metric
func (d *DatadogClient) Timing(name string, duration time.Duration, tags []string) {
	_ = d.client.Timing(name, duration, tags, 1)
}

// Close gracefully shuts down the Datadog client
func (d *DatadogClient) Close() error {
	return d.client.Close()
}

// Verify interface compliance
var _ MetricsClient = (*DatadogClient)(nil)
