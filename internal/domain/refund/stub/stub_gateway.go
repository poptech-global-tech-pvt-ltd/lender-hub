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
