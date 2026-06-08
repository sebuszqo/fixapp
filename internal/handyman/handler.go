package handyman

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

// Handler handles HTTP requests for handyman profile operations.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new handyman handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers the handyman routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// Public search
	mux.Handle("GET /handymen", http.HandlerFunc(h.SearchProfiles))
	mux.Handle("GET /handymen/{id}", http.HandlerFunc(h.GetPublicProfile))

	// Handyman's own profile management
	mux.Handle("GET /handyman/profile", middleware.RequireHandyman(http.HandlerFunc(h.GetMyProfile)))
	mux.Handle("PATCH /handyman/profile", middleware.RequireHandyman(http.HandlerFunc(h.UpdateMyProfile)))

	// Pricing management
	mux.Handle("POST /handyman/pricing", middleware.RequireHandyman(http.HandlerFunc(h.AddPricing)))
	mux.Handle("DELETE /handyman/pricing/{id}", middleware.RequireHandyman(http.HandlerFunc(h.DeletePricing)))

	// Portfolio management
	mux.Handle("POST /handyman/portfolio", middleware.RequireHandyman(http.HandlerFunc(h.AddPortfolio)))
	mux.Handle("DELETE /handyman/portfolio/{id}", middleware.RequireHandyman(http.HandlerFunc(h.DeletePortfolio)))
}

// SearchProfiles godoc
// @Summary      Search handymen
// @Description  Search for available handyman profiles with filters
// @Tags         handymen
// @Accept       json
// @Produce      json
// @Param        category_id  query     string  false  "Filter by category ID"
// @Param        district_id  query     string  false  "Filter by district ID"
// @Param        verified     query     bool    false  "Filter by verification status"
// @Param        search       query     string  false  "Search by name or bio"
// @Param        limit        query     int     false  "Limit (default 20, max 100)"
// @Param        offset       query     int     false  "Offset (default 0)"
// @Success      200  {object}  ProfileListResponse
// @Router       /handymen [get]
func (h *Handler) SearchProfiles(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	filter := DefaultSearchFilter()

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
	if v := r.URL.Query().Get("verified"); v != "" {
		b, _ := strconv.ParseBool(v)
		filter.IsVerified = &b
	}
	filter.Search = r.URL.Query().Get("search")

	if l := r.URL.Query().Get("limit"); l != "" {
		filter.Limit, _ = strconv.Atoi(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		filter.Offset, _ = strconv.Atoi(o)
	}

	profiles, total, err := h.service.Search(r.Context(), filter)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToProfileListResponse(profiles, total, filter.Limit, filter.Offset))
}

// GetPublicProfile godoc
// @Summary      Get handyman public profile
// @Description  Returns a handyman's full public profile with pricing and portfolio
// @Tags         handymen
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Profile ID"
// @Success      200  {object}  FullProfileResponse
// @Failure      404  {object}  response.Error
// @Router       /handymen/{id} [get]
func (h *Handler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid profile ID")
		return
	}

	profile, pricing, portfolio, err := h.service.GetFullProfile(r.Context(), id)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToFullProfileResponse(profile, pricing, portfolio))
}

// GetMyProfile godoc
// @Summary      Get my handyman profile
// @Description  Returns the authenticated handyman's full profile
// @Tags         handyman-profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  FullProfileResponse
// @Failure      401  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /handyman/profile [get]
func (h *Handler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	profile, pricing, portfolio, err := h.service.GetMyFullProfile(r.Context())
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToFullProfileResponse(profile, pricing, portfolio))
}

// UpdateMyProfile godoc
// @Summary      Update my handyman profile
// @Description  Updates the authenticated handyman's profile
// @Tags         handyman-profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      UpdateProfileRequest  true  "Profile update"
// @Success      200   {object}  ProfileResponse
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Failure      422   {object}  response.ValidationError
// @Router       /handyman/profile [patch]
func (h *Handler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); errs != nil {
		response.UnprocessableEntity(w, errs)
		return
	}

	profile, err := h.service.UpdateProfile(r.Context(), req)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToProfileResponse(profile))
}

// AddPricing godoc
// @Summary      Add pricing item
// @Description  Adds a new service pricing item to the handyman's profile
// @Tags         handyman-profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreatePricingRequest  true  "Pricing item"
// @Success      201   {object}  PricingItemResponse
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Failure      422   {object}  response.ValidationError
// @Router       /handyman/pricing [post]
func (h *Handler) AddPricing(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	var req CreatePricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); errs != nil {
		response.UnprocessableEntity(w, errs)
		return
	}

	item, err := h.service.AddPricing(r.Context(), req)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.Created(w, ToPricingItemResponse(item))
}

// DeletePricing godoc
// @Summary      Delete pricing item
// @Description  Removes a pricing item from the handyman's profile
// @Tags         handyman-profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Pricing item ID"
// @Success      204
// @Failure      401  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /handyman/pricing/{id} [delete]
func (h *Handler) DeletePricing(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid pricing item ID")
		return
	}

	if err := h.service.DeletePricing(r.Context(), id); err != nil {
		h.handleError(w, log, err)
		return
	}

	response.NoContent(w)
}

// AddPortfolio godoc
// @Summary      Add portfolio photo
// @Description  Adds a new photo to the handyman's portfolio
// @Tags         handyman-profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      AddPortfolioRequest  true  "Portfolio item"
// @Success      201   {object}  PortfolioItemResponse
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Failure      422   {object}  response.ValidationError
// @Router       /handyman/portfolio [post]
func (h *Handler) AddPortfolio(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	var req AddPortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); errs != nil {
		response.UnprocessableEntity(w, errs)
		return
	}

	item, err := h.service.AddPortfolio(r.Context(), req)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.Created(w, ToPortfolioItemResponse(item))
}

// DeletePortfolio godoc
// @Summary      Delete portfolio photo
// @Description  Removes a photo from the handyman's portfolio
// @Tags         handyman-profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Portfolio item ID"
// @Success      204
// @Failure      401  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /handyman/portfolio/{id} [delete]
func (h *Handler) DeletePortfolio(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid portfolio item ID")
		return
	}

	if err := h.service.DeletePortfolio(r.Context(), id); err != nil {
		h.handleError(w, log, err)
		return
	}

	response.NoContent(w)
}

// handleError maps domain errors to HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, log *zap.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrProfileNotFound):
		response.NotFound(w, "Handyman profile not found")
	case errors.Is(err, domain.ErrProfileAlreadyExists):
		response.Conflict(w, "Handyman profile already exists")
	case errors.Is(err, domain.ErrTooManyCategories):
		response.BadRequest(w, "Maximum 3 categories allowed")
	case errors.Is(err, domain.ErrPricingItemNotFound):
		response.NotFound(w, "Pricing item not found")
	case errors.Is(err, domain.ErrPortfolioItemNotFound):
		response.NotFound(w, "Portfolio item not found")
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
