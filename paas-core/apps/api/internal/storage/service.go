package storage

import (
	"context"
	"io"
	"time"
)

// FileInfo holds metadata about a stored file.
type FileInfo struct {
	Key         string `json:"key"`          // unique path/key in the bucket
	ContentType string `json:"content_type"` // MIME type
	Size        int64  `json:"size"`         // bytes
}

// Service defines the provider-agnostic storage interface (Goilerplate pattern).
type Service interface {
	// Upload stores a file and returns its storage key.
	Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (*FileInfo, error)

	// Delete removes a file by its key.
	Delete(ctx context.Context, key string) error

	// GetPresignedURL returns a time-limited URL for downloading the file.
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// GetPublicURL returns the permanent public URL (for public buckets).
	GetPublicURL(key string) string
}
