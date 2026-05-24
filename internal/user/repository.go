// Package user provides user management functionality.
package user

import (
	"context"

	"fixapp/internal/auth/role"
	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for user data access.
// This interface allows for easy testing with mocks and swapping implementations.
type Repository interface {
	// Create inserts a new user.
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// GetByEmail retrieves a user by their email.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByProvider retrieves a user by their provider and provider ID.
	GetByProvider(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error)

	// Update modifies an existing user.
	Update(ctx context.Context, user *domain.User) error

	// Delete removes a user (soft delete by setting is_active = false).
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves users with pagination and optional filters.
	List(ctx context.Context, filter ListFilter) ([]*domain.User, int64, error)

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// UpdateLastLogin updates the last login timestamp.
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// UpdateRole changes a user's role.
	UpdateRole(ctx context.Context, id uuid.UUID, role role.Role) error
}

// ListFilter contains filters for listing users.
type ListFilter struct {
	Role     *role.Role
	IsActive *bool
	Search   string // Searches email and name
	Limit    int
	Offset   int
}

// DefaultListFilter returns a filter with sensible defaults.
func DefaultListFilter() ListFilter {
	return ListFilter{
		Limit:  20,
		Offset: 0,
	}
}

