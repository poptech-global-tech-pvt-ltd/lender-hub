package entity

// RefundStatus represents the state of a refund
type RefundStatus string

const (
	RefundStatusPending    RefundStatus = "PENDING"
	RefundStatusProcessing RefundStatus = "PROCESSING"
	RefundStatusSuccess    RefundStatus = "SUCCESS"
	RefundStatusFailed     RefundStatus = "FAILED"
	RefundStatusUnknown    RefundStatus = "UNKNOWN"
)

// IsTerminal returns true if the status is a terminal state (SUCCESS or FAILED)
func (s RefundStatus) IsTerminal() bool {
	return s == RefundStatusSuccess || s == RefundStatusFailed
}

// IsResolvable returns true if the status can be resolved via enquiry (PENDING, UNKNOWN, PROCESSING)
func (s RefundStatus) IsResolvable() bool {
	return s == RefundStatusUnknown || s == RefundStatusProcessing || s == RefundStatusPending
}

// OrDefault returns PENDING if empty
func (s RefundStatus) OrDefault() RefundStatus {
	if s == "" {
		return RefundStatusPending
	}
	return s
}
