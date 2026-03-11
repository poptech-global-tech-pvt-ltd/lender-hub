package entity

// CreditLine is a value object representing a user's credit line
type CreditLine struct {
	Limit          float64
	AvailableLimit float64
	Currency       string // default "INR"
}
