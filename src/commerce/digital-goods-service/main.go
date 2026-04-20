package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/shopos/digital-goods-service/handler"
	"github.com/shopos/digital-goods-service/service"
	"github.com/shopos/digital-goods-service/store"
)

func main() {
	httpPort := getenv("HTTP_PORT", "8147")
	minioEndpoint := getenv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getenv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getenv("MINIO_SECRET_KEY", "minioadmin")
	minioBucket := getenv("MINIO_BUCKET", "digital-goods")
	minioUseSSL := getenvBool("MINIO_USE_SSL", false)

	assetStore, err := store.NewMinIOStore(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, minioUseSSL)
	if err != nil {
		log.Fatalf("failed to initialise MinIO store: %v", err)
	}

	svc := service.New(assetStore)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%s", httpPort)
	log.Printf("digital-goods-service listening on %s", addr)
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

func getenvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}
