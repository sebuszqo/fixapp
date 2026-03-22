package auth

import (
	"context"
	"testing"
	"time"

	"fixapp/internal/auth/permission"
	"fixapp/internal/auth/role"
)

func TestNewContext(t *testing.T) {
	ctx := context.Background()
	user := &User{
		ID:    "123",
		Email: "test@example.com",
		Name:  "Test User",
		Role:  role.User,
	}

	ctx = NewContext(ctx, user)

	retrieved := FromContext(ctx)
	if retrieved == nil {
		t.Fatal("FromContext returned nil")
	}

	if retrieved.ID != user.ID {
		t.Errorf("FromContext().ID = %v, want %v", retrieved.ID, user.ID)
	}
}

func TestFromContextNil(t *testing.T) {
	ctx := context.Background()

	user := FromContext(ctx)
	if user != nil {
		t.Error("FromContext should return nil for context without user")
	}
}

func TestMustFromContext(t *testing.T) {
	ctx := context.Background()
	user := &User{ID: "123", Role: role.User}
	ctx = NewContext(ctx, user)

	// Should not panic
	retrieved := MustFromContext(ctx)
	if retrieved.ID != "123" {
		t.Errorf("MustFromContext().ID = %v, want 123", retrieved.ID)
	}

	// Should panic with no user
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustFromContext should panic when no user in context")
		}
	}()
	MustFromContext(context.Background())
}

func TestIsAuthenticated(t *testing.T) {
	ctx := context.Background()
	if IsAuthenticated(ctx) {
		t.Error("IsAuthenticated should return false for empty context")
	}

	ctx = NewContext(ctx, &User{ID: "123"})
	if !IsAuthenticated(ctx) {
		t.Error("IsAuthenticated should return true when user is present")
	}
}

func TestHasRole(t *testing.T) {
	ctx := context.Background()
	user := &User{ID: "123", Role: role.Handyman}
	ctx = NewContext(ctx, user)

	if !HasRole(ctx, role.Handyman) {
		t.Error("HasRole should return true for matching role")
	}

	if HasRole(ctx, role.Admin) {
		t.Error("HasRole should return false for non-matching role")
	}

	// Empty context
	if HasRole(context.Background(), role.User) {
		t.Error("HasRole should return false for empty context")
	}
}

func TestHasRoleAtLeast(t *testing.T) {
	ctx := context.Background()
	user := &User{ID: "123", Role: role.Handyman}
	ctx = NewContext(ctx, user)

	if !HasRoleAtLeast(ctx, role.User) {
		t.Error("Handyman should have at least User role")
	}

	if !HasRoleAtLeast(ctx, role.Handyman) {
		t.Error("Handyman should have at least Handyman role")
	}

	if HasRoleAtLeast(ctx, role.Admin) {
		t.Error("Handyman should not have Admin role")
	}
}

func TestHasPermission(t *testing.T) {
	ctx := context.Background()
	user := &User{ID: "123", Role: role.User}
	ctx = NewContext(ctx, user)

	if !HasPermission(ctx, permission.ProfileRead) {
		t.Error("User should have ProfileRead permission")
	}

	if HasPermission(ctx, permission.AdminAccess) {
		t.Error("User should not have AdminAccess permission")
	}

	// Empty context
	if HasPermission(context.Background(), permission.ProfileRead) {
		t.Error("HasPermission should return false for empty context")
	}
}

func TestHasAnyPermission(t *testing.T) {
	ctx := context.Background()
	user := &User{ID: "123", Role: role.User}
	ctx = NewContext(ctx, user)

	if !HasAnyPermission(ctx, permission.ProfileRead, permission.AdminAccess) {
		t.Error("User should have ProfileRead")
	}

	if HasAnyPermission(ctx, permission.AdminAccess, permission.AdminUserMgmt) {
		t.Error("User should not have any admin permissions")
	}
}

func TestHasAllPermissions(t *testing.T) {
	ctx := context.Background()
	user := &User{ID: "123", Role: role.Admin}
	ctx = NewContext(ctx, user)

	if !HasAllPermissions(ctx, permission.AdminAccess, permission.AdminUserMgmt) {
		t.Error("Admin should have all admin permissions")
	}
}

func TestUserMethods(t *testing.T) {
	user := &User{
		ID:        "123",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      role.Admin,
		Provider:  "google",
		CreatedAt: time.Now(),
	}

	if !user.IsAdmin() {
		t.Error("User with Admin role should return true for IsAdmin()")
	}

	if user.IsHandyman() {
		t.Error("Admin should not return true for IsHandyman()")
	}

	if user.IsUser() {
		t.Error("Admin should not return true for IsUser()")
	}

	if !user.HasPermission(permission.AdminAccess) {
		t.Error("Admin should have AdminAccess permission")
	}

	if !user.HasAnyPermission(permission.AdminAccess, permission.ProfileRead) {
		t.Error("Admin should have at least one of these permissions")
	}

	if !user.HasAllPermissions(permission.AdminAccess, permission.AdminUserMgmt) {
		t.Error("Admin should have all admin permissions")
	}
}

func TestUserIsHandyman(t *testing.T) {
	user := &User{ID: "123", Role: role.Handyman}
	if !user.IsHandyman() {
		t.Error("Handyman user should return true for IsHandyman()")
	}
}

func TestUserIsUser(t *testing.T) {
	user := &User{ID: "123", Role: role.User}
	if !user.IsUser() {
		t.Error("User with User role should return true for IsUser()")
	}
}


