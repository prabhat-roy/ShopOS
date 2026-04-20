package store

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/shopos/video-service/internal/domain"
)

// VideoStore defines the persistence contract for video assets.
type VideoStore interface {
	UploadVideo(ctx context.Context, video domain.Video, reader io.Reader, size int64) error
	GetStreamURL(ctx context.Context, id string, expiry time.Duration) (string, error)
	GetMetadata(id string) (domain.Video, error)
	UpdateStatus(id string, status domain.VideoStatus) error
	Delete(ctx context.Context, id string) error
	ListByOwner(ownerID string) ([]domain.Video, error)
}

// MinIOStore implements VideoStore backed by MinIO.
// Video metadata is kept in an in-memory map protected by a read/write mutex.
type MinIOStore struct {
	client   *minio.Client
	bucket   string
	mu       sync.RWMutex
	metadata map[string]domain.Video // keyed by video.ID
}

// NewMinIOStore creates a MinIOStore and ensures the target bucket exists.
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
		metadata: make(map[string]domain.Video),
	}, nil
}

// UploadVideo streams video bytes to MinIO and indexes its metadata.
func (s *MinIOStore) UploadVideo(ctx context.Context, video domain.Video, reader io.Reader, size int64) error {
	key := fmt.Sprintf("%s/%s", video.ID, video.OriginalName)
	video.StoredKey = key

	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: video.ContentType,
	})
	if err != nil {
		return fmt.Errorf("minio put object: %w", err)
	}

	s.mu.Lock()
	s.metadata[video.ID] = video
	s.mu.Unlock()

	return nil
}

// GetStreamURL returns a presigned URL valid for the given expiry duration.
func (s *MinIOStore) GetStreamURL(ctx context.Context, id string, expiry time.Duration) (string, error) {
	video, err := s.GetMetadata(id)
	if err != nil {
		return "", err
	}

	u, err := s.client.PresignedGetObject(ctx, s.bucket, video.StoredKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}
	return u.String(), nil
}

// GetMetadata returns the in-memory metadata for a given video ID.
func (s *MinIOStore) GetMetadata(id string) (domain.Video, error) {
	s.mu.RLock()
	video, ok := s.metadata[id]
	s.mu.RUnlock()

	if !ok {
		return domain.Video{}, domain.ErrNotFound
	}
	return video, nil
}

// UpdateStatus changes the processing status of a video in the metadata index.
func (s *MinIOStore) UpdateStatus(id string, status domain.VideoStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	video, ok := s.metadata[id]
	if !ok {
		return domain.ErrNotFound
	}
	video.Status = status
	video.UpdatedAt = time.Now().UTC()
	s.metadata[id] = video
	return nil
}

// Delete removes the video object from MinIO and purges its metadata entry.
func (s *MinIOStore) Delete(ctx context.Context, id string) error {
	video, err := s.GetMetadata(id)
	if err != nil {
		return err
	}

	if err := s.client.RemoveObject(ctx, s.bucket, video.StoredKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("minio remove object: %w", err)
	}

	s.mu.Lock()
	delete(s.metadata, id)
	s.mu.Unlock()

	return nil
}

// ListByOwner returns all videos belonging to a given owner ID.
func (s *MinIOStore) ListByOwner(ownerID string) ([]domain.Video, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.Video
	for _, v := range s.metadata {
		if v.OwnerID == ownerID {
			result = append(result, v)
		}
	}
	if result == nil {
		result = []domain.Video{}
	}
	return result, nil
}
