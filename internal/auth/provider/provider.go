// Package provider defines the interface for authentication providers.
// This allows adding new providers (Google, Facebook, Email) without changing core auth logic.
package provider

import (
	"context"
)

// UserInfo contains standardized user data from any authentication provider.
// All providers (OAuth or email/password) return this common structure.
type UserInfo struct {
	ProviderID    string // Unique ID from the provider (e.g., Google's sub claim)
	Email         string
	Name          string
	AvatarURL     string
	EmailVerified bool
}

// Provider defines the interface for authentication providers.
// Implement this interface to add new auth methods (OAuth providers, email/password, etc.)
type Provider interface {
	// Name returns the provider identifier (e.g., "google", "facebook", "email").
	// This is stored in the database and used for routing.
	Name() string

	// Type returns the provider type for categorization.
	Type() Type
}

// OAuthProvider extends Provider for OAuth2-based authentication.
// Google, Facebook, GitHub, etc. implement this interface.
type OAuthProvider interface {
	Provider

	// AuthURL returns the URL to redirect users for OAuth authentication.
	// The state parameter should be a random string to prevent CSRF.
	AuthURL(state string) string

	// Exchange trades an authorization code for user information.
	// Called after the user is redirected back from the OAuth provider.
	Exchange(ctx context.Context, code string) (*UserInfo, error)
}

// CredentialsProvider extends Provider for username/password authentication.
// Email/password login would implement this interface (future).
type CredentialsProvider interface {
	Provider

	// Authenticate validates credentials and returns user info.
	Authenticate(ctx context.Context, email, password string) (*UserInfo, error)

	// Register creates a new user with email/password.
	Register(ctx context.Context, email, password, name string) (*UserInfo, error)

	// VerifyEmail confirms the user's email address.
	VerifyEmail(ctx context.Context, token string) error

	// RequestPasswordReset initiates password reset flow.
	RequestPasswordReset(ctx context.Context, email string) error

	// ResetPassword completes password reset with token.
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// Type represents the category of authentication provider.
type Type string

const (
	TypeOAuth       Type = "oauth"
	TypeCredentials Type = "credentials"
)

// Registry holds all registered authentication providers.
// Use this to look up providers by name.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// GetOAuth retrieves an OAuth provider by name.
// Returns nil if not found or not an OAuth provider.
func (r *Registry) GetOAuth(name string) (OAuthProvider, bool) {
	p, ok := r.providers[name]
	if !ok {
		return nil, false
	}
	oauth, ok := p.(OAuthProvider)
	return oauth, ok
}

// GetCredentials retrieves a credentials provider by name.
// Returns nil if not found or not a credentials provider.
func (r *Registry) GetCredentials(name string) (CredentialsProvider, bool) {
	p, ok := r.providers[name]
	if !ok {
		return nil, false
	}
	creds, ok := p.(CredentialsProvider)
	return creds, ok
}

// List returns all registered provider names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ListOAuth returns names of all OAuth providers.
func (r *Registry) ListOAuth() []string {
	var names []string
	for name, p := range r.providers {
		if _, ok := p.(OAuthProvider); ok {
			names = append(names, name)
		}
	}
	return names
}

