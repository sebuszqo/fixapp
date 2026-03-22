// Package auth provides authentication and authorization functionality.
package auth

import (
	"context"
	"time"

	"fixapp/internal/auth/permission"
	"fixapp/internal/auth/role"
)

// ctxKeyUser is a private type for the user context key to avoid collisions.
type ctxKeyUser struct{}

// userKey is the context key for storing User.
var userKey ctxKeyUser

// User represents an authenticated user in the system.
// This struct is attached to the request context after successful authentication.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      role.Role `json:"role"`
	Provider  string    `json:"provider"` // "google", "facebook", "email"
	CreatedAt time.Time `json:"created_at"`
}

// NewContext returns a new context with the user attached.
func NewContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// FromContext retrieves the user from the context.
// Returns nil if no user is present (unauthenticated request).
func FromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userKey).(*User)
	return u
}

// MustFromContext retrieves the user from the context.
// Panics if no user is present. Use only when authentication is guaranteed.
func MustFromContext(ctx context.Context) *User {
	u := FromContext(ctx)
	if u == nil {
		panic("auth: user not found in context")
	}
	return u
}

// IsAuthenticated returns true if a user is present in the context.
func IsAuthenticated(ctx context.Context) bool {
	return FromContext(ctx) != nil
}

// HasRole checks if the authenticated user has the specified role.
func HasRole(ctx context.Context, r role.Role) bool {
	u := FromContext(ctx)
	if u == nil {
		return false
	}
	return u.Role == r
}

// HasRoleAtLeast checks if the authenticated user has at least the specified role level.
func HasRoleAtLeast(ctx context.Context, r role.Role) bool {
	u := FromContext(ctx)
	if u == nil {
		return false
	}
	return u.Role.IsAtLeast(r)
}

// HasPermission checks if the authenticated user has the specified permission.
func HasPermission(ctx context.Context, p permission.Permission) bool {
	u := FromContext(ctx)
	if u == nil {
		return false
	}
	return permission.Has(u.Role, p)
}

// HasAnyPermission checks if the authenticated user has at least one of the permissions.
func HasAnyPermission(ctx context.Context, perms ...permission.Permission) bool {
	u := FromContext(ctx)
	if u == nil {
		return false
	}
	return permission.HasAny(u.Role, perms...)
}

// HasAllPermissions checks if the authenticated user has all of the permissions.
func HasAllPermissions(ctx context.Context, perms ...permission.Permission) bool {
	u := FromContext(ctx)
	if u == nil {
		return false
	}
	return permission.HasAll(u.Role, perms...)
}

// Methods on User for convenience

// HasPermission checks if the user has the specified permission.
func (u *User) HasPermission(p permission.Permission) bool {
	return permission.Has(u.Role, p)
}

// HasAnyPermission checks if the user has at least one of the permissions.
func (u *User) HasAnyPermission(perms ...permission.Permission) bool {
	return permission.HasAny(u.Role, perms...)
}

// HasAllPermissions checks if the user has all of the permissions.
func (u *User) HasAllPermissions(perms ...permission.Permission) bool {
	return permission.HasAll(u.Role, perms...)
}

// IsAdmin returns true if the user has the Admin role.
func (u *User) IsAdmin() bool {
	return u.Role == role.Admin
}

// IsHandyman returns true if the user has the Handyman role.
func (u *User) IsHandyman() bool {
	return u.Role == role.Handyman
}

// IsUser returns true if the user has the User role.
func (u *User) IsUser() bool {
	return u.Role == role.User
}


