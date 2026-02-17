package authprovider

import (
	"context"

	"github.com/google/uuid"

	"paas-core/apps/api/internal/auth"
)

// AuthProvider abstracts authentication so the API can use either the built-in
// local auth (bcrypt + HS256 JWT) or Supabase GoTrue without changing handlers.
type AuthProvider interface {
	// Register creates a new user account and returns tokens.
	Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, error)

	// Login authenticates a user and returns tokens.
	Login(ctx context.Context, req auth.LoginRequest) (*auth.AuthResponse, error)

	// ValidateToken verifies an access token and returns its claims.
	ValidateToken(tokenString string) (*auth.Claims, error)

	// RefreshToken exchanges a refresh token for a new token pair.
	RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error)

	// Logout invalidates all sessions / refresh tokens for the user.
	Logout(ctx context.Context, userID uuid.UUID) error

	// Name returns the provider identifier ("local" or "supabase").
	Name() string
}
