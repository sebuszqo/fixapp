package domain

import (
	"time"

	"github.com/google/uuid"
)

// HandymanProfile is the extended profile for handyman users.
type HandymanProfile struct {
	ID     uuid.UUID
	UserID uuid.UUID

	// Company info
	CompanyName    string
	NIP            string // Polish tax ID
	CompanyAddress string
	PKDCode        string
	GUSStatus      string // e.g. "Aktywna", "Zawieszona"
	GUSVerified    bool
	Phone          string
	Email          string
	BusinessType   string // "jdg" or "company"
	IsVATPayer     bool   // true = VAT payer, false = exempt

	// Consents
	ConsentIdentityVerification bool
	ConsentMarketing            bool

	// Public profile
	Bio             string
	AvatarURL       string
	ExperienceYears string // e.g. "1-2 lata", "3-5 lata", "6-10 lat", "10+ lat"
	WorkingHours    string // JSON encoded weekly schedule

	// Service configuration
	Categories []uuid.UUID // max 3 category IDs
	Districts  []uuid.UUID // served district IDs

	// Availability
	IsAvailable        bool // false = vacation/pause mode
	EmergencyAvailable bool // accepts emergency jobs

	// Verification
	IsVerified                bool
	VerificationStatus        string     // "unverified", "pending_review", "verified", "rejected"
	CEIDGDocumentURL          string     // uploaded PDF URL
	CEIDGUploadedAt           *time.Time
	MicrotransferStatus       string     // "none", "pending", "confirmed"
	MicrotransferCode         string     // e.g. "FIXAPP-7823-JAN"
	MicrotransferConfirmedAt *time.Time

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewHandymanProfile creates a new profile for a handyman.
func NewHandymanProfile(userID uuid.UUID) *HandymanProfile {
	now := time.Now()
	return &HandymanProfile{
		ID:                 uuid.New(),
		UserID:             userID,
		IsAvailable:        true,
		GUSStatus:          "Aktywna",
		BusinessType:       "jdg",
		VerificationStatus: "unverified",
		MicrotransferStatus: "none",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// CompletionPercentage calculates profile completeness (0-100).
func (p *HandymanProfile) CompletionPercentage() int {
	score := 0
	total := 10

	if p.CompanyName != "" {
		score++
	}
	if p.Phone != "" {
		score++
	}
	if p.Bio != "" {
		score++
	}
	if p.AvatarURL != "" {
		score++
	}
	if len(p.Categories) > 0 {
		score++
	}
	if len(p.Districts) > 0 {
		score++
	}
	if p.NIP != "" {
		score++
	}
	if p.GUSVerified || p.IsVerified {
		score++
	}
	if p.WorkingHours != "" {
		score++
	}
	if p.ExperienceYears != "" {
		score++
	}

	return (score * 100) / total
}

// PricingItem represents a single service with price.
type PricingItem struct {
	ID                uuid.UUID
	ProfileID         uuid.UUID
	ServiceName       string
	PriceFrom         int    // minimum price in PLN
	PriceTo           *int   // maximum price in PLN (optional, for ranges)
	Unit              string // e.g., "per hour", "per service", "per m2"
	EstimatedDuration string // e.g. "~1h", "~2h"
	SortOrder         int
}

// PortfolioItem represents a photo in the handyman's portfolio.
type PortfolioItem struct {
	ID        uuid.UUID
	ProfileID uuid.UUID
	ImageURL  string
	Caption   string
	SortOrder int
	CreatedAt time.Time
}

