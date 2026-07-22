package message

import (
	"context"
	"database/sql"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// PostgresRepository implements message Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new postgres repository for messages.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, m *domain.Message) error {
	query := `
		INSERT INTO messages (id, job_id, sender_id, receiver_id, content, image_url, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		m.ID,
		m.JobID,
		m.SenderID,
		m.ReceiverID,
		m.Content,
		m.ImageURL,
		m.IsRead,
		m.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetConversations(ctx context.Context, userID uuid.UUID) ([]*domain.Conversation, error) {
	query := `
		WITH active_pairs AS (
			SELECT 
				CASE WHEN m.sender_id = $1 THEN m.receiver_id ELSE m.sender_id END AS counterparty_id,
				m.job_id,
				MAX(m.created_at) AS last_act
			FROM messages m
			WHERE m.sender_id = $1 OR m.receiver_id = $1
			GROUP BY 1, 2

			UNION

			SELECT 
				CASE WHEN l.handyman_id = $1 THEN j.client_id ELSE l.handyman_id END AS counterparty_id,
				j.id AS job_id,
				COALESCE(l.accepted_at, l.created_at) AS last_act
			FROM leads l
			JOIN jobs j ON j.id = l.job_id
			WHERE (l.handyman_id = $1 OR j.client_id = $1)
			  AND j.client_id IS NOT NULL
			  AND l.handyman_id IS NOT NULL
			  AND (CASE WHEN l.handyman_id = $1 THEN j.client_id ELSE l.handyman_id END) != $1
		),
		distinct_pairs AS (
			SELECT DISTINCT ON (counterparty_id, COALESCE(job_id, '00000000-0000-0000-0000-000000000000'::uuid))
				counterparty_id,
				job_id,
				last_act
			FROM active_pairs
			WHERE counterparty_id != '00000000-0000-0000-0000-000000000000'::uuid
			ORDER BY counterparty_id, COALESCE(job_id, '00000000-0000-0000-0000-000000000000'::uuid), last_act DESC
		),
		latest_messages AS (
			SELECT DISTINCT ON (counterparty_id, COALESCE(job_id, '00000000-0000-0000-0000-000000000000'::uuid))
				counterparty_id,
				job_id,
				content AS last_message_content,
				image_url AS last_message_image,
				created_at AS last_message_time
			FROM (
				SELECT 
					CASE WHEN m.sender_id = $1 THEN m.receiver_id ELSE m.sender_id END AS counterparty_id,
					m.job_id,
					m.content,
					m.image_url,
					m.created_at
				FROM messages m
				WHERE m.sender_id = $1 OR m.receiver_id = $1
			) msg_sub
			ORDER BY counterparty_id, COALESCE(job_id, '00000000-0000-0000-0000-000000000000'::uuid), created_at DESC
		),
		unread_counts AS (
			SELECT 
				sender_id AS counterparty_id,
				job_id,
				COUNT(*) AS unread_cnt
			FROM messages
			WHERE receiver_id = $1 AND is_read = false
			GROUP BY sender_id, job_id
		)
		SELECT 
			dp.counterparty_id,
			COALESCE(NULLIF(hp.company_name, ''), u.name, 'Użytkownik') AS counterparty_name,
			u.role AS counterparty_role,
			COALESCE(hp.avatar_url, u.avatar_url, '') AS avatar_url,
			dp.job_id,
			COALESCE(j.title, 'Zlecenie') AS job_title,
			COALESCE(cat.name, '') AS category_name,
			COALESCE(l.status, j.status, '') AS lead_status,
			CASE 
				WHEN lm.last_message_content IS NOT NULL AND lm.last_message_content <> '' THEN lm.last_message_content
				WHEN lm.last_message_image IS NOT NULL AND lm.last_message_image <> '' THEN 'Przesłano zdjęcie'
				ELSE 'Rozpocznij konwersację...'
			END AS last_message,
			COALESCE(lm.last_message_time, dp.last_act) AS last_message_time,
			COALESCE(uc.unread_cnt, 0) AS unread_count
		FROM distinct_pairs dp
		JOIN users u ON u.id = dp.counterparty_id
		LEFT JOIN handyman_profiles hp ON hp.user_id = dp.counterparty_id
		LEFT JOIN jobs j ON j.id = dp.job_id
		LEFT JOIN leads l ON l.job_id = dp.job_id AND (l.handyman_id = dp.counterparty_id OR l.handyman_id = $1)
		LEFT JOIN service_categories cat ON cat.id = j.category_id
		LEFT JOIN latest_messages lm ON lm.counterparty_id = dp.counterparty_id AND lm.job_id IS NOT DISTINCT FROM dp.job_id
		LEFT JOIN unread_counts uc ON uc.counterparty_id = dp.counterparty_id AND uc.job_id IS NOT DISTINCT FROM dp.job_id
		ORDER BY COALESCE(lm.last_message_time, dp.last_act) DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*domain.Conversation
	for rows.Next() {
		c := &domain.Conversation{}
		var jobID sql.NullString
		if err := rows.Scan(
			&c.CounterpartyID,
			&c.CounterpartyName,
			&c.CounterpartyRole,
			&c.AvatarURL,
			&jobID,
			&c.JobTitle,
			&c.CategoryName,
			&c.LeadStatus,
			&c.LastMessage,
			&c.LastMessageTime,
			&c.UnreadCount,
		); err != nil {
			return nil, err
		}
		if jobID.Valid {
			parsedJobID, err := uuid.Parse(jobID.String)
			if err == nil {
				c.JobID = &parsedJobID
			}
		}
		c.IsOnline = true // Mark active status
		conversations = append(conversations, c)
	}

	return conversations, rows.Err()
}

func (r *PostgresRepository) GetThread(ctx context.Context, userID, counterpartyID uuid.UUID, jobID *uuid.UUID, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT id, job_id, sender_id, receiver_id, content, image_url, is_read, created_at
		FROM messages
		WHERE ((sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1))`

	args := []interface{}{userID, counterpartyID}
	if jobID != nil {
		query += ` AND job_id = $3`
		args = append(args, *jobID)
	}

	query += ` ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		m := &domain.Message{}
		var jID sql.NullString
		if err := rows.Scan(
			&m.ID,
			&jID,
			&m.SenderID,
			&m.ReceiverID,
			&m.Content,
			&m.ImageURL,
			&m.IsRead,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		if jID.Valid {
			parsed, err := uuid.Parse(jID.String)
			if err == nil {
				m.JobID = &parsed
			}
		}
		messages = append(messages, m)
	}

	return messages, rows.Err()
}

func (r *PostgresRepository) MarkAsRead(ctx context.Context, userID, counterpartyID uuid.UUID) error {
	query := `UPDATE messages SET is_read = true WHERE receiver_id = $1 AND sender_id = $2 AND is_read = false`
	_, err := r.db.ExecContext(ctx, query, userID, counterpartyID)
	return err
}

func (r *PostgresRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM messages WHERE receiver_id = $1 AND is_read = false`
	var count int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}
