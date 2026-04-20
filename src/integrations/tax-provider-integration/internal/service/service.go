package service

import (
	"fmt"

	"github.com/shopos/tax-provider-integration/internal/adapter"
	"github.com/shopos/tax-provider-integration/internal/domain"
)

// Servicer is the interface that defines the tax business operations.
// Exposed as an interface so handlers can be tested with a mock.
type Servicer interface {
	CalculateTax(req domain.TaxCalculationRequest) (domain.TaxCalculationResponse, error)
	CommitTransaction(req domain.CommitRequest) (domain.CommitResponse, error)
	GetProviderInfo(provider domain.TaxProvider) (map[string]interface{}, error)
	ValidateAddress(provider domain.TaxProvider, addr domain.TaxAddress) (bool, string, error)
	ListProviders() []domain.TaxProvider
}

// Service implements Servicer by delegating to the TaxProviderAdapter.
type Service struct {
	adapter *adapter.TaxProviderAdapter
}

// New creates a Service backed by the given adapter.
func New(a *adapter.TaxProviderAdapter) *Service {
	return &Service{adapter: a}
}

// CalculateTax validates the request and delegates to the adapter.
func (svc *Service) CalculateTax(req domain.TaxCalculationRequest) (domain.TaxCalculationResponse, error) {
	if err := validateCalculationRequest(req); err != nil {
		return domain.TaxCalculationResponse{}, err
	}
	return svc.adapter.Calculate(req)
}

// CommitTransaction commits a previously calculated transaction at the provider.
func (svc *Service) CommitTransaction(req domain.CommitRequest) (domain.CommitResponse, error) {
	if req.TransactionID == "" {
		return domain.CommitResponse{}, fmt.Errorf("transactionId is required")
	}
	return svc.adapter.Commit(req)
}

// GetProviderInfo returns metadata about a given provider.
func (svc *Service) GetProviderInfo(provider domain.TaxProvider) (map[string]interface{}, error) {
	return svc.adapter.GetProviderInfo(provider)
}

// ValidateAddress checks whether an address is valid according to the given provider.
func (svc *Service) ValidateAddress(provider domain.TaxProvider, addr domain.TaxAddress) (bool, string, error) {
	if !isValidProvider(provider) {
		return false, "", fmt.Errorf("unsupported provider: %s", provider)
	}
	valid, msg := svc.adapter.ValidateAddress(provider, addr)
	return valid, msg, nil
}

// ListProviders returns all supported tax providers.
func (svc *Service) ListProviders() []domain.TaxProvider {
	return domain.AllProviders
}

// ---------------------------------------------------------------------------
// validation helpers
// ---------------------------------------------------------------------------

func validateCalculationRequest(req domain.TaxCalculationRequest) error {
	if !isValidProvider(req.Provider) {
		return fmt.Errorf("unsupported provider: %q", req.Provider)
	}
	if req.TransactionID == "" {
		return fmt.Errorf("transactionId is required")
	}
	if len(req.LineItems) == 0 {
		return fmt.Errorf("at least one lineItem is required")
	}
	for i, item := range req.LineItems {
		if item.Quantity <= 0 {
			return fmt.Errorf("lineItems[%d].quantity must be > 0", i)
		}
		if item.UnitPrice < 0 {
			return fmt.Errorf("lineItems[%d].unitPrice must be >= 0", i)
		}
	}
	if req.ToAddress.PostalCode == "" {
		return fmt.Errorf("toAddress.postalCode is required")
	}
	return nil
}

func isValidProvider(p domain.TaxProvider) bool {
	for _, v := range domain.AllProviders {
		if v == p {
			return true
		}
	}
	return false
}
