package lender

// Lender represents a supported BNPL/loan provider
type Lender string

const (
	Lazypay Lender = "LAZYPAY"
)

func (l Lender) String() string { return string(l) }
