package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	googleOAuth "golang.org/x/oauth2/google"

	"paas-core/apps/api/internal/config"
)

// ProviderUser holds the user profile returned by an OAuth provider.
type ProviderUser struct {
	ID        string // provider's unique user ID
	Email     string
	Name      string
	AvatarURL string
}

// Provider defines the interface for an OAuth identity provider.
type Provider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*ProviderUser, *oauth2.Token, error)
	Name() string
}

// --- Google Provider ---

type googleProvider struct {
	config *oauth2.Config
}

// NewGoogleProvider creates a Google OAuth provider.
func NewGoogleProvider(cfg config.OAuthProviderConfig, baseURL string) Provider {
	return &googleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  baseURL + "/api/v1/auth/oauth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     googleOAuth.Endpoint,
		},
	}
}

func (g *googleProvider) Name() string { return "google" }

func (g *googleProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
}

func (g *googleProvider) ExchangeCode(ctx context.Context, code string) (*ProviderUser, *oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("google code exchange failed: %w", err)
	}

	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, nil, fmt.Errorf("google userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		Picture   string `json:"picture"`
		Verified  bool   `json:"verified_email"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, nil, fmt.Errorf("google userinfo parse failed: %w", err)
	}

	return &ProviderUser{
		ID:        info.ID,
		Email:     info.Email,
		Name:      info.Name,
		AvatarURL: info.Picture,
	}, token, nil
}

// --- GitHub Provider ---

type githubProvider struct {
	config *oauth2.Config
}

// NewGitHubProvider creates a GitHub OAuth provider.
func NewGitHubProvider(cfg config.OAuthProviderConfig, baseURL string) Provider {
	return &githubProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  baseURL + "/api/v1/auth/oauth/github/callback",
			Scopes:       []string{"user:email", "read:user"},
			Endpoint:     github.Endpoint,
		},
	}
}

func (g *githubProvider) Name() string { return "github" }

func (g *githubProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}

func (g *githubProvider) ExchangeCode(ctx context.Context, code string) (*ProviderUser, *oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("github code exchange failed: %w", err)
	}

	client := g.config.Client(ctx, token)

	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, nil, fmt.Errorf("github user request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var user struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, nil, fmt.Errorf("github user parse failed: %w", err)
	}

	// If email is private, fetch from /user/emails
	email := user.Email
	if email == "" {
		email, _ = g.fetchPrimaryEmail(ctx, client)
	}

	name := user.Name
	if name == "" {
		name = user.Login
	}

	return &ProviderUser{
		ID:        fmt.Sprintf("%d", user.ID),
		Email:     email,
		Name:      name,
		AvatarURL: user.AvatarURL,
	}, token, nil
}

func (g *githubProvider) fetchPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no verified email found")
}
