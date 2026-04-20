package generator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/shopos/sitemap-service/internal/domain"
)

const (
	xmlHeader      = `<?xml version="1.0" encoding="UTF-8"?>`
	urlsetOpen     = `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
	urlsetClose    = `</urlset>`
	sitemapIdxOpen = `<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
	sitemapIdxClose = `</sitemapindex>`
)

// SitemapGenerator generates XML sitemaps, sitemap indexes, and robots.txt content.
type SitemapGenerator struct {
	maxURLs int
}

// New creates a SitemapGenerator with the given per-sitemap URL limit.
func New(maxURLs int) *SitemapGenerator {
	if maxURLs <= 0 {
		maxURLs = 50000
	}
	return &SitemapGenerator{maxURLs: maxURLs}
}

// GenerateSitemap produces a valid XML sitemap string from the given URLs.
// If the URL count exceeds maxURLs, only the first maxURLs entries are included.
// Callers that need to split large URL sets should call GenerateSitemapChunks
// and wrap the results in a sitemap index.
func (g *SitemapGenerator) GenerateSitemap(urls []domain.SitemapURL) string {
	chunk := urls
	if len(chunk) > g.maxURLs {
		chunk = chunk[:g.maxURLs]
	}
	return g.buildURLSet(chunk)
}

// GenerateSitemapChunks splits a URL slice into chunks of at most maxURLs and
// returns one XML sitemap string per chunk.
func (g *SitemapGenerator) GenerateSitemapChunks(urls []domain.SitemapURL) []string {
	var chunks []string
	for i := 0; i < len(urls); i += g.maxURLs {
		end := i + g.maxURLs
		if end > len(urls) {
			end = len(urls)
		}
		chunks = append(chunks, g.buildURLSet(urls[i:end]))
	}
	if chunks == nil {
		chunks = []string{g.buildURLSet(nil)}
	}
	return chunks
}

// buildURLSet assembles the XML urlset document for a slice of URLs.
func (g *SitemapGenerator) buildURLSet(urls []domain.SitemapURL) string {
	var sb strings.Builder
	sb.WriteString(xmlHeader)
	sb.WriteByte('\n')
	sb.WriteString(urlsetOpen)
	sb.WriteByte('\n')

	for _, u := range urls {
		sb.WriteString("  <url>\n")
		sb.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", escapeXML(u.Loc)))
		if u.Lastmod != "" {
			sb.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", escapeXML(u.Lastmod)))
		}
		if u.Changefreq != "" && domain.ValidChangefreqs[u.Changefreq] {
			sb.WriteString(fmt.Sprintf("    <changefreq>%s</changefreq>\n", u.Changefreq))
		}
		if u.Priority > 0 {
			priority := u.Priority
			if priority > 1.0 {
				priority = 1.0
			}
			sb.WriteString(fmt.Sprintf("    <priority>%.1f</priority>\n", priority))
		}
		sb.WriteString("  </url>\n")
	}

	sb.WriteString(urlsetClose)
	return sb.String()
}

// GenerateSitemapIndex produces a <sitemapindex> XML document from a list of
// sitemap URLs (e.g., https://example.com/sitemap-products.xml).
func (g *SitemapGenerator) GenerateSitemapIndex(sitemapURLs []string) string {
	var sb strings.Builder
	sb.WriteString(xmlHeader)
	sb.WriteByte('\n')
	sb.WriteString(sitemapIdxOpen)
	sb.WriteByte('\n')

	for _, u := range sitemapURLs {
		sb.WriteString("  <sitemap>\n")
		sb.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", escapeXML(u)))
		sb.WriteString("  </sitemap>\n")
	}

	sb.WriteString(sitemapIdxClose)
	return sb.String()
}

// GenerateRobotsTxt produces a robots.txt file that references the given
// sitemap URL and lists the given disallow paths for all user-agents.
func (g *SitemapGenerator) GenerateRobotsTxt(sitemapURL string, disallowPaths []string) string {
	var sb strings.Builder
	sb.WriteString("User-agent: *\n")
	for _, p := range disallowPaths {
		sb.WriteString(fmt.Sprintf("Disallow: %s\n", p))
	}
	if len(disallowPaths) == 0 {
		sb.WriteString("Disallow:\n")
	}
	sb.WriteByte('\n')
	sb.WriteString(fmt.Sprintf("Sitemap: %s\n", sitemapURL))
	return sb.String()
}

// ValidateURL reports whether rawURL is a well-formed absolute HTTP/HTTPS URL.
func (g *SitemapGenerator) ValidateURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

// escapeXML replaces XML special characters with their entity equivalents.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
