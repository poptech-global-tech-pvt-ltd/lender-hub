package constants

// Lazypay API endpoints (appended to BaseURL - matching Postman contracts)
const (
	PathOnboarding       = "/api/lazypay/cof/v0/standalone/initiate-onboarding"
	PathEligibility      = "/api/lazypay/v7/payment/eligibility"
	PathCustomerStatus   = "/api/lazypay/cof/v0/customer-status"
	PathCreateOrder      = "/api/lazypay/cof/v0/payment/order"
	PathOrderEnquiry     = "/api/lazypay/v3/enquiry"
	PathRefund           = "/api/lazypay/v0/refund"
	PathRefundEnquiry    = "/api/lazypay/v3/enquiry" // Same as OrderEnquiry
	PathOnboardingStatus = "/v7/onboarding/status"   // Keep old path if still used
)

// Lazypay error codes (provider-specific)
const (
	LPErrUserBlocked          = "LP_USER_BLOCKED"
	LPErrInsufficientBalance  = "COF_INSUFFICIENT_BALANCE"
	LPErrInvalidMobile        = "INVALID_MOBILE_FORMAT"
	LPErrInvalidPAN           = "INVALID_PAN_FORMAT"
	LPErrPANAlreadyRegistered = "PAN_ALREADY_REGISTERED"
	LPErrUserIneligible       = "USER_INELIGIBLE"
	LPErrBureauTimeout        = "BUREAU_TIMEOUT"
	LPErrServiceUnavailable   = "SERVICE_UNAVAILABLE"
	LPErrInternalError        = "INTERNAL_ERROR"
	LPErrRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"
	LPErrKYCFailed            = "KYC_FAILED"
	LPErrPANVerificationLimit = "PAN_VERIFICATION_LIMIT_EXHAUSTED"
	LPErrDuplicateRefund      = "LP_DUPLICATE_REFUND"
)

// HTTP headers
const (
	// Core Lazypay headers (all required)
	HeaderAccessKey     = "accessKey"
	HeaderSignature     = "signature"
	HeaderContentType   = "Content-Type"
	HeaderDeviceID      = "deviceId"
	HeaderPlatform      = "platform"
	HeaderUserIPAddress = "userIpAddress"

	// Webhook or internal headers (if used)
	HeaderWebhookSignature = "X-Webhook-Signature"

	ContentTypeJSON = "application/json"
)
