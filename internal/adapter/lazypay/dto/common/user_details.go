package common

// LPUserDetails matches Lazypay's user object shape
type LPUserDetails struct {
	Mobile    string `json:"mobile"`
	Email     string `json:"email"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}
