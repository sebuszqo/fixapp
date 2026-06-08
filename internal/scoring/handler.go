package scoring

import (
	"errors"
	"net/http"

	"fixapp/internal/domain"
	"fixapp/pkg/ctxlog"
	"fixapp/pkg/middleware"
	"fixapp/pkg/response"

	"go.uber.org/zap"
)

// Handler handles HTTP requests for scoring operations.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new scoring handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers the scoring routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// Client sees their Commit Score
	mux.Handle("GET /score/commit", middleware.RequireAuth(http.HandlerFunc(h.GetMyCommitScore)))
	// Handyman sees their ProScore
	mux.Handle("GET /score/pro", middleware.RequireHandyman(http.HandlerFunc(h.GetMyProScore)))
}

// GetMyCommitScore godoc
// @Summary      Get my Commit Score
// @Description  Returns the authenticated client's Commit Score with breakdown
// @Tags         scoring
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  CommitScoreResponse
// @Failure      401  {object}  response.Error
// @Router       /score/commit [get]
func (h *Handler) GetMyCommitScore(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	cs, err := h.service.GetMyCommitScore(r.Context())
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToCommitScoreResponse(cs))
}

// GetMyProScore godoc
// @Summary      Get my ProScore
// @Description  Returns the authenticated handyman's ProScore with breakdown
// @Tags         scoring
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  ProScoreResponse
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Router       /score/pro [get]
func (h *Handler) GetMyProScore(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	ps, err := h.service.GetMyProScore(r.Context())
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToProScoreResponse(ps))
}

// handleError maps domain errors to HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, log *zap.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		response.Unauthorized(w, "Authentication required")
	case errors.Is(err, domain.ErrForbidden):
		response.Forbidden(w, "Access denied")
	default:
		log.Error("internal error", zap.Error(err))
		response.InternalServerError(w, "")
	}
}
