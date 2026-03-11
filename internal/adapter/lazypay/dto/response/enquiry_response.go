package response

// LPEnquiryResponse matches Lazypay /v3/enquiry response structure
type LPEnquiryResponse struct {
	Order        LPEnquiryOrder         `json:"order"`
	Transactions []LPEnquiryTransaction `json:"transactions"`
}

// LPEnquiryOrder represents order information from enquiry
type LPEnquiryOrder struct {
	OrderID         string           `json:"orderId"`
	Status          string           `json:"status"`
	Message         string           `json:"message"`
	SelectedEmiPlan *SelectedEmiPlan `json:"selectedEmiPlan,omitempty"`
}

// LPEnquiryTransaction represents a transaction from enquiry response
type LPEnquiryTransaction struct {
	Status      string `json:"status"`
	RespMessage string `json:"respMessage"`
	LpTxnID     string `json:"lpTxnId"`
	TxnType     string `json:"txnType"`
	TxnRefNo    string `json:"txnRefNo"`
	TxnDateTime string `json:"txnDateTime"`
	Amount      string `json:"amount"`
}

// SelectedEmiPlan represents EMI plan details (optional field)
type SelectedEmiPlan struct {
	InterestRate float64 `json:"interestRate"`
	Tenure       int     `json:"tenure"`
	Emi          float64 `json:"emi"`
	// ... other fields as needed
}
