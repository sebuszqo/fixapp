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
	r.ensureCategories(ctx)
	query := `SELECT id, name, slug, icon, image_url, base_price, is_active FROM service_categories`
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
		var icon, imageURL sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &icon, &imageURL, &c.BasePrice, &c.IsActive); err != nil {
			return nil, err
		}
		c.Icon = icon.String
		c.ImageURL = imageURL.String
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// GetCategoryByID retrieves a category by ID.
func (r *PostgresRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.ServiceCategory, error) {
	query := `SELECT id, name, slug, icon, image_url, base_price, is_active FROM service_categories WHERE id = $1`

	c := &domain.ServiceCategory{}
	var icon, imageURL sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Name, &c.Slug, &icon, &imageURL, &c.BasePrice, &c.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	c.Icon = icon.String
	c.ImageURL = imageURL.String
	return c, nil
}

// GetCategoryBySlug retrieves a category by slug.
func (r *PostgresRepository) GetCategoryBySlug(ctx context.Context, slug string) (*domain.ServiceCategory, error) {
	query := `SELECT id, name, slug, icon, image_url, base_price, is_active FROM service_categories WHERE slug = $1`

	c := &domain.ServiceCategory{}
	var icon, imageURL sql.NullString
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&c.ID, &c.Name, &c.Slug, &icon, &imageURL, &c.BasePrice, &c.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	c.Icon = icon.String
	c.ImageURL = imageURL.String
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

func (r *PostgresRepository) ensureCategories(ctx context.Context) {
	sql := `
INSERT INTO service_categories (name, slug, icon, base_price, image_url) VALUES
    ('Hydraulik', 'hydraulik', 'droplets', 28, 'https://images.unsplash.com/photo-1585704032915-c3400ca199e7?auto=format&fit=crop&w=800&q=80'),
    ('Elektryk', 'elektryk', 'zap', 28, 'https://images.unsplash.com/photo-1621905252507-b35492cc74b4?auto=format&fit=crop&w=800&q=80'),
    ('Złota rączka', 'zlota-raczka', 'hammer', 15, 'https://images.unsplash.com/photo-1581783898377-1c85bf937427?auto=format&fit=crop&w=800&q=80'),
    ('Naprawa AGD', 'agd', 'tv', 20, 'https://images.unsplash.com/photo-1584622650111-993a426fbf0a?auto=format&fit=crop&w=800&q=80'),
    ('Malarz', 'malarz', 'paintbrush', 22, 'https://images.unsplash.com/photo-1562259929-b4e1fd3aef09?auto=format&fit=crop&w=800&q=80'),
    ('Stolarz', 'stolarz', 'box', 22, 'https://images.unsplash.com/photo-1504148455328-c376907d081c?auto=format&fit=crop&w=800&q=80'),
    ('Ślusarz', 'slusarz', 'key', 25, 'https://images.unsplash.com/photo-1582139329536-e7284fece509?auto=format&fit=crop&w=800&q=80'),
    ('Ogrodnik', 'ogrodnik', 'flower', 18, 'https://images.unsplash.com/photo-1416879595882-3373a0480b5b?auto=format&fit=crop&w=800&q=80'),
    ('Klimatyzacja', 'klimatyzacja', 'wind', 30, 'https://images.unsplash.com/photo-1621905251189-08b45d6a269e?auto=format&fit=crop&w=800&q=80'),
    ('Sprzątanie', 'sprzatanie', 'sparkles', 15, 'https://images.unsplash.com/photo-1581578731548-c64695cc6952?auto=format&fit=crop&w=800&q=80'),
    ('Przeprowadzki', 'przeprowadzki', 'truck', 25, 'https://images.unsplash.com/photo-1600518464441-9154a4dea21b?auto=format&fit=crop&w=800&q=80'),
    ('Dekarz', 'dacharz', 'home', 30, 'https://images.unsplash.com/photo-1632759145351-1d592919f522?auto=format&fit=crop&w=800&q=80')
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    icon = EXCLUDED.icon,
    image_url = EXCLUDED.image_url,
    base_price = EXCLUDED.base_price;`
	_, _ = r.db.ExecContext(ctx, sql)
}

