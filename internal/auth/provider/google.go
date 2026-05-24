package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Ensure GoogleProvider implements OAuthProvider.
var _ OAuthProvider = (*GoogleProvider)(nil)

// GoogleProvider implements OAuth2 authentication with Google.
type GoogleProvider struct {
	config *oauth2.Config
}

// GoogleConfig holds configuration for Google OAuth.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// DefaultGoogleScopes returns the default scopes for Google OAuth.
func DefaultGoogleScopes() []string {
	return []string{
		"openid",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	}
}

// NewGoogleProvider creates a new Google OAuth provider.
func NewGoogleProvider(cfg GoogleConfig) *GoogleProvider {
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = DefaultGoogleScopes()
	}

	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

// Name returns the provider identifier.
func (p *GoogleProvider) Name() string {
	return "google"
}

// Type returns the provider type.
func (p *GoogleProvider) Type() Type {
	return TypeOAuth
}

// AuthURL returns the Google OAuth authorization URL.
func (p *GoogleProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange trades an authorization code for user information.
func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	// Exchange code for token
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google: failed to exchange code: %w", err)
	}

	// Use token to get user info
	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("google: failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google: user info request failed with status %d", resp.StatusCode)
	}

	var googleUser googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("google: failed to decode user info: %w", err)
	}

	return &UserInfo{
		ProviderID:    googleUser.ID,
		Email:         googleUser.Email,
		Name:          googleUser.Name,
		AvatarURL:     googleUser.Picture,
		EmailVerified: googleUser.VerifiedEmail,
	}, nil
}

// googleUserInfo represents the response from Google's userinfo endpoint.
type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

