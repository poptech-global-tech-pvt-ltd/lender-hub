package port

import "context"

// ProcessRefundRequest holds data for a refund API call
type ProcessRefundRequest struct {
	MerchantTxnID string // order's loanId = merchantTxnId at Lazypay
	Amount        float64
	Currency      string
	RefundTxnID   string // our refundId = refundTxnId sent to Lazypay
}

// ProcessRefundResponse holds the Lazypay refund API response
type ProcessRefundResponse struct {
	Status      string // REFUND_SUCCESS or error code
	LpTxnID     string // Lazypay's REFUND transaction lpTxnId
	ParentTxnID string // Lazypay's SALE transaction lpTxnId
	RespMessage string
	ErrorCode   string // LPDUPLICATEREFUND etc.
	IsTimeout   bool
}

// RefundGateway abstracts external refund provider calls
type RefundGateway interface {
	// ProcessRefund single attempt — no retry
	ProcessRefund(ctx context.Context, req ProcessRefundRequest) (*ProcessRefundResponse, error)
	// EnquireRefund queries the provider using order's loanId as merchantTxnId
	EnquireRefund(ctx context.Context, merchantTxnID string) (*EnquiryResponse, error)
}

// EnquiryResponse represents the response from enquiry API
type EnquiryResponse struct {
	Order        EnquiryOrder
	Transactions []EnquiryTransaction
}

// EnquiryOrder represents order information from enquiry
type EnquiryOrder struct {
	OrderID string
	Status  string
	Message string
}

// EnquiryTransaction represents a transaction from enquiry response
type EnquiryTransaction struct {
	Status      string
	RespMessage string
	LpTxnID     string
	TxnType     string
	TxnRefNo    string
	TxnDateTime string
	Amount      string
}
