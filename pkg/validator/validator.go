package validator

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	v := validator.New()
	// Register custom tags as needed, e.g.:
	// v.RegisterValidation("currency", validateCurrency)
	return &Validator{v: v}
}

func (v *Validator) Struct(s any) error {
	return v.v.Struct(s)
}
