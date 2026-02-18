package constants

// Lazypay API endpoints (appended to BaseURL)
const (
	PathEligibility      = "/v7/payment/eligibility"
	PathCustomerStatus   = "/v7/payment/customerStatus"
	PathCreateOnboarding = "/v7/createStandaloneOnboarding"
	PathOnboardingStatus = "/v7/onboarding/status"
	PathCreateOrder      = "/cof/v0/payment/order"
	PathOrderEnquiry     = "/cof/v0/payment/enquiry"
	PathRefund           = "/v7/refund"
	PathRefundEnquiry    = "/v3/enquiry"
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
	HeaderPlatform      = "platform"
	HeaderUserIPAddress = "userIpAddress"

	// Webhook or internal headers (if used)
	HeaderWebhookSignature = "X-Webhook-Signature"

	ContentTypeJSON = "application/json"
)
