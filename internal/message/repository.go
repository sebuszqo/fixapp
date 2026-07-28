package message

import (
	"context"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Repository defines database operations for messages and conversations.
type Repository interface {
	Create(ctx context.Context, message *domain.Message) error
	GetConversations(ctx context.Context, userID uuid.UUID) ([]*domain.Conversation, error)
	GetThread(ctx context.Context, userID, counterpartyID uuid.UUID, jobID *uuid.UUID, limit, offset int) ([]*domain.Message, error)
	MarkAsRead(ctx context.Context, userID, counterpartyID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	IsChatAllowed(ctx context.Context, userID1, userID2 uuid.UUID, jobID *uuid.UUID) (bool, error)
}
