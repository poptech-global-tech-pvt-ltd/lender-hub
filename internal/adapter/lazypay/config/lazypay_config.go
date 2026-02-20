package config

// LazypayConfig holds Lazypay provider configuration
type LazypayConfig struct {
	BaseURL        string `yaml:"baseUrl"`
	AccessKey      string `yaml:"accessKey"`
	SecretKey      string `yaml:"secretKey"`
	MerchantID     string `yaml:"merchantId"`     // Optional - use SubMerchantID if empty
	SubMerchantID  string `yaml:"subMerchantId"`  // Used in onboarding and as fallback
	ReturnURL      string `yaml:"returnUrl"`      // Callback URL for redirects
	ProfileTimeout int    `yaml:"profileTimeout"` // seconds
	PaymentTimeout int    `yaml:"paymentTimeout"` // seconds
	WebhookSecret  string `yaml:"webhookSecret"`  // for verifying inbound webhooks
}

// GetMerchantID returns MerchantID if set, otherwise SubMerchantID as fallback
func (c *LazypayConfig) GetMerchantID() string {
	if c.MerchantID != "" {
		return c.MerchantID
	}
	return c.SubMerchantID
}

// DefaultConfig returns default Lazypay configuration
func DefaultConfig() LazypayConfig {
	return LazypayConfig{
		BaseURL:        "https://sboxapi.lazypay.in",
		ProfileTimeout: 10,
		PaymentTimeout: 5,
		MerchantID:     "270",
		SubMerchantID:  "270",
	}
}
