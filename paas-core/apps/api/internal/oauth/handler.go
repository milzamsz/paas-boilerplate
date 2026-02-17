package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"paas-core/apps/api/internal/auth"
	apiErrors "paas-core/apps/api/internal/errors"
)

// Sentinel errors
var (
	ErrLastAuthMethod = errors.New("cannot unlink the last authentication method")
	ErrAccountNotLinked = errors.New("oauth account not linked")
)

// OAuthAccountResponse is the public DTO for a linked OAuth account.
type OAuthAccountResponse struct {
	Provider  string    `json:"provider"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	LinkedAt  time.Time `json:"linked_at"`
}

// Handler handles OAuth HTTP routes.
type Handler struct {
	providers   map[string]Provider
	service     *OAuthService
	authService auth.Service
	frontendURL string
}

// NewHandler creates a new OAuth handler.
func NewHandler(providers map[string]Provider, service *OAuthService, authService auth.Service, frontendURL string) *Handler {
	return &Handler{
		providers:   providers,
		service:     service,
		authService: authService,
		frontendURL: frontendURL,
	}
}

// Initiate redirects the user to the provider's consent screen.
// GET /auth/oauth/:provider
func (h *Handler) Initiate(c *gin.Context) {
	providerName := c.Param("provider")
	provider, ok := h.providers[providerName]
	if !ok {
		_ = c.Error(apiErrors.BadRequest(fmt.Sprintf("Unsupported provider: %s", providerName)))
		return
	}

	// Generate cryptographic state token for CSRF protection
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store state in a secure cookie (5 minutes TTL)
	isProduction := c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetCookie(
		"oauth_state",
		state,
		300, // 5 minutes
		"/",
		"",
		isProduction,
		true, // HttpOnly
	)

	url := provider.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Callback handles the provider's redirect after consent.
// GET /auth/oauth/:provider/callback
func (h *Handler) Callback(c *gin.Context) {
	providerName := c.Param("provider")
	provider, ok := h.providers[providerName]
	if !ok {
		h.redirectError(c, "unsupported_provider", fmt.Sprintf("Unsupported provider: %s", providerName))
		return
	}

	// Verify state token
	state := c.Query("state")
	savedState, err := c.Cookie("oauth_state")
	if err != nil || state == "" || state != savedState {
		h.redirectError(c, "invalid_state", "Invalid or missing state token")
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Check for error from provider
	if errCode := c.Query("error"); errCode != "" {
		errDesc := c.DefaultQuery("error_description", "OAuth authorization was denied")
		h.redirectError(c, errCode, errDesc)
		return
	}

	// Exchange authorization code for tokens
	code := c.Query("code")
	if code == "" {
		h.redirectError(c, "missing_code", "Authorization code is missing")
		return
	}

	providerUser, _, err := provider.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		slog.Error("OAuth code exchange failed", "provider", providerName, "error", err)
		h.redirectError(c, "exchange_failed", "Failed to exchange authorization code")
		return
	}

	if providerUser.Email == "" {
		h.redirectError(c, "no_email", "No email address was provided by the OAuth provider")
		return
	}

	// Find or create user
	user, roles, _, err := h.service.FindOrCreateUser(c.Request.Context(), providerName, providerUser)
	if err != nil {
		slog.Error("OAuth user creation failed", "provider", providerName, "error", err)
		h.redirectError(c, "user_error", "Failed to create or link user account")
		return
	}

	// Generate JWT token pair
	tokenPair, err := h.authService.GenerateTokenPair(
		c.Request.Context(),
		user.ID,
		user.Email,
		user.Name,
		roles,
	)
	if err != nil {
		slog.Error("OAuth token generation failed", "provider", providerName, "error", err)
		h.redirectError(c, "token_error", "Failed to generate authentication tokens")
		return
	}

	// Redirect to frontend with tokens in URL fragment (not query params for security)
	redirectURL := fmt.Sprintf(
		"%s/auth/oauth/callback#access_token=%s&refresh_token=%s&token_type=%s&expires_in=%d",
		h.frontendURL,
		tokenPair.AccessToken,
		tokenPair.RefreshToken,
		tokenPair.TokenType,
		tokenPair.ExpiresIn,
	)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GetLinkedAccounts returns all OAuth accounts linked to the current user.
// GET /users/me/oauth-accounts
func (h *Handler) GetLinkedAccounts(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		_ = c.Error(apiErrors.Unauthorized(""))
		return
	}
	userID := userIDVal.(uuid.UUID)

	accounts, err := h.service.GetLinkedAccounts(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": accounts})
}

// UnlinkAccount removes an OAuth provider link from the current user.
// DELETE /users/me/oauth-accounts/:provider
func (h *Handler) UnlinkAccount(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		_ = c.Error(apiErrors.Unauthorized(""))
		return
	}
	userID := userIDVal.(uuid.UUID)
	provider := c.Param("provider")

	err := h.service.UnlinkAccount(c.Request.Context(), userID, provider)
	if err != nil {
		if errors.Is(err, ErrLastAuthMethod) {
			_ = c.Error(apiErrors.BadRequest("Cannot unlink the last authentication method. Please set a password first."))
			return
		}
		if errors.Is(err, ErrAccountNotLinked) {
			_ = c.Error(apiErrors.NotFound("OAuth account not linked"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s account unlinked successfully", provider)})
}

// redirectError redirects to the frontend with an error code and message.
func (h *Handler) redirectError(c *gin.Context, code, message string) {
	redirectURL := fmt.Sprintf(
		"%s/auth/oauth/callback?error=%s&error_description=%s",
		h.frontendURL,
		code,
		message,
	)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
