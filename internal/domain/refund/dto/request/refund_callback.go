package request

import "time"

// RefundCallbackRequest represents a refund callback from Kafka (Ingestion Service)
type RefundCallbackRequest struct {
	RefundID         string    `json:"refundId"`
	PaymentRefundID  string    `json:"paymentRefundId"`
	LoanID           string    `json:"loanId"`
	PaymentID        string    `json:"paymentId"`
	LenderStatus     string    `json:"lenderStatus"`
	LenderTxnID      string    `json:"lenderTxnId"`
	LenderTxnStatus  string    `json:"lenderTxnStatus"`
	LenderTxnMessage string    `json:"lenderTxnMessage"`
	EventTime        time.Time `json:"eventTime"`
}
