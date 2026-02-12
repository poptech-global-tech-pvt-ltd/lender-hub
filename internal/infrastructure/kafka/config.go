package kafka

// ProducerConfig holds Kafka producer configuration
type ProducerConfig struct {
	Brokers         []string `yaml:"brokers"`
	Async           bool     `yaml:"async"`
	CompressionType string   `yaml:"compressionType"` // "snappy", "gzip", "none"
	BatchSize       int      `yaml:"batchSize"`       // messages per batch
	LingerMs        int      `yaml:"lingerMs"`        // batch linger
	Retries         int      `yaml:"retries"`
	RequiredAcks    string   `yaml:"requiredAcks"`    // "all", "leader", "none"
}

// ConsumerConfig holds Kafka consumer configuration
type ConsumerConfig struct {
	Brokers     []string `yaml:"brokers"`
	GroupID     string   `yaml:"groupId"`
	AutoCommit bool     `yaml:"autoCommit"`
	OffsetReset string   `yaml:"offsetReset"` // "earliest", "latest"
}

// DefaultProducerConfig returns default producer configuration
func DefaultProducerConfig() ProducerConfig {
	return ProducerConfig{
		Brokers:         []string{"localhost:9092"},
		Async:           true,
		CompressionType: "snappy",
		BatchSize:       100,
		LingerMs:        5,
		Retries:         3,
		RequiredAcks:    "all",
	}
}

// DefaultConsumerConfig returns default consumer configuration
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		Brokers:     []string{"localhost:9092"},
		GroupID:     "payin3-service",
		AutoCommit:  true,
		OffsetReset: "earliest",
	}
}
