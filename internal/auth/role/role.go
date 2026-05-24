// Package role defines user roles for the application.
// Roles are hierarchical with Admin > Handyman > User.
package role

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// Role represents a user's role in the system.
// Using a custom type provides type safety and prevents invalid role assignments.
type Role uint8

// Role constants using iota for efficient storage and comparison.
// The order matters for hierarchy checks (higher value = more privileges).
const (
	Unknown  Role = iota // Zero value, indicates unset or invalid role
	User                 // Regular user with basic access
	Handyman             // Service provider with additional capabilities
	Admin                // Full administrative access
)

// String returns the string representation of the role.
// Implements fmt.Stringer interface.
func (r Role) String() string {
	switch r {
	case User:
		return "user"
	case Handyman:
		return "handyman"
	case Admin:
		return "admin"
	default:
		return "unknown"
	}
}

// Parse converts a string to a Role.
// Returns Unknown for unrecognized strings.
func Parse(s string) Role {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "user":
		return User
	case "handyman":
		return Handyman
	case "admin":
		return Admin
	default:
		return Unknown
	}
}

// MustParse converts a string to a Role, panics on invalid input.
// Use only when the input is guaranteed to be valid (e.g., hardcoded values).
func MustParse(s string) Role {
	r := Parse(s)
	if r == Unknown {
		panic(fmt.Sprintf("invalid role: %q", s))
	}
	return r
}

// IsValid returns true if the role is a known, valid role.
func (r Role) IsValid() bool {
	return r >= User && r <= Admin
}

// IsAtLeast returns true if this role has at least the privileges of the given role.
// Useful for hierarchical permission checks.
func (r Role) IsAtLeast(other Role) bool {
	return r >= other
}

// CanActAs returns true if this role can perform actions as the given role.
// Admin can act as any role, others can only act as themselves or lower.
func (r Role) CanActAs(target Role) bool {
	return r.IsAtLeast(target)
}

// All returns all valid roles.
func All() []Role {
	return []Role{User, Handyman, Admin}
}

// MarshalText implements encoding.TextMarshaler for JSON/YAML serialization.
func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON/YAML deserialization.
func (r *Role) UnmarshalText(text []byte) error {
	*r = Parse(string(text))
	if *r == Unknown && len(text) > 0 && string(text) != "unknown" {
		return fmt.Errorf("invalid role: %q", string(text))
	}
	return nil
}

// Value implements driver.Valuer for database storage.
func (r Role) Value() (driver.Value, error) {
	return r.String(), nil
}

// Scan implements sql.Scanner for database retrieval.
func (r *Role) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		*r = Parse(v)
	case []byte:
		*r = Parse(string(v))
	case nil:
		*r = Unknown
	default:
		return fmt.Errorf("cannot scan %T into Role", src)
	}
	return nil
}

