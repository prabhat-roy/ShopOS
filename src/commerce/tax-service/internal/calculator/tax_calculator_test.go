package calculator_test

import (
	"math"
	"testing"

	"github.com/shopos/tax-service/internal/calculator"
	"github.com/shopos/tax-service/internal/domain"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func newCalc() *calculator.Calculator { return calculator.New() }

func singleItemReq(country, state, customerType string, amount float64) domain.TaxRequest {
	return domain.TaxRequest{
		Items: []domain.TaxLineItem{
			{ProductID: "p-1", Category: "general", Amount: amount, Quantity: 1},
		},
		ShipTo: domain.Address{
			Country: country,
			State:   state,
		},
		Currency:     "USD",
		CustomerType: customerType,
	}
}

// approxEqual returns true if a and b are within 0.01 of each other.
func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

// ─── US tests ────────────────────────────────────────────────────────────────

func TestUS_CA_Rate(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("US", "CA", "b2c", 100.00)
	resp := calc.Calculate(req)

	// CA = 7.25%  => tax on $100 = $7.25
	if !approxEqual(resp.TaxAmount, 7.25) {
		t.Errorf("US-CA: expected tax 7.25, got %.4f", resp.TaxAmount)
	}
	if !approxEqual(resp.TaxRate, 0.0725) {
		t.Errorf("US-CA: expected rate 0.0725, got %.4f", resp.TaxRate)
	}
	if !approxEqual(resp.Total, 107.25) {
		t.Errorf("US-CA: expected total 107.25, got %.4f", resp.Total)
	}
}

func TestUS_NY_Rate(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("US", "NY", "b2c", 200.00)
	resp := calc.Calculate(req)

	// NY = 4%  => tax on $200 = $8.00
	if !approxEqual(resp.TaxAmount, 8.00) {
		t.Errorf("US-NY: expected tax 8.00, got %.4f", resp.TaxAmount)
	}
}

func TestUS_TX_Rate(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("US", "TX", "b2c", 100.00)
	resp := calc.Calculate(req)

	// TX = 6.25%  => $6.25
	if !approxEqual(resp.TaxAmount, 6.25) {
		t.Errorf("US-TX: expected tax 6.25, got %.4f", resp.TaxAmount)
	}
}

func TestUS_WA_Rate(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("US", "WA", "b2c", 100.00)
	resp := calc.Calculate(req)

	// WA = 6.5%  => $6.50
	if !approxEqual(resp.TaxAmount, 6.50) {
		t.Errorf("US-WA: expected tax 6.50, got %.4f", resp.TaxAmount)
	}
}

func TestUS_UnknownState_DefaultRate(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("US", "XX", "b2c", 100.00)
	resp := calc.Calculate(req)

	// Default US = 5%  => $5.00
	if !approxEqual(resp.TaxAmount, 5.00) {
		t.Errorf("US-XX default: expected tax 5.00, got %.4f", resp.TaxAmount)
	}
}

// ─── EU tests ────────────────────────────────────────────────────────────────

func TestEU_DE_VAT(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("DE", "", "b2c", 100.00)
	resp := calc.Calculate(req)

	// DE VAT = 19%  => $19.00
	if !approxEqual(resp.TaxAmount, 19.00) {
		t.Errorf("EU-DE: expected tax 19.00, got %.4f", resp.TaxAmount)
	}
	if !approxEqual(resp.TaxRate, 0.19) {
		t.Errorf("EU-DE: expected rate 0.19, got %.4f", resp.TaxRate)
	}
}

func TestEU_FR_VAT(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("FR", "", "b2c", 50.00)
	resp := calc.Calculate(req)

	// FR VAT = 20%  => $10.00
	if !approxEqual(resp.TaxAmount, 10.00) {
		t.Errorf("EU-FR: expected tax 10.00, got %.4f", resp.TaxAmount)
	}
}

func TestEU_NL_VAT(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("NL", "", "b2c", 100.00)
	resp := calc.Calculate(req)

	// NL VAT = 21%  => $21.00
	if !approxEqual(resp.TaxAmount, 21.00) {
		t.Errorf("EU-NL: expected tax 21.00, got %.4f", resp.TaxAmount)
	}
}

func TestEU_IE_VAT(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("IE", "", "b2c", 100.00)
	resp := calc.Calculate(req)

	// IE VAT = 23%  => $23.00
	if !approxEqual(resp.TaxAmount, 23.00) {
		t.Errorf("EU-IE: expected tax 23.00, got %.4f", resp.TaxAmount)
	}
}

func TestEU_IT_VAT(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("IT", "", "b2c", 100.00)
	resp := calc.Calculate(req)

	// IT VAT = 22%  => $22.00
	if !approxEqual(resp.TaxAmount, 22.00) {
		t.Errorf("EU-IT: expected tax 22.00, got %.4f", resp.TaxAmount)
	}
}

func TestEU_GB_VAT(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("GB", "", "b2c", 100.00)
	resp := calc.Calculate(req)

	// GB VAT = 20%  => $20.00
	if !approxEqual(resp.TaxAmount, 20.00) {
		t.Errorf("EU-GB: expected tax 20.00, got %.4f", resp.TaxAmount)
	}
}

// ─── EU B2B reverse charge ────────────────────────────────────────────────────

func TestEU_B2B_ReverseCharge_Zero(t *testing.T) {
	calc := newCalc()
	// DE, FR, GB — all should be 0% when customer_type=b2b
	for _, country := range []string{"DE", "FR", "GB", "NL", "IE", "IT"} {
		req := singleItemReq(country, "", "b2b", 100.00)
		resp := calc.Calculate(req)
		if resp.TaxAmount != 0 {
			t.Errorf("EU B2B reverse charge %s: expected 0 tax, got %.4f", country, resp.TaxAmount)
		}
		if resp.TaxRate != 0 {
			t.Errorf("EU B2B reverse charge %s: expected 0 rate, got %.4f", country, resp.TaxRate)
		}
	}
}

// ─── Canada tests ─────────────────────────────────────────────────────────────

func TestCA_GST_Only(t *testing.T) {
	calc := newCalc()
	// Alberta has no PST — only GST 5%
	req := singleItemReq("CA", "AB", "b2c", 100.00)
	resp := calc.Calculate(req)

	if !approxEqual(resp.TaxAmount, 5.00) {
		t.Errorf("CA-AB GST only: expected 5.00, got %.4f", resp.TaxAmount)
	}
}

func TestCA_BC_GST_PST(t *testing.T) {
	calc := newCalc()
	// BC: GST 5% + PST 7% = 12%
	req := singleItemReq("CA", "BC", "b2c", 100.00)
	resp := calc.Calculate(req)

	if !approxEqual(resp.TaxAmount, 12.00) {
		t.Errorf("CA-BC: expected 12.00 tax, got %.4f", resp.TaxAmount)
	}
	if len(resp.Breakdown) != 2 {
		t.Errorf("CA-BC: expected 2 breakdown entries, got %d", len(resp.Breakdown))
	}
}

// ─── Australia ────────────────────────────────────────────────────────────────

func TestAU_GST(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("AU", "", "b2c", 100.00)
	resp := calc.Calculate(req)

	// AU GST = 10%  => $10.00
	if !approxEqual(resp.TaxAmount, 10.00) {
		t.Errorf("AU GST: expected 10.00, got %.4f", resp.TaxAmount)
	}
	if !approxEqual(resp.TaxRate, 0.10) {
		t.Errorf("AU GST: expected rate 0.10, got %.4f", resp.TaxRate)
	}
}

// ─── Unknown country defaults to 0% ──────────────────────────────────────────

func TestUnknownCountry_ZeroTax(t *testing.T) {
	calc := newCalc()
	req := singleItemReq("XX", "", "b2c", 100.00)
	resp := calc.Calculate(req)

	if resp.TaxAmount != 0 {
		t.Errorf("unknown country: expected 0 tax, got %.4f", resp.TaxAmount)
	}
	if len(resp.Breakdown) != 0 {
		t.Errorf("unknown country: expected empty breakdown, got %d entries", len(resp.Breakdown))
	}
}

// ─── Multi-item subtotal ──────────────────────────────────────────────────────

func TestMultipleItems_Subtotal(t *testing.T) {
	calc := newCalc()
	req := domain.TaxRequest{
		Items: []domain.TaxLineItem{
			{ProductID: "p-1", Amount: 10.00, Quantity: 3},  // $30
			{ProductID: "p-2", Amount: 25.50, Quantity: 2},  // $51
			{ProductID: "p-3", Amount: 5.00, Quantity: 1},   // $5
		},
		ShipTo:       domain.Address{Country: "US", State: "CA"},
		Currency:     "USD",
		CustomerType: "b2c",
	}
	// subtotal = $86  tax = 86 * 7.25% = 6.235  => rounds to $6.24
	resp := calc.Calculate(req)
	if !approxEqual(resp.Subtotal, 86.00) {
		t.Errorf("multi-item subtotal: expected 86.00, got %.4f", resp.Subtotal)
	}
	if !approxEqual(resp.TaxAmount, 6.24) {
		t.Errorf("multi-item tax: expected 6.24, got %.4f", resp.TaxAmount)
	}
}

// ─── RateInfo ─────────────────────────────────────────────────────────────────

func TestRateInfo_US_CA(t *testing.T) {
	calc := newCalc()
	info := calc.RateInfo("US", "CA")
	if !approxEqual(info.EffectiveRate, 0.0725) {
		t.Errorf("RateInfo US-CA: expected 0.0725, got %.4f", info.EffectiveRate)
	}
}

func TestRateInfo_DE(t *testing.T) {
	calc := newCalc()
	info := calc.RateInfo("DE", "")
	if !approxEqual(info.EffectiveRate, 0.19) {
		t.Errorf("RateInfo DE: expected 0.19, got %.4f", info.EffectiveRate)
	}
}
