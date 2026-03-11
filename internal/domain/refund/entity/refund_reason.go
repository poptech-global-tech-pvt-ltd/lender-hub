package entity

// RefundReason represents the reason for a refund
type RefundReason string

const (
	RefundReasonUserCancelled  RefundReason = "USER_CANCELLED"
	RefundReasonProductReturn  RefundReason = "PRODUCT_RETURN"
	RefundReasonOrderCancelled RefundReason = "ORDER_CANCELLED"
)
