package idgen

const (
	// Core domain entities
	PrefixPayment    = "PAY" // Payment/Order IDs
	PrefixRefund     = "REF" // Refund IDs
	PrefixOnboarding = "ONB" // Onboarding IDs

	// Infrastructure
	PrefixRequest     = "REQ" // Request IDs (for tracing)
	PrefixWebhook     = "WBH" // Webhook event IDs
	PrefixIdempotency = "IDM" // Idempotency keys (if generated)

	// Supporting entities
	PrefixUserProfile = "USR"  // User profile IDs (internal)
	PrefixTransaction = "TXN"  // Generic transaction reference
)

// PrefixMap for validation and documentation
var PrefixMap = map[string]string{
	PrefixPayment:     "Payment/Order identifier",
	PrefixRefund:      "Refund identifier",
	PrefixOnboarding:  "Onboarding session identifier",
	PrefixRequest:     "Request correlation identifier",
	PrefixWebhook:     "Webhook event identifier",
	PrefixIdempotency: "Idempotency key",
	PrefixUserProfile: "User profile identifier",
	PrefixTransaction: "Transaction reference",
}
