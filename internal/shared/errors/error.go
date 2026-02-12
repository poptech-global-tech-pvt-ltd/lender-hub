package errors

import "fmt"

// DomainError represents a business-level error with HTTP status mapping
type DomainError struct {
	Code       string
	Message    string
	Status     int
	Retryable  bool
	Suggestion string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code string, status int, msg string) *DomainError {
	return &DomainError{Code: code, Status: status, Message: msg}
}

func NewRetryable(code string, status int, msg, suggestion string) *DomainError {
	return &DomainError{
		Code: code, Status: status, Message: msg,
		Retryable: true, Suggestion: suggestion,
	}
}
