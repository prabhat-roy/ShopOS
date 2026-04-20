package domain

import "time"

// TaxProvider represents a supported external or internal tax calculation engine.
type TaxProvider string

const (
	ProviderAvalara  TaxProvider = "AVALARA"
	ProviderTaxJar   TaxProvider = "TAXJAR"
	ProviderVertex   TaxProvider = "VERTEX"
	ProviderInternal TaxProvider = "INTERNAL"
)

// AllProviders lists every supported tax provider.
var AllProviders = []TaxProvider{
	ProviderAvalara,
	ProviderTaxJar,
	ProviderVertex,
	ProviderInternal,
}

// TaxAddress is the address model used for tax jurisdiction lookups.
type TaxAddress struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

// TaxLineItem represents a single purchasable item in a tax calculation request.
type TaxLineItem struct {
	ProductID   string  `json:"productId"`
	SKU         string  `json:"sku"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Amount      float64 `json:"amount"`
	TaxCode     string  `json:"taxCode"`
}

// TaxCalculationRequest is the unified input for calculating tax with any provider.
type TaxCalculationRequest struct {
	Provider      TaxProvider   `json:"provider"`
	TransactionID string        `json:"transactionId"`
	FromAddress   TaxAddress    `json:"fromAddress"`
	ToAddress     TaxAddress    `json:"toAddress"`
	LineItems     []TaxLineItem `json:"lineItems"`
	Currency      string        `json:"currency"`
}

// TaxBreakdownItem represents a single tax jurisdiction's contribution.
type TaxBreakdownItem struct {
	Jurisdiction string  `json:"jurisdiction"`
	TaxType      string  `json:"taxType"`
	Rate         float64 `json:"rate"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
}

// TaxCalculationResponse is the unified output after a successful tax calculation.
type TaxCalculationResponse struct {
	Provider      TaxProvider        `json:"provider"`
	TransactionID string             `json:"transactionId"`
	Subtotal      float64            `json:"subtotal"`
	TotalTax      float64            `json:"totalTax"`
	Total         float64            `json:"total"`
	Currency      string             `json:"currency"`
	Breakdown     []TaxBreakdownItem `json:"breakdown"`
	CalculatedAt  time.Time          `json:"calculatedAt"`
}

// CommitRequest asks a provider to commit (record) a previously calculated transaction.
type CommitRequest struct {
	TransactionID string      `json:"transactionId"`
	Provider      TaxProvider `json:"provider"`
}

// CommitResponse confirms that the transaction was committed.
type CommitResponse struct {
	Committed   bool      `json:"committed"`
	CommittedAt time.Time `json:"committedAt"`
}
