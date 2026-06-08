package lead

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL lead repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create inserts a new lead.
func (r *PostgresRepository) Create(ctx context.Context, lead *domain.Lead) error {
	query := `
		INSERT INTO leads (
			id, job_id, handyman_id, status,
			price, client_commit_score,
			created_at, updated_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		lead.ID, lead.JobID, lead.HandymanID, lead.Status,
		lead.Price, lead.ClientCommitScore,
		lead.CreatedAt, lead.UpdatedAt, lead.ExpiresAt,
	)
	return err
}

// GetByID retrieves a lead by its ID.
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	query := `
		SELECT id, job_id, handyman_id, status,
			price, client_commit_score,
			created_at, updated_at, expires_at, accepted_at, rejected_at
		FROM leads WHERE id = $1`

	lead := &domain.Lead{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lead.ID, &lead.JobID, &lead.HandymanID, &lead.Status,
		&lead.Price, &lead.ClientCommitScore,
		&lead.CreatedAt, &lead.UpdatedAt, &lead.ExpiresAt, &lead.AcceptedAt, &lead.RejectedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrLeadNotFound
		}
		return nil, err
	}
	return lead, nil
}

// Update modifies an existing lead.
func (r *PostgresRepository) Update(ctx context.Context, lead *domain.Lead) error {
	query := `
		UPDATE leads SET
			status = $2, accepted_at = $3, rejected_at = $4
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		lead.ID, lead.Status, lead.AcceptedAt, lead.RejectedAt,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrLeadNotFound
	}
	return nil
}

// ListByHandyman retrieves leads for a specific handyman with filters.
func (r *PostgresRepository) ListByHandyman(ctx context.Context, handymanID uuid.UUID, filter ListFilter) ([]*domain.Lead, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("handyman_id = $%d", argIdx))
	args = append(args, handymanID)
	argIdx++

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM leads %s", where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch leads
	query := fmt.Sprintf(`
		SELECT id, job_id, handyman_id, status,
			price, client_commit_score,
			created_at, updated_at, expires_at, accepted_at, rejected_at
		FROM leads %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var leads []*domain.Lead
	for rows.Next() {
		lead := &domain.Lead{}
		if err := rows.Scan(
			&lead.ID, &lead.JobID, &lead.HandymanID, &lead.Status,
			&lead.Price, &lead.ClientCommitScore,
			&lead.CreatedAt, &lead.UpdatedAt, &lead.ExpiresAt, &lead.AcceptedAt, &lead.RejectedAt,
		); err != nil {
			return nil, 0, err
		}
		leads = append(leads, lead)
	}

	return leads, total, rows.Err()
}

// ListByJob retrieves all leads for a specific job.
func (r *PostgresRepository) ListByJob(ctx context.Context, jobID uuid.UUID) ([]*domain.Lead, error) {
	query := `
		SELECT id, job_id, handyman_id, status,
			price, client_commit_score,
			created_at, updated_at, expires_at, accepted_at, rejected_at
		FROM leads WHERE job_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leads []*domain.Lead
	for rows.Next() {
		lead := &domain.Lead{}
		if err := rows.Scan(
			&lead.ID, &lead.JobID, &lead.HandymanID, &lead.Status,
			&lead.Price, &lead.ClientCommitScore,
			&lead.CreatedAt, &lead.UpdatedAt, &lead.ExpiresAt, &lead.AcceptedAt, &lead.RejectedAt,
		); err != nil {
			return nil, err
		}
		leads = append(leads, lead)
	}

	return leads, rows.Err()
}

// CountAcceptedByJob returns the number of accepted leads for a job.
func (r *PostgresRepository) CountAcceptedByJob(ctx context.Context, jobID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM leads WHERE job_id = $1 AND status = 'accepted'`
	var count int
	err := r.db.QueryRowContext(ctx, query, jobID).Scan(&count)
	return count, err
}

// ExpirePendingLeads marks all expired pending leads as expired.
func (r *PostgresRepository) ExpirePendingLeads(ctx context.Context) (int64, error) {
	query := `
		UPDATE leads SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
