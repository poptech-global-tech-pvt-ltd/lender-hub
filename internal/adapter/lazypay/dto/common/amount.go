package common

// LPAmount matches Lazypay's amount object shape
type LPAmount struct {
	Value    string `json:"value"`    // String representation "1000.00"
	Currency string `json:"currency"` // "INR"
}
