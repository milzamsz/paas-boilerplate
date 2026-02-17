package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"paas-core/apps/api/internal/email"
	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/model"
)

const (
	verificationTokenExpiry  = 24 * time.Hour  // 24 hours for email verification
	passwordResetTokenExpiry = 15 * time.Minute // 15 minutes for password reset (Goilerplate pattern)
	tokenByteLength          = 32
)

// VerificationService handles email verification and password reset flows.
type VerificationService struct {
	db           *gorm.DB
	emailService email.Service
	appName      string
	appURL       string // e.g. "https://app.example.com"
}

// NewVerificationService creates a new verification service.
func NewVerificationService(db *gorm.DB, emailService email.Service, appName, appURL string) *VerificationService {
	return &VerificationService{
		db:           db,
		emailService: emailService,
		appName:      appName,
		appURL:       appURL,
	}
}

// SendVerificationEmail generates a token and sends a verification email.
func (s *VerificationService) SendVerificationEmail(ctx context.Context, usr *model.User) error {
	rawToken, tokenHash := generateTokenPair()

	record := &model.EmailVerificationToken{
		UserID:    usr.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(verificationTokenExpiry),
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	link := fmt.Sprintf("%s/auth/verify-email?token=%s", s.appURL, rawToken)

	msg := email.RenderVerificationEmail(email.TemplateData{
		AppName:   s.appName,
		AppURL:    s.appURL,
		UserName:  usr.Name,
		UserEmail: usr.Email,
		Token:     rawToken,
		Link:      link,
		ExpiresIn: "24 hours",
	})

	return s.emailService.Send(ctx, msg)
}

// VerifyEmail validates a verification token and marks the user's email as verified.
func (s *VerificationService) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash := hashToken(rawToken)

	var token model.EmailVerificationToken
	if err := s.db.WithContext(ctx).
		Where("token_hash = ? AND expires_at > ? AND used_at IS NULL", tokenHash, time.Now()).
		First(&token).Error; err != nil {
		return apiErrors.BadRequest("Invalid or expired verification token")
	}

	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&token).Update("used_at", now).Error; err != nil {
			return fmt.Errorf("failed to mark token as used: %w", err)
		}
		if err := tx.Model(&model.User{}).Where("id = ?", token.UserID).Update("email_verified", true).Error; err != nil {
			return fmt.Errorf("failed to verify user email: %w", err)
		}
		return nil
	})
}

// SendPasswordResetEmail generates a reset token and sends a password reset email.
// If the email doesn't exist, returns nil silently (security: don't reveal existence).
func (s *VerificationService) SendPasswordResetEmail(ctx context.Context, emailAddr string) error {
	var usr model.User
	if err := s.db.WithContext(ctx).Where("email = ?", emailAddr).First(&usr).Error; err != nil {
		return nil // don't reveal whether email exists
	}

	rawToken, tokenHash := generateTokenPair()

	record := &model.PasswordResetToken{
		UserID:    usr.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(passwordResetTokenExpiry),
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	link := fmt.Sprintf("%s/auth/reset-password?token=%s", s.appURL, rawToken)

	msg := email.RenderPasswordResetEmail(email.TemplateData{
		AppName:   s.appName,
		AppURL:    s.appURL,
		UserName:  usr.Name,
		UserEmail: usr.Email,
		Token:     rawToken,
		Link:      link,
		ExpiresIn: "15 minutes",
	})

	return s.emailService.Send(ctx, msg)
}

// ResetPassword validates a reset token and updates the user's password.
func (s *VerificationService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if err := ValidatePasswordNIST(newPassword); err != nil {
		return apiErrors.BadRequest(err.Error())
	}

	tokenHash := hashToken(rawToken)

	var token model.PasswordResetToken
	if err := s.db.WithContext(ctx).
		Where("token_hash = ? AND expires_at > ? AND used_at IS NULL", tokenHash, time.Now()).
		First(&token).Error; err != nil {
		return apiErrors.BadRequest("Invalid or expired reset token")
	}

	hashedPw, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&token).Update("used_at", now).Error; err != nil {
			return fmt.Errorf("failed to mark token as used: %w", err)
		}
		if err := tx.Model(&model.User{}).Where("id = ?", token.UserID).Update("password_hash", string(hashedPw)).Error; err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}
		// Invalidate all other reset tokens for this user
		if err := tx.Model(&model.PasswordResetToken{}).
			Where("user_id = ? AND used_at IS NULL AND id != ?", token.UserID, token.ID).
			Update("used_at", now).Error; err != nil {
			return fmt.Errorf("failed to invalidate other reset tokens: %w", err)
		}
		return nil
	})
}

// --- Helpers ---

// generateTokenPair returns a raw token (for the user) and its SHA-256 hash (for DB storage).
func generateTokenPair() (rawToken, tokenHash string) {
	b := make([]byte, tokenByteLength)
	_, _ = rand.Read(b)
	raw := hex.EncodeToString(b)
	return raw, hashToken(raw)
}

// hashToken returns the SHA-256 hex digest of a token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
