package user

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

// Handler handles HTTP requests for user operations.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new user handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers the user routes on the given mux.
// It uses middleware for authentication and authorization.
func (h *Handler) Register(mux *http.ServeMux) {
	// Profile routes (authenticated users)
	mux.Handle("GET /profile", middleware.RequireAuth(http.HandlerFunc(h.GetProfile)))
	mux.Handle("PATCH /profile", middleware.RequireAuth(http.HandlerFunc(h.UpdateProfile)))

	// Admin routes
	mux.Handle("GET /admin/users", middleware.RequireAdmin(http.HandlerFunc(h.ListUsers)))
	mux.Handle("GET /admin/users/{id}", middleware.RequireAdmin(http.HandlerFunc(h.GetUser)))
	mux.Handle("PATCH /admin/users/{id}/role", middleware.RequireAdmin(http.HandlerFunc(h.UpdateUserRole)))
	mux.Handle("DELETE /admin/users/{id}", middleware.RequireAdmin(http.HandlerFunc(h.DeactivateUser)))
}

// GetProfile godoc
// @Summary      Get current user's profile
// @Description  Returns the profile of the authenticated user
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  ProfileResponse
// @Failure      401  {object}  response.Error
// @Router       /profile [get]
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	user, err := h.service.GetProfile(r.Context())
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToProfileResponse(user))
}

// UpdateProfile godoc
// @Summary      Update current user's profile
// @Description  Updates the profile of the authenticated user
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      UpdateProfileRequest  true  "Profile update"
// @Success      200   {object}  ProfileResponse
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Failure      422   {object}  response.ValidationError
// @Router       /profile [patch]
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.service.UpdateProfile(r.Context(), req.Name, req.Phone, req.AvatarURL)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToProfileResponse(user))
}

// ListUsers godoc
// @Summary      List all users
// @Description  Returns a paginated list of users (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        role      query     string  false  "Filter by role"
// @Param        is_active query     bool    false  "Filter by active status"
// @Param        search    query     string  false  "Search by email or name"
// @Param        limit     query     int     false  "Limit (default 20, max 100)"
// @Param        offset    query     int     false  "Offset (default 0)"
// @Success      200  {object}  UserListResponse
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Router       /admin/users [get]
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	req := ListUsersRequest{
		Role:   r.URL.Query().Get("role"),
		Search: r.URL.Query().Get("search"),
	}

	if isActive := r.URL.Query().Get("is_active"); isActive != "" {
		b, _ := strconv.ParseBool(isActive)
		req.IsActive = &b
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		req.Limit, _ = strconv.Atoi(limit)
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		req.Offset, _ = strconv.Atoi(offset)
	}

	filter := req.ToFilter()
	users, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToUserListResponse(users, total, filter.Limit, filter.Offset))
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Returns a user by their ID (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /admin/users/{id} [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	user, err := h.service.GetByIDForAdmin(r.Context(), id)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToUserResponse(user))
}

// UpdateUserRole godoc
// @Summary      Update user role
// @Description  Changes a user's role (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string            true  "User ID"
// @Param        body  body      UpdateRoleRequest true  "Role update"
// @Success      200   {object}  response.Success
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Failure      403   {object}  response.Error
// @Failure      404   {object}  response.Error
// @Failure      422   {object}  response.ValidationError
// @Router       /admin/users/{id}/role [patch]
func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); errs != nil {
		response.UnprocessableEntity(w, errs)
		return
	}

	if err := h.service.UpdateRole(r.Context(), id, req.GetRole()); err != nil {
		h.handleError(w, log, err)
		return
	}

	response.JSON(w, http.StatusOK, response.Success{
		Success: true,
		Message: "Role updated successfully",
	})
}

// DeactivateUser godoc
// @Summary      Deactivate user
// @Description  Disables a user account (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      204
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /admin/users/{id} [delete]
func (h *Handler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	if err := h.service.Deactivate(r.Context(), id); err != nil {
		h.handleError(w, log, err)
		return
	}

	response.NoContent(w)
}

// handleError maps domain errors to HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, log *zap.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		response.NotFound(w, "User not found")

	case errors.Is(err, domain.ErrUserAlreadyExists):
		response.Conflict(w, "User already exists")

	case errors.Is(err, domain.ErrUnauthorized):
		response.Unauthorized(w, "Authentication required")

	case errors.Is(err, domain.ErrForbidden),
		errors.Is(err, domain.ErrInsufficientRole),
		errors.Is(err, domain.ErrMissingPermission):
		response.Forbidden(w, "Access denied")

	case errors.Is(err, domain.ErrInvalidInput):
		response.BadRequest(w, "Invalid input")

	case errors.Is(err, domain.ErrUserDisabled):
		response.ForbiddenWithCode(w, "Account is disabled", response.CodeAccountDisabled)

	default:
		log.Error("internal error", zap.Error(err))
		response.InternalServerError(w, "")
	}
}
