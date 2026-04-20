package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/web-bff/internal/config"
	"github.com/shopos/web-bff/internal/router"
)

func testConfig() *config.Config {
	return &config.Config{
		Port:             "8081",
		Env:              "development",
		GRPCTimeout:      5 * time.Second,
		AuthServiceAddr:  "localhost:50060",
		UserServiceAddr:  "localhost:50061",
		CatalogAddr:      "localhost:50070",
		InventoryAddr:    "localhost:50074",
		CartServiceAddr:  "localhost:50080",
		OrderServiceAddr: "localhost:50082",
		PricingAddr:      "localhost:50073",
		SearchAddr:       "localhost:50078",
	}
}

func TestHealth_Returns200(t *testing.T) {
	h := router.New(testConfig())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]string
	json.NewDecoder(rr.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %s", body["status"])
	}
}

func TestProductRoutes_ReturnJSON(t *testing.T) {
	h := router.New(testConfig())
	routes := []string{"/products", "/categories"}

	for _, path := range routes {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("path %s: expected application/json, got %s", path, ct)
		}
	}
}

func TestSearchRoute_MissingQuery_Returns400(t *testing.T) {
	h := router.New(testConfig())
	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestUnknownRoute_Returns404(t *testing.T) {
	h := router.New(testConfig())
	req := httptest.NewRequest(http.MethodGet, "/unknown-path", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
