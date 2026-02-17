package idgen

const (
	PrefixPayment    = "PAY" // payment/order id
	PrefixRefund     = "REF" // refund id
	PrefixOnboarding = "ONB" // onboarding id

	PrefixRequest     = "REQ" // request correlation id
	PrefixWebhook     = "WBH" // webhook event id
	PrefixIdempotency = "IDM" // idempotency key

	PrefixUserProfile = "USR" // user profile id (internal)
	PrefixTransaction = "TXN" // generic transaction reference
	PrefixLocking     = "LCK" // locking IDs
)

var PrefixDoc = map[string]string{
	PrefixPayment:     "Payment/Order identifier",
	PrefixRefund:      "Refund identifier",
	PrefixOnboarding:  "Onboarding session identifier",
	PrefixRequest:     "Request correlation identifier",
	PrefixWebhook:     "Webhook event identifier",
	PrefixIdempotency: "Idempotency key",
	PrefixUserProfile: "User profile identifier",
	PrefixTransaction: "Transaction reference",
	PrefixLocking:     "Locking identifier",
}
