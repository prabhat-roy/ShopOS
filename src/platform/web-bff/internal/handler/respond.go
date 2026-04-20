package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopos/web-bff/internal/service"
)

func JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrUnauthorized):
		JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidInput):
		JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrNotImplemented):
		JSON(w, http.StatusNotImplemented, map[string]string{"error": "service not yet available"})
	default:
		JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func userID(r *http.Request) string {
	return r.Header.Get("X-User-ID")
}
