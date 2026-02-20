package stub

import (
	"context"

	"lending-hub-service/internal/domain/refund/port"
)

// StubRefundGateway implements port.RefundGateway for local development
type StubRefundGateway struct{}

// NewStubRefundGateway creates a new stub gateway
func NewStubRefundGateway() port.RefundGateway {
	return &StubRefundGateway{}
}

// ProcessRefund returns a fake REFUND_SUCCESS response
func (g *StubRefundGateway) ProcessRefund(ctx context.Context, req port.ProcessRefundRequest) (*port.ProcessRefundResponse, error) {
	return &port.ProcessRefundResponse{
		Status:      "REFUND_SUCCESS",
		LpTxnID:     "LP-TXN-STUB-" + req.RefundTxnID,
		ParentTxnID: "LP-PARENT-STUB",
		RespMessage: "Stub refund success",
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
