package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiErrors "paas-core/apps/api/internal/errors"
)

// AuthProvider is a minimal interface that the auth handler needs.
// Both the local and Supabase providers satisfy it.
type AuthProvider interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	Name() string
}

// Handler handles authentication HTTP requests.
type Handler struct {
	provider AuthProvider
}

// NewHandler creates a new auth handler using the given provider.
func NewHandler(provider AuthProvider) *Handler {
	return &Handler{provider: provider}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email and password, returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 200 {object} errors.Response{data=AuthResponse} "Success"
// @Failure 400 {object} errors.Response "Validation error"
// @Failure 409 {object} errors.Response "Email already exists"
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	authResp, err := h.provider.Register(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(authResp))
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} errors.Response{data=AuthResponse} "Success"
// @Failure 401 {object} errors.Response "Invalid credentials"
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	authResp, err := h.provider.Login(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(authResp))
}

// Refresh godoc
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh request"
// @Success 200 {object} errors.Response{data=TokenPair} "Success"
// @Failure 401 {object} errors.Response "Invalid or expired refresh token"
// @Router /api/v1/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	tokenPair, err := h.provider.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		_ = c.Error(apiErrors.Unauthorized(err.Error()))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(tokenPair))
}

// Logout godoc
// @Summary Logout user
// @Description Revoke all refresh tokens for the authenticated user
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} errors.Response "Success"
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		_ = c.Error(apiErrors.Unauthorized(""))
		return
	}
	authClaims := claims.(*Claims)

	if err := h.provider.Logout(c.Request.Context(), authClaims.UserID); err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Logged out successfully"}))
}
