package domain

// TaxRequest is the input payload for a tax calculation.
type TaxRequest struct {
	Items        []TaxLineItem `json:"items"`
	ShipTo       Address       `json:"ship_to"`
	Currency     string        `json:"currency"`
	CustomerType string        `json:"customer_type"` // "b2c" | "b2b"
}

// TaxLineItem represents one taxable line in an order.
type TaxLineItem struct {
	ProductID string  `json:"product_id"`
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
	Quantity  int     `json:"quantity"`
}

// Address is the destination address used to look up the applicable tax jurisdiction.
type Address struct {
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// TaxResponse is the output of a tax calculation.
type TaxResponse struct {
	Subtotal  float64        `json:"subtotal"`
	TaxAmount float64        `json:"tax_amount"`
	TaxRate   float64        `json:"tax_rate"`
	Total     float64        `json:"total"`
	Currency  string         `json:"currency"`
	Breakdown []TaxBreakdown `json:"breakdown"`
}

// TaxBreakdown describes the tax levied by a single jurisdiction.
type TaxBreakdown struct {
	Jurisdiction string  `json:"jurisdiction"`
	Rate         float64 `json:"rate"`
	Amount       float64 `json:"amount"`
}

// RateInfo is the response shape for GET /tax/rates.
type RateInfo struct {
	Country       string         `json:"country"`
	State         string         `json:"state,omitempty"`
	Jurisdictions []TaxBreakdown `json:"jurisdictions"`
	EffectiveRate float64        `json:"effective_rate"`
}

// ErrorResponse is the JSON shape returned on error.
type ErrorResponse struct {
	Error string `json:"error"`
}
