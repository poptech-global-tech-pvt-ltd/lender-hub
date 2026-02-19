package stub

import (
	"context"
	"fmt"

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
func (g *StubOrderGateway) CreateOrder(ctx context.Context, input port.OrderInput) (*res.OrderResponse, error) {
	// Generate a fake lender order ID
	lenderOrderID := fmt.Sprintf("LP-ORDER-%s", input.MerchantTxnID)

	// Return fake redirect URL
	redirectURL := fmt.Sprintf("https://stub.lazypay.in/payment/%s", lenderOrderID)

	return &res.OrderResponse{
		PaymentID:     "", // Set by service layer
		Status:        "PENDING",
		LenderOrderID: &lenderOrderID,
		RedirectURL:   &redirectURL,
	}, nil
}

// GetOrderStatus implements OrderGateway.GetOrderStatus
func (g *StubOrderGateway) GetOrderStatus(ctx context.Context, merchantTxnID string) (*res.OrderStatusResponse, error) {
	return &res.OrderStatusResponse{
		PaymentID:     "", // Set by service layer
		UserID:        "stub-user",
		MerchantID:    "",
		Status:        "SUCCESS",
		LenderOrderID: &merchantTxnID,
		Amount:        0,
		Currency:      "INR",
		CreatedAt:     "",
		UpdatedAt:     "",
	}, nil
}
