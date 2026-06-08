package domain

import (
	"time"

	"github.com/google/uuid"
)

// LeadStatus represents the lifecycle state of a lead.
type LeadStatus string

const (
	LeadStatusPending  LeadStatus = "pending"
	LeadStatusAccepted LeadStatus = "accepted"
	LeadStatusRejected LeadStatus = "rejected"
	LeadStatusExpired  LeadStatus = "expired"
)

func (s LeadStatus) String() string {
	return string(s)
}

func (s LeadStatus) IsValid() bool {
	switch s {
	case LeadStatusPending, LeadStatusAccepted, LeadStatusRejected, LeadStatusExpired:
		return true
	default:
		return false
	}
}

// Lead represents a job opportunity sent to a specific handyman.
// A single Job can generate multiple Leads (one per matching handyman).
type Lead struct {
	ID         uuid.UUID
	JobID      uuid.UUID
	HandymanID uuid.UUID

	Status LeadStatus

	// Pricing
	Price int // cost in credits to accept this lead (calculated by dynamic pricing)

	// Client quality indicator (snapshot at lead creation time)
	ClientCommitScore int

	// Timestamps
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiresAt  time.Time // when the lead expires if not acted upon
	AcceptedAt *time.Time
	RejectedAt *time.Time
}

// NewLead creates a new pending lead for a handyman.
func NewLead(jobID, handymanID uuid.UUID, price, clientCommitScore int, expiresIn time.Duration) *Lead {
	now := time.Now()
	return &Lead{
		ID:                uuid.New(),
		JobID:             jobID,
		HandymanID:        handymanID,
		Status:            LeadStatusPending,
		Price:             price,
		ClientCommitScore: clientCommitScore,
		CreatedAt:         now,
		UpdatedAt:         now,
		ExpiresAt:         now.Add(expiresIn),
	}
}

// Accept marks the lead as accepted (handyman pays credits).
func (l *Lead) Accept() error {
	if l.Status != LeadStatusPending {
		return ErrInvalidLeadTransition
	}
	if l.IsExpired() {
		return ErrLeadExpired
	}
	l.Status = LeadStatusAccepted
	now := time.Now()
	l.AcceptedAt = &now
	l.UpdatedAt = now
	return nil
}

// Reject marks the lead as rejected by the handyman.
func (l *Lead) Reject() error {
	if l.Status != LeadStatusPending {
		return ErrInvalidLeadTransition
	}
	l.Status = LeadStatusRejected
	now := time.Now()
	l.RejectedAt = &now
	l.UpdatedAt = now
	return nil
}

// Expire marks the lead as expired.
func (l *Lead) Expire() error {
	if l.Status != LeadStatusPending {
		return ErrInvalidLeadTransition
	}
	l.Status = LeadStatusExpired
	l.UpdatedAt = time.Now()
	return nil
}

// IsExpired checks if the lead has passed its expiration time.
func (l *Lead) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}
