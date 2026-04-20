package domain

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a video cannot be located.
var ErrNotFound = errors.New("video not found")

// VideoStatus represents the processing lifecycle of a video asset.
type VideoStatus string

const (
	VideoStatusUploading  VideoStatus = "UPLOADING"
	VideoStatusProcessing VideoStatus = "PROCESSING"
	VideoStatusReady      VideoStatus = "READY"
	VideoStatusFailed     VideoStatus = "FAILED"
)

// Video holds all metadata for a stored video asset.
type Video struct {
	ID           string      `json:"id"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	OriginalName string      `json:"originalName"`
	StoredKey    string      `json:"storedKey"`
	ContentType  string      `json:"contentType"`
	Size         int64       `json:"size"`
	Duration     int         `json:"duration"` // seconds; 0 until known
	Status       VideoStatus `json:"status"`
	ThumbnailURL string      `json:"thumbnailUrl"`
	OwnerID      string      `json:"ownerId"`
	Tags         []string    `json:"tags"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}
