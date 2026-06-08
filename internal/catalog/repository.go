// Package catalog provides reference data (categories, districts) access.
package catalog

import (
	"context"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for catalog data access.
type Repository interface {
	// Categories
	ListCategories(ctx context.Context, activeOnly bool) ([]*domain.ServiceCategory, error)
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.ServiceCategory, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*domain.ServiceCategory, error)

	// Districts
	ListDistricts(ctx context.Context, activeOnly bool) ([]*domain.District, error)
	GetDistrictByID(ctx context.Context, id uuid.UUID) (*domain.District, error)
	GetDistrictBySlug(ctx context.Context, slug string) (*domain.District, error)
}
