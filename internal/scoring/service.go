package scoring

import (
	"context"

	"fixapp/internal/auth"
	"fixapp/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ReviewCounter provides review statistics needed for scoring.
type ReviewCounter interface {
	CountByRating(ctx context.Context, revieweeID uuid.UUID, minRating, maxRating int) (int, error)
}

// Service handles scoring business logic.
type Service struct {
	repo          Repository
	reviewCounter ReviewCounter
	logger        *zap.Logger
}

// NewService creates a new scoring service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// SetReviewCounter sets the review counter (called after init to break circular dep).
func (s *Service) SetReviewCounter(rc ReviewCounter) {
	s.reviewCounter = rc
}

// OnReviewSubmitted recalculates the reviewee's ProScore after a new review.
func (s *Service) OnReviewSubmitted(ctx context.Context, revieweeID uuid.UUID) error {
	if s.reviewCounter == nil {
		return nil
	}

	ps, err := s.repo.GetProScore(ctx, revieweeID)
	if err != nil {
		return err
	}

	// Update review-based factors
	fiveStarCount, err := s.reviewCounter.CountByRating(ctx, revieweeID, 5, 5)
	if err != nil {
		return err
	}
	ps.FiveStarReviews = fiveStarCount

	lowRatingCount, err := s.reviewCounter.CountByRating(ctx, revieweeID, 1, 2)
	if err != nil {
		return err
	}
	ps.LowRatingCount = lowRatingCount

	// Recalculate
	ps.Score = CalculateProScore(ps)

	if err := s.repo.UpsertProScore(ctx, ps); err != nil {
		return err
	}

	s.logger.Info("pro score updated after review",
		zap.String("user_id", revieweeID.String()),
		zap.Int("new_score", ps.Score),
		zap.Int("five_star_reviews", fiveStarCount),
		zap.Int("low_rating_count", lowRatingCount),
	)

	return nil
}

// GetMyCommitScore retrieves the authenticated client's Commit Score.
func (s *Service) GetMyCommitScore(ctx context.Context) (*domain.CommitScore, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.GetCommitScore(ctx, userID)
}

// GetMyProScore retrieves the authenticated handyman's ProScore.
func (s *Service) GetMyProScore(ctx context.Context) (*domain.ProScore, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.GetProScore(ctx, userID)
}

// GetCommitScoreByUserID retrieves a user's Commit Score (for lead creation).
func (s *Service) GetCommitScoreByUserID(ctx context.Context, userID uuid.UUID) (*domain.CommitScore, error) {
	return s.repo.GetCommitScore(ctx, userID)
}

// GetProScoreByUserID retrieves a handyman's ProScore (for lead pricing).
func (s *Service) GetProScoreByUserID(ctx context.Context, userID uuid.UUID) (*domain.ProScore, error) {
	return s.repo.GetProScore(ctx, userID)
}

// RecalculateCommitScore recalculates and saves a client's Commit Score.
// Call this after events that affect the score (job completed, no-show, profile update).
func (s *Service) RecalculateCommitScore(ctx context.Context, userID uuid.UUID, factors CommitScoreFactors) error {
	cs, err := s.repo.GetCommitScore(ctx, userID)
	if err != nil {
		return err
	}

	// Update factors
	cs.UserID = userID
	cs.PhoneVerified = factors.PhoneVerified
	cs.ProfileComplete = factors.ProfileComplete
	cs.HasAvatar = factors.HasAvatar
	cs.HasJobHistory = factors.JobsCompleted > 0
	cs.NoNoShows = factors.NoShowCount == 0
	cs.NoExcessCancels = factors.JobsCancelled < 2
	cs.JobsCompleted = factors.JobsCompleted
	cs.JobsCancelled = factors.JobsCancelled
	cs.NoShowCount = factors.NoShowCount

	// Calculate new score
	cs.Score = CalculateCommitScore(cs)

	if err := s.repo.UpsertCommitScore(ctx, cs); err != nil {
		return err
	}

	s.logger.Info("commit score recalculated",
		zap.String("user_id", userID.String()),
		zap.Int("new_score", cs.Score),
		zap.String("level", string(cs.Level())),
	)

	return nil
}

// RecalculateProScore recalculates and saves a handyman's ProScore.
// Call this after events that affect the score (job done, review received, etc.).
func (s *Service) RecalculateProScore(ctx context.Context, userID uuid.UUID, factors ProScoreFactors) error {
	ps, err := s.repo.GetProScore(ctx, userID)
	if err != nil {
		return err
	}

	// Update factors
	ps.UserID = userID
	ps.JobsCompleted = factors.JobsCompleted
	ps.FiveStarReviews = factors.FiveStarReviews
	ps.AvgResponseMins = factors.AvgResponseMins
	ps.ProfileComplete = factors.ProfileComplete
	ps.ActiveLast7Days = factors.ActiveLast7Days
	ps.PortfolioCount = factors.PortfolioCount
	ps.NoShowCount = factors.NoShowCount
	ps.CancelledAfterAccept = factors.CancelledAfterAccept
	ps.SlowResponseCount = factors.SlowResponseCount
	ps.LowRatingCount = factors.LowRatingCount

	// Calculate new score
	ps.Score = CalculateProScore(ps)

	if err := s.repo.UpsertProScore(ctx, ps); err != nil {
		return err
	}

	s.logger.Info("pro score recalculated",
		zap.String("user_id", userID.String()),
		zap.Int("new_score", ps.Score),
		zap.String("level", string(ps.Level())),
	)

	return nil
}

// CalculateLeadPrice computes the dynamic lead price.
// Price = BasePrice × ClientMultiplier × HandymanMultiplier
func (s *Service) CalculateLeadPrice(ctx context.Context, basePriceCredits int, clientID, handymanID uuid.UUID) (int, error) {
	cs, err := s.repo.GetCommitScore(ctx, clientID)
	if err != nil {
		return 0, err
	}

	ps, err := s.repo.GetProScore(ctx, handymanID)
	if err != nil {
		return 0, err
	}

	price := float64(basePriceCredits) * cs.ClientMultiplier() * ps.HandymanMultiplier()

	// Round to nearest integer, minimum 1
	result := int(price + 0.5)
	if result < 1 {
		result = 1
	}

	return result, nil
}

// CommitScoreFactors are the inputs needed to recalculate a Commit Score.
type CommitScoreFactors struct {
	PhoneVerified   bool
	ProfileComplete bool
	HasAvatar       bool
	JobsCompleted   int
	JobsCancelled   int
	NoShowCount     int
}

// ProScoreFactors are the inputs needed to recalculate a ProScore.
type ProScoreFactors struct {
	JobsCompleted        int
	FiveStarReviews      int
	AvgResponseMins      int
	ProfileComplete      bool
	ActiveLast7Days      bool
	PortfolioCount       int
	NoShowCount          int
	CancelledAfterAccept int
	SlowResponseCount    int
	LowRatingCount       int
}
