package middleware

import (
	"net/http"
	"strings"

	"fixapp/internal/auth"
	"fixapp/internal/auth/role"
	"fixapp/internal/auth/token"
	"fixapp/pkg/response"
)

// JWTAuth is middleware that extracts and validates JWT tokens.
// It attaches the authenticated user to the request context.
type JWTAuth struct {
	tokenService *token.Service
}

// NewJWTAuth creates a new JWT authentication middleware.
func NewJWTAuth(tokenService *token.Service) *JWTAuth {
	return &JWTAuth{tokenService: tokenService}
}

// Middleware returns the HTTP middleware handler.
// It extracts the JWT from the Authorization header and validates it.
// If valid, the user is attached to the context.
// If invalid or missing, the request continues without a user (for optional auth).
func (j *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := extractToken(r)
		if tokenString == "" {
			// No token - continue without auth (handler can check if auth required)
			next.ServeHTTP(w, r)
			return
		}

		claims, err := j.tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			// Invalid token - continue without auth
			// Specific endpoints will return 401 if they require auth
			next.ServeHTTP(w, r)
			return
		}

		// Create auth user from claims
		user := &auth.User{
			ID:       claims.UserID,
			Email:    claims.Email,
			Name:     claims.Name,
			Role:     claims.Role,
			Provider: claims.Provider,
		}

		// Attach user to context
		ctx := auth.NewContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Required returns middleware that requires valid authentication.
// Returns 401 if no valid token is present.
func (j *JWTAuth) Required(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := extractToken(r)
		if tokenString == "" {
			response.UnauthorizedWithCode(w, "Authentication required", response.CodeTokenMissing)
			return
		}

		claims, err := j.tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			switch err {
			case token.ErrTokenExpired:
				response.UnauthorizedWithCode(w, "Token has expired", response.CodeTokenExpired)
			case token.ErrTokenMalformed:
				response.UnauthorizedWithCode(w, "Malformed token", response.CodeTokenInvalid)
			default:
				response.UnauthorizedWithCode(w, "Invalid token", response.CodeTokenInvalid)
			}
			return
		}

		// Create auth user from claims
		user := &auth.User{
			ID:       claims.UserID,
			Email:    claims.Email,
			Name:     claims.Name,
			Role:     claims.Role,
			Provider: claims.Provider,
		}

		// Attach user to context
		ctx := auth.NewContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns middleware that requires a specific role.
func (j *JWTAuth) RequireRole(r role.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return j.Required(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user := auth.FromContext(req.Context())
			if user == nil || user.Role != r {
				response.Forbidden(w, "Access denied: insufficient role")
				return
			}
			next.ServeHTTP(w, req)
		}))
	}
}

// RequireRoleAtLeast returns middleware that requires at least a specific role level.
func (j *JWTAuth) RequireRoleAtLeast(r role.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return j.Required(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user := auth.FromContext(req.Context())
			if user == nil || !user.Role.IsAtLeast(r) {
				response.Forbidden(w, "Access denied: insufficient privileges")
				return
			}
			next.ServeHTTP(w, req)
		}))
	}
}

// RequireAdmin returns middleware that requires Admin role.
func (j *JWTAuth) RequireAdmin() func(http.Handler) http.Handler {
	return j.RequireRole(role.Admin)
}

// extractToken gets the JWT from the Authorization header.
// Expects format: "Bearer <token>"
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return parts[1]
}


