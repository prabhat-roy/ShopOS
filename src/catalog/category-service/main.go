package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/shopos/category-service/handler"
	"github.com/shopos/category-service/service"
	"github.com/shopos/category-service/store"
)

func main() {
	dbURL := mustEnv("DATABASE_URL")
	httpPort := envOr("HTTP_PORT", "8111")

	st, err := store.New(dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer st.Close()
	log.Println("connected to database")

	svc := service.New(st)
	h := handler.New(svc)

	addr := fmt.Sprintf(":%s", httpPort)
	log.Printf("category-service listening on %s", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
