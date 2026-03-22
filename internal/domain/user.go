// Package domain contains the core business entities.
// These models represent the business domain, independent of infrastructure.
package domain

import (
	"time"

	"fixapp/internal/auth/role"

	"github.com/google/uuid"
)

// AuthProvider represents the authentication provider.
type AuthProvider string

const (
	AuthProviderGoogle   AuthProvider = "google"
	AuthProviderFacebook AuthProvider = "facebook"
	AuthProviderEmail    AuthProvider = "email"
)

// String implements fmt.Stringer.
func (p AuthProvider) String() string {
	return string(p)
}

// IsValid returns true if the provider is known.
func (p AuthProvider) IsValid() bool {
	switch p {
	case AuthProviderGoogle, AuthProviderFacebook, AuthProviderEmail:
		return true
	default:
		return false
	}
}

// User represents a user in the domain layer.
// This is the core business entity, used by services and repositories.
type User struct {
	ID            uuid.UUID
	Email         string
	Name          string
	Role          role.Role
	Provider      AuthProvider
	ProviderID    string // External provider's user ID
	PasswordHash  string // Only for email provider
	AvatarURL     string
	Phone         string
	IsActive      bool
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
}

// NewUser creates a new user with sensible defaults.
func NewUser(email, name string, provider AuthProvider, providerID string) *User {
	now := time.Now()
	return &User{
		ID:            uuid.New(),
		Email:         email,
		Name:          name,
		Role:          role.User, // Default role
		Provider:      provider,
		ProviderID:    providerID,
		IsActive:      true,
		EmailVerified: provider != AuthProviderEmail, // SSO providers are pre-verified
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// CanLogin returns true if the user is allowed to authenticate.
func (u *User) CanLogin() bool {
	return u.IsActive
}

// RequiresEmailVerification returns true if the user needs to verify email.
func (u *User) RequiresEmailVerification() bool {
	return u.Provider == AuthProviderEmail && !u.EmailVerified
}

// UpdateLastLogin sets the last login timestamp.
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// SetRole changes the user's role.
func (u *User) SetRole(r role.Role) {
	u.Role = r
}

// Deactivate disables the user account.
func (u *User) Deactivate() {
	u.IsActive = false
}

// Activate enables the user account.
func (u *User) Activate() {
	u.IsActive = true
}

// VerifyEmail marks the email as verified.
func (u *User) VerifyEmail() {
	u.EmailVerified = true
}

