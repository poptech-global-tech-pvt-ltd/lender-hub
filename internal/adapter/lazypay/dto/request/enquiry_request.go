package request

// LPEnquiryRequest represents the enquiry API request
// Enquiry API uses query param merchantTxnId, no JSON body needed
// This is a placeholder struct if needed for signature generation

type LPEnquiryRequest struct {
	MerchantTxnID string `json:"merchantTxnId"`
}
