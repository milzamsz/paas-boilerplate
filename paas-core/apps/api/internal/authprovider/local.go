package authprovider

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"paas-core/apps/api/internal/auth"
)

// UserService is the minimal interface the local provider needs from the user domain.
type UserService interface {
	RegisterUser(ctx *gin.Context, req auth.RegisterRequest) (*auth.UserResponse, []string, error)
	AuthenticateUser(ctx *gin.Context, req auth.LoginRequest) (*auth.UserResponse, []string, error)
}

// LocalProvider wraps the existing auth.Service + UserService to implement
// AuthProvider. This is the default provider when Supabase is not enabled.
type LocalProvider struct {
	authService auth.Service
	userService UserService
}

// NewLocalProvider creates a local auth provider from the existing services.
func NewLocalProvider(authService auth.Service, userService UserService) *LocalProvider {
	return &LocalProvider{
		authService: authService,
		userService: userService,
	}
}

func (p *LocalProvider) Name() string { return "local" }

// Register delegates to the existing user service and generates JWT tokens.
func (p *LocalProvider) Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, error) {
	// The existing user service expects a *gin.Context for validation.
	// We create a minimal gin context wrapping the real context.
	ginCtx := newGinContext(ctx)

	userResp, roles, err := p.userService.RegisterUser(ginCtx, req)
	if err != nil {
		return nil, err
	}

	tokenPair, err := p.authService.GenerateTokenPair(ctx, userResp.ID, userResp.Email, userResp.Name, roles)
	if err != nil {
		return nil, err
	}

	return &auth.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         *userResp,
	}, nil
}

// Login delegates to the existing user service and generates JWT tokens.
func (p *LocalProvider) Login(ctx context.Context, req auth.LoginRequest) (*auth.AuthResponse, error) {
	ginCtx := newGinContext(ctx)

	userResp, roles, err := p.userService.AuthenticateUser(ginCtx, req)
	if err != nil {
		return nil, err
	}

	tokenPair, err := p.authService.GenerateTokenPair(ctx, userResp.ID, userResp.Email, userResp.Name, roles)
	if err != nil {
		return nil, err
	}

	return &auth.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         *userResp,
	}, nil
}

// ValidateToken delegates to the existing auth service.
func (p *LocalProvider) ValidateToken(tokenString string) (*auth.Claims, error) {
	return p.authService.ValidateToken(tokenString)
}

// RefreshToken delegates to the existing auth service.
func (p *LocalProvider) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	return p.authService.RefreshAccessToken(ctx, refreshToken)
}

// Logout revokes all refresh tokens for the user.
func (p *LocalProvider) Logout(ctx context.Context, userID uuid.UUID) error {
	return p.authService.RevokeAllUserTokens(ctx, userID)
}

// newGinContext creates a minimal *gin.Context that wraps a context.Context.
// This bridges the gap between the provider interface (context.Context) and
// the existing user service (which expects *gin.Context).
func newGinContext(ctx context.Context) *gin.Context {
	ginCtx := &gin.Context{
		Request: (&http.Request{}).WithContext(ctx),
	}
	return ginCtx
}
