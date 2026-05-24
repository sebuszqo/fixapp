// Package permission defines granular permissions for the application.
// Permissions are mapped to roles, allowing fine-grained access control.
package permission

import (
	"fixapp/internal/auth/role"
)

// Permission represents a specific action that can be performed.
// Using string type for readability in logs and debugging.
type Permission string

// Permission constants grouped by domain.
// Naming convention: <domain>:<action>
const (
	// User permissions
	UserRead   Permission = "user:read"
	UserUpdate Permission = "user:update"
	UserDelete Permission = "user:delete"
	UserList   Permission = "user:list"

	// Profile permissions (self-service)
	ProfileRead   Permission = "profile:read"
	ProfileUpdate Permission = "profile:update"

	// Job/Task permissions (for handyman domain)
	JobCreate Permission = "job:create"
	JobRead   Permission = "job:read"
	JobUpdate Permission = "job:update"
	JobDelete Permission = "job:delete"
	JobList   Permission = "job:list"
	JobAccept Permission = "job:accept"

	// Admin permissions
	AdminAccess    Permission = "admin:access"
	AdminUserMgmt  Permission = "admin:user_management"
	AdminRoleMgmt  Permission = "admin:role_management"
	AdminSystemCfg Permission = "admin:system_config"
)

// rolePermissions maps each role to its allowed permissions.
// Higher roles inherit lower role permissions through the Has function.
var rolePermissions = map[role.Role][]Permission{
	role.User: {
		ProfileRead,
		ProfileUpdate,
		JobCreate, // Users can create job requests
		JobRead,
		JobList,
	},
	role.Handyman: {
		ProfileRead,
		ProfileUpdate,
		JobRead,
		JobList,
		JobAccept, // Handymen can accept jobs
		JobUpdate, // Handymen can update job status
	},
	role.Admin: {
		// Admins have all permissions
		UserRead,
		UserUpdate,
		UserDelete,
		UserList,
		ProfileRead,
		ProfileUpdate,
		JobCreate,
		JobRead,
		JobUpdate,
		JobDelete,
		JobList,
		JobAccept,
		AdminAccess,
		AdminUserMgmt,
		AdminRoleMgmt,
		AdminSystemCfg,
	},
}

// Has checks if the given role has the specified permission.
func Has(r role.Role, p Permission) bool {
	perms, ok := rolePermissions[r]
	if !ok {
		return false
	}

	for _, perm := range perms {
		if perm == p {
			return true
		}
	}
	return false
}

// HasAny checks if the role has at least one of the specified permissions.
func HasAny(r role.Role, permissions ...Permission) bool {
	for _, p := range permissions {
		if Has(r, p) {
			return true
		}
	}
	return false
}

// HasAll checks if the role has all of the specified permissions.
func HasAll(r role.Role, permissions ...Permission) bool {
	for _, p := range permissions {
		if !Has(r, p) {
			return false
		}
	}
	return true
}

// ForRole returns all permissions for a given role.
func ForRole(r role.Role) []Permission {
	perms := rolePermissions[r]
	// Return a copy to prevent modification
	result := make([]Permission, len(perms))
	copy(result, perms)
	return result
}

// String returns the string representation of the permission.
func (p Permission) String() string {
	return string(p)
}

// IsValid checks if the permission is a known permission.
func (p Permission) IsValid() bool {
	allPerms := []Permission{
		UserRead, UserUpdate, UserDelete, UserList,
		ProfileRead, ProfileUpdate,
		JobCreate, JobRead, JobUpdate, JobDelete, JobList, JobAccept,
		AdminAccess, AdminUserMgmt, AdminRoleMgmt, AdminSystemCfg,
	}

	for _, valid := range allPerms {
		if p == valid {
			return true
		}
	}
	return false
}

