package store

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/shopos/media-asset-service/internal/domain"
)

// AssetStore defines the persistence contract for media assets.
type AssetStore interface {
	Upload(ctx context.Context, asset domain.Asset, reader io.Reader, size int64) error
	Download(ctx context.Context, assetID string) (io.ReadCloser, domain.Asset, error)
	GetPresignedURL(ctx context.Context, assetID string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, assetID string) error
	GetMetadata(assetID string) (domain.Asset, error)
}

// MinIOStore implements AssetStore backed by MinIO object storage.
// Asset metadata is kept in an in-memory map protected by a read/write mutex.
type MinIOStore struct {
	client   *minio.Client
	bucket   string
	mu       sync.RWMutex
	metadata map[string]domain.Asset // keyed by asset.ID
}

// NewMinIOStore creates a new MinIOStore and ensures the target bucket exists.
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
		client:   client,
		bucket:   bucket,
		metadata: make(map[string]domain.Asset),
	}, nil
}

// objectKey returns the MinIO object key for a given asset.
// Objects are stored as "{id}/{originalName}" to keep filenames human-readable.
func objectKey(asset domain.Asset) string {
	return fmt.Sprintf("%s/%s", asset.ID, asset.OriginalName)
}

// Upload stores the asset bytes in MinIO and indexes its metadata in memory.
func (s *MinIOStore) Upload(ctx context.Context, asset domain.Asset, reader io.Reader, size int64) error {
	key := objectKey(asset)
	asset.StoredName = key

	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: asset.ContentType,
	})
	if err != nil {
		return fmt.Errorf("minio put object: %w", err)
	}

	s.mu.Lock()
	s.metadata[asset.ID] = asset
	s.mu.Unlock()

	return nil
}

// Download retrieves the asset bytes and metadata from MinIO.
func (s *MinIOStore) Download(ctx context.Context, assetID string) (io.ReadCloser, domain.Asset, error) {
	asset, err := s.GetMetadata(assetID)
	if err != nil {
		return nil, domain.Asset{}, err
	}

	obj, err := s.client.GetObject(ctx, s.bucket, asset.StoredName, minio.GetObjectOptions{})
	if err != nil {
		return nil, domain.Asset{}, fmt.Errorf("minio get object: %w", err)
	}

	return obj, asset, nil
}

// GetPresignedURL generates a time-limited pre-signed download URL for the asset.
func (s *MinIOStore) GetPresignedURL(ctx context.Context, assetID string, expiry time.Duration) (string, error) {
	asset, err := s.GetMetadata(assetID)
	if err != nil {
		return "", err
	}

	u, err := s.client.PresignedGetObject(ctx, s.bucket, asset.StoredName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}

	return u.String(), nil
}

// Delete removes the object from MinIO and purges its metadata entry.
func (s *MinIOStore) Delete(ctx context.Context, assetID string) error {
	asset, err := s.GetMetadata(assetID)
	if err != nil {
		return err
	}

	if err := s.client.RemoveObject(ctx, s.bucket, asset.StoredName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("minio remove object: %w", err)
	}

	s.mu.Lock()
	delete(s.metadata, assetID)
	s.mu.Unlock()

	return nil
}

// GetMetadata returns the in-memory metadata for the given asset ID.
func (s *MinIOStore) GetMetadata(assetID string) (domain.Asset, error) {
	s.mu.RLock()
	asset, ok := s.metadata[assetID]
	s.mu.RUnlock()

	if !ok {
		return domain.Asset{}, domain.ErrNotFound
	}
	return asset, nil
}
