package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/shopos/subscription-billing-service/domain"
	"github.com/shopos/subscription-billing-service/service"
)

// Svc is the interface the handler depends on (enables mocking in tests).
type Svc interface {
	Subscribe(req service.SubscribeRequest) (*domain.Subscription, error)
	GetSubscription(id string) (*domain.Subscription, error)
	ListSubscriptions(customerID string) ([]*domain.Subscription, error)
	Cancel(id string) error
	Pause(id string) error
	Resume(id string) error
	ListBillingRecords(subID string) ([]*domain.BillingRecord, error)
}

// Handler holds dependencies for the HTTP layer.
type Handler struct {
	svc Svc
}

// New constructs an HTTP handler.
func New(svc Svc) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires up all HTTP routes onto mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/subscriptions", h.subscriptions)
	mux.HandleFunc("/subscriptions/", h.subscriptionByID)
}

// -------------------------------------------------------------------------
// Route: GET /healthz
// -------------------------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// -------------------------------------------------------------------------
// Route: /subscriptions  (no trailing ID segment)
// -------------------------------------------------------------------------

func (h *Handler) subscriptions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createSubscription(w, r)
	case http.MethodGet:
		h.listSubscriptions(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// POST /subscriptions
func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID string              `json:"customer_id"`
		PlanID     string              `json:"plan_id"`
		ProductID  string              `json:"product_id"`
		Cycle      domain.BillingCycle `json:"cycle"`
		Price      float64             `json:"price"`
		Currency   string              `json:"currency"`
		TrialDays  int                 `json:"trial_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.CustomerID == "" || body.PlanID == "" {
		writeError(w, http.StatusBadRequest, "customer_id and plan_id are required")
		return
	}
	if body.Currency == "" {
		body.Currency = "USD"
	}

	req := service.SubscribeRequest{
		CustomerID: body.CustomerID,
		PlanID:     body.PlanID,
		ProductID:  body.ProductID,
		Cycle:      body.Cycle,
		Price:      body.Price,
		Currency:   body.Currency,
		TrialDays:  body.TrialDays,
	}

	sub, err := h.svc.Subscribe(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, subToResponse(sub))
}

// GET /subscriptions?customer_id=
func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	if customerID == "" {
		writeError(w, http.StatusBadRequest, "customer_id query param is required")
		return
	}
	subs, err := h.svc.ListSubscriptions(customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]subscriptionResponse, 0, len(subs))
	for _, s := range subs {
		resp = append(resp, subToResponse(s))
	}
	writeJSON(w, http.StatusOK, resp)
}

// -------------------------------------------------------------------------
// Route: /subscriptions/{id}[/action]
// -------------------------------------------------------------------------

func (h *Handler) subscriptionByID(w http.ResponseWriter, r *http.Request) {
	// path:  /subscriptions/{id}
	//        /subscriptions/{id}/cancel
	//        /subscriptions/{id}/pause
	//        /subscriptions/{id}/resume
	//        /subscriptions/{id}/billing
	parts := splitPath(r.URL.Path) // ["subscriptions", "{id}", optional...]
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "missing subscription id")
		return
	}
	id := parts[1]

	if len(parts) == 2 {
		// /subscriptions/{id}
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getSubscription(w, r, id)
		return
	}

	action := parts[2]
	switch action {
	case "cancel":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.cancelSubscription(w, r, id)
	case "pause":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.pauseSubscription(w, r, id)
	case "resume":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.resumeSubscription(w, r, id)
	case "billing":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.listBilling(w, r, id)
	default:
		writeError(w, http.StatusNotFound, "unknown action")
	}
}

func (h *Handler) getSubscription(w http.ResponseWriter, _ *http.Request, id string) {
	sub, err := h.svc.GetSubscription(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, subToResponse(sub))
}

func (h *Handler) cancelSubscription(w http.ResponseWriter, _ *http.Request, id string) {
	if err := h.svc.Cancel(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) pauseSubscription(w http.ResponseWriter, _ *http.Request, id string) {
	if err := h.svc.Pause(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) resumeSubscription(w http.ResponseWriter, _ *http.Request, id string) {
	if err := h.svc.Resume(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listBilling(w http.ResponseWriter, _ *http.Request, id string) {
	records, err := h.svc.ListBillingRecords(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]billingResponse, 0, len(records))
	for _, rec := range records {
		resp = append(resp, billingToResponse(rec))
	}
	writeJSON(w, http.StatusOK, resp)
}

// -------------------------------------------------------------------------
// Response types
// -------------------------------------------------------------------------

type subscriptionResponse struct {
	ID            string      `json:"id"`
	CustomerID    string      `json:"customer_id"`
	PlanID        string      `json:"plan_id"`
	ProductID     string      `json:"product_id"`
	Status        string      `json:"status"`
	Cycle         string      `json:"cycle"`
	Price         float64     `json:"price"`
	Currency      string      `json:"currency"`
	TrialEndsAt   *time.Time  `json:"trial_ends_at,omitempty"`
	NextBillingAt time.Time   `json:"next_billing_at"`
	StartedAt     time.Time   `json:"started_at"`
	CancelledAt   *time.Time  `json:"cancelled_at,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

type billingResponse struct {
	ID             string    `json:"id"`
	SubscriptionID string    `json:"subscription_id"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

func subToResponse(s *domain.Subscription) subscriptionResponse {
	return subscriptionResponse{
		ID:            s.ID,
		CustomerID:    s.CustomerID,
		PlanID:        s.PlanID,
		ProductID:     s.ProductID,
		Status:        string(s.Status),
		Cycle:         string(s.Cycle),
		Price:         s.Price,
		Currency:      s.Currency,
		TrialEndsAt:   s.TrialEndsAt,
		NextBillingAt: s.NextBillingAt,
		StartedAt:     s.StartedAt,
		CancelledAt:   s.CancelledAt,
		CreatedAt:     s.CreatedAt,
	}
}

func billingToResponse(r *domain.BillingRecord) billingResponse {
	return billingResponse{
		ID:             r.ID,
		SubscriptionID: r.SubscriptionID,
		Amount:         r.Amount,
		Currency:       r.Currency,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
	}
}

// -------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// splitPath removes empty segments and the leading slash, returning path parts.
// e.g. "/subscriptions/abc-123/cancel" → ["subscriptions","abc-123","cancel"]
func splitPath(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	filtered := parts[:0]
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
