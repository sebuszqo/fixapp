package notification

import (
	"context"

	"fixapp/internal/auth"
	"fixapp/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles notification business logic.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService creates a new notification service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateNotification creates and saves a new notification for a specific user.
func (s *Service) CreateNotification(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, content, linkURL string) (*domain.Notification, error) {
	n := domain.NewNotification(userID, notifType, title, content, linkURL)
	if err := s.repo.Create(ctx, n); err != nil {
		s.logger.Error("failed to create notification", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}
	return n, nil
}

// ListMyNotifications lists notifications for the current authenticated user.
func (s *Service) ListMyNotifications(ctx context.Context, notifType *domain.NotificationType, limit, offset int) ([]*domain.Notification, int64, int64, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, 0, 0, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, 0, 0, domain.ErrInvalidInput
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	filter := ListFilter{
		Type:   notifType,
		Limit:  limit,
		Offset: offset,
	}

	return s.repo.ListByUser(ctx, userID, filter)
}

// MarkAllAsRead marks all notifications as read for the current user.
func (s *Service) MarkAllAsRead(ctx context.Context) error {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return domain.ErrInvalidInput
	}

	return s.repo.MarkAllAsRead(ctx, userID)
}

// MarkAsRead marks a specific notification as read.
func (s *Service) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return domain.ErrInvalidInput
	}

	return s.repo.MarkAsRead(ctx, userID, notificationID)
}

// GetUnreadCount returns the count of unread notifications for the current user.
func (s *Service) GetUnreadCount(ctx context.Context) (int64, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return 0, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return 0, domain.ErrInvalidInput
	}

	return s.repo.GetUnreadCount(ctx, userID)
}
