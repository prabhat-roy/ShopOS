package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/customs-duties-service/internal/calculator"
	"github.com/shopos/customs-duties-service/internal/domain"
)

// Calculator is the interface the handler depends on.
type Calculator interface {
	Calculate(req domain.DutyRequest) (*domain.DutyResponse, error)
	GetHSCode(code string) (*domain.HSCodeInfo, error)
	ListHSCodes() []*domain.HSCodeInfo
	GetCountryRates(country string) (*domain.CountryRates, error)
}

// Handler holds the HTTP handler state.
type Handler struct {
	calc Calculator
	mux  *http.ServeMux
}

// New creates a Handler with all routes registered.
func New(calc *calculator.DutyCalculator) *Handler {
	h := &Handler{calc: calc}
	h.mux = http.NewServeMux()
	h.registerRoutes()
	return h
}

// NewWithCalculator creates a Handler using the Calculator interface (for tests).
func NewWithCalculator(calc Calculator) *Handler {
	h := &Handler{calc: calc}
	h.mux = http.NewServeMux()
	h.registerRoutes()
	return h
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/healthz", h.handleHealthz)
	h.mux.HandleFunc("/customs/calculate", h.handleCalculate)
	h.mux.HandleFunc("/customs/hs-codes", h.handleHSCodes)
	h.mux.HandleFunc("/customs/hs-codes/", h.handleHSCodeByCode)
	h.mux.HandleFunc("/customs/countries/", h.handleCountryRates)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ─── route handlers ───────────────────────────────────────────────────────────

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleCalculate handles POST /customs/calculate.
func (h *Handler) handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req domain.DutyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid JSON body: "+err.Error())
		return
	}
	if req.FromCountry == "" || req.ToCountry == "" {
		badRequest(w, "fromCountry and toCountry are required")
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1 // default to 1 unit
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	resp, err := h.calc.Calculate(req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRequest) {
			badRequest(w, err.Error())
			return
		}
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleHSCodes handles GET /customs/hs-codes — list all codes.
func (h *Handler) handleHSCodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	codes := h.calc.ListHSCodes()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"hsCodes": codes,
		"total":   len(codes),
	})
}

// handleHSCodeByCode handles GET /customs/hs-codes/{code}.
func (h *Handler) handleHSCodeByCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	code := strings.TrimPrefix(r.URL.Path, "/customs/hs-codes/")
	if code == "" {
		badRequest(w, "hs code is required")
		return
	}
	info, err := h.calc.GetHSCode(code)
	if err != nil {
		if errors.Is(err, domain.ErrHSCodeNotFound) {
			notFound(w, "HS code not found")
			return
		}
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}

// handleCountryRates handles GET /customs/countries/{country}/rates.
func (h *Handler) handleCountryRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	// path: /customs/countries/{country}/rates
	path := strings.TrimPrefix(r.URL.Path, "/customs/countries/")
	parts := strings.SplitN(path, "/", 2)
	country := parts[0]
	if country == "" {
		badRequest(w, "country code is required")
		return
	}

	rates, err := h.calc.GetCountryRates(country)
	if err != nil {
		if errors.Is(err, domain.ErrCountryNotFound) {
			notFound(w, "country rates not found")
			return
		}
		internalError(w, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rates)
}

// ─── response helpers ─────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func notFound(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": msg})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func internalError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
}
