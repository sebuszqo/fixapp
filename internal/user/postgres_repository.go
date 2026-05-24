package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"fixapp/internal/auth/role"
	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Ensure PostgresRepository implements Repository.
var _ Repository = (*PostgresRepository)(nil)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create inserts a new user into the database.
func (r *PostgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, email, name, role, provider, provider_id, password_hash,
			avatar_url, phone, is_active, email_verified, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Role.String(),
		user.Provider.String(),
		nullString(user.ProviderID),
		nullString(user.PasswordHash),
		nullString(user.AvatarURL),
		nullString(user.Phone),
		user.IsActive,
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their ID.
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, name, role, provider, provider_id, password_hash,
		       avatar_url, phone, is_active, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	user, err := r.scanUser(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by their email.
func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, name, role, provider, provider_id, password_hash,
		       avatar_url, phone, is_active, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`

	user, err := r.scanUser(r.db.QueryRowContext(ctx, query, email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

// GetByProvider retrieves a user by their provider and provider ID.
func (r *PostgresRepository) GetByProvider(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error) {
	query := `
		SELECT id, email, name, role, provider, provider_id, password_hash,
		       avatar_url, phone, is_active, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE provider = $1 AND provider_id = $2
	`

	user, err := r.scanUser(r.db.QueryRowContext(ctx, query, provider.String(), providerID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by provider: %w", err)
	}

	return user, nil
}

// Update modifies an existing user.
func (r *PostgresRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET
			email = $2,
			name = $3,
			role = $4,
			avatar_url = $5,
			phone = $6,
			is_active = $7,
			email_verified = $8,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Role.String(),
		nullString(user.AvatarURL),
		nullString(user.Phone),
		user.IsActive,
		user.EmailVerified,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("update user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Delete performs a soft delete by setting is_active = false.
func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// List retrieves users with pagination and optional filters.
func (r *PostgresRepository) List(ctx context.Context, filter ListFilter) ([]*domain.User, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Role != nil {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, filter.Role.String())
		argIdx++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(LOWER(email) LIKE $%d OR LOWER(name) LIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(filter.Search)+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	// Fetch users
	query := fmt.Sprintf(`
		SELECT id, email, name, role, provider, provider_id, password_hash,
		       avatar_url, phone, is_active, email_verified, created_at, updated_at, last_login_at
		FROM users
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user, err := r.scanUserFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return users, total, nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *PostgresRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1))`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("exists by email: %w", err)
	}

	return exists, nil
}

// UpdateLastLogin updates the last login timestamp.
func (r *PostgresRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	return nil
}

// UpdateRole changes a user's role.
func (r *PostgresRepository) UpdateRole(ctx context.Context, id uuid.UUID, r2 role.Role) error {
	query := `UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, r2.String())
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Helper functions

type scanner interface {
	Scan(dest ...interface{}) error
}

func (r *PostgresRepository) scanUser(row *sql.Row) (*domain.User, error) {
	return scanUserFrom(row)
}

func (r *PostgresRepository) scanUserFromRows(rows *sql.Rows) (*domain.User, error) {
	return scanUserFrom(rows)
}

func scanUserFrom(s scanner) (*domain.User, error) {
	var user domain.User
	var roleStr, providerStr string
	var providerID, passwordHash, avatarURL, phone sql.NullString
	var lastLoginAt sql.NullTime

	err := s.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&roleStr,
		&providerStr,
		&providerID,
		&passwordHash,
		&avatarURL,
		&phone,
		&user.IsActive,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err != nil {
		return nil, err
	}

	user.Role = role.Parse(roleStr)
	user.Provider = domain.AuthProvider(providerStr)
	user.ProviderID = providerID.String
	user.PasswordHash = passwordHash.String
	user.AvatarURL = avatarURL.String
	user.Phone = phone.String

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	// PostgreSQL error code 23505 = unique_violation
	return err != nil && strings.Contains(err.Error(), "23505")
}

