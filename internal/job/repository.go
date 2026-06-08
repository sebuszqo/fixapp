// Package job provides job (service request) management functionality.
package job

import (
	"context"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for job data access.
type Repository interface {
	// Create inserts a new job.
	Create(ctx context.Context, job *domain.Job) error

	// GetByID retrieves a job by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)

	// Update modifies an existing job.
	Update(ctx context.Context, job *domain.Job) error

	// List retrieves jobs with pagination and filters.
	List(ctx context.Context, filter ListFilter) ([]*domain.Job, int64, error)

	// ListActiveForMatching retrieves active jobs matching category and district.
	ListActiveForMatching(ctx context.Context, categoryID, districtID uuid.UUID) ([]*domain.Job, error)
}

// ListFilter contains filters for listing jobs.
type ListFilter struct {
	ClientID   *uuid.UUID
	CategoryID *uuid.UUID
	DistrictID *uuid.UUID
	Status     *domain.JobStatus
	Limit      int
	Offset     int
}

// DefaultListFilter returns a filter with sensible defaults.
func DefaultListFilter() ListFilter {
	return ListFilter{
		Limit:  20,
		Offset: 0,
	}
}
