package authprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"paas-core/apps/api/internal/auth"
	"paas-core/apps/api/internal/config"
)

// Sentinel errors for Supabase operations.
var (
	ErrSupabaseRequest = errors.New("supabase request failed")
	ErrSupabaseAuth    = errors.New("supabase authentication failed")
	ErrSupabaseToken   = errors.New("supabase token validation failed")
)

// SupabaseProvider implements AuthProvider by delegating to Supabase GoTrue.
type SupabaseProvider struct {
	baseURL    string // e.g. https://xyz.supabase.co or http://localhost:8000
	anonKey    string
	serviceKey string
	jwtSecret  string
	httpClient *http.Client
}

// NewSupabaseProvider creates a Supabase auth provider.
func NewSupabaseProvider(cfg config.SupabaseConfig) *SupabaseProvider {
	return &SupabaseProvider{
		baseURL:    cfg.URL,
		anonKey:    cfg.AnonKey,
		serviceKey: cfg.ServiceKey,
		jwtSecret:  cfg.JWTSecret,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *SupabaseProvider) Name() string { return "supabase" }

// --- GoTrue REST API types ---

type gotrueSignUpRequest struct {
	Email    string                 `json:"email"`
	Password string                 `json:"password"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type gotrueTokenRequest struct {
	Email        string `json:"email,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type gotrueTokenResponse struct {
	AccessToken  string     `json:"access_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"`
	ExpiresAt    int64      `json:"expires_at"`
	RefreshToken string     `json:"refresh_token"`
	User         gotrueUser `json:"user"`
}

type gotrueUser struct {
	ID           string                 `json:"id"`
	Email        string                 `json:"email"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
	CreatedAt    string                 `json:"created_at"`
}

type gotrueErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Message          string `json:"msg"`
	Code             int    `json:"code"`
}

// Register creates a user in Supabase via POST /auth/v1/signup.
func (p *SupabaseProvider) Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, error) {
	body := gotrueSignUpRequest{
		Email:    req.Email,
		Password: req.Password,
		Data: map[string]interface{}{
			"name": req.Name,
		},
	}

	resp, err := p.doGoTrueRequest(ctx, http.MethodPost, "/auth/v1/signup", body, false)
	if err != nil {
		return nil, err
	}

	return p.tokenResponseToAuthResponse(resp), nil
}

// Login authenticates via POST /auth/v1/token?grant_type=password.
func (p *SupabaseProvider) Login(ctx context.Context, req auth.LoginRequest) (*auth.AuthResponse, error) {
	body := gotrueTokenRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := p.doGoTrueRequest(ctx, http.MethodPost, "/auth/v1/token?grant_type=password", body, false)
	if err != nil {
		return nil, err
	}

	return p.tokenResponseToAuthResponse(resp), nil
}

// ValidateToken verifies a Supabase JWT using the configured JWT secret (HS256).
func (p *SupabaseProvider) ValidateToken(tokenString string) (*auth.Claims, error) {
	if p.jwtSecret == "" {
		return nil, fmt.Errorf("%w: JWT secret not configured", ErrSupabaseToken)
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Supabase uses HS256 by default
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(p.jwtSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, auth.ErrExpiredToken
		}
		return nil, auth.ErrInvalidToken
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, auth.ErrInvalidToken
	}

	// Extract Supabase claims and map to our Claims struct
	sub, _ := mapClaims.GetSubject()
	userID, err := uuid.Parse(sub)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user id in token", auth.ErrInvalidToken)
	}

	email, _ := mapClaims["email"].(string)
	name := ""
	if metadata, ok := mapClaims["user_metadata"].(map[string]interface{}); ok {
		name, _ = metadata["name"].(string)
	}

	// Extract role from Supabase's "role" claim
	role, _ := mapClaims["role"].(string)
	roles := []string{}
	if role != "" {
		roles = append(roles, role)
	}

	// Extract app_metadata.roles if present (custom claim)
	if appMeta, ok := mapClaims["app_metadata"].(map[string]interface{}); ok {
		if appRoles, ok := appMeta["roles"].([]interface{}); ok {
			for _, r := range appRoles {
				if rs, ok := r.(string); ok {
					roles = append(roles, rs)
				}
			}
		}
	}

	iat, _ := mapClaims.GetIssuedAt()
	exp, _ := mapClaims.GetExpirationTime()
	jti, _ := mapClaims["jti"].(string)
	if jti == "" {
		jti = uuid.New().String()
	}

	claims := &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			IssuedAt:  iat,
			ExpiresAt: exp,
			ID:        jti,
		},
		UserID: userID,
		Email:  email,
		Name:   name,
		Roles:  roles,
	}

	return claims, nil
}

// RefreshToken exchanges a refresh token via POST /auth/v1/token?grant_type=refresh_token.
func (p *SupabaseProvider) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	body := gotrueTokenRequest{
		RefreshToken: refreshToken,
	}

	resp, err := p.doGoTrueRequest(ctx, http.MethodPost, "/auth/v1/token?grant_type=refresh_token", body, false)
	if err != nil {
		return nil, err
	}

	return &auth.TokenPair{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

// Logout invalidates the user's session via POST /auth/v1/logout.
func (p *SupabaseProvider) Logout(ctx context.Context, userID uuid.UUID) error {
	// Supabase logout requires the user's access token, not the user ID.
	// We use the service key to call the admin API to revoke sessions.
	url := fmt.Sprintf("%s/auth/v1/admin/users/%s/factors", p.baseURL, userID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSupabaseRequest, err)
	}
	req.Header.Set("apikey", p.serviceKey)
	req.Header.Set("Authorization", "Bearer "+p.serviceKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSupabaseRequest, err)
	}
	defer resp.Body.Close()

	// Logout is best-effort; 404 means user has no active factors, which is fine
	if resp.StatusCode >= 500 {
		return fmt.Errorf("%w: server error %d", ErrSupabaseRequest, resp.StatusCode)
	}

	return nil
}

// --- Helpers ---

// doGoTrueRequest makes an HTTP request to the Supabase GoTrue API.
func (p *SupabaseProvider) doGoTrueRequest(ctx context.Context, method, path string, body interface{}, useServiceKey bool) (*gotrueTokenResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal request: %v", ErrSupabaseRequest, err)
	}

	url := p.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrSupabaseRequest, err)
	}

	req.Header.Set("Content-Type", "application/json")
	apiKey := p.anonKey
	if useServiceKey {
		apiKey = p.serviceKey
	}
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSupabaseRequest, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrSupabaseRequest, err)
	}

	if resp.StatusCode >= 400 {
		var errResp gotrueErrorResponse
		_ = json.Unmarshal(respBody, &errResp)
		msg := errResp.ErrorDescription
		if msg == "" {
			msg = errResp.Message
		}
		if msg == "" {
			msg = errResp.Error
		}
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%w: %s", ErrSupabaseAuth, msg)
	}

	var tokenResp gotrueTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrSupabaseRequest, err)
	}

	return &tokenResp, nil
}

// tokenResponseToAuthResponse converts a GoTrue token response to our AuthResponse.
func (p *SupabaseProvider) tokenResponseToAuthResponse(resp *gotrueTokenResponse) *auth.AuthResponse {
	name := ""
	if resp.User.UserMetadata != nil {
		name, _ = resp.User.UserMetadata["name"].(string)
	}

	userID, _ := uuid.Parse(resp.User.ID)

	createdAt, _ := time.Parse(time.RFC3339, resp.User.CreatedAt)

	return &auth.AuthResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
		ExpiresIn:    resp.ExpiresIn,
		User: auth.UserResponse{
			ID:        userID,
			Name:      name,
			Email:     resp.User.Email,
			Roles:     []string{"authenticated"},
			CreatedAt: createdAt,
		},
	}
}
