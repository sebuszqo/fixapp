package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fixapp/internal/auth"
	"fixapp/internal/auth/permission"
	"fixapp/internal/auth/role"
)

func TestRequireAuth(t *testing.T) {
	// Handler that should only be reached if authenticated
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows authenticated request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.User})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireAuth(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("RequireAuth allowed request, got status %d", rr.Code)
		}
	})

	t.Run("blocks unauthenticated request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		RequireAuth(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("RequireAuth should block, got status %d", rr.Code)
		}
	})
}

func TestRequireRole(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows matching role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.Admin})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireRole(role.Admin)(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("RequireRole should allow matching role, got status %d", rr.Code)
		}
	})

	t.Run("blocks non-matching role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.User})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireRole(role.Admin)(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("RequireRole should block, got status %d", rr.Code)
		}
	})

	t.Run("blocks unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		RequireRole(role.Admin)(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("RequireRole should return 401, got status %d", rr.Code)
		}
	})
}

func TestRequireRoleAtLeast(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		userRole     role.Role
		requiredRole role.Role
		wantStatus   int
	}{
		{"admin passes admin check", role.Admin, role.Admin, http.StatusOK},
		{"admin passes handyman check", role.Admin, role.Handyman, http.StatusOK},
		{"admin passes user check", role.Admin, role.User, http.StatusOK},
		{"handyman passes handyman check", role.Handyman, role.Handyman, http.StatusOK},
		{"handyman passes user check", role.Handyman, role.User, http.StatusOK},
		{"handyman fails admin check", role.Handyman, role.Admin, http.StatusForbidden},
		{"user passes user check", role.User, role.User, http.StatusOK},
		{"user fails handyman check", role.User, role.Handyman, http.StatusForbidden},
		{"user fails admin check", role.User, role.Admin, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: tt.userRole})
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			RequireRoleAtLeast(tt.requiredRole)(handler).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows with permission", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.Admin})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequirePermission(permission.AdminAccess)(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("should allow with permission, got status %d", rr.Code)
		}
	})

	t.Run("blocks without permission", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.User})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequirePermission(permission.AdminAccess)(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("should block without permission, got status %d", rr.Code)
		}
	})
}

func TestRequireAnyPermission(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows with any permission", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.User})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireAnyPermission(permission.ProfileRead, permission.AdminAccess)(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("should allow with any permission, got status %d", rr.Code)
		}
	})

	t.Run("blocks without any permission", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.User})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireAnyPermission(permission.AdminAccess, permission.AdminUserMgmt)(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("should block without any permission, got status %d", rr.Code)
		}
	})
}

func TestRequireAllPermissions(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows with all permissions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.Admin})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireAllPermissions(permission.AdminAccess, permission.AdminUserMgmt)(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("should allow with all permissions, got status %d", rr.Code)
		}
	})

	t.Run("blocks without all permissions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.User})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireAllPermissions(permission.ProfileRead, permission.AdminAccess)(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("should block without all permissions, got status %d", rr.Code)
		}
	})
}

func TestRequireAdmin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.Admin})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireAdmin(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("should allow admin, got status %d", rr.Code)
		}
	})

	t.Run("blocks non-admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.Handyman})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireAdmin(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("should block non-admin, got status %d", rr.Code)
		}
	})
}

func TestRequireHandyman(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.Admin})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireHandyman(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("should allow admin (higher role), got status %d", rr.Code)
		}
	})

	t.Run("allows handyman", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.Handyman})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireHandyman(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("should allow handyman, got status %d", rr.Code)
		}
	})

	t.Run("blocks user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := auth.NewContext(req.Context(), &auth.User{ID: "123", Role: role.User})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		RequireHandyman(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("should block user, got status %d", rr.Code)
		}
	})
}


