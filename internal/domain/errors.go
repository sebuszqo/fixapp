// Package domain contains the core business entities and errors.
package domain

import "errors"

// Domain errors are used across the application.
// Services return these errors, handlers map them to HTTP responses.
var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Authorization errors
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInsufficientRole  = errors.New("insufficient role")
	ErrMissingPermission = errors.New("missing permission")

	// Validation errors
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidProvider = errors.New("invalid auth provider")

	// Token errors
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

