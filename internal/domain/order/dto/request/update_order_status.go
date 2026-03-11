package request

// UpdateOrderStatusRequest — used by PATCH /v1/payin3/order/:paymentId/status
// Only support/CX tool calls this endpoint
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=FAILED CANCELLED"`
	Reason string `json:"reason" binding:"required"` // mandatory audit reason
	Actor  string `json:"actor"  binding:"required"` // e.g. "support@popclub.in"
}
