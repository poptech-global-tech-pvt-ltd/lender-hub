package entity

type EmiSelection struct {
	Tenure             int     `json:"tenure"`
	EMI                float64 `json:"emi"`
	InterestRate       float64 `json:"interestRate"`
	Principal          float64 `json:"principal"`
	TotalPayableAmount float64 `json:"totalPayableAmount"`
}
