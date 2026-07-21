package domain

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message between client and handyman.
type Message struct {
	ID         uuid.UUID  `json:"id"`
	JobID      *uuid.UUID `json:"job_id,omitempty"`
	SenderID   uuid.UUID  `json:"sender_id"`
	ReceiverID uuid.UUID  `json:"receiver_id"`
	Content    string     `json:"content"`
	ImageURL   string     `json:"image_url,omitempty"`
	IsRead     bool       `json:"is_read"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Conversation represents a conversation summary with another user.
type Conversation struct {
	CounterpartyID   uuid.UUID  `json:"counterparty_id"`
	CounterpartyName string     `json:"counterparty_name"`
	CounterpartyRole string     `json:"counterparty_role"`
	AvatarURL        string     `json:"avatar_url"`
	JobID            *uuid.UUID `json:"job_id,omitempty"`
	JobTitle         string     `json:"job_title,omitempty"`
	CategoryName     string     `json:"category_name,omitempty"`
	LastMessage      string     `json:"last_message"`
	LastMessageTime  time.Time  `json:"last_message_time"`
	UnreadCount      int        `json:"unread_count"`
	IsOnline         bool       `json:"is_online"`
}

// NewMessage creates a new message.
func NewMessage(senderID, receiverID uuid.UUID, jobID *uuid.UUID, content, imageURL string) *Message {
	return &Message{
		ID:         uuid.New(),
		JobID:      jobID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		ImageURL:   imageURL,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}
}
