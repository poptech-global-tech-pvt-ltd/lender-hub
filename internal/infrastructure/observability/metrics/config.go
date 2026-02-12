package metrics

// DatadogConfig holds Datadog statsd configuration
type DatadogConfig struct {
	Address   string   `yaml:"address"`   // "localhost:8125"
	Namespace string   `yaml:"namespace"` // "payin3"
	Enabled   bool     `yaml:"enabled"`
	Tags      []string `yaml:"tags"`      // global tags: env, service, version
}

// DefaultConfig returns default Datadog configuration
func DefaultConfig() DatadogConfig {
	return DatadogConfig{
		Address:   "localhost:8125",
		Namespace: "payin3",
		Enabled:   false,
		Tags:      []string{"service:payin3-service"},
	}
}
