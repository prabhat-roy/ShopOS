// Package validator provides rule-based postal address validation and normalization.
// It requires no external API calls and operates entirely from local rules.
package validator

import (
	"regexp"
	"strings"

	"github.com/shopos/address-validation-service/domain"
)

// Compiled regular expressions for postal code patterns.
var (
	reUSZip    = regexp.MustCompile(`^\d{5}(-\d{4})?$`)
	reUKPost   = regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]?\s*\d[A-Z]{2}$`)
	reCAPost   = regexp.MustCompile(`^[A-Z]\d[A-Z]\s*\d[A-Z]\d$`)
	reUS2State = regexp.MustCompile(`^[A-Z]{2}$`)
)

// Valid US state abbreviations.
var usStates = map[string]struct{}{
	"AL": {}, "AK": {}, "AZ": {}, "AR": {}, "CA": {}, "CO": {}, "CT": {}, "DE": {},
	"FL": {}, "GA": {}, "HI": {}, "ID": {}, "IL": {}, "IN": {}, "IA": {}, "KS": {},
	"KY": {}, "LA": {}, "ME": {}, "MD": {}, "MA": {}, "MI": {}, "MN": {}, "MS": {},
	"MO": {}, "MT": {}, "NE": {}, "NV": {}, "NH": {}, "NJ": {}, "NM": {}, "NY": {},
	"NC": {}, "ND": {}, "OH": {}, "OK": {}, "OR": {}, "PA": {}, "RI": {}, "SC": {},
	"SD": {}, "TN": {}, "TX": {}, "UT": {}, "VT": {}, "VA": {}, "WA": {}, "WV": {},
	"WI": {}, "WY": {}, "DC": {}, "PR": {}, "GU": {}, "VI": {}, "AS": {}, "MP": {},
}

// Validate checks and normalizes the provided address.
func Validate(addr domain.Address) domain.ValidationResult {
	issues := []string{}

	// ── Normalize ────────────────────────────────────────────────────────────
	norm := domain.Address{
		Line1:      strings.TrimSpace(addr.Line1),
		Line2:      strings.TrimSpace(addr.Line2),
		City:       strings.TrimSpace(addr.City),
		State:      strings.ToUpper(strings.TrimSpace(addr.State)),
		PostalCode: strings.TrimSpace(addr.PostalCode),
		Country:    strings.ToUpper(strings.TrimSpace(addr.Country)),
	}

	// ── Required field checks ────────────────────────────────────────────────
	if norm.Line1 == "" {
		issues = append(issues, "line1 is required")
	}
	if norm.City == "" {
		issues = append(issues, "city is required")
	}
	if norm.Country == "" {
		issues = append(issues, "country is required")
	}

	// If we are missing a required field we can stop early.
	if len(issues) > 0 {
		return domain.ValidationResult{
			Valid:      false,
			Normalized: &norm,
			Issues:     issues,
			Confidence: 0.0,
		}
	}

	// ── Country-specific rules ───────────────────────────────────────────────
	switch norm.Country {
	case "US":
		validateUS(&norm, &issues)
	case "GB":
		validateUK(&norm, &issues)
	case "CA":
		validateCA(&norm, &issues)
	default:
		// Generic: no country-specific rules; postal code optional
	}

	confidence := computeConfidence(norm, len(issues))
	return domain.ValidationResult{
		Valid:      len(issues) == 0,
		Normalized: &norm,
		Issues:     issues,
		Confidence: confidence,
	}
}

// validateUS applies US-specific postal and state rules.
func validateUS(addr *domain.Address, issues *[]string) {
	// State must be present and a valid 2-letter code.
	if addr.State == "" {
		*issues = append(*issues, "state is required for US addresses")
	} else if !reUS2State.MatchString(addr.State) {
		*issues = append(*issues, "state must be a 2-letter US abbreviation")
	} else if _, ok := usStates[addr.State]; !ok {
		*issues = append(*issues, "unrecognized US state abbreviation: "+addr.State)
	}

	// Postal code must match 5-digit or ZIP+4 format.
	if addr.PostalCode == "" {
		*issues = append(*issues, "postal_code is required for US addresses")
	} else {
		// Normalize ZIP+4 separator
		normalized := strings.ReplaceAll(addr.PostalCode, " ", "")
		addr.PostalCode = normalized
		if !reUSZip.MatchString(normalized) {
			*issues = append(*issues, "postal_code must match US ZIP format (12345 or 12345-6789)")
		}
	}
}

// validateUK applies UK-specific postal code rules.
func validateUK(addr *domain.Address, issues *[]string) {
	if addr.PostalCode == "" {
		*issues = append(*issues, "postal_code is required for GB addresses")
		return
	}
	// Normalize: uppercase, collapse internal whitespace to single space
	pc := strings.ToUpper(strings.Join(strings.Fields(addr.PostalCode), " "))
	addr.PostalCode = pc
	if !reUKPost.MatchString(pc) {
		*issues = append(*issues, "postal_code does not match UK postcode format (e.g. SW1A 1AA)")
	}
}

// validateCA applies Canadian postal code rules.
func validateCA(addr *domain.Address, issues *[]string) {
	if addr.PostalCode == "" {
		*issues = append(*issues, "postal_code is required for CA addresses")
		return
	}
	// Normalize: uppercase, single space in the middle (A1A 1A1)
	pc := strings.ToUpper(strings.ReplaceAll(addr.PostalCode, " ", ""))
	if len(pc) == 6 {
		pc = pc[:3] + " " + pc[3:]
	}
	addr.PostalCode = pc
	if !reCAPost.MatchString(pc) {
		*issues = append(*issues, "postal_code does not match Canadian format (e.g. M5V 2T6)")
	}
}

// computeConfidence returns a simple heuristic confidence score.
func computeConfidence(addr domain.Address, issueCount int) float64 {
	if issueCount > 0 {
		return 0.0
	}
	score := 0.6 // baseline for passing required fields
	if addr.State != "" {
		score += 0.1
	}
	if addr.PostalCode != "" {
		score += 0.2
	}
	if addr.Line2 != "" {
		score += 0.1
	}
	if score > 1.0 {
		score = 1.0
	}
	return score
}
