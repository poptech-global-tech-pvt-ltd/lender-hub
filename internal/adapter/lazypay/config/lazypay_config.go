package config

// LazypayConfig holds Lazypay provider configuration
type LazypayConfig struct {
	BaseURL        string `yaml:"baseUrl"`
	AccessKey      string `yaml:"accessKey"`
	SecretKey      string `yaml:"secretKey"`
	MerchantID     string `yaml:"merchantId"`
	ProfileTimeout int    `yaml:"profileTimeout"` // seconds
	PaymentTimeout int    `yaml:"paymentTimeout"` // seconds
	WebhookSecret  string `yaml:"webhookSecret"`   // for verifying inbound webhooks
}

// DefaultConfig returns default Lazypay configuration
func DefaultConfig() LazypayConfig {
	return LazypayConfig{
		BaseURL:        "https://sandbox.lazypay.in",
		ProfileTimeout: 10,
		PaymentTimeout: 5,
	}
}
