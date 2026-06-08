package catalog

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

// NewPostgresRepository creates a new PostgreSQL catalog repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ListCategories retrieves all service categories.
func (r *PostgresRepository) ListCategories(ctx context.Context, activeOnly bool) ([]*domain.ServiceCategory, error) {
	query := `SELECT id, name, slug, icon, base_price, is_active FROM service_categories`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.ServiceCategory
	for rows.Next() {
		c := &domain.ServiceCategory{}
		var icon sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &icon, &c.BasePrice, &c.IsActive); err != nil {
			return nil, err
		}
		c.Icon = icon.String
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// GetCategoryByID retrieves a category by ID.
func (r *PostgresRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.ServiceCategory, error) {
	query := `SELECT id, name, slug, icon, base_price, is_active FROM service_categories WHERE id = $1`

	c := &domain.ServiceCategory{}
	var icon sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Name, &c.Slug, &icon, &c.BasePrice, &c.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	c.Icon = icon.String
	return c, nil
}

// GetCategoryBySlug retrieves a category by slug.
func (r *PostgresRepository) GetCategoryBySlug(ctx context.Context, slug string) (*domain.ServiceCategory, error) {
	query := `SELECT id, name, slug, icon, base_price, is_active FROM service_categories WHERE slug = $1`

	c := &domain.ServiceCategory{}
	var icon sql.NullString
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&c.ID, &c.Name, &c.Slug, &icon, &c.BasePrice, &c.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	c.Icon = icon.String
	return c, nil
}

// ListDistricts retrieves all districts.
func (r *PostgresRepository) ListDistricts(ctx context.Context, activeOnly bool) ([]*domain.District, error) {
	query := `SELECT id, name, slug, city_name, is_active FROM districts`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var districts []*domain.District
	for rows.Next() {
		d := &domain.District{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Slug, &d.CityName, &d.IsActive); err != nil {
			return nil, err
		}
		districts = append(districts, d)
	}
	return districts, rows.Err()
}

// GetDistrictByID retrieves a district by ID.
func (r *PostgresRepository) GetDistrictByID(ctx context.Context, id uuid.UUID) (*domain.District, error) {
	query := `SELECT id, name, slug, city_name, is_active FROM districts WHERE id = $1`

	d := &domain.District{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&d.ID, &d.Name, &d.Slug, &d.CityName, &d.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDistrictNotFound
		}
		return nil, err
	}
	return d, nil
}

// GetDistrictBySlug retrieves a district by slug.
func (r *PostgresRepository) GetDistrictBySlug(ctx context.Context, slug string) (*domain.District, error) {
	query := `SELECT id, name, slug, city_name, is_active FROM districts WHERE slug = $1`

	d := &domain.District{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&d.ID, &d.Name, &d.Slug, &d.CityName, &d.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDistrictNotFound
		}
		return nil, err
	}
	return d, nil
}
