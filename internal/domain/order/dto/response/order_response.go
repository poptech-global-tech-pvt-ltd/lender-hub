package response

import "time"

// OrderResponse (CreateOrderResponse) is returned by POST /v1/payin3/order
type OrderResponse struct {
	PaymentID     string    `json:"paymentId"`     // server-generated lps_01JN...
	Status        string    `json:"status"`        // always "PENDING"
	LenderOrderID string    `json:"lenderOrderId"` // Lazypay orderId
	RedirectURL   string    `json:"redirectUrl"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"createdAt"`
	// Legacy fields for errors (optional)
	ErrorCode    *string `json:"errorCode,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}
