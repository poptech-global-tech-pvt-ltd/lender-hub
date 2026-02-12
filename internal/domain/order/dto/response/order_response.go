package response

type OrderResponse struct {
	PaymentID         string  `json:"paymentId"`
	Status            string  `json:"status"`
	LenderOrderID     *string `json:"lenderOrderId,omitempty"`
	RedirectURL       *string `json:"redirectUrl,omitempty"`
	ErrorCode         *string `json:"errorCode,omitempty"`
	ErrorMessage      *string `json:"errorMessage,omitempty"`
}
