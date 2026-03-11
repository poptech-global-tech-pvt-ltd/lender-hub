package events

import "time"

// OrderCallbackEvent — consumed from lsp.lazypay.order.callback (published by Ingestion Service)
type OrderCallbackEvent struct {
	BaseEvent
	LoanID           string    `json:"loanId"`
	PaymentID        string    `json:"paymentId"`
	LenderOrderID    string    `json:"lenderOrderId"`
	LenderStatus     string    `json:"lenderStatus"`
	LenderTxnID      string    `json:"lenderTxnId"`
	LenderTxnStatus  string    `json:"lenderTxnStatus"`
	LenderTxnMessage string    `json:"lenderTxnMessage"`
	LenderTxnTime    time.Time `json:"lenderTxnTime"`
	RawPayload       string    `json:"rawPayload,omitempty"`
	EventTime        time.Time `json:"eventTime"`
}

// RefundCallbackEvent — consumed from lsp.lazypay.refund.callback
type RefundCallbackEvent struct {
	BaseEvent
	RefundID         string    `json:"refundId"`
	PaymentRefundID  string    `json:"paymentRefundId"`
	LoanID           string    `json:"loanId"`
	PaymentID        string    `json:"paymentId"`
	LenderStatus     string    `json:"lenderStatus"`
	LenderTxnID      string    `json:"lenderTxnId"`
	LenderTxnStatus  string    `json:"lenderTxnStatus"`
	LenderTxnMessage string    `json:"lenderTxnMessage"`
	RawPayload       string    `json:"rawPayload,omitempty"`
	EventTime        time.Time `json:"eventTime"`
}
