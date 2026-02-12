package port

import (
	"context"

	req "lending-hub-service/internal/domain/refund/dto/request"
	res "lending-hub-service/internal/domain/refund/dto/response"
)

// RefundGateway abstracts external refund provider calls
type RefundGateway interface {
	ProcessRefund(ctx context.Context, req req.CreateRefundRequest) (*res.RefundResponse, error)
}
