package common

import "fmt"

// LPAmount matches Lazypay's amount object shape
type LPAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// LPUserDetails matches Lazypay's user object shape
// Email is nullable: nil → JSON null, &email → JSON "email@..."
type LPUserDetails struct {
	Mobile string  `json:"mobile"`
	Email  *string `json:"email"`
}

// LPCustomParams contains subMerchantId for onboarding
type LPCustomParams struct {
	SubMerchantID string `json:"subMerchantId"`
}

// NewLPUserDetails creates LPUserDetails with nullable email
func NewLPUserDetails(mobile, email string) LPUserDetails {
	ud := LPUserDetails{Mobile: mobile}
	if email != "" {
		ud.Email = &email
	}
	return ud
}

// FormatAmount converts float64 to string with 2 decimal places
func FormatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}
