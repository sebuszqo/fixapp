package lead

import (
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

// Handler handles HTTP requests for lead operations.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new lead handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers the lead routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// Handyman lead routes
	mux.Handle("GET /leads", middleware.RequireHandyman(http.HandlerFunc(h.ListMyLeads)))
	mux.Handle("GET /leads/{id}", middleware.RequireHandyman(http.HandlerFunc(h.GetLead)))
	mux.Handle("POST /leads/{id}/accept", middleware.RequireHandyman(http.HandlerFunc(h.AcceptLead)))
	mux.Handle("POST /leads/{id}/reject", middleware.RequireHandyman(http.HandlerFunc(h.RejectLead)))
}

// ListMyLeads godoc
// @Summary      List my leads
// @Description  Returns the authenticated handyman's leads with optional status filter
// @Tags         leads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        status  query     string  false  "Filter by status (pending, accepted, rejected, expired)"
// @Param        limit   query     int     false  "Limit (default 20, max 100)"
// @Param        offset  query     int     false  "Offset (default 0)"
// @Success      200  {object}  LeadListResponse
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Router       /leads [get]
func (h *Handler) ListMyLeads(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	var status *domain.LeadStatus
	if s := r.URL.Query().Get("status"); s != "" {
		ls := domain.LeadStatus(s)
		if !ls.IsValid() {
			response.BadRequest(w, "Invalid status filter")
			return
		}
		status = &ls
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	leads, total, err := h.service.ListMyLeads(r.Context(), status, limit, offset)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToLeadListResponse(leads, total, limit, offset))
}

// GetLead godoc
// @Summary      Get lead details
// @Description  Returns a lead with full job details
// @Tags         leads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Lead ID"
// @Success      200  {object}  LeadDetailResponse
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /leads/{id} [get]
func (h *Handler) GetLead(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid lead ID")
		return
	}

	lead, job, err := h.service.GetDetail(r.Context(), id)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	// Only reveal address after lead is accepted
	revealAddress := lead.Status == domain.LeadStatusAccepted
	response.OK(w, ToLeadDetailResponse(lead, job, revealAddress))
}

// AcceptLead godoc
// @Summary      Accept a lead
// @Description  Accepts a lead, deducting credits and updating the job status
// @Tags         leads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Lead ID"
// @Success      200  {object}  LeadResponse
// @Failure      400  {object}  response.Error
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Failure      409  {object}  response.Error
// @Router       /leads/{id}/accept [post]
func (h *Handler) AcceptLead(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid lead ID")
		return
	}

	lead, err := h.service.Accept(r.Context(), id)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToLeadResponse(lead))
}

// RejectLead godoc
// @Summary      Reject a lead
// @Description  Rejects a lead (no credits are charged)
// @Tags         leads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Lead ID"
// @Success      200  {object}  LeadResponse
// @Failure      400  {object}  response.Error
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /leads/{id}/reject [post]
func (h *Handler) RejectLead(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid lead ID")
		return
	}

	lead, err := h.service.Reject(r.Context(), id)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToLeadResponse(lead))
}

// handleError maps domain errors to HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, log *zap.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrLeadNotFound):
		response.NotFound(w, "Lead not found")
	case errors.Is(err, domain.ErrInvalidLeadTransition):
		response.BadRequest(w, "Invalid status transition for this lead")
	case errors.Is(err, domain.ErrLeadExpired):
		response.BadRequest(w, "Lead has expired")
	case errors.Is(err, domain.ErrLeadAlreadyAccepted):
		response.Conflict(w, "Lead already accepted")
	case errors.Is(err, domain.ErrJobAlreadyAccepted):
		response.Conflict(w, "Job already accepted by another handyman")
	case errors.Is(err, domain.ErrInsufficientCredits):
		response.ForbiddenWithCode(w, "Insufficient credits to accept this lead", "INSUFFICIENT_CREDITS")
	case errors.Is(err, domain.ErrUnauthorized):
		response.Unauthorized(w, "Authentication required")
	case errors.Is(err, domain.ErrForbidden):
		response.Forbidden(w, "Access denied")
	case errors.Is(err, domain.ErrJobNotFound):
		response.NotFound(w, "Associated job not found")
	default:
		log.Error("internal error", zap.Error(err))
		response.InternalServerError(w, "")
	}
}
