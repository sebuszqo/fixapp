package notification

import (
	"context"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// ListFilter specifies parameters for querying notifications.
type ListFilter struct {
	Type   *domain.NotificationType
	Limit  int
	Offset int
}

// Repository defines data access operations for notifications.
type Repository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	ListByUser(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]*domain.Notification, int64, int64, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
}
