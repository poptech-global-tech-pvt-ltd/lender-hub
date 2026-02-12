package entity

type OrderStatus string

const (
	OrderPending   OrderStatus = "PENDING"
	OrderSuccess   OrderStatus = "SUCCESS"
	OrderFailed    OrderStatus = "FAILED"
	OrderRefunded  OrderStatus = "REFUNDED"
	OrderExpired   OrderStatus = "EXPIRED"
	OrderCancelled OrderStatus = "CANCELLED"
)

func (s OrderStatus) IsTerminal() bool {
	return s == OrderSuccess || s == OrderFailed || s == OrderRefunded ||
		s == OrderExpired || s == OrderCancelled
}
