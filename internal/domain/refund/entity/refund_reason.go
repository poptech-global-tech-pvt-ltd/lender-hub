package entity

type RefundReason string

const (
	ReasonUserCancelled  RefundReason = "USER_CANCELLED"
	ReasonProductReturn  RefundReason = "PRODUCT_RETURN"
	ReasonOrderCancelled RefundReason = "ORDER_CANCELLED"
)
