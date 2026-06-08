package wallet

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

// Handler handles HTTP requests for wallet operations.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new wallet handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers the wallet routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// User wallet routes
	mux.Handle("GET /wallet", middleware.RequireAuth(http.HandlerFunc(h.GetMyWallet)))
	mux.Handle("GET /wallet/transactions", middleware.RequireAuth(http.HandlerFunc(h.GetMyTransactions)))

	// Admin wallet routes
	mux.Handle("POST /admin/wallet/top-up", middleware.RequireAdmin(http.HandlerFunc(h.AdminTopUp)))
	mux.Handle("GET /admin/wallet/{user_id}", middleware.RequireAdmin(http.HandlerFunc(h.GetUserWallet)))
}

// GetMyWallet godoc
// @Summary      Get my wallet
// @Description  Returns the authenticated user's wallet balance
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  WalletResponse
// @Failure      401  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /wallet [get]
func (h *Handler) GetMyWallet(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	wallet, err := h.service.GetMyWallet(r.Context())
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToWalletResponse(wallet))
}

// GetMyTransactions godoc
// @Summary      Get my transactions
// @Description  Returns the authenticated user's wallet transaction history
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query     int  false  "Limit (default 20, max 100)"
// @Param        offset  query     int  false  "Offset (default 0)"
// @Success      200  {object}  TransactionListResponse
// @Failure      401  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /wallet/transactions [get]
func (h *Handler) GetMyTransactions(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	txs, total, err := h.service.GetMyTransactions(r.Context(), limit, offset)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToTransactionListResponse(txs, total, limit, offset))
}

// AdminTopUp godoc
// @Summary      Top up user wallet (admin)
// @Description  Adds credits to a user's wallet (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      AdminTopUpRequest  true  "Top-up details"
// @Success      200   {object}  WalletResponse
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Failure      403   {object}  response.Error
// @Failure      404   {object}  response.Error
// @Failure      422   {object}  response.ValidationError
// @Router       /admin/wallet/top-up [post]
func (h *Handler) AdminTopUp(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	var req AdminTopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if errs := req.Validate(); errs != nil {
		response.UnprocessableEntity(w, errs)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.BadRequest(w, "Invalid user_id")
		return
	}

	wallet, err := h.service.AdminTopUp(r.Context(), userID, req.Amount, req.Description)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToWalletResponse(wallet))
}

// GetUserWallet godoc
// @Summary      Get user wallet (admin)
// @Description  Returns a user's wallet balance (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path      string  true  "User ID"
// @Success      200  {object}  WalletResponse
// @Failure      401  {object}  response.Error
// @Failure      403  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Router       /admin/wallet/{user_id} [get]
func (h *Handler) GetUserWallet(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	userID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	wallet, err := h.service.GetWalletByUserID(r.Context(), userID)
	if err != nil {
		h.handleError(w, log, err)
		return
	}

	response.OK(w, ToWalletResponse(wallet))
}

// handleError maps domain errors to HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, log *zap.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrWalletNotFound):
		response.NotFound(w, "Wallet not found")
	case errors.Is(err, domain.ErrWalletAlreadyExists):
		response.Conflict(w, "Wallet already exists for this user")
	case errors.Is(err, domain.ErrInsufficientCredits):
		response.ForbiddenWithCode(w, "Insufficient credits", "INSUFFICIENT_CREDITS")
	case errors.Is(err, domain.ErrUnauthorized):
		response.Unauthorized(w, "Authentication required")
	case errors.Is(err, domain.ErrForbidden):
		response.Forbidden(w, "Access denied")
	default:
		log.Error("internal error", zap.Error(err))
		response.InternalServerError(w, "")
	}
}
