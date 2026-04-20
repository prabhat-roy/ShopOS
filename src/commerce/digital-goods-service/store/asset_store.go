package store

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/shopos/digital-goods-service/domain"
)

// AssetStorer defines the persistence interface for digital assets.
type AssetStorer interface {
	UploadAsset(ctx context.Context, productID, name, fileName, contentType string, reader io.Reader, size int64, downloadLimit int) (*domain.DigitalAsset, error)
	GetAsset(ctx context.Context, id string) (*domain.DigitalAsset, error)
	ListByProduct(ctx context.Context, productID string) ([]*domain.DigitalAsset, error)
	GenerateDownloadURL(ctx context.Context, id string, expiry time.Duration) (*domain.DownloadLink, error)
	DeleteAsset(ctx context.Context, id string) error
	IncrementDownloadCount(ctx context.Context, id string) error
}

// MinIOStore is a MinIO-backed implementation of AssetStorer.
// Asset metadata is kept in an in-memory map (no relational DB required).
type MinIOStore struct {
	client *minio.Client
	bucket string

	mu     sync.RWMutex
	assets map[string]*domain.DigitalAsset // keyed by asset ID
}

// NewMinIOStore creates and initializes a MinIOStore.
func NewMinIOStore(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOStore, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	ctx := context.Background()
	if exists, err := client.BucketExists(ctx, bucket); err == nil && !exists {
		client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}) //nolint:errcheck
	}

	return &MinIOStore{
		client: client,
		bucket: bucket,
		assets: make(map[string]*domain.DigitalAsset),
	}, nil
}

// UploadAsset streams content to MinIO and records metadata.
func (s *MinIOStore) UploadAsset(
	ctx context.Context,
	productID, name, fileName, contentType string,
	reader io.Reader,
	size int64,
	downloadLimit int,
) (*domain.DigitalAsset, error) {
	id := uuid.New().String()
	bucketKey := fmt.Sprintf("products/%s/%s/%s", productID, id, fileName)

	_, err := s.client.PutObject(ctx, s.bucket, bucketKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("put object: %w", err)
	}

	asset := &domain.DigitalAsset{
		ID:            id,
		ProductID:     productID,
		Name:          name,
		FileName:      fileName,
		ContentType:   contentType,
		SizeBytes:     size,
		BucketKey:     bucketKey,
		DownloadLimit: downloadLimit,
		DownloadCount: 0,
		Active:        true,
		CreatedAt:     time.Now().UTC(),
	}

	s.mu.Lock()
	s.assets[id] = asset
	s.mu.Unlock()

	return asset, nil
}

// GetAsset returns an asset by ID.
func (s *MinIOStore) GetAsset(_ context.Context, id string) (*domain.DigitalAsset, error) {
	s.mu.RLock()
	a, ok := s.assets[id]
	s.mu.RUnlock()
	if !ok {
		return nil, domain.ErrNotFound
	}
	// Return a copy to avoid data races.
	copy := *a
	return &copy, nil
}

// ListByProduct returns all active assets for a product.
func (s *MinIOStore) ListByProduct(_ context.Context, productID string) ([]*domain.DigitalAsset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*domain.DigitalAsset
	for _, a := range s.assets {
		if a.ProductID == productID && a.Active {
			copy := *a
			result = append(result, &copy)
		}
	}
	return result, nil
}

// GenerateDownloadURL creates a pre-signed MinIO URL for the asset.
func (s *MinIOStore) GenerateDownloadURL(ctx context.Context, id string, expiry time.Duration) (*domain.DownloadLink, error) {
	asset, err := s.GetAsset(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset.DownloadLimit > 0 && asset.DownloadCount >= asset.DownloadLimit {
		return nil, domain.ErrDownloadLimitExceeded
	}

	params := url.Values{}
	presigned, err := s.client.PresignedGetObject(ctx, s.bucket, asset.BucketKey, expiry, params)
	if err != nil {
		return nil, fmt.Errorf("presign: %w", err)
	}

	return &domain.DownloadLink{
		URL:       presigned.String(),
		ExpiresAt: time.Now().UTC().Add(expiry),
		Token:     uuid.New().String(),
	}, nil
}

// DeleteAsset soft-deletes an asset (marks inactive) and removes it from MinIO.
func (s *MinIOStore) DeleteAsset(ctx context.Context, id string) error {
	asset, err := s.GetAsset(ctx, id)
	if err != nil {
		return err
	}
	if err := s.client.RemoveObject(ctx, s.bucket, asset.BucketKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	s.mu.Lock()
	if a, ok := s.assets[id]; ok {
		a.Active = false
	}
	s.mu.Unlock()
	return nil
}

// IncrementDownloadCount atomically bumps the download counter.
func (s *MinIOStore) IncrementDownloadCount(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.assets[id]
	if !ok {
		return domain.ErrNotFound
	}
	a.DownloadCount++
	return nil
}
