package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/shopos/voucher-service/domain"
	"github.com/shopos/voucher-service/service"
)

// Handler holds HTTP route logic for the voucher-service.
type Handler struct {
	svc *service.VoucherService
}

// New creates a Handler.
func New(svc *service.VoucherService) *Handler { return &Handler{svc: svc} }

// RegisterRoutes wires all routes onto mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.Healthz)
	mux.HandleFunc("/vouchers", h.vouchersCollection)
	mux.HandleFunc("/vouchers/", h.vouchersResource)
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// vouchersCollection handles POST /vouchers (issue).
func (h *Handler) vouchersCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.issueVoucher(w, r)
}

// vouchersResource routes sub-paths.
func (h *Handler) vouchersResource(w http.ResponseWriter, r *http.Request) {
	// Strip "/vouchers/"
	path := strings.TrimPrefix(r.URL.Path, "/vouchers/")
	parts := strings.SplitN(path, "/", 3)

	// /vouchers/customer/{customerID}
	if parts[0] == "customer" && len(parts) == 2 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.listByCustomer(w, r, parts[1])
		return
	}

	code := parts[0]
	if code == "" {
		http.NotFound(w, r)
		return
	}

	// /vouchers/{code}
	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			h.getVoucher(w, r, code)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// /vouchers/{code}/use
	if parts[1] == "use" && r.Method == http.MethodPost {
		h.useVoucher(w, r, code)
		return
	}

	http.NotFound(w, r)
}

func (h *Handler) issueVoucher(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID string    `json:"customer_id"`
		Amount     float64   `json:"amount"`
		Currency   string    `json:"currency"`
		ExpiresAt  time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.CustomerID == "" {
		writeError(w, http.StatusBadRequest, "customer_id is required")
		return
	}
	if body.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if body.ExpiresAt.IsZero() {
		body.ExpiresAt = time.Now().UTC().AddDate(1, 0, 0) // default 1 year
	}
	v, err := h.svc.IssueVoucher(r.Context(), body.CustomerID, body.Amount, body.Currency, body.ExpiresAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

func (h *Handler) getVoucher(w http.ResponseWriter, r *http.Request, code string) {
	v, err := h.svc.GetVoucher(r.Context(), code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "voucher not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) useVoucher(w http.ResponseWriter, r *http.Request, code string) {
	var body struct {
		OrderID string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.OrderID == "" {
		writeError(w, http.StatusBadRequest, "order_id is required")
		return
	}
	v, err := h.svc.UseVoucher(r.Context(), code, body.OrderID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, http.StatusNotFound, "voucher not found")
		case errors.Is(err, domain.ErrAlreadyUsed):
			writeError(w, http.StatusConflict, "voucher has already been used")
		case errors.Is(err, domain.ErrExpired):
			writeError(w, http.StatusUnprocessableEntity, "voucher has expired")
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) listByCustomer(w http.ResponseWriter, r *http.Request, customerID string) {
	vouchers, err := h.svc.ListCustomerVouchers(r.Context(), customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if vouchers == nil {
		vouchers = []*domain.Voucher{}
	}
	writeJSON(w, http.StatusOK, vouchers)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
