package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"paas-core/apps/api/internal/model"
)

// RefreshTokenRepository defines the storage interface for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	RevokeByHash(ctx context.Context, tokenHash string) error
	RevokeByFamily(ctx context.Context, family uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new repository backed by GORM.
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Updates(map[string]interface{}{"revoked": true, "revoked_at": &now}).Error
}

func (r *refreshTokenRepository) RevokeByFamily(ctx context.Context, family uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("family = ? AND revoked = false", family).
		Updates(map[string]interface{}{"revoked": true, "revoked_at": &now}).Error
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Updates(map[string]interface{}{"revoked": true, "revoked_at": &now}).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&model.RefreshToken{}).Error
}
