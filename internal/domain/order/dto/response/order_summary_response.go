package response

import "time"

// OrderSummaryResponse is returned by GET /v1/payin3/order/:paymentId/summary
type OrderSummaryResponse struct {
	Order   OrderStatusResponse     `json:"order"`
	Refunds []RefundSummaryItem     `json:"refunds"`
	Summary OrderFinancialSummary   `json:"summary"`
}

// RefundSummaryItem represents a refund in the summary
type RefundSummaryItem struct {
	RefundID            string    `json:"refundId"`
	PaymentRefundID     string    `json:"paymentRefundId"`
	ProviderRefundTxnID string    `json:"providerRefundTxnId,omitempty"`
	Status              string    `json:"status"`
	Amount              float64   `json:"amount"`
	Currency            string    `json:"currency"`
	Reason              string    `json:"reason,omitempty"`
	LenderStatus        string    `json:"lenderStatus,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// OrderFinancialSummary holds computed financial totals
type OrderFinancialSummary struct {
	TotalAmount   float64 `json:"totalAmount"`
	TotalRefunded float64 `json:"totalRefunded"`
	NetAmount     float64 `json:"netAmount"`
	RefundCount   int     `json:"refundCount"`
	FullyRefunded bool    `json:"fullyRefunded"`
}
