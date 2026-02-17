package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"paas-core/apps/api/internal/config"
	"paas-core/apps/api/internal/model"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
	ErrTokenReuse   = errors.New("token reuse detected")
	ErrTokenRevoked = errors.New("token has been revoked")
)

// Service defines the authentication service interface.
type Service interface {
	GenerateTokenPair(ctx context.Context, userID uuid.UUID, email, name string, roles []string) (*TokenPair, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	ValidateToken(tokenString string) (*Claims, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	jwtSecret        string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	refreshTokenRepo RefreshTokenRepository
	db               *gorm.DB
}

// NewService creates a new authentication service.
func NewService(cfg *config.JWTConfig, db *gorm.DB) Service {
	jwtSecret := cfg.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	accessTokenTTL := cfg.AccessTokenTTL
	if accessTokenTTL == 0 {
		accessTokenTTL = 15 * time.Minute
	}

	refreshTokenTTL := cfg.RefreshTokenTTL
	if refreshTokenTTL == 0 {
		refreshTokenTTL = 168 * time.Hour // 7 days
	}

	return &service{
		jwtSecret:        jwtSecret,
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		refreshTokenRepo: NewRefreshTokenRepository(db),
		db:               db,
	}
}

// GenerateTokenPair creates an access token and a refresh token with a shared family.
func (s *service) GenerateTokenPair(ctx context.Context, userID uuid.UUID, email, name string, roles []string) (*TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTokenTTL)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.New().String(),
		},
		UserID: userID,
		Email:  email,
		Name:   name,
		Roles:  roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate opaque refresh token
	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshToken := base64.URLEncoding.EncodeToString(refreshBytes)

	family := uuid.New()

	// Store refresh token hash
	tokenHash := hashToken(refreshToken)
	rtRecord := &model.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		Family:    family,
		ExpiresAt: now.Add(s.refreshTokenTTL),
	}
	if err := s.refreshTokenRepo.Create(ctx, rtRecord); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		TokenFamily:  family,
	}, nil
}

// RefreshAccessToken validates a refresh token and issues a new pair (rotation).
func (s *service) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	tokenHash := hashToken(refreshToken)

	stored, err := s.refreshTokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}
	if stored == nil {
		return nil, ErrInvalidToken
	}

	// Token reuse detection: if already revoked, revoke entire family
	if stored.Revoked {
		_ = s.refreshTokenRepo.RevokeByFamily(ctx, stored.Family)
		return nil, ErrTokenReuse
	}

	if time.Now().After(stored.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	// Revoke the old refresh token
	if err := s.refreshTokenRepo.RevokeByHash(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Fetch user roles for the new access token
	var roles []string
	if s.db != nil {
		err := s.db.Table("roles").
			Select("roles.name").
			Joins("JOIN user_roles ON user_roles.role_id = roles.id").
			Where("user_roles.user_id = ?", stored.UserID).
			Find(&roles).Error
		if err != nil {
			return nil, fmt.Errorf("failed to fetch roles: %w", err)
		}
	}

	// Fetch user info
	var user model.User
	if err := s.db.First(&user, "id = ?", stored.UserID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Issue new pair in the same family
	now := time.Now()
	expiresAt := now.Add(s.accessTokenTTL)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   stored.UserID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.New().String(),
		},
		UserID: stored.UserID,
		Email:  user.Email,
		Name:   user.Name,
		Roles:  roles,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newAccessToken, err := jwtToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign new access token: %w", err)
	}

	// New refresh token in same family
	newRefreshBytes := make([]byte, 32)
	if _, err := rand.Read(newRefreshBytes); err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}
	newRefreshToken := base64.URLEncoding.EncodeToString(newRefreshBytes)
	newTokenHash := hashToken(newRefreshToken)

	newRT := &model.RefreshToken{
		UserID:    stored.UserID,
		TokenHash: newTokenHash,
		Family:    stored.Family, // same family for rotation
		ExpiresAt: now.Add(s.refreshTokenTTL),
	}
	if err := s.refreshTokenRepo.Create(ctx, newRT); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		TokenFamily:  stored.Family,
	}, nil
}

// ValidateToken parses and validates a JWT access token.
func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RevokeRefreshToken revokes a single refresh token by its raw value.
func (s *service) RevokeRefreshToken(ctx context.Context, rawToken string) error {
	return s.refreshTokenRepo.RevokeByHash(ctx, hashToken(rawToken))
}

// RevokeAllUserTokens revokes all refresh tokens for a user.
func (s *service) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return s.refreshTokenRepo.RevokeAllForUser(ctx, userID)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
