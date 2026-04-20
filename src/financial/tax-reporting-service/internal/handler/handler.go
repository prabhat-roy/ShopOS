package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopos/tax-reporting-service/internal/domain"
	"github.com/shopos/tax-reporting-service/internal/service"
)

// Handler wires HTTP routes to the service layer.
type Handler struct {
	svc service.Servicer
	mux *http.ServeMux
}

// New creates and registers all routes.
func New(svc service.Servicer) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/tax-records", h.taxRecordsCollection)
	h.mux.HandleFunc("/tax-records/", h.taxRecordItem)
	h.mux.HandleFunc("/tax-summary", h.taxSummary)
	h.mux.HandleFunc("/tax-reports/generate", h.generateReport)
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// GET /healthz
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// /tax-records  (POST + GET with query params)
func (h *Handler) taxRecordsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTaxRecord(w, r)
	case http.MethodGet:
		h.listTaxRecords(w, r)
	default:
		methodNotAllowed(w)
	}
}

// POST /tax-records — creates a new tax record, returns 201.
func (h *Handler) createTaxRecord(w http.ResponseWriter, r *http.Request) {
	var rec domain.TaxRecord
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	created, err := h.svc.RecordTax(&rec)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// GET /tax-records?jurisdiction=&taxType=&start=&end=&limit=
func (h *Handler) listTaxRecords(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	f := domain.ListFilter{
		Jurisdiction: q.Get("jurisdiction"),
		TaxType:      q.Get("taxType"),
		StartDate:    q.Get("start"),
		EndDate:      q.Get("end"),
		Limit:        limit,
	}
	records, err := h.svc.ListTaxRecords(f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if records == nil {
		records = []*domain.TaxRecord{}
	}
	writeJSON(w, http.StatusOK, records)
}

// /tax-records/{id}  (GET)
func (h *Handler) taxRecordItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/tax-records/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	rec, err := h.svc.GetTaxRecord(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "tax record not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

// GET /tax-summary?jurisdiction=&period=YYYY-MM
func (h *Handler) taxSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	q := r.URL.Query()
	jurisdiction := q.Get("jurisdiction")
	period := q.Get("period")
	if period == "" {
		writeError(w, http.StatusBadRequest, "period query parameter is required (YYYY-MM)")
		return
	}
	summaries, err := h.svc.GetTaxSummary(jurisdiction, period)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if summaries == nil {
		summaries = []*domain.TaxSummary{}
	}
	writeJSON(w, http.StatusOK, summaries)
}

// POST /tax-reports/generate
func (h *Handler) generateReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req domain.GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	summaries, err := h.svc.GenerateReport(req.StartDate, req.EndDate, req.Jurisdiction)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if summaries == nil {
		summaries = []*domain.TaxSummary{}
	}
	writeJSON(w, http.StatusOK, summaries)
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
