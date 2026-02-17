package validator

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTTPValidator struct {
	v *Validator
}

func NewHTTPValidator(v *Validator) *HTTPValidator {
	return &HTTPValidator{v: v}
}

// BindAndValidate binds JSON into dst, validates it, and writes
// a standardized error response if invalid.
// Returns true if OK to proceed, false if response already sent.
func (h *HTTPValidator) BindAndValidate(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		resp := &ValidationErrorResponse{
			Code:    "PAYIN3_INVALID_JSON",
			Message: "Malformed JSON body",
			Errors:  nil,
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    nil,
			"error":   resp,
		})
		return false
	}

	if err := h.v.Struct(dst); err != nil {
		resp := MapError(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    nil,
			"error":   resp,
		})
		return false
	}

	return true
}
