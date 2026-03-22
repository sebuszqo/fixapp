package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

// Ensure FacebookProvider implements OAuthProvider.
var _ OAuthProvider = (*FacebookProvider)(nil)

// FacebookProvider implements OAuth2 authentication with Facebook.
type FacebookProvider struct {
	config *oauth2.Config
}

// FacebookConfig holds configuration for Facebook OAuth.
type FacebookConfig struct {
	ClientID     string // Facebook App ID
	ClientSecret string // Facebook App Secret
	RedirectURL  string
	Scopes       []string
}

// DefaultFacebookScopes returns the default scopes for Facebook OAuth.
func DefaultFacebookScopes() []string {
	return []string{
		"email",
		"public_profile",
	}
}

// NewFacebookProvider creates a new Facebook OAuth provider.
func NewFacebookProvider(cfg FacebookConfig) *FacebookProvider {
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = DefaultFacebookScopes()
	}

	return &FacebookProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       scopes,
			Endpoint:     facebook.Endpoint,
		},
	}
}

// Name returns the provider identifier.
func (p *FacebookProvider) Name() string {
	return "facebook"
}

// Type returns the provider type.
func (p *FacebookProvider) Type() Type {
	return TypeOAuth
}

// AuthURL returns the Facebook OAuth authorization URL.
func (p *FacebookProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

// Exchange trades an authorization code for user information.
func (p *FacebookProvider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	// Exchange code for token
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("facebook: failed to exchange code: %w", err)
	}

	// Use token to get user info
	// Facebook requires fields parameter to specify what data to return
	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,picture.type(large)")
	if err != nil {
		return nil, fmt.Errorf("facebook: failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("facebook: user info request failed with status %d", resp.StatusCode)
	}

	var fbUser facebookUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&fbUser); err != nil {
		return nil, fmt.Errorf("facebook: failed to decode user info: %w", err)
	}

	// Extract avatar URL from nested picture object
	avatarURL := ""
	if fbUser.Picture.Data.URL != "" {
		avatarURL = fbUser.Picture.Data.URL
	}

	return &UserInfo{
		ProviderID:    fbUser.ID,
		Email:         fbUser.Email,
		Name:          fbUser.Name,
		AvatarURL:     avatarURL,
		EmailVerified: true, // Facebook emails are verified
	}, nil
}

// facebookUserInfo represents the response from Facebook's me endpoint.
type facebookUserInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

