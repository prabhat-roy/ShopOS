package lookup

import (
	"errors"
	"testing"

	"github.com/shopos/geolocation-service/internal/domain"
)

func TestResolve_KnownIP_US(t *testing.T) {
	l := New()

	// 8.8.8.8 falls inside 8.0.0.0/8 — should resolve to US / Mountain View.
	loc, err := l.Resolve("8.8.8.8")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if loc.CountryCode != "US" {
		t.Errorf("expected country code US, got %q", loc.CountryCode)
	}
	if loc.City != "Mountain View" {
		t.Errorf("expected city Mountain View, got %q", loc.City)
	}
	if loc.Timezone != "America/Los_Angeles" {
		t.Errorf("expected timezone America/Los_Angeles, got %q", loc.Timezone)
	}
	if loc.IP != "8.8.8.8" {
		t.Errorf("expected IP field 8.8.8.8, got %q", loc.IP)
	}
}

func TestResolve_KnownIP_UK(t *testing.T) {
	l := New()

	// 51.100.200.1 falls inside 51.0.0.0/8 — should resolve to GB / London.
	loc, err := l.Resolve("51.100.200.1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if loc.CountryCode != "GB" {
		t.Errorf("expected country code GB, got %q", loc.CountryCode)
	}
	if loc.City != "London" {
		t.Errorf("expected city London, got %q", loc.City)
	}
}

func TestResolve_KnownIP_IN(t *testing.T) {
	l := New()

	// 117.55.66.77 falls inside 117.0.0.0/8 — should resolve to IN / Mumbai.
	loc, err := l.Resolve("117.55.66.77")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if loc.CountryCode != "IN" {
		t.Errorf("expected country code IN, got %q", loc.CountryCode)
	}
}

func TestResolve_InvalidIP(t *testing.T) {
	l := New()

	_, err := l.Resolve("not-an-ip")
	if !errors.Is(err, domain.ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP, got %v", err)
	}
}

func TestResolve_EmptyString(t *testing.T) {
	l := New()

	_, err := l.Resolve("")
	if !errors.Is(err, domain.ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP for empty string, got %v", err)
	}
}

func TestResolve_UnknownIP(t *testing.T) {
	l := New()

	// 10.0.0.1 is RFC-1918 private space; not in any seeded CIDR.
	_, err := l.Resolve("10.0.0.1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestResolve_UnknownIP_Loopback(t *testing.T) {
	l := New()

	_, err := l.Resolve("127.0.0.1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for loopback, got %v", err)
	}
}

func TestResolveMany_PartialResults(t *testing.T) {
	l := New()

	ips := []string{
		"8.8.8.8",    // US — should resolve
		"10.0.0.1",   // private — not found, skipped
		"not-an-ip",  // invalid — skipped
		"202.10.5.6", // JP — should resolve
	}

	results, err := l.ResolveMany(ips)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 resolved results, got %d", len(results))
	}
	if results[0].CountryCode != "US" {
		t.Errorf("first result: expected US, got %q", results[0].CountryCode)
	}
	if results[1].CountryCode != "JP" {
		t.Errorf("second result: expected JP, got %q", results[1].CountryCode)
	}
}

func TestResolveMany_AllInvalid(t *testing.T) {
	l := New()

	results, err := l.ResolveMany([]string{"bad", "also-bad"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestResolveMany_Empty(t *testing.T) {
	l := New()

	results, err := l.ResolveMany([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestResolve_IPFieldPopulated(t *testing.T) {
	l := New()

	// Verify the IP field on the returned location equals the queried address.
	ip := "64.10.20.30"
	loc, err := l.Resolve(ip)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc.IP != ip {
		t.Errorf("expected IP field %q, got %q", ip, loc.IP)
	}
}
