package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopos/subscription-billing-service/handler"
	"github.com/shopos/subscription-billing-service/service"
	"github.com/shopos/subscription-billing-service/store"
)

func main() {
	httpPort := getEnv("HTTP_PORT", "8135")
	dsn      := buildDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	pgStore := store.NewPostgresStore(db)
	svc     := service.New(pgStore)
	h       := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("subscription-billing-service listening on :%s", httpPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	log.Println("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}

func buildDSN() string {
	url := os.Getenv("DATABASE_URL")
	if url != "" {
		return url
	}
	host     := getEnv("DB_HOST", "localhost")
	port     := getEnv("DB_PORT", "5432")
	user     := getEnv("DB_USER", "shopos")
	password := getEnv("DB_PASSWORD", "changeme")
	dbname   := getEnv("DB_NAME", "subscription_billing_db")
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func waitForDB(db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := db.Ping(); err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return db.Ping()
}
