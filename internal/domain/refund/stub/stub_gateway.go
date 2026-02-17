package stub

import (
	"context"

	req "lending-hub-service/internal/domain/refund/dto/request"
	res "lending-hub-service/internal/domain/refund/dto/response"
	"lending-hub-service/internal/domain/refund/port"
)

// StubRefundGateway implements port.RefundGateway for local development
type StubRefundGateway struct{}

// NewStubRefundGateway creates a new stub gateway
func NewStubRefundGateway() port.RefundGateway {
	return &StubRefundGateway{}
}

// ProcessRefund returns a fake refund response with PENDING status
func (g *StubRefundGateway) ProcessRefund(ctx context.Context, req req.CreateRefundRequest) (*res.RefundResponse, error) {
	// Generate a fake lender refund ID
	lenderRefID := "LP-REFUND-" + req.RefundID

	return &res.RefundResponse{
		RefundID:    req.RefundID,
		PaymentID:   req.PaymentID,
		Provider:    "LAZYPAY",
		Status:      "PENDING",
		Amount:      req.Amount,
		Currency:    req.Currency,
		LenderRefID: &lenderRefID,
	}, nil
}

// EnquireRefund returns a stub enquiry response
func (g *StubRefundGateway) EnquireRefund(ctx context.Context, merchantTxnID string) (*port.EnquiryResponse, error) {
	return &port.EnquiryResponse{
		Order: port.EnquiryOrder{
			OrderID: merchantTxnID,
			Status:  "SUCCESS",
			Message: "Stub enquiry response",
		},
		Transactions: []port.EnquiryTransaction{
			{
				Status:      "SUCCESS",
				RespMessage: "Stub refund transaction",
				LpTxnID:     "LP-TXN-STUB",
				TxnType:     "REFUND",
				TxnRefNo:    merchantTxnID,
				TxnDateTime: "2024-01-01T00:00:00Z",
				Amount:      "100.00",
			},
		},
	}, nil
}
