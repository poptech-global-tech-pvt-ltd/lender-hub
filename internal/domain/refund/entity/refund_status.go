package entity

type RefundStatus string

const (
	RefundPending RefundStatus = "PENDING"
	RefundSuccess RefundStatus = "SUCCESS"
	RefundFailed  RefundStatus = "FAILED"
)
