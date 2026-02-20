package entity

type OrderStatus string

const (
	OrderPending   OrderStatus = "PENDING"
	OrderSuccess   OrderStatus = "SUCCESS"
	OrderComplete  OrderStatus = "COMPLETE" // Lazypay returns COMPLETE for successful orders
	OrderFailed    OrderStatus = "FAILED"
	OrderRefunded  OrderStatus = "REFUNDED"
	OrderExpired   OrderStatus = "EXPIRED"
	OrderCancelled OrderStatus = "CANCELLED"
)

func (s OrderStatus) IsTerminal() bool {
	return s == OrderSuccess || s == OrderComplete || s == OrderFailed || s == OrderRefunded ||
		s == OrderExpired || s == OrderCancelled
}

// IsRefundable returns true if order can be refunded (SUCCESS or COMPLETE at Lazypay)
func (s OrderStatus) IsRefundable() bool {
	return s == OrderSuccess || s == OrderComplete
}

// NormalizeForDB maps Lazypay status to DB enum (lender_payment_status has no COMPLETE)
func (s OrderStatus) NormalizeForDB() OrderStatus {
	if s == OrderComplete {
		return OrderSuccess
	}
	return s
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
