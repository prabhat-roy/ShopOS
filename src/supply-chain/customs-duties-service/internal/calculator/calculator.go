package calculator

import (
	"math"
	"strings"

	"github.com/shopos/customs-duties-service/internal/domain"
)

// ─── static data tables ───────────────────────────────────────────────────────

// hsCodes is an in-memory HS tariff table covering 20+ common categories.
var hsCodes = map[string]*domain.HSCodeInfo{
	// Electronics
	"8471.30": {Code: "8471.30", Description: "Portable computers (laptops/tablets)", GeneralRate: 0.00},
	"8517.12": {Code: "8517.12", Description: "Mobile phones / smartphones", GeneralRate: 0.00},
	"8528.72": {Code: "8528.72", Description: "Television sets / monitors", GeneralRate: 0.03},
	"8519.81": {Code: "8519.81", Description: "Audio players / headphones", GeneralRate: 0.02},
	"8543.70": {Code: "8543.70", Description: "Electrical machines NES (misc electronics)", GeneralRate: 0.02},
	// Clothing & Apparel
	"6109.10": {Code: "6109.10", Description: "T-shirts and singlets — cotton", GeneralRate: 0.16},
	"6203.42": {Code: "6203.42", Description: "Men's trousers — cotton", GeneralRate: 0.165},
	"6204.62": {Code: "6204.62", Description: "Women's trousers — cotton", GeneralRate: 0.165},
	"6401.92": {Code: "6401.92", Description: "Waterproof footwear", GeneralRate: 0.37},
	"6403.99": {Code: "6403.99", Description: "Other footwear — leather uppers", GeneralRate: 0.085},
	// Food & Beverages
	"1806.32": {Code: "1806.32", Description: "Chocolate (not filled)", GeneralRate: 0.06},
	"2101.11": {Code: "2101.11", Description: "Instant coffee preparations", GeneralRate: 0.01},
	"0901.11": {Code: "0901.11", Description: "Coffee, not roasted", GeneralRate: 0.00},
	"1905.31": {Code: "1905.31", Description: "Sweet biscuits / cookies", GeneralRate: 0.08},
	// Machinery & Equipment
	"8421.39": {Code: "8421.39", Description: "Filtering machinery NES", GeneralRate: 0.03},
	"8479.89": {Code: "8479.89", Description: "Machines for particular industries NES", GeneralRate: 0.035},
	"8413.70": {Code: "8413.70", Description: "Centrifugal pumps NES", GeneralRate: 0.00},
	// Automotive
	"8703.23": {Code: "8703.23", Description: "Passenger vehicles — 1500-3000cc", GeneralRate: 0.025},
	"8708.99": {Code: "8708.99", Description: "Vehicle parts NES", GeneralRate: 0.025},
	// Cosmetics & Personal Care
	"3304.99": {Code: "3304.99", Description: "Beauty/cosmetics preparations NES", GeneralRate: 0.05},
	"3305.10": {Code: "3305.10", Description: "Shampoos", GeneralRate: 0.065},
	// Sports & Toys
	"9503.00": {Code: "9503.00", Description: "Toys NES", GeneralRate: 0.00},
	"9506.91": {Code: "9506.91", Description: "Exercise / sports equipment NES", GeneralRate: 0.04},
	// Books / Printed Matter
	"4901.99": {Code: "4901.99", Description: "Printed books / brochures", GeneralRate: 0.00},
}

// countryRates holds per-destination VAT/GST rates and de minimis thresholds.
// De minimis values are expressed in USD equivalents for simplicity.
var countryRates = map[string]*domain.CountryRates{
	// North America
	"US": {Country: "US", VATRate: 0.00, DeMinimiisUSD: 800, Notes: "No federal VAT; duty waived below $800 de minimis"},
	"CA": {Country: "CA", VATRate: 0.05, DeMinimiisUSD: 20, Notes: "5% GST; provincial taxes not included"},
	"MX": {Country: "MX", VATRate: 0.16, DeMinimiisUSD: 50, Notes: "16% IVA"},
	// European Union (representative members)
	"DE": {Country: "DE", VATRate: 0.19, DeMinimiisUSD: 162, Notes: "19% MwSt; EU de minimis €150 (~$162)"},
	"FR": {Country: "FR", VATRate: 0.20, DeMinimiisUSD: 162, Notes: "20% TVA; EU de minimis €150 (~$162)"},
	"ES": {Country: "ES", VATRate: 0.21, DeMinimiisUSD: 162, Notes: "21% IVA; EU de minimis €150 (~$162)"},
	"IT": {Country: "IT", VATRate: 0.22, DeMinimiisUSD: 162, Notes: "22% IVA; EU de minimis €150 (~$162)"},
	"NL": {Country: "NL", VATRate: 0.21, DeMinimiisUSD: 162, Notes: "21% BTW; EU de minimis €150 (~$162)"},
	// United Kingdom
	"GB": {Country: "GB", VATRate: 0.20, DeMinimiisUSD: 170, Notes: "20% VAT; de minimis £135 (~$170)"},
	// Asia-Pacific
	"AU": {Country: "AU", VATRate: 0.10, DeMinimiisUSD: 1000, Notes: "10% GST; A$1000 (~USD) de minimis"},
	"JP": {Country: "JP", VATRate: 0.10, DeMinimiisUSD: 160, Notes: "10% consumption tax"},
	"SG": {Country: "SG", VATRate: 0.09, DeMinimiisUSD: 400, Notes: "9% GST"},
	"KR": {Country: "KR", VATRate: 0.10, DeMinimiisUSD: 150, Notes: "10% VAT"},
	"IN": {Country: "IN", VATRate: 0.18, DeMinimiisUSD: 0, Notes: "18% GST; no de minimis"},
	"CN": {Country: "CN", VATRate: 0.13, DeMinimiisUSD: 50, Notes: "13% VAT"},
	// South America
	"BR": {Country: "BR", VATRate: 0.12, DeMinimiisUSD: 50, Notes: "12% IOF + 35% II (high tariff country); threshold $50"},
	"AR": {Country: "AR", VATRate: 0.21, DeMinimiisUSD: 25, Notes: "21% IVA"},
	// Middle East / Africa
	"AE": {Country: "AE", VATRate: 0.05, DeMinimiisUSD: 272, Notes: "5% VAT; AED 1000 (~$272) de minimis"},
	"SA": {Country: "SA", VATRate: 0.15, DeMinimiisUSD: 266, Notes: "15% VAT"},
	"ZA": {Country: "ZA", VATRate: 0.15, DeMinimiisUSD: 40, Notes: "15% VAT; R500 (~$40) de minimis"},
}

// prohibitedCountryPairs lists origin→destination pairs where goods are typically prohibited.
// This is a simplified sample list; real implementations would use regulatory databases.
var prohibitedCountryPairs = map[string]bool{
	"US→KP": true,
	"US→IR": true,
	"US→CU": true,
	"US→SY": true,
}

// brazilHighTariffRate is the additional Brazilian import tariff applied on top of the HS general rate.
const brazilHighTariffRate = 0.35

// ─── DutyCalculator ───────────────────────────────────────────────────────────

// DutyCalculator performs customs duty calculations.
type DutyCalculator struct{}

// New creates a new DutyCalculator.
func New() *DutyCalculator {
	return &DutyCalculator{}
}

// Calculate returns a full DutyResponse for the supplied request.
func (dc *DutyCalculator) Calculate(req domain.DutyRequest) (*domain.DutyResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	// Normalise country codes to uppercase.
	from := strings.ToUpper(req.FromCountry)
	to := strings.ToUpper(req.ToCountry)

	// Same-country shipments have no import duties.
	if from == to {
		return &domain.DutyResponse{
			FromCountry:     from,
			ToCountry:       to,
			HSCode:          req.HSCode,
			DeclaredValue:   req.DeclaredValue,
			Currency:        strings.ToUpper(req.Currency),
			DutyAmount:      0,
			VATAmount:       0,
			TotalLandedCost: req.DeclaredValue,
			Breakdown:       []domain.DutyLineItem{},
			Notes:           "Domestic shipment — no customs duties apply.",
		}, nil
	}

	// Check for prohibited country pairs.
	pairKey := from + "→" + to
	if prohibitedCountryPairs[pairKey] {
		return &domain.DutyResponse{
			FromCountry:     from,
			ToCountry:       to,
			HSCode:          req.HSCode,
			DeclaredValue:   req.DeclaredValue,
			Currency:        strings.ToUpper(req.Currency),
			ProhibitedItems: true,
			Notes:           "Shipments from " + from + " to " + to + " are restricted under trade sanctions.",
		}, nil
	}

	// Look up destination country rates.
	destRates, ok := countryRates[to]
	if !ok {
		// Unknown destination — apply a generic 5% duty + 0% VAT with no de minimis.
		destRates = &domain.CountryRates{Country: to, VATRate: 0.05, DeMinimiisUSD: 0}
	}

	// Check de minimis threshold (declared value in USD assumed for simplicity).
	if destRates.DeMinimiisUSD > 0 && req.DeclaredValue <= destRates.DeMinimiisUSD {
		return &domain.DutyResponse{
			FromCountry:     from,
			ToCountry:       to,
			HSCode:          req.HSCode,
			DeclaredValue:   req.DeclaredValue,
			Currency:        strings.ToUpper(req.Currency),
			DutyAmount:      0,
			VATAmount:       0,
			TotalLandedCost: req.DeclaredValue,
			Breakdown:       []domain.DutyLineItem{},
			DeMinimisMet:    true,
			Notes:           "Declared value is below de minimis threshold — no duties collected.",
		}, nil
	}

	// Resolve the HS code duty rate.
	hsInfo, hsFound := hsCodes[req.HSCode]
	var baseDutyRate float64
	if hsFound {
		baseDutyRate = hsInfo.GeneralRate
	} else {
		// Default fallback rate for unknown HS codes.
		baseDutyRate = 0.05
	}

	// Brazil applies an additional high import tariff.
	effectiveDutyRate := baseDutyRate
	if to == "BR" {
		effectiveDutyRate = baseDutyRate + brazilHighTariffRate
	}

	dutyAmount := round2(req.DeclaredValue * effectiveDutyRate)
	// VAT/GST is calculated on declared value + duty (CIF+duty basis).
	vatBase := req.DeclaredValue + dutyAmount
	vatAmount := round2(vatBase * destRates.VATRate)
	totalLanded := round2(req.DeclaredValue + dutyAmount + vatAmount)

	breakdown := buildBreakdown(req, hsInfo, effectiveDutyRate, dutyAmount, destRates, vatAmount)

	requiresForm := req.DeclaredValue > 2500 // simplified rule: >$2500 requires full customs declaration

	return &domain.DutyResponse{
		FromCountry:         from,
		ToCountry:           to,
		HSCode:              req.HSCode,
		DeclaredValue:       req.DeclaredValue,
		Currency:            strings.ToUpper(req.Currency),
		DutyAmount:          dutyAmount,
		VATAmount:           vatAmount,
		TotalLandedCost:     totalLanded,
		Breakdown:           breakdown,
		RequiresCustomsForm: requiresForm,
		ProhibitedItems:     false,
		DeMinimisMet:        false,
		Notes:               destRates.Notes,
	}, nil
}

// GetHSCode returns the HS code info for the given code string.
func (dc *DutyCalculator) GetHSCode(code string) (*domain.HSCodeInfo, error) {
	info, ok := hsCodes[code]
	if !ok {
		return nil, domain.ErrHSCodeNotFound
	}
	return info, nil
}

// ListHSCodes returns all known HS codes.
func (dc *DutyCalculator) ListHSCodes() []*domain.HSCodeInfo {
	out := make([]*domain.HSCodeInfo, 0, len(hsCodes))
	for _, v := range hsCodes {
		out = append(out, v)
	}
	return out
}

// GetCountryRates returns the import rate configuration for the given country code.
func (dc *DutyCalculator) GetCountryRates(country string) (*domain.CountryRates, error) {
	rates, ok := countryRates[strings.ToUpper(country)]
	if !ok {
		return nil, domain.ErrCountryNotFound
	}
	return rates, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func validateRequest(req domain.DutyRequest) error {
	if req.FromCountry == "" || req.ToCountry == "" {
		return domain.ErrInvalidRequest
	}
	if req.DeclaredValue < 0 {
		return domain.ErrInvalidRequest
	}
	if req.Quantity <= 0 {
		return domain.ErrInvalidRequest
	}
	return nil
}

func buildBreakdown(
	req domain.DutyRequest,
	hsInfo *domain.HSCodeInfo,
	dutyRate, dutyAmount float64,
	destRates *domain.CountryRates,
	vatAmount float64,
) []domain.DutyLineItem {
	description := req.Description
	if hsInfo != nil && description == "" {
		description = hsInfo.Description
	}

	items := []domain.DutyLineItem{
		{
			Description: "Import duty — " + description,
			DutyRate:    dutyRate,
			DutyAmount:  dutyAmount,
			TaxRate:     0,
			TaxAmount:   0,
		},
	}

	if destRates.VATRate > 0 {
		items = append(items, domain.DutyLineItem{
			Description: "VAT/GST — " + destRates.Country,
			DutyRate:    0,
			DutyAmount:  0,
			TaxRate:     destRates.VATRate,
			TaxAmount:   vatAmount,
		})
	}

	return items
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
