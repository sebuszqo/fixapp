package review

import (
	"context"
	"database/sql"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// PostgresRepository implements Repository with PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL review repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, job_id, reviewer_id, reviewee_id, type, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		review.ID,
		review.JobID,
		review.ReviewerID,
		review.RevieweeID,
		review.Type,
		review.Rating,
		review.Comment,
		review.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrReviewAlreadyExists
		}
		return err
	}
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	query := `
		SELECT id, job_id, reviewer_id, reviewee_id, type, rating, comment, created_at
		FROM reviews WHERE id = $1`

	review := &domain.Review{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&review.ID,
		&review.JobID,
		&review.ReviewerID,
		&review.RevieweeID,
		&review.Type,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrReviewNotFound
	}
	if err != nil {
		return nil, err
	}
	return review, nil
}

func (r *PostgresRepository) GetByJobAndReviewer(ctx context.Context, jobID, reviewerID uuid.UUID) (*domain.Review, error) {
	query := `
		SELECT id, job_id, reviewer_id, reviewee_id, type, rating, comment, created_at
		FROM reviews WHERE job_id = $1 AND reviewer_id = $2`

	review := &domain.Review{}
	err := r.db.QueryRowContext(ctx, query, jobID, reviewerID).Scan(
		&review.ID,
		&review.JobID,
		&review.ReviewerID,
		&review.RevieweeID,
		&review.Type,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrReviewNotFound
	}
	if err != nil {
		return nil, err
	}
	return review, nil
}

func (r *PostgresRepository) ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit, offset int) ([]*domain.Review, int64, error) {
	countQuery := `SELECT COUNT(*) FROM reviews WHERE reviewee_id = $1`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, revieweeID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, job_id, reviewer_id, reviewee_id, type, rating, comment, created_at
		FROM reviews WHERE reviewee_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, revieweeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		review := &domain.Review{}
		if err := rows.Scan(
			&review.ID,
			&review.JobID,
			&review.ReviewerID,
			&review.RevieweeID,
			&review.Type,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, review)
	}
	return reviews, total, rows.Err()
}

func (r *PostgresRepository) ListByJob(ctx context.Context, jobID uuid.UUID) ([]*domain.Review, error) {
	query := `
		SELECT id, job_id, reviewer_id, reviewee_id, type, rating, comment, created_at
		FROM reviews WHERE job_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		review := &domain.Review{}
		if err := rows.Scan(
			&review.ID,
			&review.JobID,
			&review.ReviewerID,
			&review.RevieweeID,
			&review.Type,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

func (r *PostgresRepository) GetAverageRating(ctx context.Context, revieweeID uuid.UUID) (float64, int, error) {
	query := `SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM reviews WHERE reviewee_id = $1`
	var avg float64
	var count int
	err := r.db.QueryRowContext(ctx, query, revieweeID).Scan(&avg, &count)
	return avg, count, err
}

func (r *PostgresRepository) CountByRating(ctx context.Context, revieweeID uuid.UUID, minRating, maxRating int) (int, error) {
	query := `SELECT COUNT(*) FROM reviews WHERE reviewee_id = $1 AND rating >= $2 AND rating <= $3`
	var count int
	err := r.db.QueryRowContext(ctx, query, revieweeID, minRating, maxRating).Scan(&count)
	return count, err
}

// isDuplicateKeyError checks for PostgreSQL unique constraint violation.
func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "unique_review_per_job"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
