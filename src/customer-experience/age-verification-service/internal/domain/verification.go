package domain

import "time"

// VerificationRequest carries the data needed to verify a customer's age.
type VerificationRequest struct {
	CustomerID      string `json:"customerId"`
	DateOfBirth     string `json:"dateOfBirth"`     // format: YYYY-MM-DD
	Country         string `json:"country"`         // ISO 3166-1 alpha-2 (e.g. "US", "UK")
	ProductCategory string `json:"productCategory"` // alcohol, tobacco, gambling, adult_content, firearms
}

// VerificationResult is the outcome of an age verification check.
type VerificationResult struct {
	CustomerID      string    `json:"customerId"`
	Verified        bool      `json:"verified"`
	Age             int       `json:"age"`
	MinAge          int       `json:"minAge"`
	Country         string    `json:"country"`
	ProductCategory string    `json:"productCategory"`
	Reason          string    `json:"reason"`
	VerifiedAt      time.Time `json:"verifiedAt"`
}

// CountryMinAges holds the minimum age for alcohol by country code.
// For non-alcohol categories, the category minimum applies instead.
// The effective minimum is max(countryMin, categoryMin).
var CountryMinAges = map[string]int{
	"US": 21, // alcohol legal age in the US
	"UK": 18,
	"DE": 18,
	"JP": 20,
	"KR": 19,
	"AU": 18,
	"CA": 19,
	"FR": 18,
	"IT": 18,
	"ES": 18,
	"NL": 18,
	"BR": 18,
	"MX": 18,
	"IN": 18,
}

// DefaultMinAge is used when a country is not in the lookup table.
const DefaultMinAge = 18

// CategoryMinAges defines the baseline minimum age for each restricted category.
var CategoryMinAges = map[string]int{
	"alcohol":       18,
	"tobacco":       18,
	"gambling":      21,
	"adult_content": 18,
	"firearms":      21,
}
