package validator

import "github.com/go-playground/validator/v10"

type FieldError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Param string `json:"param,omitempty"`
}

type ValidationErrorResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors"`
}

func MapError(err error) *ValidationErrorResponse {
	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return &ValidationErrorResponse{
			Code:    "PAYIN3_INVALID_REQUEST",
			Message: "Invalid request payload",
			Errors:  nil,
		}
	}

	out := make([]FieldError, 0, len(verrs))
	for _, fe := range verrs {
		out = append(out, FieldError{
			Field: fe.Field(),
			Tag:   fe.Tag(),
			Param: fe.Param(),
		})
	}

	return &ValidationErrorResponse{
		Code:    "PAYIN3_VALIDATION_FAILED",
		Message: "One or more fields are invalid",
		Errors:  out,
	}
}
