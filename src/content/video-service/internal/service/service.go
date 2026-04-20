package service

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/video-service/internal/domain"
	"github.com/shopos/video-service/internal/store"
)

// Servicer defines the business-logic contract for video operations.
type Servicer interface {
	UploadVideo(ctx context.Context, ownerID, title, description, originalName, contentType string, reader io.Reader, size int64) (domain.Video, error)
	GetVideo(ctx context.Context, id string) (domain.Video, error)
	GetStreamURL(ctx context.Context, id string) (string, error)
	UpdateVideoStatus(ctx context.Context, id string, status domain.VideoStatus) error
	DeleteVideo(ctx context.Context, id string) error
	ListOwnerVideos(ctx context.Context, ownerID string) ([]domain.Video, error)
}

// Service implements Servicer using a VideoStore for persistence.
type Service struct {
	store        store.VideoStore
	bucket       string
	streamExpiry time.Duration
}

// New creates a new Service.
func New(s store.VideoStore, bucket string, streamExpiry time.Duration) *Service {
	return &Service{
		store:        s,
		bucket:       bucket,
		streamExpiry: streamExpiry,
	}
}

// UploadVideo builds a Video record in UPLOADING state and streams it to the store.
func (s *Service) UploadVideo(ctx context.Context, ownerID, title, description, originalName, contentType string, reader io.Reader, size int64) (domain.Video, error) {
	now := time.Now().UTC()
	video := domain.Video{
		ID:           uuid.New().String(),
		Title:        title,
		Description:  description,
		OriginalName: originalName,
		ContentType:  contentType,
		Size:         size,
		Status:       domain.VideoStatusUploading,
		OwnerID:      ownerID,
		Tags:         []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.store.UploadVideo(ctx, video, reader, size); err != nil {
		return domain.Video{}, err
	}

	// Transition to READY after a successful upload.
	// (Real pipeline would dispatch a processing job; here we go straight to READY.)
	if err := s.store.UpdateStatus(video.ID, domain.VideoStatusReady); err != nil {
		return domain.Video{}, err
	}

	return s.store.GetMetadata(video.ID)
}

// GetVideo returns video metadata by ID.
func (s *Service) GetVideo(ctx context.Context, id string) (domain.Video, error) {
	return s.store.GetMetadata(id)
}

// GetStreamURL returns a presigned streaming URL for the video.
func (s *Service) GetStreamURL(ctx context.Context, id string) (string, error) {
	return s.store.GetStreamURL(ctx, id, s.streamExpiry)
}

// UpdateVideoStatus changes the processing status of a video.
func (s *Service) UpdateVideoStatus(ctx context.Context, id string, status domain.VideoStatus) error {
	return s.store.UpdateStatus(id, status)
}

// DeleteVideo removes a video from storage.
func (s *Service) DeleteVideo(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// ListOwnerVideos returns all videos owned by ownerID.
func (s *Service) ListOwnerVideos(ctx context.Context, ownerID string) ([]domain.Video, error) {
	return s.store.ListByOwner(ownerID)
}
