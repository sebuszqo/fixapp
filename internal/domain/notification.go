package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents category of notification.
type NotificationType string

const (
	NotificationTypeMessage   NotificationType = "message"
	NotificationTypeJobStatus NotificationType = "job_status"
	NotificationTypePayment   NotificationType = "payment"
	NotificationTypeInquiry   NotificationType = "inquiry"
	NotificationTypeSystem    NotificationType = "system"
)

func (t NotificationType) String() string {
	return string(t)
}

func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeMessage, NotificationTypeJobStatus, NotificationTypePayment, NotificationTypeInquiry, NotificationTypeSystem:
		return true
	default:
		return false
	}
}

// Notification represents an in-app alert for a user.
type Notification struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Content   string           `json:"content"`
	LinkURL   string           `json:"link_url,omitempty"`
	IsRead    bool             `json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
}

// NewNotification creates a new notification entity.
func NewNotification(userID uuid.UUID, notifType NotificationType, title, content, linkURL string) *Notification {
	return &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Content:   content,
		LinkURL:   linkURL,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
}
