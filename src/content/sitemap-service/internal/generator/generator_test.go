package generator_test

import (
	"strings"
	"testing"

	"github.com/shopos/sitemap-service/internal/domain"
	"github.com/shopos/sitemap-service/internal/generator"
)

func makeURLs(n int) []domain.SitemapURL {
	urls := make([]domain.SitemapURL, n)
	for i := 0; i < n; i++ {
		urls[i] = domain.SitemapURL{
			Loc:        "https://example.com/page/" + string(rune('a'+i%26)),
			Changefreq: "daily",
			Priority:   0.5,
		}
	}
	return urls
}

// Test 1: GenerateSitemap produces well-formed XML with the required XML declaration.
func TestGenerateSitemapXMLHeader(t *testing.T) {
	g := generator.New(50000)
	urls := []domain.SitemapURL{{Loc: "https://example.com/page1"}}
	result := g.GenerateSitemap(urls)

	if !strings.HasPrefix(result, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Fatalf("missing XML declaration, got: %s", result[:min(len(result), 80)])
	}
}

// Test 2: GenerateSitemap output contains the required Sitemaps namespace.
func TestGenerateSitemapNamespace(t *testing.T) {
	g := generator.New(50000)
	urls := []domain.SitemapURL{{Loc: "https://example.com/"}}
	result := g.GenerateSitemap(urls)

	if !strings.Contains(result, "http://www.sitemaps.org/schemas/sitemap/0.9") {
		t.Fatal("sitemap namespace missing from output")
	}
}

// Test 3: GenerateSitemap output contains the provided URL.
func TestGenerateSitemapContainsURL(t *testing.T) {
	g := generator.New(50000)
	urls := []domain.SitemapURL{{Loc: "https://example.com/products/shoe"}}
	result := g.GenerateSitemap(urls)

	if !strings.Contains(result, "https://example.com/products/shoe") {
		t.Fatalf("expected URL in output, got: %s", result)
	}
}

// Test 4: GenerateSitemap respects maxURLs — output is capped at the limit.
func TestGenerateSitemapMaxURLsCap(t *testing.T) {
	limit := 10
	g := generator.New(limit)
	urls := makeURLs(20)
	result := g.GenerateSitemap(urls)

	count := strings.Count(result, "<url>")
	if count != limit {
		t.Fatalf("expected %d <url> entries, got %d", limit, count)
	}
}

// Test 5: GenerateSitemapChunks splits at exactly 50000 URLs per chunk.
func TestGenerateSitemapChunksSplit(t *testing.T) {
	limit := 50000
	g := generator.New(limit)
	urls := makeURLs(75000)
	chunks := g.GenerateSitemapChunks(urls)

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	first := strings.Count(chunks[0], "<url>")
	second := strings.Count(chunks[1], "<url>")
	if first != 50000 {
		t.Fatalf("first chunk: expected 50000 URLs, got %d", first)
	}
	if second != 25000 {
		t.Fatalf("second chunk: expected 25000 URLs, got %d", second)
	}
}

// Test 6: GenerateSitemapIndex produces valid <sitemapindex> XML.
func TestGenerateSitemapIndexFormat(t *testing.T) {
	g := generator.New(50000)
	urls := []string{
		"https://example.com/sitemap-products.xml",
		"https://example.com/sitemap-categories.xml",
	}
	result := g.GenerateSitemapIndex(urls)

	if !strings.Contains(result, "<sitemapindex") {
		t.Fatal("missing <sitemapindex> tag")
	}
	if !strings.Contains(result, "http://www.sitemaps.org/schemas/sitemap/0.9") {
		t.Fatal("missing namespace in sitemap index")
	}
	if strings.Count(result, "<sitemap>") != 2 {
		t.Fatalf("expected 2 <sitemap> entries, got %d", strings.Count(result, "<sitemap>"))
	}
	for _, u := range urls {
		if !strings.Contains(result, u) {
			t.Fatalf("URL %q missing from sitemap index", u)
		}
	}
}

// Test 7: GenerateRobotsTxt produces correctly formatted robots.txt.
func TestGenerateRobotsTxt(t *testing.T) {
	g := generator.New(50000)
	result := g.GenerateRobotsTxt("https://example.com/sitemap.xml", []string{"/admin/", "/private/"})

	if !strings.Contains(result, "User-agent: *") {
		t.Fatal("missing User-agent directive")
	}
	if !strings.Contains(result, "Disallow: /admin/") {
		t.Fatal("missing /admin/ disallow")
	}
	if !strings.Contains(result, "Sitemap: https://example.com/sitemap.xml") {
		t.Fatal("missing Sitemap directive")
	}
}

// Test 8: ValidateURL rejects relative and malformed URLs.
func TestValidateURLInvalid(t *testing.T) {
	g := generator.New(50000)
	cases := []string{
		"",
		"/relative/path",
		"not-a-url",
		"ftp://example.com/file",
		"http://",
	}
	for _, u := range cases {
		if g.ValidateURL(u) {
			t.Errorf("expected invalid for %q, but got valid", u)
		}
	}
}

// Test 9: Changefreq values outside the allowed set are omitted from output.
func TestChangefreqValidation(t *testing.T) {
	g := generator.New(50000)
	urls := []domain.SitemapURL{
		{Loc: "https://example.com/a", Changefreq: "invalid-freq"},
		{Loc: "https://example.com/b", Changefreq: "daily"},
	}
	result := g.GenerateSitemap(urls)

	if strings.Contains(result, "invalid-freq") {
		t.Fatal("invalid changefreq should not appear in output")
	}
	if !strings.Contains(result, "<changefreq>daily</changefreq>") {
		t.Fatal("valid changefreq 'daily' missing from output")
	}
}

// Test 10: Priority values above 1.0 are clamped to 1.0 in the output.
func TestPriorityBoundsClamped(t *testing.T) {
	g := generator.New(50000)
	urls := []domain.SitemapURL{
		{Loc: "https://example.com/high", Priority: 1.5},
		{Loc: "https://example.com/normal", Priority: 0.8},
	}
	result := g.GenerateSitemap(urls)

	if strings.Contains(result, "1.5") {
		t.Fatal("priority 1.5 should be clamped; must not appear in output")
	}
	if !strings.Contains(result, "<priority>1.0</priority>") {
		t.Fatal("clamped priority should be 1.0")
	}
	if !strings.Contains(result, "<priority>0.8</priority>") {
		t.Fatal("priority 0.8 missing from output")
	}
}

