// Package calculator provides rule-based tax calculation with in-memory rate tables.
// Supported jurisdictions: US (federal + per-state), EU VAT, Canada (GST + PST),
// Australia (GST), and a 0% default for all other countries.
package calculator

import (
	"strings"

	"github.com/shopos/tax-service/internal/domain"
)

// ─── rate tables ─────────────────────────────────────────────────────────────

// usStateTaxRates maps US state codes (upper-case) to their state sales-tax rate.
// Federal rate is 0% in the US — states carry the full tax burden.
var usStateTaxRates = map[string]float64{
	"AL": 0.04,
	"AK": 0.00,
	"AZ": 0.056,
	"AR": 0.065,
	"CA": 0.0725,
	"CO": 0.029,
	"CT": 0.0635,
	"DE": 0.00,
	"FL": 0.06,
	"GA": 0.04,
	"HI": 0.04,
	"ID": 0.06,
	"IL": 0.0625,
	"IN": 0.07,
	"IA": 0.06,
	"KS": 0.065,
	"KY": 0.06,
	"LA": 0.0445,
	"ME": 0.055,
	"MD": 0.06,
	"MA": 0.0625,
	"MI": 0.06,
	"MN": 0.06875,
	"MS": 0.07,
	"MO": 0.04225,
	"MT": 0.00,
	"NE": 0.055,
	"NV": 0.0685,
	"NH": 0.00,
	"NJ": 0.06625,
	"NM": 0.05125,
	"NY": 0.04,
	"NC": 0.0475,
	"ND": 0.05,
	"OH": 0.0575,
	"OK": 0.045,
	"OR": 0.00,
	"PA": 0.06,
	"RI": 0.07,
	"SC": 0.06,
	"SD": 0.045,
	"TN": 0.07,
	"TX": 0.0625,
	"UT": 0.0485,
	"VT": 0.06,
	"VA": 0.053,
	"WA": 0.065,
	"WV": 0.06,
	"WI": 0.05,
	"WY": 0.04,
	"DC": 0.06,
}

// defaultUSRate is applied when the state code is not in usStateTaxRates.
const defaultUSRate = 0.05

// euVATRates maps ISO-3166-1 alpha-2 country codes to their standard VAT rate.
// EU reverse-charge (B2B zero-rating) is handled in Calculate().
var euVATRates = map[string]float64{
	"AT": 0.20,
	"BE": 0.21,
	"BG": 0.20,
	"CY": 0.19,
	"CZ": 0.21,
	"DE": 0.19,
	"DK": 0.25,
	"EE": 0.22,
	"ES": 0.21,
	"FI": 0.24,
	"FR": 0.20,
	"GR": 0.24,
	"HR": 0.25,
	"HU": 0.27,
	"IE": 0.23,
	"IT": 0.22,
	"LT": 0.21,
	"LU": 0.17,
	"LV": 0.21,
	"MT": 0.18,
	"NL": 0.21,
	"PL": 0.23,
	"PT": 0.23,
	"RO": 0.19,
	"SE": 0.25,
	"SI": 0.22,
	"SK": 0.20,
	// Post-Brexit UK is treated like EU for simplicity
	"GB": 0.20,
}

// canadaPSTRates maps Canadian province codes to their Provincial Sales Tax rate.
// GST (5%) is applied on top in every province.
var canadaPSTRates = map[string]float64{
	"BC": 0.07,
	"MB": 0.07,
	"SK": 0.06,
	// QC has QST (9.975%), not PST — close enough for demo purposes
	"QC": 0.09975,
	// ON, AB, NL, NS, NB, PE, NT, NU, YT use HST or GST-only — PST = 0
}

// canadaGST is the federal Goods and Services Tax rate for Canada.
const canadaGST = 0.05

// auGST is the Australian Goods and Services Tax rate.
const auGST = 0.10

// ─── Calculator ───────────────────────────────────────────────────────────────

// Calculator provides stateless tax calculation.
type Calculator struct{}

// New returns a new Calculator.
func New() *Calculator { return &Calculator{} }

// Calculate computes taxes for the given TaxRequest using in-memory rate rules.
func (c *Calculator) Calculate(req domain.TaxRequest) domain.TaxResponse {
	country := strings.ToUpper(strings.TrimSpace(req.ShipTo.Country))
	state := strings.ToUpper(strings.TrimSpace(req.ShipTo.State))
	isB2B := strings.ToLower(strings.TrimSpace(req.CustomerType)) == "b2b"

	// Compute subtotal
	var subtotal float64
	for _, item := range req.Items {
		subtotal += item.Amount * float64(item.Quantity)
	}
	subtotal = round2(subtotal)

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Determine breakdown slices per jurisdiction
	breakdown := c.buildBreakdown(country, state, isB2B, subtotal)

	// Sum tax from all jurisdictions
	var totalTax float64
	for _, b := range breakdown {
		totalTax += b.Amount
	}
	totalTax = round2(totalTax)

	// Effective blended rate
	var effectiveRate float64
	if subtotal > 0 {
		effectiveRate = round4(totalTax / subtotal)
	}

	return domain.TaxResponse{
		Subtotal:  subtotal,
		TaxAmount: totalTax,
		TaxRate:   effectiveRate,
		Total:     round2(subtotal + totalTax),
		Currency:  currency,
		Breakdown: breakdown,
	}
}

// ─── buildBreakdown ───────────────────────────────────────────────────────────

func (c *Calculator) buildBreakdown(country, state string, isB2B bool, subtotal float64) []domain.TaxBreakdown {
	switch country {
	case "US":
		return c.usBreakdown(state, subtotal)
	case "CA":
		return c.canadaBreakdown(state, subtotal)
	case "AU":
		return c.australiaBreakdown(subtotal)
	default:
		if rate, ok := euVATRates[country]; ok {
			return c.euBreakdown(country, rate, isB2B, subtotal)
		}
		// No tax for unknown jurisdictions
		return []domain.TaxBreakdown{}
	}
}

func (c *Calculator) usBreakdown(state string, subtotal float64) []domain.TaxBreakdown {
	rate, ok := usStateTaxRates[state]
	if !ok {
		rate = defaultUSRate
	}
	if rate == 0 {
		return []domain.TaxBreakdown{
			{Jurisdiction: "US-" + state, Rate: 0, Amount: 0},
		}
	}
	jurisdiction := "US"
	if state != "" {
		jurisdiction = "US-" + state
	}
	return []domain.TaxBreakdown{
		{Jurisdiction: jurisdiction, Rate: rate, Amount: round2(subtotal * rate)},
	}
}

func (c *Calculator) euBreakdown(country string, rate float64, isB2B bool, subtotal float64) []domain.TaxBreakdown {
	// EU B2B reverse charge — zero-rated at the point of sale
	if isB2B {
		return []domain.TaxBreakdown{
			{Jurisdiction: country + "-VAT-ReverseCharge", Rate: 0, Amount: 0},
		}
	}
	return []domain.TaxBreakdown{
		{Jurisdiction: country + "-VAT", Rate: rate, Amount: round2(subtotal * rate)},
	}
}

func (c *Calculator) canadaBreakdown(province string, subtotal float64) []domain.TaxBreakdown {
	bd := []domain.TaxBreakdown{
		{Jurisdiction: "CA-GST", Rate: canadaGST, Amount: round2(subtotal * canadaGST)},
	}
	if pst, ok := canadaPSTRates[province]; ok && pst > 0 {
		label := "CA-" + province + "-PST"
		if province == "QC" {
			label = "CA-QC-QST"
		}
		bd = append(bd, domain.TaxBreakdown{
			Jurisdiction: label,
			Rate:         pst,
			Amount:       round2(subtotal * pst),
		})
	}
	return bd
}

func (c *Calculator) australiaBreakdown(subtotal float64) []domain.TaxBreakdown {
	return []domain.TaxBreakdown{
		{Jurisdiction: "AU-GST", Rate: auGST, Amount: round2(subtotal * auGST)},
	}
}

// ─── RateInfo ─────────────────────────────────────────────────────────────────

// RateInfo returns the applicable tax breakdown for a given country/state WITHOUT
// an order amount (rate query only — amounts will all be zero).
func (c *Calculator) RateInfo(country, state string) domain.RateInfo {
	country = strings.ToUpper(strings.TrimSpace(country))
	state = strings.ToUpper(strings.TrimSpace(state))

	bd := c.buildBreakdown(country, state, false, 0)

	var effective float64
	for _, b := range bd {
		effective += b.Rate
	}

	return domain.RateInfo{
		Country:       country,
		State:         state,
		Jurisdictions: bd,
		EffectiveRate: round4(effective),
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func round2(f float64) float64 {
	v := int64(f*100 + 0.5)
	return float64(v) / 100
}

func round4(f float64) float64 {
	v := int64(f*10000 + 0.5)
	return float64(v) / 10000
}
