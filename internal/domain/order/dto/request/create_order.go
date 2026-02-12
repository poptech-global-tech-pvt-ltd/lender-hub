package request

type Address struct {
	Street1 string `json:"street1"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type ProductLine struct {
	Name     string  `json:"name"`
	SKU      string  `json:"sku"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type EmiSelection struct {
	Tenure             int     `json:"tenure" binding:"required"`
	EMI                float64 `json:"emi" binding:"required"`
	InterestRate       float64 `json:"interestRate"`
	Principal          float64 `json:"principal"`
	TotalPayableAmount float64 `json:"totalPayableAmount" binding:"required"`
}

type CreateOrderRequest struct {
	PaymentID    string        `json:"paymentId" binding:"required"`
	UserID       string        `json:"userId" binding:"required"`
	Mobile       string        `json:"mobile" binding:"required"`
	Amount       float64       `json:"amount" binding:"required"`
	Currency     string        `json:"currency" binding:"required"`
	MerchantID   string        `json:"merchantId" binding:"required"`
	ReturnURL    string        `json:"returnUrl" binding:"required"`
	Address      *Address      `json:"address"`
	ProductLines []ProductLine `json:"productLines"`
	EmiSelection EmiSelection  `json:"emiSelection" binding:"required"`
}
