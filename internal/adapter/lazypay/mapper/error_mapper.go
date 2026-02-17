package mapper

import (
	sharedErrors "lending-hub-service/internal/shared/errors"
)

// LPErrorMapping maps Lazypay error codes → canonical DomainErrors
var LPErrorMapping = map[string]*sharedErrors.DomainError{
	"LP_USER_BLOCKED":                  sharedErrors.New(sharedErrors.CodeUserBlocked, 422, "User is blocked for Pay-in-3"),
	"COF_INSUFFICIENT_BALANCE":         sharedErrors.New(sharedErrors.CodeInsufficientLimit, 422, "Insufficient credit limit"),
	"INVALID_MOBILE_FORMAT":            sharedErrors.New(sharedErrors.CodeInvalidRequest, 400, "Invalid mobile number format"),
	"INVALID_PAN_FORMAT":                sharedErrors.New(sharedErrors.CodeInvalidRequest, 400, "Invalid PAN format"),
	"PAN_ALREADY_REGISTERED":           sharedErrors.New(sharedErrors.CodeDuplicateIdentity, 422, "PAN already registered with another account"),
	"USER_INELIGIBLE":                  sharedErrors.New(sharedErrors.CodeUserIneligible, 422, "User does not meet eligibility criteria"),
	"BUREAU_TIMEOUT":                   sharedErrors.NewRetryable(sharedErrors.CodeBureauTimeout, 500, "Credit bureau check timed out", "Please try again in a few minutes"),
	"SERVICE_UNAVAILABLE":              sharedErrors.NewRetryable(sharedErrors.CodeServiceUnavailable, 503, "Service temporarily unavailable", "Please try again shortly"),
	"INTERNAL_ERROR":                   sharedErrors.NewRetryable(sharedErrors.CodeProviderError, 500, "Provider internal error", "Please try again"),
	"RATE_LIMIT_EXCEEDED":              sharedErrors.NewRetryable(sharedErrors.CodeRateLimited, 429, "Rate limit exceeded", "Please wait before retrying"),
	"KYC_FAILED":                       sharedErrors.NewRetryable(sharedErrors.CodeKYCFailed, 422, "KYC verification failed", "Please verify your documents and try again"),
	"PAN_VERIFICATION_LIMIT_EXHAUSTED": sharedErrors.NewRetryable(sharedErrors.CodeVerificationExhaust, 422, "Verification attempt limit reached", "Please try again after cooldown period"),
	"LP_DUPLICATE_REFUND":              sharedErrors.New(sharedErrors.CodeDuplicateRefund, 409, "Refund with this ID already exists"),
}

// MapLPError converts a Lazypay error code to a canonical DomainError
func MapLPError(lpErrorCode string) *sharedErrors.DomainError {
	if de, ok := LPErrorMapping[lpErrorCode]; ok {
		return de
	}
	return sharedErrors.New(sharedErrors.CodeInternalError, 500, "Unknown provider error: "+lpErrorCode)
}
