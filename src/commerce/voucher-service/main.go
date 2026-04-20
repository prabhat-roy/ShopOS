package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/shopos/voucher-service/handler"
	"github.com/shopos/voucher-service/service"
	"github.com/shopos/voucher-service/store"
)

func main() {
	databaseURL := getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/voucher_service?sslmode=disable")
	httpPort := getenv("HTTP_PORT", "8146")

	st, err := store.New(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%s", httpPort)
	log.Printf("voucher-service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
