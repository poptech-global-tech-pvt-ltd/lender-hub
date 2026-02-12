package kafka

import "time"

// DomainEvent is the standard event envelope published to Kafka
type DomainEvent struct {
	EventID    string      `json:"eventId"`    // UUID
	EventType  string      `json:"eventType"`  // e.g. "ProfileActivated", "OrderCompleted"
	Source     string      `json:"source"`     // "payin3-service"
	OccurredAt time.Time   `json:"occurredAt"`
	UserID     string      `json:"userId"`
	Lender     string      `json:"lender"`
	Data       interface{} `json:"data"` // Event-specific payload
}

// ProfileEventData represents profile event payload
type ProfileEventData struct {
	PreviousStatus string  `json:"previousStatus"`
	NewStatus      string  `json:"newStatus"`
	CreditLimit    float64 `json:"creditLimit"`
	AvailableLimit float64 `json:"availableLimit"`
	IsBlocked      bool    `json:"isBlocked"`
	BlockReason    string  `json:"blockReason,omitempty"`
	BlockSource    string  `json:"blockSource,omitempty"`
}

// OrderEventData represents order event payload
type OrderEventData struct {
	PaymentID     string  `json:"paymentId"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	LenderOrderID string  `json:"lenderOrderId,omitempty"`
	ErrorCode     string  `json:"errorCode,omitempty"`
}

// RefundEventData represents refund event payload
type RefundEventData struct {
	RefundID    string  `json:"refundId"`
	PaymentID   string  `json:"paymentId"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	LenderRefID string  `json:"lenderRefId,omitempty"`
}
