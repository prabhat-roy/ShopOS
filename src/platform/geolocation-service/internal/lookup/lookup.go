package lookup

import (
	"net"

	"github.com/shopos/geolocation-service/internal/domain"
)

// entry pairs a CIDR network block with the Location data for that block.
type entry struct {
	network *net.IPNet
	loc     domain.Location
}

// Lookup performs in-memory CIDR-based IP geolocation.
type Lookup struct {
	entries []entry
}

// New constructs a Lookup pre-seeded with representative CIDR blocks across multiple continents.
func New() *Lookup {
	seeds := []struct {
		cidr string
		loc  domain.Location
	}{
		// United States — West Coast
		{
			cidr: "8.0.0.0/8",
			loc: domain.Location{
				CountryCode: "US", CountryName: "United States",
				Region: "California", City: "Mountain View",
				PostalCode: "94043", Timezone: "America/Los_Angeles",
				ISP:       "Google LLC",
				Latitude:  37.3861, Longitude: -122.0839,
			},
		},
		// United States — East Coast
		{
			cidr: "64.0.0.0/8",
			loc: domain.Location{
				CountryCode: "US", CountryName: "United States",
				Region: "New York", City: "New York City",
				PostalCode: "10001", Timezone: "America/New_York",
				ISP:       "Verizon Communications",
				Latitude:  40.7128, Longitude: -74.0060,
			},
		},
		// United Kingdom
		{
			cidr: "51.0.0.0/8",
			loc: domain.Location{
				CountryCode: "GB", CountryName: "United Kingdom",
				Region: "England", City: "London",
				PostalCode: "EC1A 1BB", Timezone: "Europe/London",
				ISP:       "British Telecom",
				Latitude:  51.5074, Longitude: -0.1278,
			},
		},
		// Germany
		{
			cidr: "85.0.0.0/8",
			loc: domain.Location{
				CountryCode: "DE", CountryName: "Germany",
				Region: "Berlin", City: "Berlin",
				PostalCode: "10115", Timezone: "Europe/Berlin",
				ISP:       "Deutsche Telekom AG",
				Latitude:  52.5200, Longitude: 13.4050,
			},
		},
		// India
		{
			cidr: "117.0.0.0/8",
			loc: domain.Location{
				CountryCode: "IN", CountryName: "India",
				Region: "Maharashtra", City: "Mumbai",
				PostalCode: "400001", Timezone: "Asia/Kolkata",
				ISP:       "Reliance Jio",
				Latitude:  19.0760, Longitude: 72.8777,
			},
		},
		// Japan
		{
			cidr: "202.0.0.0/8",
			loc: domain.Location{
				CountryCode: "JP", CountryName: "Japan",
				Region: "Tokyo", City: "Tokyo",
				PostalCode: "100-0001", Timezone: "Asia/Tokyo",
				ISP:       "NTT Communications",
				Latitude:  35.6895, Longitude: 139.6917,
			},
		},
		// Canada
		{
			cidr: "24.0.0.0/8",
			loc: domain.Location{
				CountryCode: "CA", CountryName: "Canada",
				Region: "Ontario", City: "Toronto",
				PostalCode: "M5H 2N2", Timezone: "America/Toronto",
				ISP:       "Rogers Communications",
				Latitude:  43.6532, Longitude: -79.3832,
			},
		},
		// Australia
		{
			cidr: "203.0.0.0/8",
			loc: domain.Location{
				CountryCode: "AU", CountryName: "Australia",
				Region: "New South Wales", City: "Sydney",
				PostalCode: "2000", Timezone: "Australia/Sydney",
				ISP:       "Telstra Corporation",
				Latitude:  -33.8688, Longitude: 151.2093,
			},
		},
		// France
		{
			cidr: "90.0.0.0/8",
			loc: domain.Location{
				CountryCode: "FR", CountryName: "France",
				Region: "Île-de-France", City: "Paris",
				PostalCode: "75001", Timezone: "Europe/Paris",
				ISP:       "Orange S.A.",
				Latitude:  48.8566, Longitude: 2.3522,
			},
		},
		// Singapore
		{
			cidr: "175.0.0.0/8",
			loc: domain.Location{
				CountryCode: "SG", CountryName: "Singapore",
				Region: "Singapore", City: "Singapore",
				PostalCode: "018989", Timezone: "Asia/Singapore",
				ISP:       "Singtel",
				Latitude:  1.3521, Longitude: 103.8198,
			},
		},
		// Brazil
		{
			cidr: "177.0.0.0/8",
			loc: domain.Location{
				CountryCode: "BR", CountryName: "Brazil",
				Region: "São Paulo", City: "São Paulo",
				PostalCode: "01310-100", Timezone: "America/Sao_Paulo",
				ISP:       "Claro Brasil",
				Latitude:  -23.5505, Longitude: -46.6333,
			},
		},
		// Netherlands
		{
			cidr: "80.0.0.0/8",
			loc: domain.Location{
				CountryCode: "NL", CountryName: "Netherlands",
				Region: "North Holland", City: "Amsterdam",
				PostalCode: "1012 AB", Timezone: "Europe/Amsterdam",
				ISP:       "KPN",
				Latitude:  52.3676, Longitude: 4.9041,
			},
		},
		// South Africa
		{
			cidr: "196.0.0.0/8",
			loc: domain.Location{
				CountryCode: "ZA", CountryName: "South Africa",
				Region: "Gauteng", City: "Johannesburg",
				PostalCode: "2000", Timezone: "Africa/Johannesburg",
				ISP:       "MTN Group",
				Latitude:  -26.2041, Longitude: 28.0473,
			},
		},
		// United Arab Emirates
		{
			cidr: "194.0.0.0/8",
			loc: domain.Location{
				CountryCode: "AE", CountryName: "United Arab Emirates",
				Region: "Dubai", City: "Dubai",
				PostalCode: "00000", Timezone: "Asia/Dubai",
				ISP:       "Etisalat",
				Latitude:  25.2048, Longitude: 55.2708,
			},
		},
		// South Korea
		{
			cidr: "211.0.0.0/8",
			loc: domain.Location{
				CountryCode: "KR", CountryName: "South Korea",
				Region: "Seoul", City: "Seoul",
				PostalCode: "03000", Timezone: "Asia/Seoul",
				ISP:       "SK Broadband",
				Latitude:  37.5665, Longitude: 126.9780,
			},
		},
	}

	entries := make([]entry, 0, len(seeds))
	for _, s := range seeds {
		_, network, err := net.ParseCIDR(s.cidr)
		if err != nil {
			// Seeds are compile-time constants; a parse error is a programming mistake.
			panic("geolocation-service: invalid seed CIDR " + s.cidr + ": " + err.Error())
		}
		entries = append(entries, entry{network: network, loc: s.loc})
	}

	return &Lookup{entries: entries}
}

// Resolve looks up geolocation data for a single IP address string.
// It returns ErrInvalidIP if ip cannot be parsed, or ErrNotFound if no CIDR
// block in the seed data matches.
func (l *Lookup) Resolve(ip string) (*domain.Location, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil, domain.ErrInvalidIP
	}

	for _, e := range l.entries {
		if e.network.Contains(parsed) {
			result := e.loc
			result.IP = ip
			return &result, nil
		}
	}

	return nil, domain.ErrNotFound
}

// ResolveMany resolves a slice of IP addresses. Entries that fail to resolve are
// omitted from the result; the returned slice may therefore be shorter than ips.
func (l *Lookup) ResolveMany(ips []string) ([]*domain.Location, error) {
	results := make([]*domain.Location, 0, len(ips))
	for _, ip := range ips {
		loc, err := l.Resolve(ip)
		if err != nil {
			// Skip unresolvable entries; partial results are valid.
			continue
		}
		results = append(results, loc)
	}
	return results, nil
}
