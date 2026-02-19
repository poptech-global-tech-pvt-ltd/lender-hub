package request

import "lending-hub-service/internal/adapter/lazypay/dto/common"

// LPCreateOrderRequest matches Lazypay POST /api/lazypay/cof/v0/payment/order
type LPCreateOrderRequest struct {
	MerchantTxnID string               `json:"merchantTxnId"`
	Amount        common.LPAmount      `json:"amount"`
	UserDetails   common.LPUserDetails `json:"userDetails"`
	Source        string               `json:"source"`
	ReturnURL     string               `json:"returnUrl"`
	EmiPlans      []LPEmiPlan          `json:"emiPlans"`
}

// LPEmiPlan represents an EMI plan option
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
	SubventionTag            *string `json:"subventionTag,omitempty"`
	DiscountedInterestAmount float64 `json:"discountedInterestAmount"`
	Schedule                 *string `json:"schedule,omitempty"`
	Type                     string  `json:"type"`
}
