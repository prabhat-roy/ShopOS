package domain

import (
	"errors"
	"time"
)

// ErrNotFound is returned when an asset cannot be located.
var ErrNotFound = errors.New("asset not found")

// AssetType classifies the media asset.
type AssetType string

const (
	AssetTypeImage    AssetType = "IMAGE"
	AssetTypeDocument AssetType = "DOCUMENT"
	AssetTypeVideo    AssetType = "VIDEO"
	AssetTypeAudio    AssetType = "AUDIO"
	AssetTypeOther    AssetType = "OTHER"
)

// AssetTypeFromContentType derives an AssetType from a MIME content-type string.
func AssetTypeFromContentType(ct string) AssetType {
	switch {
	case len(ct) >= 6 && ct[:6] == "image/":
		return AssetTypeImage
	case len(ct) >= 6 && ct[:6] == "video/":
		return AssetTypeVideo
	case len(ct) >= 6 && ct[:6] == "audio/":
		return AssetTypeAudio
	case ct == "application/pdf" ||
		ct == "application/msword" ||
		ct == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" ||
		ct == "text/plain":
		return AssetTypeDocument
	default:
		return AssetTypeOther
	}
}

// Asset represents a media file stored in MinIO with its associated metadata.
type Asset struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"originalName"`
	StoredName   string    `json:"storedName"`
	ContentType  string    `json:"contentType"`
	Size         int64     `json:"size"`
	AssetType    AssetType `json:"assetType"`
	Bucket       string    `json:"bucket"`
	Tags         []string  `json:"tags"`
	OwnerID      string    `json:"ownerId"`
	UploadedAt   time.Time `json:"uploadedAt"`
}
