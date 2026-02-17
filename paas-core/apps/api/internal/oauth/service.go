package oauth

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"paas-core/apps/api/internal/model"
)

// OAuthService handles user lookup/creation and account linking for OAuth flows.
type OAuthService struct {
	db *gorm.DB
}

// NewOAuthService creates a new OAuth service.
func NewOAuthService(db *gorm.DB) *OAuthService {
	return &OAuthService{db: db}
}

// FindOrCreateUser finds an existing user by OAuth link or email, or creates a new one.
// Returns the user, their roles, and whether the account is newly created.
func (s *OAuthService) FindOrCreateUser(ctx context.Context, provider string, pu *ProviderUser) (*model.User, []string, bool, error) {
	// 1. Check if an OAuth account already exists for this provider + provider ID
	var oauthAccount model.OAuthAccount
	err := s.db.Where("provider = ? AND provider_id = ?", provider, pu.ID).First(&oauthAccount).Error
	if err == nil {
		// Found existing link — load user
		var user model.User
		if err := s.db.Preload("Roles").First(&user, "id = ?", oauthAccount.UserID).Error; err != nil {
			return nil, nil, false, err
		}
		roles := make([]string, len(user.Roles))
		for i, r := range user.Roles {
			roles[i] = r.Name
		}
		return &user, roles, false, nil
	}

	// 2. Check if a user with this email already exists → auto-link
	if pu.Email != "" {
		var existingUser model.User
		err := s.db.Preload("Roles").Where("email = ?", pu.Email).First(&existingUser).Error
		if err == nil {
			// Auto-link the OAuth account to the existing user
			link := model.OAuthAccount{
				UserID:     existingUser.ID,
				Provider:   provider,
				ProviderID: pu.ID,
				Email:      pu.Email,
				AvatarURL:  pu.AvatarURL,
			}
			if err := s.db.Create(&link).Error; err != nil {
				return nil, nil, false, err
			}
			// Auto-verify email since OAuth provider already verified it
			if !existingUser.EmailVerified {
				s.db.Model(&existingUser).Update("email_verified", true)
				existingUser.EmailVerified = true
			}
			// Update avatar if empty
			if existingUser.AvatarURL == "" && pu.AvatarURL != "" {
				s.db.Model(&existingUser).Update("avatar_url", pu.AvatarURL)
				existingUser.AvatarURL = pu.AvatarURL
			}
			roles := make([]string, len(existingUser.Roles))
			for i, r := range existingUser.Roles {
				roles[i] = r.Name
			}
			slog.Info("OAuth account auto-linked to existing user", "provider", provider, "email", pu.Email)
			return &existingUser, roles, false, nil
		}
	}

	// 3. Create a new user + OAuth link
	newUser := model.User{
		Name:          pu.Name,
		Email:         pu.Email,
		PasswordHash:  "", // OAuth-only user, no password
		AvatarURL:     pu.AvatarURL,
		EmailVerified: true, // OAuth provider already verified the email
	}
	if err := s.db.Create(&newUser).Error; err != nil {
		return nil, nil, false, err
	}

	// Assign default "user" role
	var userRole model.Role
	if err := s.db.Where("name = ?", "user").First(&userRole).Error; err == nil {
		s.db.Create(&model.UserRole{UserID: newUser.ID, RoleID: userRole.ID})
	}

	// Create OAuth link
	link := model.OAuthAccount{
		UserID:     newUser.ID,
		Provider:   provider,
		ProviderID: pu.ID,
		Email:      pu.Email,
		AvatarURL:  pu.AvatarURL,
	}
	if err := s.db.Create(&link).Error; err != nil {
		return nil, nil, false, err
	}

	slog.Info("New user created via OAuth", "provider", provider, "email", pu.Email, "userId", newUser.ID)
	return &newUser, []string{"user"}, true, nil
}

// GetLinkedAccounts returns all OAuth accounts linked to a user.
func (s *OAuthService) GetLinkedAccounts(ctx context.Context, userID uuid.UUID) ([]OAuthAccountResponse, error) {
	var accounts []model.OAuthAccount
	if err := s.db.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	result := make([]OAuthAccountResponse, len(accounts))
	for i, a := range accounts {
		result[i] = OAuthAccountResponse{
			Provider:  a.Provider,
			Email:     a.Email,
			AvatarURL: a.AvatarURL,
			LinkedAt:  a.CreatedAt,
		}
	}
	return result, nil
}

// UnlinkAccount removes an OAuth provider link from a user.
// Prevents unlinking the last auth method.
func (s *OAuthService) UnlinkAccount(ctx context.Context, userID uuid.UUID, provider string) error {
	// Check that user has another auth method
	var user model.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	var linkCount int64
	s.db.Model(&model.OAuthAccount{}).Where("user_id = ?", userID).Count(&linkCount)

	// A valid bcrypt hash is always 60 characters long
	hasPassword := len(user.PasswordHash) >= 60

	if linkCount <= 1 && !hasPassword {
		return ErrLastAuthMethod
	}

	result := s.db.Where("user_id = ? AND provider = ?", userID, provider).Delete(&model.OAuthAccount{})
	if result.RowsAffected == 0 {
		return ErrAccountNotLinked
	}
	return result.Error
}
