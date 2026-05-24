package permission

import (
	"testing"

	"fixapp/internal/auth/role"
)

func TestHas(t *testing.T) {
	tests := []struct {
		name       string
		role       role.Role
		permission Permission
		want       bool
	}{
		{"user can read profile", role.User, ProfileRead, true},
		{"user can update profile", role.User, ProfileUpdate, true},
		{"user can create job", role.User, JobCreate, true},
		{"user cannot access admin", role.User, AdminAccess, false},
		{"user cannot delete job", role.User, JobDelete, false},

		{"handyman can read profile", role.Handyman, ProfileRead, true},
		{"handyman can accept job", role.Handyman, JobAccept, true},
		{"handyman cannot access admin", role.Handyman, AdminAccess, false},
		{"handyman cannot create job", role.Handyman, JobCreate, false},

		{"admin can access admin", role.Admin, AdminAccess, true},
		{"admin can delete job", role.Admin, JobDelete, true},
		{"admin can manage users", role.Admin, AdminUserMgmt, true},

		{"unknown role has no permissions", role.Unknown, ProfileRead, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Has(tt.role, tt.permission); got != tt.want {
				t.Errorf("Has(%v, %v) = %v, want %v", tt.role, tt.permission, got, tt.want)
			}
		})
	}
}

func TestHasAny(t *testing.T) {
	// User should have at least one of ProfileRead or AdminAccess
	if !HasAny(role.User, ProfileRead, AdminAccess) {
		t.Error("User should have ProfileRead")
	}

	// User should not have any admin permissions
	if HasAny(role.User, AdminAccess, AdminUserMgmt) {
		t.Error("User should not have any admin permissions")
	}
}

func TestHasAll(t *testing.T) {
	// User should have both ProfileRead and ProfileUpdate
	if !HasAll(role.User, ProfileRead, ProfileUpdate) {
		t.Error("User should have both ProfileRead and ProfileUpdate")
	}

	// User should not have ProfileRead and AdminAccess
	if HasAll(role.User, ProfileRead, AdminAccess) {
		t.Error("User should not have both ProfileRead and AdminAccess")
	}

	// Admin should have all admin permissions
	if !HasAll(role.Admin, AdminAccess, AdminUserMgmt, AdminRoleMgmt, AdminSystemCfg) {
		t.Error("Admin should have all admin permissions")
	}
}

func TestForRole(t *testing.T) {
	userPerms := ForRole(role.User)
	if len(userPerms) == 0 {
		t.Error("User should have some permissions")
	}

	// Verify it returns a copy (modify and check original)
	userPerms[0] = "modified:permission"
	originalPerms := ForRole(role.User)
	if originalPerms[0] == "modified:permission" {
		t.Error("ForRole should return a copy, not the original slice")
	}
}

func TestPermissionString(t *testing.T) {
	if ProfileRead.String() != "profile:read" {
		t.Errorf("ProfileRead.String() = %v, want profile:read", ProfileRead.String())
	}
}

func TestPermissionIsValid(t *testing.T) {
	tests := []struct {
		perm Permission
		want bool
	}{
		{ProfileRead, true},
		{AdminAccess, true},
		{JobCreate, true},
		{Permission("invalid:permission"), false},
		{Permission(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.perm), func(t *testing.T) {
			if got := tt.perm.IsValid(); got != tt.want {
				t.Errorf("Permission(%q).IsValid() = %v, want %v", tt.perm, got, tt.want)
			}
		})
	}
}


