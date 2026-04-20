package domain

import (
	"errors"
	"time"
)

// Sentinel errors.
var (
	ErrNotFound              = errors.New("asset not found")
	ErrDownloadLimitExceeded = errors.New("download limit exceeded")
)

// DigitalAsset represents a file stored in MinIO and associated with a product.
type DigitalAsset struct {
	ID             string    `json:"id"`
	ProductID      string    `json:"product_id"`
	Name           string    `json:"name"`
	FileName       string    `json:"file_name"`
	ContentType    string    `json:"content_type"`
	SizeBytes      int64     `json:"size_bytes"`
	BucketKey      string    `json:"bucket_key"`
	DownloadLimit  int       `json:"download_limit"`  // 0 = unlimited
	DownloadCount  int       `json:"download_count"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
}

// DownloadLink is a pre-signed URL returned to a purchaser.
type DownloadLink struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	Token     string    `json:"token"`
}
