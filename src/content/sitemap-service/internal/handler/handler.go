package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/shopos/sitemap-service/internal/domain"
	"github.com/shopos/sitemap-service/internal/generator"
)

// Handler holds the HTTP mux, generator, and service configuration.
type Handler struct {
	gen     *generator.SitemapGenerator
	mux     *http.ServeMux
	baseURL string
}

// New wires up all routes and returns an http.Handler.
func New(gen *generator.SitemapGenerator, baseURL string) http.Handler {
	h := &Handler{gen: gen, mux: http.NewServeMux(), baseURL: baseURL}
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/sitemaps/generate", h.generateSitemap)
	h.mux.HandleFunc("/sitemaps/index", h.generateSitemapIndex)
	h.mux.HandleFunc("/sitemaps/validate", h.validateURLs)
	h.mux.HandleFunc("/robots.txt", h.robotsTxt)
	return h.mux
}

// ServeHTTP satisfies http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func writeXML(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprint(w, body)
}

// ---- route handlers ---------------------------------------------------------

// healthz returns a liveness check response.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// generateSitemap handles POST /sitemaps/generate.
// Accepts a JSON body matching GenerateRequest; returns XML sitemap.
func (h *Handler) generateSitemap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req domain.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("decode body: %s", err))
		return
	}
	if len(req.URLs) == 0 {
		writeError(w, http.StatusBadRequest, "urls must not be empty")
		return
	}

	xml := h.gen.GenerateSitemap(req.URLs)
	writeXML(w, http.StatusOK, xml)
}

// generateSitemapIndex handles POST /sitemaps/index.
// Accepts JSON body: {"sitemapUrls":["https://...","https://..."]}
// Returns XML sitemap index.
func (h *Handler) generateSitemapIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		SitemapURLs []string `json:"sitemapUrls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("decode body: %s", err))
		return
	}
	if len(body.SitemapURLs) == 0 {
		writeError(w, http.StatusBadRequest, "sitemapUrls must not be empty")
		return
	}

	xml := h.gen.GenerateSitemapIndex(body.SitemapURLs)
	writeXML(w, http.StatusOK, xml)
}

// robotsTxt handles GET /robots.txt.
// Generates robots.txt content referencing BASE_URL/sitemap.xml.
// Optional query param: disallow — comma-separated paths (e.g. /admin/,/private/)
func (h *Handler) robotsTxt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sitemapURL := strings.TrimRight(h.baseURL, "/") + "/sitemap.xml"

	var disallowPaths []string
	if raw := r.URL.Query().Get("disallow"); raw != "" {
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				disallowPaths = append(disallowPaths, p)
			}
		}
	}

	content := h.gen.GenerateRobotsTxt(sitemapURL, disallowPaths)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, content)
}

// validateURLs handles POST /sitemaps/validate.
// Accepts JSON body: {"urls":["https://...","bad-url",...]}
// Returns: {"valid":["https://..."],"invalid":["bad-url"]}
func (h *Handler) validateURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		URLs []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("decode body: %s", err))
		return
	}

	valid := []string{}
	invalid := []string{}
	for _, u := range body.URLs {
		if h.gen.ValidateURL(u) {
			valid = append(valid, u)
		} else {
			invalid = append(invalid, u)
		}
	}

	writeJSON(w, http.StatusOK, map[string][]string{
		"valid":   valid,
		"invalid": invalid,
	})
}
