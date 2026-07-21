package message

import (
	"time"

	"fixapp/internal/domain"
)

type SendMessageRequest struct {
	ReceiverID string  `json:"receiver_id"`
	JobID      *string `json:"job_id,omitempty"`
	Content    string  `json:"content"`
	ImageURL   string  `json:"image_url,omitempty"`
}

type MessageResponse struct {
	ID         string  `json:"id"`
	JobID      *string `json:"job_id,omitempty"`
	SenderID   string  `json:"sender_id"`
	ReceiverID string  `json:"receiver_id"`
	Content    string  `json:"content"`
	ImageURL   string  `json:"image_url,omitempty"`
	IsRead     bool    `json:"is_read"`
	CreatedAt  string  `json:"created_at"`
}

func ToMessageResponse(m *domain.Message) MessageResponse {
	resp := MessageResponse{
		ID:         m.ID.String(),
		SenderID:   m.SenderID.String(),
		ReceiverID: m.ReceiverID.String(),
		Content:    m.Content,
		ImageURL:   m.ImageURL,
		IsRead:     m.IsRead,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
	}
	if m.JobID != nil {
		str := m.JobID.String()
		resp.JobID = &str
	}
	return resp
}

type ConversationResponse struct {
	CounterpartyID   string  `json:"counterparty_id"`
	CounterpartyName string  `json:"counterparty_name"`
	CounterpartyRole string  `json:"counterparty_role"`
	AvatarURL        string  `json:"avatar_url"`
	JobID            *string `json:"job_id,omitempty"`
	JobTitle         string  `json:"job_title,omitempty"`
	CategoryName     string  `json:"category_name,omitempty"`
	LastMessage      string  `json:"last_message"`
	LastMessageTime  string  `json:"last_message_time"`
	UnreadCount      int     `json:"unread_count"`
	IsOnline         bool    `json:"is_online"`
}

func ToConversationResponse(c *domain.Conversation) ConversationResponse {
	resp := ConversationResponse{
		CounterpartyID:   c.CounterpartyID.String(),
		CounterpartyName: c.CounterpartyName,
		CounterpartyRole: c.CounterpartyRole,
		AvatarURL:        c.AvatarURL,
		JobTitle:         c.JobTitle,
		CategoryName:     c.CategoryName,
		LastMessage:      c.LastMessage,
		LastMessageTime:  c.LastMessageTime.Format(time.RFC3339),
		UnreadCount:      c.UnreadCount,
		IsOnline:         c.IsOnline,
	}
	if c.JobID != nil {
		str := c.JobID.String()
		resp.JobID = &str
	}
	return resp
}

type UnreadMessageCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}
