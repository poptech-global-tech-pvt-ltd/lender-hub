package response

import "time"

// OrderResponse (CreateOrderResponse) is returned by POST /v1/payin3/order
type OrderResponse struct {
	LoanID        string    `json:"loanId"`        // our ID (lps_xxx) = merchantTxnId to Lazypay
	PaymentID     string    `json:"paymentId"`     // POP's ID — primary for polling
	LenderOrderID string    `json:"lenderOrderId"` // Lazypay orderId
	Status        string    `json:"status"`        // always "PENDING"
	RedirectURL   string    `json:"redirectUrl"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"createdAt"`
	ErrorCode     *string   `json:"errorCode,omitempty"`
	ErrorMessage  *string   `json:"errorMessage,omitempty"`
}
