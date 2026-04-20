package domain

import "errors"

// DutyRequest contains all input data required to calculate customs duties.
type DutyRequest struct {
	FromCountry    string  `json:"fromCountry"`    // ISO 3166-1 alpha-2
	ToCountry      string  `json:"toCountry"`      // ISO 3166-1 alpha-2
	HSCode         string  `json:"hsCode"`         // HS tariff classification code
	DeclaredValue  float64 `json:"declaredValue"`  // value in the supplied currency
	Currency       string  `json:"currency"`       // ISO 4217 currency code
	Quantity       int     `json:"quantity"`       // number of units
	Description    string  `json:"description"`    // free-text item description
}

// DutyLineItem represents a single component of the total duty/tax breakdown.
type DutyLineItem struct {
	Description string  `json:"description"`
	DutyRate    float64 `json:"dutyRate"`    // as a decimal fraction, e.g. 0.05 = 5%
	DutyAmount  float64 `json:"dutyAmount"`
	TaxRate     float64 `json:"taxRate"`
	TaxAmount   float64 `json:"taxAmount"`
}

// DutyResponse is the complete result of a customs duty calculation.
type DutyResponse struct {
	FromCountry         string         `json:"fromCountry"`
	ToCountry           string         `json:"toCountry"`
	HSCode              string         `json:"hsCode"`
	DeclaredValue       float64        `json:"declaredValue"`
	Currency            string         `json:"currency"`
	DutyAmount          float64        `json:"dutyAmount"`
	VATAmount           float64        `json:"vatAmount"`
	TotalLandedCost     float64        `json:"totalLandedCost"`
	Breakdown           []DutyLineItem `json:"breakdown"`
	RequiresCustomsForm bool           `json:"requiresCustomsForm"`
	ProhibitedItems     bool           `json:"prohibitedItems"`
	DeMinimisMet        bool           `json:"deMinimisMet"`   // true when below de minimis → no duty
	Notes               string         `json:"notes,omitempty"`
}

// HSCodeInfo describes a single entry in the HS tariff table.
type HSCodeInfo struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	GeneralRate float64 `json:"generalRate"` // base duty rate as decimal fraction
}

// CountryRates holds the import tax/VAT rate configuration for a single country.
type CountryRates struct {
	Country          string  `json:"country"`
	VATRate          float64 `json:"vatRate"`          // decimal fraction
	DeMinimiisUSD    float64 `json:"deMinimisUsd"`     // threshold below which no duty is collected
	Notes            string  `json:"notes,omitempty"`
}

// Sentinel errors for the customs domain.
var (
	ErrHSCodeNotFound  = errors.New("HS code not found")
	ErrCountryNotFound = errors.New("country rates not found")
	ErrInvalidRequest  = errors.New("invalid request")
)
