package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/tax-reporting-service/internal/domain"
	"github.com/shopos/tax-reporting-service/internal/store"
)

// Servicer defines the business-logic contract for tax reporting.
type Servicer interface {
	RecordTax(r *domain.TaxRecord) (*domain.TaxRecord, error)
	GetTaxRecord(id string) (*domain.TaxRecord, error)
	ListTaxRecords(f domain.ListFilter) ([]*domain.TaxRecord, error)
	GetTaxSummary(jurisdiction, period string) ([]*domain.TaxSummary, error)
	GenerateReport(startDate, endDate, jurisdiction string) ([]*domain.TaxSummary, error)
}

// Service is the concrete implementation of Servicer.
type Service struct {
	store store.Storer
}

// New creates a new Service backed by the given Storer.
func New(s store.Storer) *Service {
	return &Service{store: s}
}

// RecordTax validates and persists a new tax record.
func (svc *Service) RecordTax(r *domain.TaxRecord) (*domain.TaxRecord, error) {
	if r.OrderID == "" {
		return nil, fmt.Errorf("orderId is required")
	}
	if r.CustomerID == "" {
		return nil, fmt.Errorf("customerId is required")
	}
	if r.Jurisdiction == "" {
		return nil, fmt.Errorf("jurisdiction is required")
	}
	if err := validateTaxType(r.TaxType); err != nil {
		return nil, err
	}
	if r.TaxableAmount < 0 {
		return nil, fmt.Errorf("taxableAmount must be non-negative")
	}
	if r.TaxRate < 0 || r.TaxRate > 100 {
		return nil, fmt.Errorf("taxRate must be between 0 and 100")
	}
	if r.Currency == "" {
		return nil, fmt.Errorf("currency is required")
	}

	// Compute tax amount when not explicitly provided.
	if r.TaxAmount == 0 {
		r.TaxAmount = roundTo2(r.TaxableAmount * r.TaxRate / 100)
	}

	now := time.Now().UTC()
	r.ID = uuid.New().String()
	r.CreatedAt = now
	if r.TransactionDate.IsZero() {
		r.TransactionDate = now
	}

	if err := svc.store.SaveRecord(r); err != nil {
		return nil, fmt.Errorf("service: record tax: %w", err)
	}
	return r, nil
}

// GetTaxRecord fetches a single tax record by ID.
func (svc *Service) GetTaxRecord(id string) (*domain.TaxRecord, error) {
	return svc.store.GetRecord(id)
}

// ListTaxRecords returns filtered tax records.
func (svc *Service) ListTaxRecords(f domain.ListFilter) ([]*domain.TaxRecord, error) {
	return svc.store.ListRecords(f)
}

// GetTaxSummary returns aggregated tax data for a jurisdiction + period (YYYY-MM).
func (svc *Service) GetTaxSummary(jurisdiction, period string) ([]*domain.TaxSummary, error) {
	if period == "" {
		return nil, fmt.Errorf("period is required (format: YYYY-MM)")
	}
	return svc.store.GetSummary(jurisdiction, period)
}

// GenerateReport produces a []TaxSummary for all jurisdictions/types within a date range.
// It iterates over each YYYY-MM period between startDate and endDate and aggregates results.
func (svc *Service) GenerateReport(startDate, endDate, jurisdiction string) ([]*domain.TaxSummary, error) {
	if startDate == "" || endDate == "" {
		return nil, fmt.Errorf("startDate and endDate are required")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("startDate must be in YYYY-MM-DD format: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("endDate must be in YYYY-MM-DD format: %w", err)
	}
	if end.Before(start) {
		return nil, fmt.Errorf("endDate must be on or after startDate")
	}

	// Collect all YYYY-MM periods between start and end (inclusive).
	periods := monthsBetween(start, end)

	// Aggregate across all periods.
	aggMap := map[string]*domain.TaxSummary{}
	for _, period := range periods {
		summaries, err := svc.store.GetSummary(jurisdiction, period)
		if err != nil {
			return nil, fmt.Errorf("service: generate report for period %s: %w", period, err)
		}
		for _, s := range summaries {
			key := s.Jurisdiction + "|" + string(s.TaxType)
			if existing, ok := aggMap[key]; ok {
				existing.TotalTaxable += s.TotalTaxable
				existing.TotalTax += s.TotalTax
				existing.TransactionCount += s.TransactionCount
				// Period becomes the range.
				existing.Period = startDate[:7] + " to " + endDate[:7]
			} else {
				copy := *s
				copy.Period = startDate[:7] + " to " + endDate[:7]
				aggMap[key] = &copy
			}
		}
	}

	result := make([]*domain.TaxSummary, 0, len(aggMap))
	for _, v := range aggMap {
		result = append(result, v)
	}
	return result, nil
}

// monthsBetween returns a slice of "YYYY-MM" strings covering [start, end].
func monthsBetween(start, end time.Time) []string {
	var months []string
	cur := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	last := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)
	for !cur.After(last) {
		months = append(months, cur.Format("2006-01"))
		cur = cur.AddDate(0, 1, 0)
	}
	return months
}

func validateTaxType(t domain.TaxType) error {
	switch t {
	case domain.TaxTypeVAT, domain.TaxTypeGST, domain.TaxTypeSalesTax, domain.TaxTypeExcise:
		return nil
	}
	return fmt.Errorf("taxType must be one of VAT, GST, SALES_TAX, EXCISE")
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
