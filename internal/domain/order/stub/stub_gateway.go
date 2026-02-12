package stub

import (
	"context"
	"fmt"

	req "lending-hub-service/internal/domain/order/dto/request"
	res "lending-hub-service/internal/domain/order/dto/response"
	"lending-hub-service/internal/domain/order/port"
)

// StubOrderGateway implements port.OrderGateway for local development
type StubOrderGateway struct{}

// NewStubOrderGateway creates a new stub gateway
func NewStubOrderGateway() port.OrderGateway {
	return &StubOrderGateway{}
}

// CreateOrder returns a fake order response with PENDING status
func (g *StubOrderGateway) CreateOrder(ctx context.Context, req req.CreateOrderRequest) (*res.OrderResponse, error) {
	// Generate a fake lender order ID
	lenderOrderID := fmt.Sprintf("LP-ORDER-%s", req.PaymentID)

	// Return fake redirect URL
	redirectURL := fmt.Sprintf("https://stub.lazypay.in/payment/%s?returnUrl=%s", lenderOrderID, req.ReturnURL)

	return &res.OrderResponse{
		PaymentID:     req.PaymentID,
		Status:        "PENDING",
		LenderOrderID: &lenderOrderID,
		RedirectURL:   &redirectURL,
	}, nil
}
