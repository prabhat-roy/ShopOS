package adapter

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/shopos/tax-provider-integration/internal/domain"
)

// TaxProviderAdapter routes tax operations to the appropriate provider simulation.
// In production each provider method would call the real provider REST/SOAP API.
type TaxProviderAdapter struct{}

// New returns a ready-to-use TaxProviderAdapter.
func New() *TaxProviderAdapter {
	return &TaxProviderAdapter{}
}

// Calculate routes to the provider-specific calculation method.
func (a *TaxProviderAdapter) Calculate(req domain.TaxCalculationRequest) (domain.TaxCalculationResponse, error) {
	subtotal := lineItemsSubtotal(req.LineItems)

	var breakdown []domain.TaxBreakdownItem
	switch req.Provider {
	case domain.ProviderAvalara:
		breakdown = avalaraBreakdown(req.ToAddress, subtotal)
	case domain.ProviderTaxJar:
		breakdown = taxJarBreakdown(req.ToAddress, subtotal)
	case domain.ProviderVertex:
		breakdown = vertexBreakdown(req.ToAddress, subtotal)
	case domain.ProviderInternal:
		breakdown = internalBreakdown(req.ToAddress, subtotal)
	default:
		return domain.TaxCalculationResponse{}, fmt.Errorf("unsupported provider: %s", req.Provider)
	}

	totalTax := sumBreakdown(breakdown)
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	return domain.TaxCalculationResponse{
		Provider:      req.Provider,
		TransactionID: req.TransactionID,
		Subtotal:      round2(subtotal),
		TotalTax:      round2(totalTax),
		Total:         round2(subtotal + totalTax),
		Currency:      currency,
		Breakdown:     breakdown,
		CalculatedAt:  time.Now().UTC(),
	}, nil
}

// Commit simulates recording a transaction at the provider.
func (a *TaxProviderAdapter) Commit(req domain.CommitRequest) (domain.CommitResponse, error) {
	if req.TransactionID == "" {
		return domain.CommitResponse{}, fmt.Errorf("transactionId is required")
	}
	if !isValidProvider(req.Provider) {
		return domain.CommitResponse{}, fmt.Errorf("unsupported provider: %s", req.Provider)
	}
	return domain.CommitResponse{
		Committed:   true,
		CommittedAt: time.Now().UTC(),
	}, nil
}

// GetProviderInfo returns metadata about a specific provider.
func (a *TaxProviderAdapter) GetProviderInfo(provider domain.TaxProvider) (map[string]interface{}, error) {
	infos := map[domain.TaxProvider]map[string]interface{}{
		domain.ProviderAvalara: {
			"name":               "Avalara AvaTax",
			"version":            "21.12.0",
			"supportedCountries": []string{"US", "CA", "GB", "AU", "DE", "FR", "JP"},
			"features":           []string{"real_time_calculation", "transaction_commit", "address_validation", "nexus_management"},
			"apiDocs":            "https://developer.avalara.com/api-reference/avatax/rest/v2/",
		},
		domain.ProviderTaxJar: {
			"name":               "TaxJar SmartCalcs",
			"version":            "2.0",
			"supportedCountries": []string{"US", "CA", "AU"},
			"features":           []string{"real_time_calculation", "transaction_commit", "address_validation"},
			"apiDocs":            "https://developers.taxjar.com/api/reference/",
		},
		domain.ProviderVertex: {
			"name":               "Vertex O Series",
			"version":            "9.0",
			"supportedCountries": []string{"US", "CA", "GB", "DE", "FR", "IT", "ES", "NL", "AU", "BR", "MX"},
			"features":           []string{"real_time_calculation", "transaction_commit", "nexus_logic", "enterprise_rates", "vat_calculation"},
			"apiDocs":            "https://tax.vertexsmb.com/vertex-cloud-api",
		},
		domain.ProviderInternal: {
			"name":               "ShopOS Internal Tax Engine",
			"version":            "1.0",
			"supportedCountries": []string{"US"},
			"features":           []string{"real_time_calculation", "basic_us_rates"},
			"apiDocs":            "https://internal.shopos/tax-service",
		},
	}
	info, ok := infos[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
	return info, nil
}

// ValidateAddress simulates provider-side address validation.
// Returns (valid bool, normalizedAddress string).
func (a *TaxProviderAdapter) ValidateAddress(provider domain.TaxProvider, addr domain.TaxAddress) (bool, string) {
	if !isValidProvider(provider) {
		return false, "unsupported provider"
	}
	// Basic sanity check — a real implementation would call the carrier API.
	if strings.TrimSpace(addr.Street) == "" ||
		strings.TrimSpace(addr.City) == "" ||
		strings.TrimSpace(addr.PostalCode) == "" {
		return false, "address is incomplete"
	}
	normalized := fmt.Sprintf("%s, %s %s, %s",
		strings.ToUpper(addr.City),
		strings.ToUpper(addr.State),
		strings.ToUpper(addr.PostalCode),
		strings.ToUpper(addr.Country),
	)
	return true, normalized
}

// ---------------------------------------------------------------------------
// provider-specific rate logic
// ---------------------------------------------------------------------------

// avalaraBreakdown applies US state + county + city rates (Avalara model).
func avalaraBreakdown(addr domain.TaxAddress, subtotal float64) []domain.TaxBreakdownItem {
	state := resolveStateRate(addr.State)
	county := round4(state * 0.15)  // county adds ~15% of state rate
	city := round4(state * 0.10)    // city adds ~10% of state rate

	return []domain.TaxBreakdownItem{
		{
			Jurisdiction: fmt.Sprintf("%s State", strings.ToUpper(addr.State)),
			TaxType:      "SALES_TAX",
			Rate:         state,
			Amount:       round2(subtotal * state),
			Description:  fmt.Sprintf("State sales tax — %s", strings.ToUpper(addr.State)),
		},
		{
			Jurisdiction: fmt.Sprintf("%s County", strings.ToUpper(addr.City)),
			TaxType:      "COUNTY_TAX",
			Rate:         county,
			Amount:       round2(subtotal * county),
			Description:  "County surtax (Avalara jurisdiction)",
		},
		{
			Jurisdiction: strings.ToUpper(addr.City),
			TaxType:      "CITY_TAX",
			Rate:         city,
			Amount:       round2(subtotal * city),
			Description:  "City/municipal tax (Avalara jurisdiction)",
		},
	}
}

// taxJarBreakdown uses simplified state-level rates (TaxJar model).
func taxJarBreakdown(addr domain.TaxAddress, subtotal float64) []domain.TaxBreakdownItem {
	rate := resolveStateRate(addr.State)
	// TaxJar combines all jurisdictions into one line.
	combined := round4(rate * 1.20) // blended rate
	return []domain.TaxBreakdownItem{
		{
			Jurisdiction: strings.ToUpper(addr.State),
			TaxType:      "COMBINED_RATE",
			Rate:         combined,
			Amount:       round2(subtotal * combined),
			Description:  fmt.Sprintf("Combined sales tax rate — %s (TaxJar SmartCalcs)", strings.ToUpper(addr.State)),
		},
	}
}

// vertexBreakdown applies enterprise rates with nexus logic (Vertex model).
func vertexBreakdown(addr domain.TaxAddress, subtotal float64) []domain.TaxBreakdownItem {
	state := resolveStateRate(addr.State)
	nexusFee := 0.0025 // Vertex adds a nominal nexus-compliance fee

	return []domain.TaxBreakdownItem{
		{
			Jurisdiction: fmt.Sprintf("%s (Vertex)", strings.ToUpper(addr.State)),
			TaxType:      "STATE_TAX",
			Rate:         state,
			Amount:       round2(subtotal * state),
			Description:  "Enterprise state rate (Vertex O Series)",
		},
		{
			Jurisdiction: "NEXUS",
			TaxType:      "NEXUS_COMPLIANCE",
			Rate:         nexusFee,
			Amount:       round2(subtotal * nexusFee),
			Description:  "Nexus compliance fee (Vertex)",
		},
	}
}

// internalBreakdown mirrors the logic used by the internal tax-service.
func internalBreakdown(addr domain.TaxAddress, subtotal float64) []domain.TaxBreakdownItem {
	rate := resolveStateRate(addr.State)
	return []domain.TaxBreakdownItem{
		{
			Jurisdiction: strings.ToUpper(addr.State),
			TaxType:      "SALES_TAX",
			Rate:         rate,
			Amount:       round2(subtotal * rate),
			Description:  fmt.Sprintf("Standard US sales tax — %s", strings.ToUpper(addr.State)),
		},
	}
}

// resolveStateRate returns a representative sales tax rate for a US state code.
func resolveStateRate(state string) float64 {
	rates := map[string]float64{
		"AL": 0.04, "AK": 0.00, "AZ": 0.056, "AR": 0.065, "CA": 0.0725,
		"CO": 0.029, "CT": 0.0635, "DE": 0.00, "FL": 0.06, "GA": 0.04,
		"HI": 0.04, "ID": 0.06, "IL": 0.0625, "IN": 0.07, "IA": 0.06,
		"KS": 0.065, "KY": 0.06, "LA": 0.0445, "ME": 0.055, "MD": 0.06,
		"MA": 0.0625, "MI": 0.06, "MN": 0.06875, "MS": 0.07, "MO": 0.04225,
		"MT": 0.00, "NE": 0.055, "NV": 0.0685, "NH": 0.00, "NJ": 0.06625,
		"NM": 0.05125, "NY": 0.04, "NC": 0.0475, "ND": 0.05, "OH": 0.0575,
		"OK": 0.045, "OR": 0.00, "PA": 0.06, "RI": 0.07, "SC": 0.06,
		"SD": 0.045, "TN": 0.07, "TX": 0.0625, "UT": 0.0485, "VT": 0.06,
		"VA": 0.053, "WA": 0.065, "WV": 0.06, "WI": 0.05, "WY": 0.04,
	}
	if r, ok := rates[strings.ToUpper(state)]; ok {
		return r
	}
	return 0.06 // default fallback
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func lineItemsSubtotal(items []domain.TaxLineItem) float64 {
	total := 0.0
	for _, item := range items {
		if item.Amount > 0 {
			total += item.Amount
		} else {
			total += float64(item.Quantity) * item.UnitPrice
		}
	}
	return total
}

func sumBreakdown(items []domain.TaxBreakdownItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Amount
	}
	return total
}

func isValidProvider(p domain.TaxProvider) bool {
	for _, v := range domain.AllProviders {
		if v == p {
			return true
		}
	}
	return false
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
func round4(v float64) float64 { return math.Round(v*10000) / 10000 }
