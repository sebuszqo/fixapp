package message

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"fixapp/internal/auth"
	"fixapp/pkg/ctxlog"
	"fixapp/pkg/middleware"
	"fixapp/pkg/response"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for messaging operations.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new message handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers messaging routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /messages/conversations", middleware.RequireAuth(http.HandlerFunc(h.GetConversations)))
	mux.Handle("GET /messages/thread", middleware.RequireAuth(http.HandlerFunc(h.GetThread)))
	mux.Handle("POST /messages", middleware.RequireAuth(http.HandlerFunc(h.SendMessage)))
	mux.Handle("PUT /messages/read", middleware.RequireAuth(http.HandlerFunc(h.MarkAsRead)))
	mux.Handle("GET /messages/unread-count", middleware.RequireAuth(http.HandlerFunc(h.GetUnreadCount)))
	mux.Handle("POST /messages/upload", middleware.RequireAuth(http.HandlerFunc(h.UploadImage)))
}

func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	conversations, err := h.service.GetConversations(r.Context())
	if err != nil {
		log.Error("failed to get conversations", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	responses := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		responses[i] = ToConversationResponse(c)
	}

	response.OK(w, map[string]interface{}{
		"conversations": responses,
	})
}

func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	counterpartyStr := r.URL.Query().Get("counterparty_id")
	if counterpartyStr == "" {
		response.BadRequest(w, "counterparty_id is required")
		return
	}

	counterpartyID, err := uuid.Parse(counterpartyStr)
	if err != nil {
		response.BadRequest(w, "invalid counterparty_id")
		return
	}

	var jobID *uuid.UUID
	if jStr := r.URL.Query().Get("job_id"); jStr != "" {
		parsed, err := uuid.Parse(jStr)
		if err == nil {
			jobID = &parsed
		}
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	messages, err := h.service.GetThread(r.Context(), counterpartyID, jobID, limit, offset)
	if err != nil {
		log.Error("failed to get message thread", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	responses := make([]MessageResponse, len(messages))
	for i, m := range messages {
		responses[i] = ToMessageResponse(m)
	}

	response.OK(w, map[string]interface{}{
		"messages": responses,
	})
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.ReceiverID == "" {
		response.BadRequest(w, "receiver_id is required")
		return
	}
	if req.Content == "" && req.ImageURL == "" {
		response.BadRequest(w, "content or image_url is required")
		return
	}

	msg, err := h.service.SendMessage(r.Context(), req)
	if err != nil {
		log.Error("failed to send message", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	response.Created(w, ToMessageResponse(msg))
}

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	counterpartyStr := r.URL.Query().Get("counterparty_id")
	if counterpartyStr == "" {
		response.BadRequest(w, "counterparty_id is required")
		return
	}

	counterpartyID, err := uuid.Parse(counterpartyStr)
	if err != nil {
		response.BadRequest(w, "invalid counterparty_id")
		return
	}

	if err := h.service.MarkAsRead(r.Context(), counterpartyID); err != nil {
		log.Error("failed to mark messages as read", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	response.OK(w, map[string]string{"message": "messages marked as read"})
}

func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	count, err := h.service.GetUnreadCount(r.Context())
	if err != nil {
		log.Error("failed to get unread count", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	response.OK(w, UnreadMessageCountResponse{UnreadCount: count})
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	authUser := auth.FromContext(r.Context())
	if authUser == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	// 10 MB max file size
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "Form file 'file' is required")
		return
	}
	defer file.Close()

	// Ensure uploads directory exists
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Error("failed to create upload directory", zap.Error(err))
		response.InternalServerError(w, "Directory creation failed")
		return
	}

	ext := filepath.Ext(handler.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("chat_%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		log.Error("failed to save uploaded file", zap.Error(err))
		response.InternalServerError(w, "File save failed")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Error("failed to copy uploaded file", zap.Error(err))
		response.InternalServerError(w, "File copy failed")
		return
	}

	imageURL := fmt.Sprintf("/uploads/%s", filename)
	response.OK(w, map[string]string{
		"url": imageURL,
	})
}
