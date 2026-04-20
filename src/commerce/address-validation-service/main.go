package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/shopos/address-validation-service/handler"
)

func main() {
	httpPort := getenv("HTTP_PORT", "8145")

	h := handler.New()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%s", httpPort)
	log.Printf("address-validation-service listening on %s", addr)
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
