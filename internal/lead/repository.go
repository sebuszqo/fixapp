// Package lead provides lead management functionality.
package lead

import (
	"context"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for lead data access.
type Repository interface {
	// Create inserts a new lead.
	Create(ctx context.Context, lead *domain.Lead) error

	// GetByID retrieves a lead by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error)

	// Update modifies an existing lead.
	Update(ctx context.Context, lead *domain.Lead) error

	// ListByHandyman retrieves leads for a specific handyman with filters.
	ListByHandyman(ctx context.Context, handymanID uuid.UUID, filter ListFilter) ([]*domain.Lead, int64, error)

	// ListByJob retrieves all leads for a specific job.
	ListByJob(ctx context.Context, jobID uuid.UUID) ([]*domain.Lead, error)

	// CountAcceptedByJob returns the number of accepted leads for a job.
	CountAcceptedByJob(ctx context.Context, jobID uuid.UUID) (int, error)

	// ExpirePendingLeads marks all expired pending leads as expired.
	ExpirePendingLeads(ctx context.Context) (int64, error)
}

// ListFilter contains filters for listing leads.
type ListFilter struct {
	Status *domain.LeadStatus
	Limit  int
	Offset int
}

// DefaultListFilter returns a filter with sensible defaults.
func DefaultListFilter() ListFilter {
	return ListFilter{
		Limit:  20,
		Offset: 0,
	}
}
