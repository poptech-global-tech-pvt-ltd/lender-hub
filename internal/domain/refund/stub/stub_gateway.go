package stub

import (
	"context"

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
func (g *StubRefundGateway) ProcessRefund(ctx context.Context, merchantTxnID string, amount float64, currency string) (*res.RefundResponse, string, error) {
	// Generate a fake lender refund ID and refundTxnId
	refundTxnID := "REF-STUB-" + merchantTxnID
	lenderRefID := "LP-REFUND-" + refundTxnID

	return &res.RefundResponse{
		RefundID:    "", // Set by service layer
		PaymentID:   "", // Set by service layer
		Provider:    "LAZYPAY",
		Status:      "PENDING",
		Amount:      amount,
		Currency:    currency,
		LenderRefID: &lenderRefID,
	}, refundTxnID, nil
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
