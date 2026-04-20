package domain

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("record not found")

// TaxType enumerates supported tax classifications.
type TaxType string

const (
	TaxTypeVAT       TaxType = "VAT"
	TaxTypeGST       TaxType = "GST"
	TaxTypeSalesTax  TaxType = "SALES_TAX"
	TaxTypeExcise    TaxType = "EXCISE"
)

// TaxRecord represents a single tax transaction captured from an order.
type TaxRecord struct {
	ID              string    `json:"id"`
	OrderID         string    `json:"orderId"`
	CustomerID      string    `json:"customerId"`
	Jurisdiction    string    `json:"jurisdiction"`
	TaxType         TaxType   `json:"taxType"`
	TaxableAmount   float64   `json:"taxableAmount"`
	TaxRate         float64   `json:"taxRate"`
	TaxAmount       float64   `json:"taxAmount"`
	Currency        string    `json:"currency"`
	TransactionDate time.Time `json:"transactionDate"`
	CreatedAt       time.Time `json:"createdAt"`
}

// TaxSummary represents aggregated tax data for a jurisdiction/type/period combination.
type TaxSummary struct {
	Jurisdiction     string  `json:"jurisdiction"`
	TaxType          TaxType `json:"taxType"`
	Period           string  `json:"period"` // YYYY-MM
	TotalTaxable     float64 `json:"totalTaxable"`
	TotalTax         float64 `json:"totalTax"`
	TransactionCount int     `json:"transactionCount"`
}

// ListFilter holds optional filter parameters for listing tax records.
type ListFilter struct {
	Jurisdiction string
	TaxType      string
	StartDate    string // RFC3339 or date string
	EndDate      string // RFC3339 or date string
	Limit        int
}

// GenerateReportRequest is the input for generating a tax report over a date range.
type GenerateReportRequest struct {
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	Jurisdiction string `json:"jurisdiction"`
}
