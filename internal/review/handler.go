package review

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"fixapp/internal/domain"
	"fixapp/pkg/ctxlog"
	"fixapp/pkg/middleware"
	"fixapp/pkg/response"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for review operations.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new review handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers the review routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// Create review (authenticated users - both clients and handymen)
	mux.Handle("POST /reviews", middleware.RequireAuth(http.HandlerFunc(h.CreateReview)))

	// List reviews for a user (public)
	mux.HandleFunc("GET /users/{id}/reviews", h.ListUserReviews)

	// Get reviews for a job (authenticated)
	mux.Handle("GET /jobs/{id}/reviews", middleware.RequireAuth(http.HandlerFunc(h.ListJobReviews)))

	// Get rating summary for a user (public)
	mux.HandleFunc("GET /users/{id}/rating", h.GetRatingSummary)
}

// CreateReview godoc
// @Summary      Create a review
// @Description  Submit a review for a completed job. Clients review handymen, handymen review clients.
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateReviewRequest  true  "Review data"
// @Success      201   {object}  ReviewResponse
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Failure      403   {object}  response.Error
// @Failure      404   {object}  response.Error
// @Failure      409   {object}  response.Error
// @Router       /reviews [post]
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.JobID == "" {
		response.BadRequest(w, "job_id is required")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		response.BadRequest(w, "rating must be between 1 and 5")
		return
	}

	review, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.Created(w, ToReviewResponse(review))
}

// ListUserReviews godoc
// @Summary      List reviews for a user
// @Description  Returns paginated reviews received by a specific user
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Param        id      path      string  true   "User ID"
// @Param        limit   query     int     false  "Limit (default 20, max 100)"
// @Param        offset  query     int     false  "Offset (default 0)"
// @Success      200  {object}  ReviewListResponse
// @Failure      400  {object}  response.Error
// @Router       /users/{id}/reviews [get]
func (h *Handler) ListUserReviews(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	reviews, total, err := h.service.ListByReviewee(r.Context(), userID, limit, offset)
	if err != nil {
		log.Error("failed to list reviews", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	response.OK(w, ToReviewListResponse(reviews, total, limit, offset))
}

// ListJobReviews godoc
// @Summary      List reviews for a job
// @Description  Returns all reviews for a specific completed job
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  []ReviewResponse
// @Failure      400  {object}  response.Error
// @Failure      401  {object}  response.Error
// @Router       /jobs/{id}/reviews [get]
func (h *Handler) ListJobReviews(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	jobID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid job ID")
		return
	}

	reviews, err := h.service.ListByJob(r.Context(), jobID)
	if err != nil {
		log.Error("failed to list job reviews", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	responses := make([]ReviewResponse, len(reviews))
	for i, rev := range reviews {
		responses[i] = ToReviewResponse(rev)
	}
	response.OK(w, responses)
}

// GetRatingSummary godoc
// @Summary      Get rating summary for a user
// @Description  Returns average rating and total review count for a user
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  RatingSummaryResponse
// @Failure      400  {object}  response.Error
// @Router       /users/{id}/rating [get]
func (h *Handler) GetRatingSummary(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	avg, count, err := h.service.GetRatingSummary(r.Context(), userID)
	if err != nil {
		log.Error("failed to get rating summary", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	response.OK(w, RatingSummaryResponse{
		AverageRating: avg,
		TotalReviews:  count,
	})
}

// handleError maps domain errors to HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, log *zap.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrReviewNotFound):
		response.NotFound(w, "Review not found")
	case errors.Is(err, domain.ErrReviewAlreadyExists):
		response.Conflict(w, "Review already submitted for this job")
	case errors.Is(err, domain.ErrInvalidRating):
		response.BadRequest(w, "Rating must be between 1 and 5")
	case errors.Is(err, domain.ErrInvalidReviewType):
		response.BadRequest(w, "Invalid review type")
	case errors.Is(err, domain.ErrCannotReviewSelf):
		response.BadRequest(w, "Cannot review yourself")
	case errors.Is(err, domain.ErrJobNotCompleted):
		response.BadRequest(w, "Can only review after job is completed")
	case errors.Is(err, domain.ErrNotJobParticipant):
		response.Forbidden(w, "Only job participants can leave reviews")
	case errors.Is(err, domain.ErrJobNotFound):
		response.NotFound(w, "Job not found")
	case errors.Is(err, domain.ErrUnauthorized):
		response.Unauthorized(w, "Authentication required")
	case errors.Is(err, domain.ErrForbidden):
		response.Forbidden(w, "Access denied")
	case errors.Is(err, domain.ErrInvalidInput):
		response.BadRequest(w, "Invalid input")
	default:
		log.Error("internal error", zap.Error(err))
		response.InternalServerError(w, "")
	}
}
