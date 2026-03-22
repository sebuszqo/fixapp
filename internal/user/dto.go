package user

import (
	"time"

	"fixapp/internal/auth/role"
	"fixapp/internal/domain"
)

// DTOs (Data Transfer Objects) for API layer.
// These structs define the API contract and are separate from domain models.

// ===== Response DTOs =====

// UserResponse is the public representation of a user.
// @Description User information
type UserResponse struct {
	ID            string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email         string     `json:"email" example:"user@example.com"`
	Name          string     `json:"name" example:"John Doe"`
	Role          string     `json:"role" example:"user"`
	Provider      string     `json:"provider" example:"google"`
	AvatarURL     string     `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	Phone         string     `json:"phone,omitempty" example:"+1234567890"`
	IsActive      bool       `json:"is_active" example:"true"`
	EmailVerified bool       `json:"email_verified" example:"true"`
	CreatedAt     time.Time  `json:"created_at" example:"2025-01-10T12:00:00Z"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" example:"2025-01-10T14:30:00Z"`
}

// ProfileResponse is a simplified user profile for the current user.
// @Description Current user's profile
type ProfileResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"user@example.com"`
	Name      string `json:"name" example:"John Doe"`
	Role      string `json:"role" example:"user"`
	AvatarURL string `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	Phone     string `json:"phone,omitempty" example:"+1234567890"`
}

// UserListResponse is the paginated list of users.
// @Description Paginated user list
type UserListResponse struct {
	Users   []UserResponse `json:"users"`
	Total   int64          `json:"total" example:"100"`
	Limit   int            `json:"limit" example:"20"`
	Offset  int            `json:"offset" example:"0"`
	HasMore bool           `json:"has_more" example:"true"`
}

// ===== Request DTOs =====

// UpdateProfileRequest is the payload for updating user profile.
// @Description Profile update request
type UpdateProfileRequest struct {
	Name      string `json:"name,omitempty" example:"John Doe"`
	Phone     string `json:"phone,omitempty" example:"+1234567890"`
	AvatarURL string `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
}

// Validate checks if the request is valid.
func (r *UpdateProfileRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if r.Name != "" && len(r.Name) > 255 {
		errors["name"] = "name must be 255 characters or less"
	}

	if r.Phone != "" && len(r.Phone) > 50 {
		errors["phone"] = "phone must be 50 characters or less"
	}

	if r.AvatarURL != "" && len(r.AvatarURL) > 500 {
		errors["avatar_url"] = "avatar_url must be 500 characters or less"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

// UpdateRoleRequest is the payload for updating user role.
// @Description Role update request
type UpdateRoleRequest struct {
	Role string `json:"role" example:"handyman"`
}

// Validate checks if the request is valid.
func (r *UpdateRoleRequest) Validate() map[string]string {
	errors := make(map[string]string)

	parsedRole := role.Parse(r.Role)
	if !parsedRole.IsValid() {
		errors["role"] = "invalid role, must be one of: user, handyman, admin"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

// GetRole returns the parsed role.
func (r *UpdateRoleRequest) GetRole() role.Role {
	return role.Parse(r.Role)
}

// ListUsersRequest contains query parameters for listing users.
type ListUsersRequest struct {
	Role     string `json:"role,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
	Search   string `json:"search,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// ToFilter converts the request to a repository filter.
func (r *ListUsersRequest) ToFilter() ListFilter {
	filter := DefaultListFilter()

	if r.Role != "" {
		parsed := role.Parse(r.Role)
		if parsed.IsValid() {
			filter.Role = &parsed
		}
	}

	filter.IsActive = r.IsActive
	filter.Search = r.Search

	if r.Limit > 0 && r.Limit <= 100 {
		filter.Limit = r.Limit
	}

	if r.Offset >= 0 {
		filter.Offset = r.Offset
	}

	return filter
}

// ===== Mappers =====

// ToUserResponse converts a domain user to API response.
func ToUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		Role:          user.Role.String(),
		Provider:      user.Provider.String(),
		AvatarURL:     user.AvatarURL,
		Phone:         user.Phone,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		LastLoginAt:   user.LastLoginAt,
	}
}

// ToProfileResponse converts a domain user to profile response.
func ToProfileResponse(user *domain.User) ProfileResponse {
	return ProfileResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role.String(),
		AvatarURL: user.AvatarURL,
		Phone:     user.Phone,
	}
}

// ToUserListResponse converts a list of domain users to paginated response.
func ToUserListResponse(users []*domain.User, total int64, limit, offset int) UserListResponse {
	responses := make([]UserResponse, len(users))
	for i, u := range users {
		responses[i] = ToUserResponse(u)
	}

	return UserListResponse{
		Users:   responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: int64(offset+len(users)) < total,
	}
}

