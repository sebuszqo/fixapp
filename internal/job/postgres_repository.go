package job

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"fixapp/internal/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL job repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create inserts a new job.
func (r *PostgresRepository) Create(ctx context.Context, job *domain.Job) error {
	photoURLs, err := json.Marshal(job.PhotoURLs)
	if err != nil {
		photoURLs = []byte("[]")
	}

	query := `
		INSERT INTO jobs (
			id, client_id, category_id, district_id,
			title, description, urgency, status,
			address, building_type, floor, has_elevator,
			preferred_date1, preferred_date2, preferred_time,
			budget, wants_invoice, contact_method,
			photo_urls, expires_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15,
			$16, $17, $18,
			$19, $20,
			$21, $22
		)`

	_, err = r.db.ExecContext(ctx, query,
		job.ID, job.ClientID, job.CategoryID, job.DistrictID,
		job.Title, job.Description, job.Urgency, job.Status,
		nullString(job.Address), nullString(job.BuildingType), job.Floor, job.HasElevator,
		job.PreferredDate1, job.PreferredDate2, nullString(job.PreferredTime),
		job.Budget, job.WantsInvoice, job.ContactMethod,
		photoURLs, job.ExpiresAt,
		job.CreatedAt, job.UpdatedAt,
	)
	return err
}

// GetByID retrieves a job by its ID.
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	query := `
		SELECT j.id, j.client_id, j.category_id, j.district_id,
			j.title, j.description, j.urgency, j.status,
			j.address, j.building_type, j.floor, j.has_elevator,
			j.preferred_date1, j.preferred_date2, j.preferred_time,
			j.budget, j.wants_invoice, j.contact_method,
			j.photo_urls, j.final_value, j.completed_at, j.completed_by_id, j.client_confirmed,
			j.created_at, j.updated_at, j.expires_at,
			l.estimated_price, COALESCE(l.arrival_time, ''), COALESCE(l.proposal_message, '')
		FROM jobs j
		LEFT JOIN leads l ON l.job_id = j.id AND l.status = 'accepted'
		WHERE j.id = $1`

	job := &domain.Job{}
	var address, buildingType, preferredTime sql.NullString
	var photoURLsJSON []byte
	var estPrice sql.NullInt64
	var arrTime, propMsg sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.ClientID, &job.CategoryID, &job.DistrictID,
		&job.Title, &job.Description, &job.Urgency, &job.Status,
		&address, &buildingType, &job.Floor, &job.HasElevator,
		&job.PreferredDate1, &job.PreferredDate2, &preferredTime,
		&job.Budget, &job.WantsInvoice, &job.ContactMethod,
		&photoURLsJSON, &job.FinalValue, &job.CompletedAt, &job.CompletedByID, &job.ClientConfirmed,
		&job.CreatedAt, &job.UpdatedAt, &job.ExpiresAt,
		&estPrice, &arrTime, &propMsg,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrJobNotFound
		}
		return nil, err
	}

	job.Address = address.String
	job.BuildingType = buildingType.String
	job.PreferredTime = preferredTime.String

	if photoURLsJSON != nil {
		_ = json.Unmarshal(photoURLsJSON, &job.PhotoURLs)
	}
	if job.PhotoURLs == nil {
		job.PhotoURLs = []string{}
	}

	if estPrice.Valid || arrTime.String != "" || propMsg.String != "" {
		var ep *int
		if estPrice.Valid {
			v := int(estPrice.Int64)
			ep = &v
		}
		job.Proposal = &domain.JobProposal{
			EstimatedPrice:  ep,
			ArrivalTime:     arrTime.String,
			ProposalMessage: propMsg.String,
		}
	}

	return job, nil
}

// Update modifies an existing job.
func (r *PostgresRepository) Update(ctx context.Context, job *domain.Job) error {
	photoURLs, err := json.Marshal(job.PhotoURLs)
	if err != nil {
		photoURLs = []byte("[]")
	}

	query := `
		UPDATE jobs SET
			title = $2, description = $3, urgency = $4, status = $5,
			address = $6, building_type = $7, floor = $8, has_elevator = $9,
			preferred_date1 = $10, preferred_date2 = $11, preferred_time = $12,
			budget = $13, wants_invoice = $14, contact_method = $15,
			photo_urls = $16, final_value = $17, completed_at = $18,
			completed_by_id = $19, client_confirmed = $20, expires_at = $21
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		job.ID,
		job.Title, job.Description, job.Urgency, job.Status,
		nullString(job.Address), nullString(job.BuildingType), job.Floor, job.HasElevator,
		job.PreferredDate1, job.PreferredDate2, nullString(job.PreferredTime),
		job.Budget, job.WantsInvoice, job.ContactMethod,
		photoURLs, job.FinalValue, job.CompletedAt,
		job.CompletedByID, job.ClientConfirmed, job.ExpiresAt,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrJobNotFound
	}
	return nil
}

// List retrieves jobs with pagination and filters.
func (r *PostgresRepository) List(ctx context.Context, filter ListFilter) ([]*domain.Job, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.ClientID != nil {
		conditions = append(conditions, fmt.Sprintf("j.client_id = $%d", argIdx))
		args = append(args, *filter.ClientID)
		argIdx++
	}
	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("j.category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}
	if filter.DistrictID != nil {
		conditions = append(conditions, fmt.Sprintf("j.district_id = $%d", argIdx))
		args = append(args, *filter.DistrictID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("j.status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM jobs j %s", where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch jobs
	query := fmt.Sprintf(`
		SELECT j.id, j.client_id, j.category_id, j.district_id,
			j.title, j.description, j.urgency, j.status,
			j.address, j.building_type, j.floor, j.has_elevator,
			j.preferred_date1, j.preferred_date2, j.preferred_time,
			j.budget, j.wants_invoice, j.contact_method,
			j.photo_urls, j.final_value, j.completed_at, j.completed_by_id, j.client_confirmed,
			j.created_at, j.updated_at, j.expires_at,
			l.estimated_price, COALESCE(l.arrival_time, ''), COALESCE(l.proposal_message, '')
		FROM jobs j
		LEFT JOIN leads l ON l.job_id = j.id AND l.status = 'accepted'
		%s
		ORDER BY j.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []*domain.Job
	for rows.Next() {
		job := &domain.Job{}
		var address, buildingType, preferredTime sql.NullString
		var photoURLsJSON []byte
		var estPrice sql.NullInt64
		var arrTime, propMsg sql.NullString

		if err := rows.Scan(
			&job.ID, &job.ClientID, &job.CategoryID, &job.DistrictID,
			&job.Title, &job.Description, &job.Urgency, &job.Status,
			&address, &buildingType, &job.Floor, &job.HasElevator,
			&job.PreferredDate1, &job.PreferredDate2, &preferredTime,
			&job.Budget, &job.WantsInvoice, &job.ContactMethod,
			&photoURLsJSON, &job.FinalValue, &job.CompletedAt, &job.CompletedByID, &job.ClientConfirmed,
			&job.CreatedAt, &job.UpdatedAt, &job.ExpiresAt,
			&estPrice, &arrTime, &propMsg,
		); err != nil {
			return nil, 0, err
		}

		job.Address = address.String
		job.BuildingType = buildingType.String
		job.PreferredTime = preferredTime.String

		if photoURLsJSON != nil {
			_ = json.Unmarshal(photoURLsJSON, &job.PhotoURLs)
		}
		if job.PhotoURLs == nil {
			job.PhotoURLs = []string{}
		}

		if estPrice.Valid || arrTime.String != "" || propMsg.String != "" {
			var ep *int
			if estPrice.Valid {
				v := int(estPrice.Int64)
				ep = &v
			}
			job.Proposal = &domain.JobProposal{
				EstimatedPrice:  ep,
				ArrivalTime:     arrTime.String,
				ProposalMessage: propMsg.String,
			}
		}

		jobs = append(jobs, job)
	}

	return jobs, total, rows.Err()
}

// ListActiveForMatching retrieves active jobs matching category and district.
func (r *PostgresRepository) ListActiveForMatching(ctx context.Context, categoryID, districtID uuid.UUID) ([]*domain.Job, error) {
	query := `
		SELECT id, client_id, category_id, district_id,
			title, description, urgency, status,
			address, building_type, floor, has_elevator,
			preferred_date1, preferred_date2, preferred_time,
			budget, wants_invoice, contact_method,
			photo_urls, final_value, completed_at, completed_by_id, client_confirmed,
			created_at, updated_at, expires_at
		FROM jobs
		WHERE status = 'active'
			AND category_id = $1
			AND district_id = $2
			AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY urgency DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, categoryID, districtID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*domain.Job
	for rows.Next() {
		job := &domain.Job{}
		var address, buildingType, preferredTime sql.NullString
		var photoURLsJSON []byte

		if err := rows.Scan(
			&job.ID, &job.ClientID, &job.CategoryID, &job.DistrictID,
			&job.Title, &job.Description, &job.Urgency, &job.Status,
			&address, &buildingType, &job.Floor, &job.HasElevator,
			&job.PreferredDate1, &job.PreferredDate2, &preferredTime,
			&job.Budget, &job.WantsInvoice, &job.ContactMethod,
			&photoURLsJSON, &job.FinalValue, &job.CompletedAt, &job.CompletedByID, &job.ClientConfirmed,
			&job.CreatedAt, &job.UpdatedAt, &job.ExpiresAt,
		); err != nil {
			return nil, err
		}

		job.Address = address.String
		job.BuildingType = buildingType.String
		job.PreferredTime = preferredTime.String

		if photoURLsJSON != nil {
			_ = json.Unmarshal(photoURLsJSON, &job.PhotoURLs)
		}
		if job.PhotoURLs == nil {
			job.PhotoURLs = []string{}
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// nullString converts an empty string to sql.NullString.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// Ensure pq is imported for array handling
var _ = pq.Array
