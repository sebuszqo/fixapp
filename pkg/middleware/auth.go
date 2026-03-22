package middleware

import (
	"net/http"

	"fixapp/internal/auth"
	"fixapp/internal/auth/permission"
	"fixapp/internal/auth/role"
	"fixapp/pkg/response"

	"go.uber.org/zap"
)

// RequireAuth is middleware that ensures the request has an authenticated user.
// Returns 401 Unauthorized if no user is present in the context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthenticated(r.Context()) {
			log := FromContextLogger(r.Context())
			log.Warn("unauthenticated request blocked")
			response.Unauthorized(w, "Authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole returns middleware that ensures the user has the exact role.
// Returns 403 Forbidden if the user doesn't have the required role.
func RequireRole(r role.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user := auth.FromContext(req.Context())
			if user == nil {
				log := FromContextLogger(req.Context())
				log.Warn("unauthenticated request blocked by role check")
				response.Unauthorized(w, "Authentication required")
				return
			}

			if user.Role != r {
				log := FromContextLogger(req.Context())
				log.Warn("access denied: insufficient role",
					zap.String("user_id", user.ID),
					zap.String("user_role", user.Role.String()),
					zap.String("required_role", r.String()),
				)
				response.Forbidden(w, "Access denied: insufficient role")
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// RequireRoleAtLeast returns middleware that ensures the user has at least the specified role level.
// Uses hierarchical role comparison (Admin > Handyman > User).
func RequireRoleAtLeast(r role.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user := auth.FromContext(req.Context())
			if user == nil {
				log := FromContextLogger(req.Context())
				log.Warn("unauthenticated request blocked by role check")
				response.Unauthorized(w, "Authentication required")
				return
			}

			if !user.Role.IsAtLeast(r) {
				log := FromContextLogger(req.Context())
				log.Warn("access denied: insufficient role level",
					zap.String("user_id", user.ID),
					zap.String("user_role", user.Role.String()),
					zap.String("minimum_role", r.String()),
				)
				response.Forbidden(w, "Access denied: insufficient privileges")
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// RequirePermission returns middleware that ensures the user has the specified permission.
func RequirePermission(p permission.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user := auth.FromContext(req.Context())
			if user == nil {
				log := FromContextLogger(req.Context())
				log.Warn("unauthenticated request blocked by permission check")
				response.Unauthorized(w, "Authentication required")
				return
			}

			if !user.HasPermission(p) {
				log := FromContextLogger(req.Context())
				log.Warn("access denied: missing permission",
					zap.String("user_id", user.ID),
					zap.String("user_role", user.Role.String()),
					zap.String("required_permission", p.String()),
				)
				response.Forbidden(w, "Access denied: missing permission")
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// RequireAnyPermission returns middleware that ensures the user has at least one of the permissions.
func RequireAnyPermission(perms ...permission.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user := auth.FromContext(req.Context())
			if user == nil {
				log := FromContextLogger(req.Context())
				log.Warn("unauthenticated request blocked by permission check")
				response.Unauthorized(w, "Authentication required")
				return
			}

			if !user.HasAnyPermission(perms...) {
				log := FromContextLogger(req.Context())
				log.Warn("access denied: missing all required permissions",
					zap.String("user_id", user.ID),
					zap.String("user_role", user.Role.String()),
				)
				response.Forbidden(w, "Access denied: missing permission")
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// RequireAllPermissions returns middleware that ensures the user has all of the permissions.
func RequireAllPermissions(perms ...permission.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user := auth.FromContext(req.Context())
			if user == nil {
				log := FromContextLogger(req.Context())
				log.Warn("unauthenticated request blocked by permission check")
				response.Unauthorized(w, "Authentication required")
				return
			}

			if !user.HasAllPermissions(perms...) {
				log := FromContextLogger(req.Context())
				log.Warn("access denied: missing some required permissions",
					zap.String("user_id", user.ID),
					zap.String("user_role", user.Role.String()),
				)
				response.Forbidden(w, "Access denied: missing permissions")
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// RequireAdmin is a convenience middleware that requires Admin role.
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(role.Admin)(next)
}

// RequireHandyman is a convenience middleware that requires at least Handyman role.
func RequireHandyman(next http.Handler) http.Handler {
	return RequireRoleAtLeast(role.Handyman)(next)
}

