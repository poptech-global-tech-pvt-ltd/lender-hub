package errors

// Shared canonical error codes used across all modules
const (
	// Generic
	CodeInvalidRequest = "INVALID_REQUEST"
	CodeInternalError  = "INTERNAL_ERROR"
	CodeNotFound       = "NOT_FOUND"
	CodeConflict       = "CONFLICT"
	CodeUnauthorized   = "AUTHENTICATION_FAILED"
	CodeRateLimited    = "RATE_LIMIT_EXCEEDED"

	// Profile-specific
	CodeUserBlocked         = "LENDER_USER_BLOCKED"
	CodeInsufficientLimit   = "PAYIN3_INSUFFICIENT_LIMIT"
	CodeUserDataInvalid     = "PAYIN3_USER_DATA_INVALID"
	CodeUserNotFound        = "PAYIN3_USER_NOT_FOUND"
	CodeInvalidTransition   = "PAYIN3_INVALID_STATE_TRANSITION"
	CodeUserContactNotFound = "PAYIN3_USER_CONTACT_NOT_FOUND"

	// Onboarding-specific
	CodeOnboardingNotFound  = "ONBOARDING_NOT_FOUND"
	CodeInvalidMobile       = "INVALID_MOBILE_FORMAT"
	CodeInvalidPAN          = "INVALID_PAN_FORMAT"
	CodeDuplicateIdentity   = "IDENTITY_DOCUMENT_ALREADY_REGISTERED"
	CodeUserIneligible      = "USER_INELIGIBLE"
	CodeBureauTimeout       = "CREDIT_BUREAU_TIMEOUT"
	CodeServiceUnavailable  = "SERVICE_TEMPORARILY_UNAVAILABLE"
	CodeProviderError       = "INTERNAL_PROVIDER_ERROR"
	CodeKYCFailed           = "KYC_VERIFICATION_FAILED"
	CodeVerificationExhaust = "VERIFICATION_ATTEMPT_LIMIT_EXHAUSTED"

	// Order-specific
	CodeIdempotencyConflict = "IDEMPOTENCY_CONFLICT"
	CodeHashMismatch        = "IDEMPOTENCY_HASH_MISMATCH"
	CodeOrderNotFound       = "ORDER_NOT_FOUND"
	CodeOrderNotRefundable  = "PAYIN3_ORDER_NOT_REFUNDABLE"

	// Refund-specific
	CodeRefundNotFound     = "REFUND_NOT_FOUND"
	CodeRefundExceedsOrder = "REFUND_EXCEEDS_ORDER_AMOUNT"
	CodeDuplicateRefund    = "DUPLICATE_REFUND"
)
