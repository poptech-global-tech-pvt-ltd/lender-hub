package port

import (
	"context"

	res "lending-hub-service/internal/domain/order/dto/response"
)

// OrderInput contains all data needed to create an order via gateway (Lazypay: merchantTxnId, amount, userDetails, source, returnUrl)
type OrderInput struct {
	MerchantTxnID string
	Mobile        string
	Email         string
	Amount        float64
	Currency      string
	Source        string
	ReturnURL     string
}

// OrderGateway abstracts external payment provider calls
type OrderGateway interface {
	CreateOrder(ctx context.Context, input OrderInput) (*res.OrderResponse, error)
	GetOrderStatus(ctx context.Context, merchantTxnID string) (*res.OrderStatusResponse, error)
}
