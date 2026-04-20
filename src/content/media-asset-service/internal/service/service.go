package service

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/media-asset-service/internal/domain"
	"github.com/shopos/media-asset-service/internal/store"
)

// Servicer defines the business-logic contract for media asset operations.
type Servicer interface {
	UploadAsset(ctx context.Context, ownerID, originalName, contentType string, reader io.Reader, size int64) (domain.Asset, error)
	GetAsset(ctx context.Context, assetID string) (domain.Asset, error)
	GetDownloadURL(ctx context.Context, assetID string, expiry time.Duration) (string, error)
	DownloadAsset(ctx context.Context, assetID string) (io.ReadCloser, domain.Asset, error)
	DeleteAsset(ctx context.Context, assetID string) error
}

// Service implements Servicer using an AssetStore for persistence.
type Service struct {
	store         store.AssetStore
	defaultBucket string
	presignExpiry time.Duration
}

// New creates a new Service.
func New(s store.AssetStore, bucket string, presignExpiry time.Duration) *Service {
	return &Service{
		store:         s,
		defaultBucket: bucket,
		presignExpiry: presignExpiry,
	}
}

// UploadAsset validates the upload request, constructs an Asset record, and
// streams the file bytes to the store.
func (s *Service) UploadAsset(ctx context.Context, ownerID, originalName, contentType string, reader io.Reader, size int64) (domain.Asset, error) {
	id := uuid.New().String()

	asset := domain.Asset{
		ID:           id,
		OriginalName: originalName,
		ContentType:  contentType,
		Size:         size,
		AssetType:    domain.AssetTypeFromContentType(contentType),
		Bucket:       s.defaultBucket,
		Tags:         []string{},
		OwnerID:      ownerID,
		UploadedAt:   time.Now().UTC(),
	}

	if err := s.store.Upload(ctx, asset, reader, size); err != nil {
		return domain.Asset{}, err
	}

	// Fetch back so StoredName is populated from the store.
	return s.store.GetMetadata(id)
}

// GetAsset retrieves asset metadata by ID.
func (s *Service) GetAsset(ctx context.Context, assetID string) (domain.Asset, error) {
	return s.store.GetMetadata(assetID)
}

// GetDownloadURL returns a pre-signed URL valid for the given expiry duration.
// If expiry is zero, the service default is used.
func (s *Service) GetDownloadURL(ctx context.Context, assetID string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = s.presignExpiry
	}
	return s.store.GetPresignedURL(ctx, assetID, expiry)
}

// DownloadAsset streams the raw asset bytes together with its metadata.
func (s *Service) DownloadAsset(ctx context.Context, assetID string) (io.ReadCloser, domain.Asset, error) {
	return s.store.Download(ctx, assetID)
}

// DeleteAsset removes an asset from storage and purges its metadata.
func (s *Service) DeleteAsset(ctx context.Context, assetID string) error {
	return s.store.Delete(ctx, assetID)
}
