package port

import (
	"context"

	req "lending-hub-service/internal/domain/order/dto/request"
	res "lending-hub-service/internal/domain/order/dto/response"
)

// OrderGateway abstracts external payment provider calls
type OrderGateway interface {
	CreateOrder(ctx context.Context, req req.CreateOrderRequest) (*res.OrderResponse, error)
}
