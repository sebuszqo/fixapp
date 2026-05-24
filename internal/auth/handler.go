package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"fixapp/internal/auth/token"
	"fixapp/pkg/ctxlog"
	"fixapp/pkg/response"

	"go.uber.org/zap"
)

// Cookie configuration for secure token storage
const (
	refreshTokenCookieName = "refresh_token"
	refreshTokenMaxAge     = 7 * 24 * 60 * 60 // 7 days in seconds
)

// isSecure returns true if running in production (HTTPS)
func isSecure() bool {
	return os.Getenv("ENV") == "production" || os.Getenv("HTTPS") == "true"
}

// Handler handles HTTP requests for authentication.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new auth handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers the auth routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// OAuth endpoints
	mux.HandleFunc("GET /auth/providers", h.ListProviders)
	mux.HandleFunc("GET /auth/{provider}/login", h.Login)
	mux.HandleFunc("GET /auth/{provider}/callback", h.Callback)

	// Token endpoints
	mux.HandleFunc("POST /auth/refresh", h.RefreshToken)
	mux.HandleFunc("POST /auth/logout", h.Logout)
}

// ===== DTOs =====

// LoginResponse contains the OAuth authorization URL.
type LoginResponse struct {
	AuthURL string `json:"auth_url" example:"https://accounts.google.com/o/oauth2/v2/auth?..."`
	State   string `json:"state" example:"random_state_string"`
}

// TokenResponse contains the authentication tokens.
// Note: refresh_token is sent via HttpOnly cookie, NOT in JSON response.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in" example:"900"`
	TokenType   string `json:"token_type" example:"Bearer"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// RefreshTokenRequest is the payload for token refresh.
// Note: In secure mode, refresh_token comes from HttpOnly cookie.
// This struct is kept for backwards compatibility with mobile apps.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// setRefreshTokenCookie sets the refresh token as an HttpOnly cookie.
func setRefreshTokenCookie(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/auth", // Only sent to auth endpoints
		MaxAge:   refreshTokenMaxAge,
		HttpOnly: true,                    // JavaScript cannot access this cookie
		Secure:   isSecure(),              // Only HTTPS in production
		SameSite: http.SameSiteStrictMode, // CSRF protection
	})
}

// clearRefreshTokenCookie removes the refresh token cookie.
func clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/auth",
		MaxAge:   -1, // Delete immediately
		HttpOnly: true,
		Secure:   isSecure(),
		SameSite: http.SameSiteStrictMode,
	})
}

// getRefreshTokenFromRequest gets refresh token from cookie or request body.
// Cookie takes priority (more secure). Body is fallback for mobile apps.
func getRefreshTokenFromRequest(r *http.Request, bodyToken string) string {
	// Try cookie first (web apps)
	if cookie, err := r.Cookie(refreshTokenCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	// Fallback to body (mobile apps)
	return bodyToken
}

// ProviderInfo describes an available auth provider.
type ProviderInfo struct {
	Name string `json:"name" example:"google"`
	Type string `json:"type" example:"oauth"`
}

// ProvidersResponse lists available auth providers.
type ProvidersResponse struct {
	Providers []ProviderInfo `json:"providers"`
}

// ===== Handlers =====

// ListProviders godoc
// @Summary      List available auth providers
// @Description  Returns a list of available authentication providers
// @Tags         auth
// @Produce      json
// @Success      200  {object}  ProvidersResponse
// @Router       /auth/providers [get]
func (h *Handler) ListProviders(w http.ResponseWriter, r *http.Request) {
	names := h.service.ListProviders()

	providers := make([]ProviderInfo, len(names))
	for i, name := range names {
		providers[i] = ProviderInfo{
			Name: name,
			Type: "oauth",
		}
	}

	response.OK(w, ProvidersResponse{Providers: providers})
}

// Login godoc
// @Summary      Start OAuth login flow
// @Description  Returns the OAuth authorization URL for the specified provider
// @Tags         auth
// @Produce      json
// @Param        provider  path      string  true   "Provider name (google, facebook)"
// @Param        redirect  query     string  false  "URL to redirect after login"
// @Param        as        query     string  false  "Register as: 'user' (default) or 'handyman'"
// @Success      200       {object}  LoginResponse
// @Failure      400       {object}  response.Error
// @Router       /auth/{provider}/login [get]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")
	redirectURL := r.URL.Query().Get("redirect")
	registerAs := r.URL.Query().Get("as") // "user" or "handyman"

	if redirectURL == "" {
		redirectURL = "/" // Default redirect
	}

	authURL, state, err := h.service.GetAuthURL(providerName, redirectURL, registerAs)
	if err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			response.BadRequest(w, "Unknown authentication provider")
			return
		}
		h.logger.Error("failed to get auth URL", zap.Error(err))
		response.InternalServerError(w, "")
		return
	}

	response.OK(w, LoginResponse{
		AuthURL: authURL,
		State:   state,
	})
}

// Callback godoc
// @Summary      OAuth callback handler
// @Description  Handles the OAuth callback and returns authentication tokens
// @Tags         auth
// @Produce      json
// @Param        provider  path      string  true  "Provider name (google, facebook)"
// @Param        code      query     string  true  "Authorization code"
// @Param        state     query     string  true  "OAuth state"
// @Success      200       {object}  TokenResponse
// @Failure      400       {object}  response.Error
// @Failure      401       {object}  response.Error
// @Router       /auth/{provider}/callback [get]
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	providerName := r.PathValue("provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	// Check for OAuth error
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		log.Warn("OAuth error",
			zap.String("error", errParam),
			zap.String("description", errDesc),
		)
		response.BadRequest(w, "Authentication failed: "+errDesc)
		return
	}

	if code == "" || state == "" {
		response.BadRequest(w, "Missing code or state parameter")
		return
	}

	tokens, redirectURL, err := h.service.HandleCallback(r.Context(), providerName, code, state)
	if err != nil {
		h.handleAuthError(w, log, err)
		return
	}

	// Set refresh token as HttpOnly cookie (secure, not accessible to JavaScript)
	setRefreshTokenCookie(w, tokens.RefreshToken)

	// Return only access token in JSON (frontend stores in memory, NOT localStorage)
	response.OK(w, TokenResponse{
		AccessToken: tokens.AccessToken,
		ExpiresIn:   900, // 15 minutes in seconds
		TokenType:   tokens.TokenType,
		RedirectURL: redirectURL,
	})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchange a refresh token for a new access token. Refresh token can come from HttpOnly cookie (web) or request body (mobile).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RefreshTokenRequest  false  "Refresh token (optional, for mobile apps)"
// @Success      200   {object}  TokenResponse
// @Failure      400   {object}  response.Error
// @Failure      401   {object}  response.Error
// @Router       /auth/refresh [post]
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.FromContext(r.Context())

	// Try to get refresh token from request body (for mobile apps)
	var req RefreshTokenRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // Ignore error - body might be empty

	// Get refresh token from cookie or body
	refreshToken := getRefreshTokenFromRequest(r, req.RefreshToken)
	if refreshToken == "" {
		response.BadRequest(w, "Refresh token is required")
		return
	}

	tokens, err := h.service.RefreshTokens(r.Context(), refreshToken)
	if err != nil {
		h.handleAuthError(w, log, err)
		return
	}

	// Update the refresh token cookie (token rotation for security)
	setRefreshTokenCookie(w, tokens.RefreshToken)

	response.OK(w, TokenResponse{
		AccessToken: tokens.AccessToken,
		ExpiresIn:   900, // 15 minutes in seconds
		TokenType:   tokens.TokenType,
	})
}

// Logout godoc
// @Summary      Logout user
// @Description  Clears the refresh token cookie and invalidates the session
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.Success
// @Router       /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the refresh token cookie
	clearRefreshTokenCookie(w)

	// For JWT-based auth, logout is primarily client-side (discard access token from memory)
	// In production, you might also want to:
	// - Add access token to a blacklist (in Redis) until it expires
	// - Revoke refresh tokens in database for full token rotation

	response.JSON(w, http.StatusOK, response.Success{
		Success: true,
		Message: "Logged out successfully",
	})
}

// handleAuthError maps auth errors to HTTP responses.
func (h *Handler) handleAuthError(w http.ResponseWriter, log *zap.Logger, err error) {
	switch {
	case errors.Is(err, ErrProviderNotFound):
		response.BadRequest(w, "Unknown authentication provider")

	case errors.Is(err, ErrInvalidState):
		response.Unauthorized(w, "Invalid or expired authentication session")

	case errors.Is(err, ErrUserDisabled):
		response.ForbiddenWithCode(w, "Account is disabled", response.CodeAccountDisabled)

	case errors.Is(err, token.ErrTokenExpired):
		response.UnauthorizedWithCode(w, "Token has expired", response.CodeTokenExpired)

	case errors.Is(err, token.ErrTokenInvalid):
		response.UnauthorizedWithCode(w, "Invalid token", response.CodeTokenInvalid)

	default:
		log.Error("auth error", zap.Error(err))
		response.InternalServerError(w, "")
	}
}
