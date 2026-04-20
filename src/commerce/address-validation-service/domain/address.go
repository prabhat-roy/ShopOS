package domain

// Address represents a postal address.
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country"` // ISO 3166-1 alpha-2
}

// ValidationResult is returned by the validator.
type ValidationResult struct {
	Valid      bool     `json:"valid"`
	Normalized *Address `json:"normalized,omitempty"`
	Issues     []string `json:"issues"`
	Confidence float64  `json:"confidence"` // 0.0 – 1.0
}
