package verifier_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/shopos/age-verification-service/internal/domain"
	"github.com/shopos/age-verification-service/internal/verifier"
)

var v = verifier.New()

// dobYearsAgo returns a YYYY-MM-DD string for a birthday exactly `years` years ago.
func dobYearsAgo(years int) string {
	t := time.Now().AddDate(-years, 0, 0)
	return t.Format("2006-01-02")
}

// dobExactlyToday returns a YYYY-MM-DD string for today (age == 0 years).
func dobExactlyToday() string {
	return time.Now().Format("2006-01-02")
}

// dobYesterdayBirthday returns someone who turned exactly `years` years old yesterday.
func dobYesterdayBirthday(years int) string {
	t := time.Now().AddDate(-years, 0, 0).AddDate(0, 0, -1)
	return t.Format("2006-01-02")
}

// Test 1: Valid adult in UK for alcohol (18 required, person is 25)
func TestVerify_UKAlcohol_Adult(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-uk-1",
		DateOfBirth:     dobYearsAgo(25),
		Country:         "UK",
		ProductCategory: "alcohol",
	}
	result, err := v.Verify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Verified {
		t.Errorf("expected verified=true for 25yo UK alcohol, reason: %s", result.Reason)
	}
	if result.MinAge != 18 {
		t.Errorf("expected minAge=18 for UK alcohol, got %d", result.MinAge)
	}
	if result.Age != 25 {
		t.Errorf("expected age=25, got %d", result.Age)
	}
}

// Test 2: Underage in US for alcohol (21 required, person is 19)
func TestVerify_USAlcohol_Underage(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-us-1",
		DateOfBirth:     dobYearsAgo(19),
		Country:         "US",
		ProductCategory: "alcohol",
	}
	result, err := v.Verify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verified {
		t.Errorf("expected verified=false for 19yo US alcohol")
	}
	if result.MinAge != 21 {
		t.Errorf("expected minAge=21 for US alcohol, got %d", result.MinAge)
	}
}

// Test 3: Japan threshold is 20 — person is exactly 20
func TestVerify_JP_Exactly20_Alcohol(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-jp-1",
		DateOfBirth:     dobYearsAgo(20),
		Country:         "JP",
		ProductCategory: "alcohol",
	}
	result, err := v.Verify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Verified {
		t.Errorf("expected verified=true for 20yo in JP (threshold 20), reason: %s", result.Reason)
	}
	if result.MinAge != 20 {
		t.Errorf("expected minAge=20 for JP, got %d", result.MinAge)
	}
}

// Test 4: Tobacco minimum is 18; person is 17 — should fail
func TestVerify_Tobacco_Underage(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-de-1",
		DateOfBirth:     dobYearsAgo(17),
		Country:         "DE",
		ProductCategory: "tobacco",
	}
	result, err := v.Verify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verified {
		t.Errorf("expected verified=false for 17yo tobacco")
	}
}

// Test 5: Gambling requires 21; person is 20 — should fail even in UK
func TestVerify_Gambling_20yo_UK(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-uk-2",
		DateOfBirth:     dobYearsAgo(20),
		Country:         "UK",
		ProductCategory: "gambling",
	}
	result, err := v.Verify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verified {
		t.Errorf("expected verified=false for 20yo gambling (min 21)")
	}
	if result.MinAge != 21 {
		t.Errorf("expected minAge=21 for gambling, got %d", result.MinAge)
	}
}

// Test 6: Invalid date format returns error
func TestVerify_InvalidDateFormat(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-x-1",
		DateOfBirth:     "25-12-1990", // DD-MM-YYYY, wrong format
		Country:         "UK",
		ProductCategory: "alcohol",
	}
	_, err := v.Verify(req)
	if err == nil {
		t.Fatal("expected error for invalid date format")
	}
}

// Test 7: Empty customerId returns validation error
func TestVerify_EmptyCustomerID(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "",
		DateOfBirth:     dobYearsAgo(25),
		Country:         "UK",
		ProductCategory: "alcohol",
	}
	_, err := v.Verify(req)
	if err == nil {
		t.Fatal("expected error for empty customerId")
	}
}

// Test 8: Edge case — birthday is exactly today (person just turned 18)
func TestVerify_ExactBirthdayToday_Passes(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-bd-1",
		DateOfBirth:     dobYearsAgo(18),
		Country:         "UK",
		ProductCategory: "alcohol",
	}
	result, err := v.Verify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Verified {
		t.Errorf("expected verified=true for person who turns 18 today, reason: %s", result.Reason)
	}
}

// Test 9: Person born today (age 0) must fail for any restricted category
func TestVerify_BornToday_Fails(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-bd-2",
		DateOfBirth:     dobExactlyToday(),
		Country:         "UK",
		ProductCategory: "alcohol",
	}
	result, err := v.Verify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verified {
		t.Errorf("expected verified=false for person born today")
	}
	if result.Age != 0 {
		t.Errorf("expected age=0, got %d", result.Age)
	}
}

// Test 10: BatchVerify returns correct results for mixed valid/invalid requests
func TestBatchVerify_MixedResults(t *testing.T) {
	reqs := []domain.VerificationRequest{
		{CustomerID: "c1", DateOfBirth: dobYearsAgo(25), Country: "UK", ProductCategory: "alcohol"},   // pass
		{CustomerID: "c2", DateOfBirth: dobYearsAgo(17), Country: "UK", ProductCategory: "alcohol"},   // fail
		{CustomerID: "c3", DateOfBirth: "bad-date", Country: "UK", ProductCategory: "alcohol"},         // error
	}
	results := v.BatchVerify(reqs)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if !results[0].Verified {
		t.Errorf("expected result[0] verified=true")
	}
	if results[1].Verified {
		t.Errorf("expected result[1] verified=false")
	}
	if results[2].Verified {
		t.Errorf("expected result[2] verified=false (invalid date)")
	}
}

// Test 11: Unknown country falls back to default min age (18)
func TestVerify_UnknownCountry_DefaultMinAge(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-zz-1",
		DateOfBirth:     dobYearsAgo(18),
		Country:         "ZZ", // not in map
		ProductCategory: "alcohol",
	}
	result, err := v.Verify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MinAge != 18 {
		t.Errorf("expected default minAge=18 for unknown country, got %d", result.MinAge)
	}
	if !result.Verified {
		t.Errorf("expected verified=true for 18yo in unknown country (default=18)")
	}
}

// Test 12: Firearms requires 21; person is 21 — should pass
func TestVerify_Firearms_Exactly21(t *testing.T) {
	req := domain.VerificationRequest{
		CustomerID:      "cust-us-2",
		DateOfBirth:     dobYearsAgo(21),
		Country:         "US",
		ProductCategory: "firearms",
	}
	result, err := v.Verify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// US country min = 21, firearms category min = 21 → effective = 21
	if result.MinAge != 21 {
		t.Errorf("expected minAge=21 for US firearms, got %d", result.MinAge)
	}
	if !result.Verified {
		t.Errorf("expected verified=true for 21yo US firearms, reason: %s", result.Reason)
	}
}

// ensure fmt is used
var _ = fmt.Sprintf
