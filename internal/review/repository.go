package review

import (
	"context"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for review data access.
type Repository interface {
	Create(ctx context.Context, review *domain.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	GetByJobAndReviewer(ctx context.Context, jobID, reviewerID uuid.UUID) (*domain.Review, error)
	ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit, offset int) ([]*domain.Review, int64, error)
	ListByJob(ctx context.Context, jobID uuid.UUID) ([]*domain.Review, error)
	GetAverageRating(ctx context.Context, revieweeID uuid.UUID) (float64, int, error) // avg, count
	CountByRating(ctx context.Context, revieweeID uuid.UUID, minRating, maxRating int) (int, error)
}
