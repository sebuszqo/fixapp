package notification

import (
	"time"

	"fixapp/internal/domain"
)

type NotificationResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	LinkURL   string `json:"link_url,omitempty"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

func ToNotificationResponse(n *domain.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		Type:      n.Type.String(),
		Title:     n.Title,
		Content:   n.Content,
		LinkURL:   n.LinkURL,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
	}
}

type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	Unread        int64                  `json:"unread"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}
