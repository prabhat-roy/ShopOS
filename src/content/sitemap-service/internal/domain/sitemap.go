package domain

// SitemapType classifies the kind of content described by a sitemap.
type SitemapType string

const (
	SitemapTypeProducts   SitemapType = "PRODUCTS"
	SitemapTypeCategories SitemapType = "CATEGORIES"
	SitemapTypePages      SitemapType = "PAGES"
	SitemapTypeCustom     SitemapType = "CUSTOM"
)

// ValidChangefreqs is the set of allowed changefreq values per the Sitemaps protocol.
var ValidChangefreqs = map[string]bool{
	"always":  true,
	"hourly":  true,
	"daily":   true,
	"weekly":  true,
	"monthly": true,
	"yearly":  true,
	"never":   true,
}

// SitemapURL represents a single URL entry in a sitemap.
type SitemapURL struct {
	Loc        string  `json:"loc"`
	Lastmod    string  `json:"lastmod,omitempty"`
	Changefreq string  `json:"changefreq,omitempty"`
	Priority   float64 `json:"priority,omitempty"`
}

// SitemapEntry is a reference entry inside a sitemap index.
type SitemapEntry struct {
	Loc     string `json:"loc"`
	Lastmod string `json:"lastmod,omitempty"`
}

// SitemapIndex groups multiple SitemapEntry values into an index document.
type SitemapIndex struct {
	Sitemaps []SitemapEntry `json:"sitemaps"`
}

// GenerateRequest carries the input data for a sitemap generation request.
type GenerateRequest struct {
	URLs []SitemapURL `json:"urls"`
	Name string       `json:"name,omitempty"`
	Type SitemapType  `json:"type,omitempty"`
}
