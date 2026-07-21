package notification

import (
	"net/http"
	"strconv"

	"fixapp/internal/domain"
	"fixapp/pkg/ctxlog"
	"fixapp/pkg/middleware"
	"fixapp/pkg/response"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for notification operations.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new notification handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers notification routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /notifications", middleware.RequireAuth(http.HandlerFunc(h.ListMyNotifications)))
	mux.Handle("PUT /notifications/read-all", middleware.RequireAuth(http.HandlerFunc(h.MarkAllAsRead)))
	mux.Handle("PUT /notifications/{id}/read", middleware.RequireAuth(http.HandlerFunc(h.MarkAsRead)))
	mux.Handle("GET /notifications/unread-count", middleware.RequireAuth(http.HandlerFunc(h.GetUnreadCount)))
}

func (h *Handler) ListMyNotifications(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	var notifType *domain.NotificationType
	if t := r.URL.Query().Get("type"); t != "" && t != "all" {
		nt := domain.NotificationType(t)
		if nt.IsValid() {
			notifType = &nt
		}
	}

	notifications, total, unread, err := h.service.ListMyNotifications(r.Context(), notifType, limit, offset)
	if err != nil {
		log.Error("failed to list notifications", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	respNotifications := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		respNotifications[i] = ToNotificationResponse(n)
	}

	response.OK(w, NotificationListResponse{
		Notifications: respNotifications,
		Total:         total,
		Unread:        unread,
		Limit:         limit,
		Offset:        offset,
	})
}

func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	if err := h.service.MarkAllAsRead(r.Context()); err != nil {
		log.Error("failed to mark all notifications read", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	response.OK(w, map[string]string{"message": "all notifications marked as read"})
}

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	notifID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid notification ID")
		return
	}

	if err := h.service.MarkAsRead(r.Context(), notifID); err != nil {
		log.Error("failed to mark notification read", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	response.OK(w, map[string]string{"message": "notification marked as read"})
}

func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	count, err := h.service.GetUnreadCount(r.Context())
	if err != nil {
		log.Error("failed to get unread count", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	response.OK(w, UnreadCountResponse{UnreadCount: count})
}
