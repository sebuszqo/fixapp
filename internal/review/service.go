package review

import (
	"context"

	"fixapp/internal/auth"
	"fixapp/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// JobRepository is the subset of job.Repository needed by the review service.
type JobRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)
}

// ScoringService is the subset of scoring needed for ProScore updates.
type ScoringService interface {
	OnReviewSubmitted(ctx context.Context, revieweeID uuid.UUID) error
}

// Service handles review business logic.
type Service struct {
	repo    Repository
	jobRepo JobRepository
	scoring ScoringService
	logger  *zap.Logger
}

// NewService creates a new review service.
func NewService(repo Repository, jobRepo JobRepository, scoring ScoringService, logger *zap.Logger) *Service {
	return &Service{
		repo:    repo,
		jobRepo: jobRepo,
		scoring: scoring,
		logger:  logger,
	}
}

// Create creates a new review for a completed job.
func (s *Service) Create(ctx context.Context, req CreateReviewRequest) (*domain.Review, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	reviewerID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	jobID, err := uuid.Parse(req.JobID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	// Load the job to validate
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Job must be completed (done or confirmed)
	if job.Status != domain.JobStatusDone {
		return nil, domain.ErrJobNotCompleted
	}

	// Determine review type and reviewee based on who the reviewer is
	var reviewType domain.ReviewType
	var revieweeID uuid.UUID

	switch {
	case job.ClientID == reviewerID:
		// Client reviewing the handyman
		if job.CompletedByID == nil {
			return nil, domain.ErrNotJobParticipant
		}
		reviewType = domain.ReviewTypeClientToHandyman
		revieweeID = *job.CompletedByID
	case job.CompletedByID != nil && *job.CompletedByID == reviewerID:
		// Handyman reviewing the client
		reviewType = domain.ReviewTypeHandymanToClient
		revieweeID = job.ClientID
	default:
		return nil, domain.ErrNotJobParticipant
	}

	// Check for duplicate review
	_, err = s.repo.GetByJobAndReviewer(ctx, jobID, reviewerID)
	if err == nil {
		return nil, domain.ErrReviewAlreadyExists
	}
	if err != domain.ErrReviewNotFound {
		return nil, err
	}

	// Create the review
	review := domain.NewReview(jobID, reviewerID, revieweeID, reviewType, req.Rating, req.Comment)
	if err := review.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}

	s.logger.Info("review created",
		zap.String("review_id", review.ID.String()),
		zap.String("job_id", jobID.String()),
		zap.String("reviewer_id", reviewerID.String()),
		zap.String("reviewee_id", revieweeID.String()),
		zap.String("type", reviewType.String()),
		zap.Int("rating", req.Rating),
	)

	// Trigger score recalculation for the reviewee
	if s.scoring != nil {
		if err := s.scoring.OnReviewSubmitted(ctx, revieweeID); err != nil {
			s.logger.Error("failed to recalculate score after review",
				zap.String("reviewee_id", revieweeID.String()),
				zap.Error(err),
			)
			// Don't fail the review creation
		}
	}

	return review, nil
}

// GetByID retrieves a review by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByReviewee lists reviews received by a user.
func (s *Service) ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit, offset int) ([]*domain.Review, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListByReviewee(ctx, revieweeID, limit, offset)
}

// ListByJob lists all reviews for a specific job.
func (s *Service) ListByJob(ctx context.Context, jobID uuid.UUID) ([]*domain.Review, error) {
	return s.repo.ListByJob(ctx, jobID)
}

// GetRatingSummary returns the average rating and count for a user.
func (s *Service) GetRatingSummary(ctx context.Context, userID uuid.UUID) (float64, int, error) {
	return s.repo.GetAverageRating(ctx, userID)
}
