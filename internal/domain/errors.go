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

	// Job errors
	ErrJobNotFound          = errors.New("job not found")
	ErrInvalidJobTransition = errors.New("invalid job status transition")
	ErrJobExpired           = errors.New("job has expired")
	ErrJobAlreadyAccepted   = errors.New("job already accepted by another handyman")

	// Lead errors
	ErrLeadNotFound          = errors.New("lead not found")
	ErrInvalidLeadTransition = errors.New("invalid lead status transition")
	ErrLeadExpired           = errors.New("lead has expired")
	ErrLeadAlreadyAccepted   = errors.New("lead already accepted")
	ErrInsufficientCredits   = errors.New("insufficient credits to accept lead")

	// Category/District errors
	ErrCategoryNotFound = errors.New("service category not found")
	ErrDistrictNotFound = errors.New("district not found")

	// Wallet errors
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletAlreadyExists = errors.New("wallet already exists for this user")

	// Handyman profile errors
	ErrProfileNotFound      = errors.New("handyman profile not found")
	ErrProfileAlreadyExists = errors.New("handyman profile already exists")
	ErrTooManyCategories    = errors.New("maximum 3 categories allowed")
	ErrPricingItemNotFound  = errors.New("pricing item not found")
	ErrPortfolioItemNotFound = errors.New("portfolio item not found")

	// Review errors
	ErrReviewNotFound      = errors.New("review not found")
	ErrReviewAlreadyExists = errors.New("review already submitted for this job")
	ErrInvalidRating       = errors.New("rating must be between 1 and 5")
	ErrInvalidReviewType   = errors.New("invalid review type")
	ErrCannotReviewSelf    = errors.New("cannot review yourself")
	ErrJobNotCompleted     = errors.New("can only review after job is completed")
	ErrNotJobParticipant   = errors.New("only job participants can leave reviews")
)

