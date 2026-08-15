package notification

import (
	"context"
	"database/sql"
	"fmt"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// PostgresRepository implements notification Repository with PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new notification PostgreSQL repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, n *domain.Notification) error {
	if n.LinkURL != "" {
		var exists bool
		err := r.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM notifications WHERE user_id = $1 AND link_url = $2)`,
			n.UserID, n.LinkURL,
		).Scan(&exists)
		if err == nil && exists {
			return nil // already has a notification for this link
		}
	}

	query := `
		INSERT INTO notifications (id, user_id, type, title, content, link_url, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		n.ID,
		n.UserID,
		n.Type.String(),
		n.Title,
		n.Content,
		n.LinkURL,
		n.IsRead,
		n.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) ListByUser(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]*domain.Notification, int64, int64, error) {
	// Total count query
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	countArgs := []interface{}{userID}
	if filter.Type != nil {
		countQuery += ` AND type = $2`
		countArgs = append(countArgs, filter.Type.String())
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, 0, err
	}

	// Unread count query
	var unread int64
	unreadQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	if err := r.db.QueryRowContext(ctx, unreadQuery, userID).Scan(&unread); err != nil {
		return nil, 0, 0, err
	}

	// List query
	listQuery := `
		SELECT id, user_id, type, title, content, link_url, is_read, created_at
		FROM notifications
		WHERE user_id = $1`

	listArgs := []interface{}{userID}
	paramIdx := 2

	if filter.Type != nil {
		listQuery += fmt.Sprintf(` AND type = $%d`, paramIdx)
		listArgs = append(listArgs, filter.Type.String())
		paramIdx++
	}

	listQuery += ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		listQuery += fmt.Sprintf(` LIMIT $%d`, paramIdx)
		listArgs = append(listArgs, filter.Limit)
		paramIdx++
	}
	if filter.Offset > 0 {
		listQuery += fmt.Sprintf(` OFFSET $%d`, paramIdx)
		listArgs = append(listArgs, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		var nType string
		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&nType,
			&n.Title,
			&n.Content,
			&n.LinkURL,
			&n.IsRead,
			&n.CreatedAt,
		); err != nil {
			return nil, 0, 0, err
		}
		n.Type = domain.NotificationType(nType)
		notifications = append(notifications, n)
	}

	return notifications, total, unread, rows.Err()
}

func (r *PostgresRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *PostgresRepository) MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = true WHERE user_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, notificationID)
	return err
}

func (r *PostgresRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	var count int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}
