package message

import (
	"context"
	"fmt"

	"fixapp/internal/auth"
	"fixapp/internal/domain"
	"fixapp/internal/notification"
	"fixapp/internal/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles messaging logic.
type Service struct {
	repo         Repository
	userRepo     user.Repository
	notifService *notification.Service
	logger       *zap.Logger
}

// NewService creates a new message service.
func NewService(repo Repository, userRepo user.Repository, notifService *notification.Service, logger *zap.Logger) *Service {
	return &Service{
		repo:         repo,
		userRepo:     userRepo,
		notifService: notifService,
		logger:       logger,
	}
}

// SendMessage sends a new message to a recipient and triggers a notification.
func (s *Service) SendMessage(ctx context.Context, req SendMessageRequest) (*domain.Message, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	senderID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	if senderID == receiverID {
		return nil, domain.ErrInvalidInput
	}

	var jobID *uuid.UUID
	if req.JobID != nil && *req.JobID != "" {
		parsed, err := uuid.Parse(*req.JobID)
		if err == nil {
			jobID = &parsed
		}
	}

	allowed, err := s.repo.IsChatAllowed(ctx, senderID, receiverID, jobID)
	if err != nil || !allowed {
		return nil, domain.ErrChatNotAllowed
	}

	msg := domain.NewMessage(senderID, receiverID, jobID, req.Content, req.ImageURL)
	if err := s.repo.Create(ctx, msg); err != nil {
		s.logger.Error("failed to create message", zap.Error(err))
		return nil, err
	}

	// Create notification for recipient
	if s.notifService != nil {
		senderName := authUser.Email
		senderUser, err := s.userRepo.GetByID(ctx, senderID)
		if err == nil && senderUser != nil {
			senderName = senderUser.Name
		}

		notifContent := req.Content
		if notifContent == "" && req.ImageURL != "" {
			notifContent = "Przesłano zdjęcie"
		}

		linkURL := fmt.Sprintf("/client/messages?counterparty_id=%s", senderID.String())
		if authUser.IsUser() {
			linkURL = fmt.Sprintf("/pro/messages?counterparty_id=%s", senderID.String())
		}

		_, _ = s.notifService.CreateNotification(
			ctx,
			receiverID,
			domain.NotificationTypeMessage,
			fmt.Sprintf("Nowa wiadomość od %s", senderName),
			notifContent,
			linkURL,
		)
	}

	return msg, nil
}

// GetConversations returns all message conversations for the authenticated user.
func (s *Service) GetConversations(ctx context.Context) ([]*domain.Conversation, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.GetConversations(ctx, userID)
}

// GetThread returns message history with a specific user.
func (s *Service) GetThread(ctx context.Context, counterpartyID uuid.UUID, jobID *uuid.UUID, limit, offset int) ([]*domain.Message, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	// Mark received messages from this counterparty as read
	_ = s.repo.MarkAsRead(ctx, userID, counterpartyID)

	return s.repo.GetThread(ctx, userID, counterpartyID, jobID, limit, offset)
}

// MarkAsRead marks all messages from counterparty as read.
func (s *Service) MarkAsRead(ctx context.Context, counterpartyID uuid.UUID) error {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return domain.ErrInvalidInput
	}

	return s.repo.MarkAsRead(ctx, userID, counterpartyID)
}

// GetUnreadCount returns total count of unread messages for user.
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
