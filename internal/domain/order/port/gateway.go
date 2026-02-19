package port

import (
	"context"

	res "lending-hub-service/internal/domain/order/dto/response"
)

// OrderInput contains all data needed to create an order via gateway
type OrderInput struct {
	MerchantTxnID string
	Mobile        string
	Email         string
	Amount        float64
	Currency      string
	EmiPlans      []LPEmiPlan // from eligibility cache
}

// LPEmiPlan represents an EMI plan for Lazypay
type LPEmiPlan struct {
	InterestRate             float64 `json:"interestRate"`
	Tenure                   int     `json:"tenure"`
	Emi                      float64 `json:"emi"`
	TotalInterestAmount      float64 `json:"totalInterestAmount"`
	Principal                float64 `json:"principal"`
	TotalProcessingFee       float64 `json:"totalProcessingFee"`
	ProcessingFeeGst         float64 `json:"processingFeeGst"`
	TotalPayableAmount       float64 `json:"totalPayableAmount"`
	FirstEmiDueDate          string  `json:"firstEmiDueDate"`
	SubventionTag            *string `json:"subventionTag"`
	DiscountedInterestAmount float64 `json:"discountedInterestAmount"`
	Schedule                 *string `json:"schedule"`
	Type                     string  `json:"type"`
}

// OrderGateway abstracts external payment provider calls
type OrderGateway interface {
	CreateOrder(ctx context.Context, input OrderInput) (*res.OrderResponse, error)
	GetOrderStatus(ctx context.Context, merchantTxnID string) (*res.OrderStatusResponse, error)
}
