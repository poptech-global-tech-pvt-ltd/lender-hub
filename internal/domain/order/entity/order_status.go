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

func (s OrderStatus) IsEmpty() bool {
	return s == ""
}

func (s OrderStatus) OrDefault() OrderStatus {
	if s.IsEmpty() {
		return OrderPending
	}
	return s
}
