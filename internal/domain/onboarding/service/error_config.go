package service

import (
	"net/http"
	"time"

	"lending-hub-service/internal/domain/onboarding/entity"
)

// ErrorClassification defines how errors are classified and handled
type ErrorClassification struct {
	CanonicalCode string
	IsRetryable   bool
	Status        entity.OnboardingStatus // FAILED or INELIGIBLE
	HTTPStatus    int
	UserMessage   string
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	MaxRetries    int
}

// ErrorConfig maps error codes to their classification
var ErrorConfig = map[string]ErrorClassification{
	"INVALID_MOBILE_FORMAT": {
		CanonicalCode: "INVALID_MOBILE_FORMAT",
		IsRetryable:   false,
		Status:        entity.OnboardingFailed,
		HTTPStatus:    http.StatusBadRequest,
		UserMessage:   "Invalid mobile number format",
		InitialDelay:  0,
		MaxDelay:      0,
		MaxRetries:    0,
	},
	"PAN_ALREADY_REGISTERED": {
		CanonicalCode: "PAN_ALREADY_REGISTERED",
		IsRetryable:   false,
		Status:        entity.OnboardingIneligible,
		HTTPStatus:    http.StatusUnprocessableEntity,
		UserMessage:   "PAN already registered with another account",
		InitialDelay:  0,
		MaxDelay:      0,
		MaxRetries:    0,
	},
	"USER_INELIGIBLE": {
		CanonicalCode: "USER_INELIGIBLE",
		IsRetryable:   false,
		Status:        entity.OnboardingIneligible,
		HTTPStatus:    http.StatusUnprocessableEntity,
		UserMessage:   "User is not eligible for Pay-in-3",
		InitialDelay:  0,
		MaxDelay:      0,
		MaxRetries:    0,
	},
	"BUREAU_TIMEOUT": {
		CanonicalCode: "BUREAU_TIMEOUT",
		IsRetryable:   true,
		Status:        entity.OnboardingFailed,
		HTTPStatus:    http.StatusInternalServerError,
		UserMessage:   "Credit bureau check timed out. Please try again later.",
		InitialDelay:  5 * time.Second,
		MaxDelay:      1 * time.Hour,
		MaxRetries:    5,
	},
	"SERVICE_UNAVAILABLE": {
		CanonicalCode: "SERVICE_UNAVAILABLE",
		IsRetryable:   true,
		Status:        entity.OnboardingFailed,
		HTTPStatus:    http.StatusServiceUnavailable,
		UserMessage:   "Service temporarily unavailable. Please try again later.",
		InitialDelay:  5 * time.Second,
		MaxDelay:      1 * time.Hour,
		MaxRetries:    5,
	},
	"KYC_FAILED": {
		CanonicalCode: "KYC_FAILED",
		IsRetryable:   true,
		Status:        entity.OnboardingFailed,
		HTTPStatus:    http.StatusUnprocessableEntity,
		UserMessage:   "KYC verification failed. Please check your details and try again.",
		InitialDelay:  30 * time.Minute,
		MaxDelay:      48 * time.Hour,
		MaxRetries:    2,
	},
	"INVALID_PAN_FORMAT": {
		CanonicalCode: "INVALID_PAN_FORMAT",
		IsRetryable:   false,
		Status:        entity.OnboardingFailed,
		HTTPStatus:    http.StatusBadRequest,
		UserMessage:   "Invalid PAN format",
		InitialDelay:  0,
		MaxDelay:      0,
		MaxRetries:    0,
	},
	"KFS_FAILED": {
		CanonicalCode: "KFS_FAILED",
		IsRetryable:   true,
		Status:        entity.OnboardingFailed,
		HTTPStatus:    http.StatusUnprocessableEntity,
		UserMessage:   "KFS verification failed. Please try again.",
		InitialDelay:  30 * time.Minute,
		MaxDelay:      48 * time.Hour,
		MaxRetries:    2,
	},
	"MITC_FAILED": {
		CanonicalCode: "MITC_FAILED",
		IsRetryable:   true,
		Status:        entity.OnboardingFailed,
		HTTPStatus:    http.StatusUnprocessableEntity,
		UserMessage:   "MITC acceptance failed. Please try again.",
		InitialDelay:  30 * time.Minute,
		MaxDelay:      48 * time.Hour,
		MaxRetries:    2,
	},
	"PROVIDER_ERROR": {
		CanonicalCode: "PROVIDER_ERROR",
		IsRetryable:   true,
		Status:        entity.OnboardingFailed,
		HTTPStatus:    http.StatusInternalServerError,
		UserMessage:   "Internal provider error. Please try again later.",
		InitialDelay:  5 * time.Second,
		MaxDelay:      1 * time.Hour,
		MaxRetries:    5,
	},
}

// ClassifyError returns the error classification for a given error code
func ClassifyError(errorCode string) (ErrorClassification, bool) {
	classification, found := ErrorConfig[errorCode]
	return classification, found
}

// GetDefaultClassification returns a default classification for unknown errors
func GetDefaultClassification() ErrorClassification {
	return ErrorClassification{
		CanonicalCode: "UNKNOWN_ERROR",
		IsRetryable:   false,
		Status:        entity.OnboardingFailed,
		HTTPStatus:    http.StatusInternalServerError,
		UserMessage:   "An unexpected error occurred",
		InitialDelay:  0,
		MaxDelay:      0,
		MaxRetries:    0,
	}
}
