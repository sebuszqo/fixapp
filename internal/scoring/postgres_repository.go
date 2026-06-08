package scoring

import (
	"context"
	"database/sql"
	"errors"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL scoring repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetCommitScore retrieves a client's commit score.
func (r *PostgresRepository) GetCommitScore(ctx context.Context, userID uuid.UUID) (*domain.CommitScore, error) {
	query := `
		SELECT id, user_id, score,
			phone_verified, profile_complete, has_avatar,
			has_job_history, no_no_shows, no_excess_cancels,
			jobs_completed, jobs_cancelled, no_show_count,
			updated_at
		FROM commit_scores WHERE user_id = $1`

	cs := &domain.CommitScore{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&cs.ID, &cs.UserID, &cs.Score,
		&cs.PhoneVerified, &cs.ProfileComplete, &cs.HasAvatar,
		&cs.HasJobHistory, &cs.NoNoShows, &cs.NoExcessCancels,
		&cs.JobsCompleted, &cs.JobsCancelled, &cs.NoShowCount,
		&cs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return a default zero score
			return &domain.CommitScore{
				ID:     uuid.New(),
				UserID: userID,
				Score:  0,
			}, nil
		}
		return nil, err
	}
	return cs, nil
}

// UpsertCommitScore creates or updates a commit score.
func (r *PostgresRepository) UpsertCommitScore(ctx context.Context, score *domain.CommitScore) error {
	query := `
		INSERT INTO commit_scores (
			id, user_id, score,
			phone_verified, profile_complete, has_avatar,
			has_job_history, no_no_shows, no_excess_cancels,
			jobs_completed, jobs_cancelled, no_show_count,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			score = EXCLUDED.score,
			phone_verified = EXCLUDED.phone_verified,
			profile_complete = EXCLUDED.profile_complete,
			has_avatar = EXCLUDED.has_avatar,
			has_job_history = EXCLUDED.has_job_history,
			no_no_shows = EXCLUDED.no_no_shows,
			no_excess_cancels = EXCLUDED.no_excess_cancels,
			jobs_completed = EXCLUDED.jobs_completed,
			jobs_cancelled = EXCLUDED.jobs_cancelled,
			no_show_count = EXCLUDED.no_show_count,
			updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query,
		score.ID, score.UserID, score.Score,
		score.PhoneVerified, score.ProfileComplete, score.HasAvatar,
		score.HasJobHistory, score.NoNoShows, score.NoExcessCancels,
		score.JobsCompleted, score.JobsCancelled, score.NoShowCount,
	)
	return err
}

// GetProScore retrieves a handyman's pro score.
func (r *PostgresRepository) GetProScore(ctx context.Context, userID uuid.UUID) (*domain.ProScore, error) {
	query := `
		SELECT id, user_id, score,
			jobs_completed, five_star_reviews, avg_response_mins,
			profile_complete, active_last_7_days, portfolio_count,
			no_show_count, cancelled_after_accept, slow_response_count, low_rating_count,
			updated_at
		FROM pro_scores WHERE user_id = $1`

	ps := &domain.ProScore{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&ps.ID, &ps.UserID, &ps.Score,
		&ps.JobsCompleted, &ps.FiveStarReviews, &ps.AvgResponseMins,
		&ps.ProfileComplete, &ps.ActiveLast7Days, &ps.PortfolioCount,
		&ps.NoShowCount, &ps.CancelledAfterAccept, &ps.SlowResponseCount, &ps.LowRatingCount,
		&ps.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.ProScore{
				ID:     uuid.New(),
				UserID: userID,
				Score:  0,
			}, nil
		}
		return nil, err
	}
	return ps, nil
}

// UpsertProScore creates or updates a pro score.
func (r *PostgresRepository) UpsertProScore(ctx context.Context, score *domain.ProScore) error {
	query := `
		INSERT INTO pro_scores (
			id, user_id, score,
			jobs_completed, five_star_reviews, avg_response_mins,
			profile_complete, active_last_7_days, portfolio_count,
			no_show_count, cancelled_after_accept, slow_response_count, low_rating_count,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			score = EXCLUDED.score,
			jobs_completed = EXCLUDED.jobs_completed,
			five_star_reviews = EXCLUDED.five_star_reviews,
			avg_response_mins = EXCLUDED.avg_response_mins,
			profile_complete = EXCLUDED.profile_complete,
			active_last_7_days = EXCLUDED.active_last_7_days,
			portfolio_count = EXCLUDED.portfolio_count,
			no_show_count = EXCLUDED.no_show_count,
			cancelled_after_accept = EXCLUDED.cancelled_after_accept,
			slow_response_count = EXCLUDED.slow_response_count,
			low_rating_count = EXCLUDED.low_rating_count,
			updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query,
		score.ID, score.UserID, score.Score,
		score.JobsCompleted, score.FiveStarReviews, score.AvgResponseMins,
		score.ProfileComplete, score.ActiveLast7Days, score.PortfolioCount,
		score.NoShowCount, score.CancelledAfterAccept, score.SlowResponseCount, score.LowRatingCount,
	)
	return err
}
