package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shopos/geolocation-service/internal/config"
	"github.com/shopos/geolocation-service/internal/handler"
	"github.com/shopos/geolocation-service/internal/lookup"
)

func main() {
	cfg := config.Load()

	lkp := lookup.New()
	h := handler.New(lkp)

	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	log.Printf("geolocation-service starting — HTTP addr=%s gRPC port=%s", addr, cfg.GRPCPort)

	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
