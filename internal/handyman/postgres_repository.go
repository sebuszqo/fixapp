package handyman

import (
	"context"
	"database/sql"
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

// NewPostgresRepository creates a new PostgreSQL handyman repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateProfile creates a new handyman profile.
func (r *PostgresRepository) CreateProfile(ctx context.Context, profile *domain.HandymanProfile) error {
	query := `
		INSERT INTO handyman_profiles (
			id, user_id, company_name, nip, phone, email,
			bio, avatar_url, categories, districts,
			is_available, emergency_available, is_verified,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.db.ExecContext(ctx, query,
		profile.ID, profile.UserID,
		nullStr(profile.CompanyName), nullStr(profile.NIP), nullStr(profile.Phone), nullStr(profile.Email),
		nullStr(profile.Bio), nullStr(profile.AvatarURL),
		pq.Array(profile.Categories), pq.Array(profile.Districts),
		profile.IsAvailable, profile.EmergencyAvailable, profile.IsVerified,
		profile.CreatedAt, profile.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrProfileAlreadyExists
		}
		return err
	}
	return nil
}

// GetByUserID retrieves a profile by user ID.
func (r *PostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.HandymanProfile, error) {
	return r.getByCondition(ctx, "user_id = $1", userID)
}

// GetByID retrieves a profile by its ID.
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.HandymanProfile, error) {
	return r.getByCondition(ctx, "id = $1", id)
}

func (r *PostgresRepository) getByCondition(ctx context.Context, condition string, arg interface{}) (*domain.HandymanProfile, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, company_name, nip, phone, email,
			bio, avatar_url, categories, districts,
			is_available, emergency_available, is_verified,
			created_at, updated_at
		FROM handyman_profiles WHERE %s`, condition)

	p := &domain.HandymanProfile{}
	var companyName, nip, phone, email, bio, avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&p.ID, &p.UserID,
		&companyName, &nip, &phone, &email,
		&bio, &avatarURL,
		pq.Array(&p.Categories), pq.Array(&p.Districts),
		&p.IsAvailable, &p.EmergencyAvailable, &p.IsVerified,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}

	p.CompanyName = companyName.String
	p.NIP = nip.String
	p.Phone = phone.String
	p.Email = email.String
	p.Bio = bio.String
	p.AvatarURL = avatarURL.String

	if p.Categories == nil {
		p.Categories = []uuid.UUID{}
	}
	if p.Districts == nil {
		p.Districts = []uuid.UUID{}
	}

	return p, nil
}

// Update modifies an existing profile.
func (r *PostgresRepository) Update(ctx context.Context, profile *domain.HandymanProfile) error {
	query := `
		UPDATE handyman_profiles SET
			company_name = $2, nip = $3, phone = $4, email = $5,
			bio = $6, avatar_url = $7,
			categories = $8, districts = $9,
			is_available = $10, emergency_available = $11, is_verified = $12
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		profile.ID,
		nullStr(profile.CompanyName), nullStr(profile.NIP), nullStr(profile.Phone), nullStr(profile.Email),
		nullStr(profile.Bio), nullStr(profile.AvatarURL),
		pq.Array(profile.Categories), pq.Array(profile.Districts),
		profile.IsAvailable, profile.EmergencyAvailable, profile.IsVerified,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProfileNotFound
	}
	return nil
}

// Search finds handyman profiles matching criteria.
func (r *PostgresRepository) Search(ctx context.Context, filter SearchFilter) ([]*domain.HandymanProfile, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(categories)", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}
	if filter.DistrictID != nil {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(districts)", argIdx))
		args = append(args, *filter.DistrictID)
		argIdx++
	}
	if filter.IsVerified != nil {
		conditions = append(conditions, fmt.Sprintf("is_verified = $%d", argIdx))
		args = append(args, *filter.IsVerified)
		argIdx++
	}
	if filter.Available != nil {
		conditions = append(conditions, fmt.Sprintf("is_available = $%d", argIdx))
		args = append(args, *filter.Available)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(company_name ILIKE $%d OR bio ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM handyman_profiles %s", where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch
	query := fmt.Sprintf(`
		SELECT id, user_id, company_name, nip, phone, email,
			bio, avatar_url, categories, districts,
			is_available, emergency_available, is_verified,
			created_at, updated_at
		FROM handyman_profiles %s
		ORDER BY is_verified DESC, created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var profiles []*domain.HandymanProfile
	for rows.Next() {
		p := &domain.HandymanProfile{}
		var companyName, nip, phone, email, bio, avatarURL sql.NullString

		if err := rows.Scan(
			&p.ID, &p.UserID,
			&companyName, &nip, &phone, &email,
			&bio, &avatarURL,
			pq.Array(&p.Categories), pq.Array(&p.Districts),
			&p.IsAvailable, &p.EmergencyAvailable, &p.IsVerified,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		p.CompanyName = companyName.String
		p.NIP = nip.String
		p.Phone = phone.String
		p.Email = email.String
		p.Bio = bio.String
		p.AvatarURL = avatarURL.String
		if p.Categories == nil {
			p.Categories = []uuid.UUID{}
		}
		if p.Districts == nil {
			p.Districts = []uuid.UUID{}
		}

		profiles = append(profiles, p)
	}

	return profiles, total, rows.Err()
}

// FindMatchingForJob finds available handymen matching a job's category and district.
func (r *PostgresRepository) FindMatchingForJob(ctx context.Context, categoryID, districtID uuid.UUID, emergency bool) ([]*domain.HandymanProfile, error) {
	query := `
		SELECT id, user_id, company_name, nip, phone, email,
			bio, avatar_url, categories, districts,
			is_available, emergency_available, is_verified,
			created_at, updated_at
		FROM handyman_profiles
		WHERE is_available = true
			AND $1 = ANY(categories)
			AND $2 = ANY(districts)`

	args := []interface{}{categoryID, districtID}

	if emergency {
		query += ` AND emergency_available = true`
	}

	query += ` ORDER BY is_verified DESC, created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*domain.HandymanProfile
	for rows.Next() {
		p := &domain.HandymanProfile{}
		var companyName, nip, phone, email, bio, avatarURL sql.NullString

		if err := rows.Scan(
			&p.ID, &p.UserID,
			&companyName, &nip, &phone, &email,
			&bio, &avatarURL,
			pq.Array(&p.Categories), pq.Array(&p.Districts),
			&p.IsAvailable, &p.EmergencyAvailable, &p.IsVerified,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		p.CompanyName = companyName.String
		p.NIP = nip.String
		p.Phone = phone.String
		p.Email = email.String
		p.Bio = bio.String
		p.AvatarURL = avatarURL.String
		if p.Categories == nil {
			p.Categories = []uuid.UUID{}
		}
		if p.Districts == nil {
			p.Districts = []uuid.UUID{}
		}

		profiles = append(profiles, p)
	}

	return profiles, rows.Err()
}

// ===== Pricing =====

// CreatePricingItem adds a pricing item.
func (r *PostgresRepository) CreatePricingItem(ctx context.Context, item *domain.PricingItem) error {
	query := `
		INSERT INTO handyman_pricing (id, profile_id, service_name, price_from, price_to, unit, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProfileID, item.ServiceName, item.PriceFrom, item.PriceTo, item.Unit, item.SortOrder,
	)
	return err
}

// UpdatePricingItem updates a pricing item.
func (r *PostgresRepository) UpdatePricingItem(ctx context.Context, item *domain.PricingItem) error {
	query := `
		UPDATE handyman_pricing SET
			service_name = $2, price_from = $3, price_to = $4, unit = $5, sort_order = $6
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.ServiceName, item.PriceFrom, item.PriceTo, item.Unit, item.SortOrder,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrPricingItemNotFound
	}
	return nil
}

// DeletePricingItem removes a pricing item.
func (r *PostgresRepository) DeletePricingItem(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM handyman_pricing WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrPricingItemNotFound
	}
	return nil
}

// ListPricing retrieves all pricing items for a profile.
func (r *PostgresRepository) ListPricing(ctx context.Context, profileID uuid.UUID) ([]*domain.PricingItem, error) {
	query := `
		SELECT id, profile_id, service_name, price_from, price_to, unit, sort_order
		FROM handyman_pricing WHERE profile_id = $1
		ORDER BY sort_order, service_name`

	rows, err := r.db.QueryContext(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.PricingItem
	for rows.Next() {
		item := &domain.PricingItem{}
		if err := rows.Scan(
			&item.ID, &item.ProfileID, &item.ServiceName,
			&item.PriceFrom, &item.PriceTo, &item.Unit, &item.SortOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ===== Portfolio =====

// CreatePortfolioItem adds a portfolio photo.
func (r *PostgresRepository) CreatePortfolioItem(ctx context.Context, item *domain.PortfolioItem) error {
	query := `
		INSERT INTO handyman_portfolio (id, profile_id, image_url, caption, sort_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProfileID, item.ImageURL, nullStr(item.Caption), item.SortOrder, item.CreatedAt,
	)
	return err
}

// DeletePortfolioItem removes a portfolio photo.
func (r *PostgresRepository) DeletePortfolioItem(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM handyman_portfolio WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrPortfolioItemNotFound
	}
	return nil
}

// ListPortfolio retrieves all portfolio items for a profile.
func (r *PostgresRepository) ListPortfolio(ctx context.Context, profileID uuid.UUID) ([]*domain.PortfolioItem, error) {
	query := `
		SELECT id, profile_id, image_url, caption, sort_order, created_at
		FROM handyman_portfolio WHERE profile_id = $1
		ORDER BY sort_order, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.PortfolioItem
	for rows.Next() {
		item := &domain.PortfolioItem{}
		var caption sql.NullString
		if err := rows.Scan(
			&item.ID, &item.ProfileID, &item.ImageURL,
			&caption, &item.SortOrder, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.Caption = caption.String
		items = append(items, item)
	}
	return items, rows.Err()
}

// Helpers

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint")
}
