package service

import (
	"context"
	"io"
	"time"

	"github.com/shopos/digital-goods-service/domain"
	"github.com/shopos/digital-goods-service/store"
)

const defaultDownloadURLExpiry = 15 * time.Minute

// AssetService implements business logic for digital goods.
type AssetService struct {
	store store.AssetStorer
}

// New creates an AssetService.
func New(s store.AssetStorer) *AssetService {
	return &AssetService{store: s}
}

// UploadAsset stores a digital file and records its metadata.
func (svc *AssetService) UploadAsset(
	ctx context.Context,
	productID, name, fileName, contentType string,
	reader io.Reader,
	size int64,
	downloadLimit int,
) (*domain.DigitalAsset, error) {
	return svc.store.UploadAsset(ctx, productID, name, fileName, contentType, reader, size, downloadLimit)
}

// GetAsset retrieves asset metadata by ID.
func (svc *AssetService) GetAsset(ctx context.Context, id string) (*domain.DigitalAsset, error) {
	return svc.store.GetAsset(ctx, id)
}

// ListByProduct returns all active assets for a product.
func (svc *AssetService) ListByProduct(ctx context.Context, productID string) ([]*domain.DigitalAsset, error) {
	return svc.store.ListByProduct(ctx, productID)
}

// GenerateDownloadLink creates a time-limited pre-signed download URL for an asset.
// It also increments the download counter.
func (svc *AssetService) GenerateDownloadLink(ctx context.Context, assetID, _ string) (*domain.DownloadLink, error) {
	link, err := svc.store.GenerateDownloadURL(ctx, assetID, defaultDownloadURLExpiry)
	if err != nil {
		return nil, err
	}
	// Best-effort increment; ignore errors (non-critical)
	_ = svc.store.IncrementDownloadCount(ctx, assetID)
	return link, nil
}

// DeleteAsset removes an asset from storage.
func (svc *AssetService) DeleteAsset(ctx context.Context, id string) error {
	return svc.store.DeleteAsset(ctx, id)
}
