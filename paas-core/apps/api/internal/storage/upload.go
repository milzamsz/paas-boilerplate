package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/model"
)

// MaxAvatarSize is the maximum avatar file size (5 MB).
const MaxAvatarSize = 5 * 1024 * 1024

// AllowedImageTypes lists accepted MIME types for avatar uploads.
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// UploadService handles file uploads and metadata tracking.
type UploadService struct {
	db      *gorm.DB
	storage Service
}

// NewUploadService creates a new upload service.
func NewUploadService(db *gorm.DB, storage Service) *UploadService {
	return &UploadService{db: db, storage: storage}
}

// UploadAvatar uploads an avatar image for a user or org and returns the public URL.
func (s *UploadService) UploadAvatar(ctx context.Context, ownerID uuid.UUID, ownerType string, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate file size
	if header.Size > MaxAvatarSize {
		return "", apiErrors.BadRequest("File size exceeds 5MB limit")
	}

	// Read first 512 bytes for content type detection
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	contentType := http.DetectContentType(buf[:n])

	// Validate content type
	if !AllowedImageTypes[contentType] {
		return "", apiErrors.BadRequest("File type not allowed. Accepted: JPEG, PNG, GIF, WebP")
	}

	// Reset read position
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to seek file: %w", err)
	}

	// Generate unique storage key: avatars/{ownerType}/{ownerID}/{uuid}.{ext}
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = extensionFromMIME(contentType)
	}
	key := fmt.Sprintf("avatars/%s/%s/%s%s", ownerType, ownerID.String(), uuid.New().String(), ext)

	// Upload to storage
	info, err := s.storage.Upload(ctx, key, file, contentType, header.Size)
	if err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Save metadata to DB
	upload := model.FileUpload{
		OwnerID:     ownerID,
		OwnerType:   ownerType,
		Key:         info.Key,
		Filename:    header.Filename,
		ContentType: info.ContentType,
		Size:        info.Size,
		Category:    "avatar",
	}
	if err := s.db.WithContext(ctx).Create(&upload).Error; err != nil {
		// Try to clean up storage on DB error
		_ = s.storage.Delete(ctx, key)
		return "", fmt.Errorf("failed to save upload metadata: %w", err)
	}

	return s.storage.GetPublicURL(key), nil
}

// DeleteAvatar removes an avatar by its storage key.
func (s *UploadService) DeleteAvatar(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}

	// Delete from storage
	if err := s.storage.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete avatar from storage: %w", err)
	}

	// Delete metadata record
	if err := s.db.WithContext(ctx).Where("key = ?", key).Delete(&model.FileUpload{}).Error; err != nil {
		return fmt.Errorf("failed to delete upload metadata: %w", err)
	}

	return nil
}

// GetPresignedURL generates a time-limited download URL for a file.
func (s *UploadService) GetPresignedURL(ctx context.Context, key string) (string, error) {
	return s.storage.GetPresignedURL(ctx, key, 15*time.Minute)
}

// extensionFromMIME returns a file extension for common MIME types.
func extensionFromMIME(mime string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}
