package response

import "net/http"

// Envelope is the standard API response wrapper for all endpoints
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Retryable  bool   `json:"retryable"`
	Suggestion string `json:"suggestion,omitempty"`
	RequestID  string `json:"requestId,omitempty"`
}

func OK(data interface{}) (int, Envelope) {
	return http.StatusOK, Envelope{Success: true, Data: data}
}

func Error(status int, code, msg, requestID string) (int, Envelope) {
	return status, Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:       code,
			Message:    msg,
			StatusCode: status,
			RequestID:  requestID,
		},
	}
}

func ErrorWithRetry(status int, code, msg, suggestion, requestID string, retryable bool) (int, Envelope) {
	return status, Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:       code,
			Message:    msg,
			StatusCode: status,
			Retryable:  retryable,
			Suggestion: suggestion,
			RequestID:  requestID,
		},
	}
}
