package common

// LPAddress matches Lazypay's address object shape
type LPAddress struct {
	Street1       string `json:"street1"`
	Street2       string `json:"street2,omitempty"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
	Zip           string `json:"zip"`
	Landmark      string `json:"landmark,omitempty"`
	ResidenceType string `json:"residenceType,omitempty"`
}
