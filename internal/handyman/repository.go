// Package handyman provides handyman profile management functionality.
package handyman

import (
	"context"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for handyman profile data access.
type Repository interface {
	// CreateProfile creates a new handyman profile.
	CreateProfile(ctx context.Context, profile *domain.HandymanProfile) error

	// GetByUserID retrieves a profile by user ID.
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.HandymanProfile, error)

	// GetByID retrieves a profile by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.HandymanProfile, error)

	// Update modifies an existing profile.
	Update(ctx context.Context, profile *domain.HandymanProfile) error

	// Search finds handyman profiles matching criteria.
	Search(ctx context.Context, filter SearchFilter) ([]*domain.HandymanProfile, int64, error)

	// FindMatchingForJob finds available handymen matching a job's category and district.
	// Used by the dispatch system to create leads.
	FindMatchingForJob(ctx context.Context, categoryID, districtID uuid.UUID, emergency bool) ([]*domain.HandymanProfile, error)

	// Pricing
	CreatePricingItem(ctx context.Context, item *domain.PricingItem) error
	UpdatePricingItem(ctx context.Context, item *domain.PricingItem) error
	DeletePricingItem(ctx context.Context, id uuid.UUID) error
	ListPricing(ctx context.Context, profileID uuid.UUID) ([]*domain.PricingItem, error)

	// Portfolio
	CreatePortfolioItem(ctx context.Context, item *domain.PortfolioItem) error
	DeletePortfolioItem(ctx context.Context, id uuid.UUID) error
	ListPortfolio(ctx context.Context, profileID uuid.UUID) ([]*domain.PortfolioItem, error)
}

// SearchFilter contains filters for searching handyman profiles.
type SearchFilter struct {
	CategoryID *uuid.UUID
	DistrictID *uuid.UUID
	IsVerified *bool
	Available  *bool
	Search     string // search by company name or bio
	Limit      int
	Offset     int
}

// DefaultSearchFilter returns a filter with sensible defaults.
func DefaultSearchFilter() SearchFilter {
	return SearchFilter{
		Limit:  20,
		Offset: 0,
	}
}
