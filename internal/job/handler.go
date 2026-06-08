package job

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

// Handler handles HTTP requests for job operations.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new job handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers the job routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// Client job routes
	mux.Handle("POST /jobs", middleware.RequireAuth(http.HandlerFunc(h.CreateJob)))
	mux.Handle("GET /jobs", middleware.RequireAuth(http.HandlerFunc(h.ListMyJobs)))
	mux.Handle("GET /jobs/{id}", middleware.RequireAuth(http.HandlerFunc(h.GetJob)))
	mux.Handle("POST /jobs/{id}/publish", middleware.RequireAuth(http.HandlerFunc(h.PublishJob)))
	mux.Handle("POST /jobs/{id}/complete", middleware.RequireHandyman(http.HandlerFunc(h.CompleteJob)))
	mux.Handle("POST /jobs/{id}/confirm", middleware.RequireAuth(http.HandlerFunc(h.ConfirmJob)))
	mux.Handle("POST /jobs/{id}/cancel", middleware.RequireAuth(http.HandlerFunc(h.CancelJob)))

	// Admin job routes
	mux.Handle("GET /admin/jobs", middleware.RequireAdmin(http.HandlerFunc(h.ListAllJobs)))
}

// CreateJob godoc
// @Summary      Create a new job
// @Description  Creates a new service request (job) for the authenticated client
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateJobRequest  true  "Job details"
// @Success      201   {object}  JobResponse
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Failure      422   {object}  response.ValidationError
// @Router       /jobs [post]
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); errs != nil {
		response.UnprocessableEntity(w, errs)
		return
	}

	job, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.Created(w, ToJobResponse(job))
}

// GetJob godoc
// @Summary      Get job by ID
// @Description  Returns a job by its ID
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  JobResponse
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /jobs/{id} [get]
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid job ID")
		return
	}

	job, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToJobResponse(job))
}

// ListMyJobs godoc
// @Summary      List my jobs
// @Description  Returns the authenticated user's jobs with optional status filter
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        status  query     string  false  "Filter by status (draft, active, accepted, in_progress, done, cancelled)"
// @Param        limit   query     int     false  "Limit (default 20, max 100)"
// @Param        offset  query     int     false  "Offset (default 0)"
// @Success      200  {object}  JobListResponse
// @Failure      401  {object}  response.Error
// @Router       /jobs [get]
func (h *Handler) ListMyJobs(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	var status *domain.JobStatus
	if s := r.URL.Query().Get("status"); s != "" {
		js := domain.JobStatus(s)
		if !js.IsValid() {
			response.BadRequest(w, "Invalid status filter")
			return
		}
		status = &js
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	jobs, total, err := h.service.ListMyJobs(r.Context(), status, limit, offset)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToJobListResponse(jobs, total, limit, offset))
}

// PublishJob godoc
// @Summary      Publish a job
// @Description  Moves a draft job to active status, making it visible to handymen
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  JobResponse
// @Failure      400  {object}  response.Error
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /jobs/{id}/publish [post]
func (h *Handler) PublishJob(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid job ID")
		return
	}

	job, err := h.service.Publish(r.Context(), id)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToJobResponse(job))
}

// CompleteJob godoc
// @Summary      Complete a job
// @Description  Marks a job as done with a declared final value (handyman only)
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string             true  "Job ID"
// @Param        body  body      CompleteJobRequest  true  "Completion details"
// @Success      200   {object}  JobResponse
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Failure      403   {object}  response.Error
// @Failure      404   {object}  response.Error
// @Router       /jobs/{id}/complete [post]
func (h *Handler) CompleteJob(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid job ID")
		return
	}

	var req CompleteJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); errs != nil {
		response.UnprocessableEntity(w, errs)
		return
	}

	job, err := h.service.Complete(r.Context(), id, req.FinalValue)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToJobResponse(job))
}

// ConfirmJob godoc
// @Summary      Confirm job completion
// @Description  Client confirms that the job was completed successfully
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  JobResponse
// @Failure      400  {object}  response.Error
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /jobs/{id}/confirm [post]
func (h *Handler) ConfirmJob(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid job ID")
		return
	}

	job, err := h.service.Confirm(r.Context(), id)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToJobResponse(job))
}

// CancelJob godoc
// @Summary      Cancel a job
// @Description  Cancels a job (client can cancel their own, admin can cancel any)
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  JobResponse
// @Failure      400  {object}  response.Error
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /jobs/{id}/cancel [post]
func (h *Handler) CancelJob(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid job ID")
		return
	}

	job, err := h.service.Cancel(r.Context(), id)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToJobResponse(job))
}

// ListAllJobs godoc
// @Summary      List all jobs (admin)
// @Description  Returns a paginated list of all jobs (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        status      query     string  false  "Filter by status"
// @Param        category_id query     string  false  "Filter by category"
// @Param        district_id query     string  false  "Filter by district"
// @Param        client_id   query     string  false  "Filter by client"
// @Param        limit       query     int     false  "Limit (default 20, max 100)"
// @Param        offset      query     int     false  "Offset (default 0)"
// @Success      200  {object}  JobListResponse
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Router       /admin/jobs [get]
func (h *Handler) ListAllJobs(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	filter := DefaultListFilter()

	if s := r.URL.Query().Get("status"); s != "" {
		js := domain.JobStatus(s)
		if js.IsValid() {
			filter.Status = &js
		}
	}
	if c := r.URL.Query().Get("category_id"); c != "" {
		if id, err := uuid.Parse(c); err == nil {
			filter.CategoryID = &id
		}
	}
	if d := r.URL.Query().Get("district_id"); d != "" {
		if id, err := uuid.Parse(d); err == nil {
			filter.DistrictID = &id
		}
	}
	if cl := r.URL.Query().Get("client_id"); cl != "" {
		if id, err := uuid.Parse(cl); err == nil {
			filter.ClientID = &id
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			filter.Limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			filter.Offset = v
		}
	}

	jobs, total, err := h.service.ListAll(r.Context(), filter)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToJobListResponse(jobs, total, filter.Limit, filter.Offset))
}

// handleError maps domain errors to HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, log *zap.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrJobNotFound):
		response.NotFound(w, "Job not found")
	case errors.Is(err, domain.ErrInvalidJobTransition):
		response.BadRequest(w, "Invalid status transition for this job")
	case errors.Is(err, domain.ErrJobExpired):
		response.BadRequest(w, "Job has expired")
	case errors.Is(err, domain.ErrJobAlreadyAccepted):
		response.Conflict(w, "Job already accepted by another handyman")
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
