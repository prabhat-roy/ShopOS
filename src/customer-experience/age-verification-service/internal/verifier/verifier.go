package verifier

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopos/age-verification-service/internal/domain"
)

const dobLayout = "2006-01-02"

// AgeVerifier implements age verification logic.
type AgeVerifier struct{}

// New returns a new AgeVerifier.
func New() *AgeVerifier {
	return &AgeVerifier{}
}

// ValidateRequest checks that all required fields are present and well-formed.
func (v *AgeVerifier) ValidateRequest(req domain.VerificationRequest) error {
	if strings.TrimSpace(req.CustomerID) == "" {
		return fmt.Errorf("customerId is required")
	}
	if strings.TrimSpace(req.DateOfBirth) == "" {
		return fmt.Errorf("dateOfBirth is required")
	}
	if _, err := time.Parse(dobLayout, req.DateOfBirth); err != nil {
		return fmt.Errorf("dateOfBirth must be in YYYY-MM-DD format")
	}
	if strings.TrimSpace(req.Country) == "" {
		return fmt.Errorf("country is required")
	}
	if strings.TrimSpace(req.ProductCategory) == "" {
		return fmt.Errorf("productCategory is required")
	}
	return nil
}

// Verify performs the age check and returns a VerificationResult.
func (v *AgeVerifier) Verify(req domain.VerificationRequest) (domain.VerificationResult, error) {
	if err := v.ValidateRequest(req); err != nil {
		return domain.VerificationResult{}, err
	}

	dob, _ := time.Parse(dobLayout, req.DateOfBirth) // already validated
	age := computeAge(dob, time.Now())
	minAge := resolveMinAge(req.Country, req.ProductCategory)

	result := domain.VerificationResult{
		CustomerID:      req.CustomerID,
		Age:             age,
		MinAge:          minAge,
		Country:         req.Country,
		ProductCategory: req.ProductCategory,
		VerifiedAt:      time.Now().UTC(),
	}

	if age >= minAge {
		result.Verified = true
		result.Reason = fmt.Sprintf("customer is %d years old, meets minimum age of %d for %s in %s",
			age, minAge, req.ProductCategory, req.Country)
	} else {
		result.Verified = false
		result.Reason = fmt.Sprintf("customer is %d years old, does not meet minimum age of %d for %s in %s",
			age, minAge, req.ProductCategory, req.Country)
	}
	return result, nil
}

// BatchVerify runs Verify for each request. Invalid requests produce a failed result rather than an error.
func (v *AgeVerifier) BatchVerify(reqs []domain.VerificationRequest) []domain.VerificationResult {
	results := make([]domain.VerificationResult, 0, len(reqs))
	for _, req := range reqs {
		res, err := v.Verify(req)
		if err != nil {
			results = append(results, domain.VerificationResult{
				CustomerID:      req.CustomerID,
				Verified:        false,
				Country:         req.Country,
				ProductCategory: req.ProductCategory,
				Reason:          err.Error(),
				VerifiedAt:      time.Now().UTC(),
			})
			continue
		}
		results = append(results, res)
	}
	return results
}

// MinAgeFor returns the effective minimum age for a given country and category combination.
func MinAgeFor(country, category string) int {
	return resolveMinAge(country, category)
}

// --- private helpers ---

// computeAge returns the number of full years between dob and now.
func computeAge(dob, now time.Time) int {
	years := now.Year() - dob.Year()
	// Check if the birthday has occurred yet this calendar year
	birthdayThisYear := time.Date(now.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, now.Location())
	if now.Before(birthdayThisYear) {
		years--
	}
	return years
}

// resolveMinAge returns the effective minimum age: max(countryMinAge, categoryMinAge).
func resolveMinAge(country, category string) int {
	countryMin, ok := domain.CountryMinAges[strings.ToUpper(country)]
	if !ok {
		countryMin = domain.DefaultMinAge
	}
	catMin, ok := domain.CategoryMinAges[strings.ToLower(category)]
	if !ok {
		catMin = domain.DefaultMinAge
	}
	if countryMin > catMin {
		return countryMin
	}
	return catMin
}
