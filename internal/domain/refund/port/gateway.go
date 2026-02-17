package port

import (
	"context"

	req "lending-hub-service/internal/domain/refund/dto/request"
	res "lending-hub-service/internal/domain/refund/dto/response"
)

// RefundGateway abstracts external refund provider calls
type RefundGateway interface {
	ProcessRefund(ctx context.Context, req req.CreateRefundRequest) (*res.RefundResponse, error)
	
	// EnquireRefund queries the provider for refund status using merchantTxnID
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
