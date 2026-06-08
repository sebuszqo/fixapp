package domain

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the lifecycle state of a job.
type JobStatus string

const (
	JobStatusDraft      JobStatus = "draft"
	JobStatusActive     JobStatus = "active"
	JobStatusAccepted   JobStatus = "accepted"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusDone       JobStatus = "done"
	JobStatusCancelled  JobStatus = "cancelled"
)

func (s JobStatus) String() string {
	return string(s)
}

func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusDraft, JobStatusActive, JobStatusAccepted,
		JobStatusInProgress, JobStatusDone, JobStatusCancelled:
		return true
	default:
		return false
	}
}

// JobUrgency represents how urgent the job is.
type JobUrgency string

const (
	JobUrgencyLow      JobUrgency = "low"
	JobUrgencyNormal   JobUrgency = "normal"
	JobUrgencyUrgent   JobUrgency = "urgent"
	JobUrgencyEmergency JobUrgency = "emergency"
)

func (u JobUrgency) String() string {
	return string(u)
}

func (u JobUrgency) IsValid() bool {
	switch u {
	case JobUrgencyLow, JobUrgencyNormal, JobUrgencyUrgent, JobUrgencyEmergency:
		return true
	default:
		return false
	}
}

// Job represents a client's service request.
type Job struct {
	ID          uuid.UUID
	ClientID    uuid.UUID
	CategoryID  uuid.UUID
	DistrictID  uuid.UUID
	Title       string
	Description string
	Urgency     JobUrgency
	Status      JobStatus

	// Location details
	Address      string
	BuildingType string // apartment, house, office
	Floor        *int
	HasElevator  *bool

	// Scheduling
	PreferredDate1 *time.Time
	PreferredDate2 *time.Time
	PreferredTime  string // morning, afternoon, evening, flexible

	// Budget
	Budget       *int // client's budget in PLN (optional)
	WantsInvoice bool

	// Contact preferences
	ContactMethod string // phone, app, any

	// Photos (stored as URLs)
	PhotoURLs []string

	// Completion
	FinalValue      *int       // declared value after completion (PLN)
	CompletedAt     *time.Time
	CompletedByID   *uuid.UUID // handyman who completed the job
	ClientConfirmed *bool

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt *time.Time // when the job posting expires
}

// NewJob creates a new job in draft status.
func NewJob(clientID, categoryID, districtID uuid.UUID, title, description string) *Job {
	now := time.Now()
	return &Job{
		ID:          uuid.New(),
		ClientID:    clientID,
		CategoryID:  categoryID,
		DistrictID:  districtID,
		Title:       title,
		Description: description,
		Urgency:     JobUrgencyNormal,
		Status:      JobStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Publish moves the job from draft to active (visible to handymen).
func (j *Job) Publish() error {
	if j.Status != JobStatusDraft {
		return ErrInvalidJobTransition
	}
	j.Status = JobStatusActive
	j.UpdatedAt = time.Now()
	// Set default expiry: 7 days from publish
	expires := time.Now().Add(7 * 24 * time.Hour)
	j.ExpiresAt = &expires
	return nil
}

// Accept marks the job as accepted by a handyman (via lead acceptance).
func (j *Job) Accept(handymanID uuid.UUID) error {
	if j.Status != JobStatusActive {
		return ErrInvalidJobTransition
	}
	j.Status = JobStatusAccepted
	j.CompletedByID = &handymanID
	j.UpdatedAt = time.Now()
	return nil
}

// StartWork moves the job to in_progress.
func (j *Job) StartWork() error {
	if j.Status != JobStatusAccepted {
		return ErrInvalidJobTransition
	}
	j.Status = JobStatusInProgress
	j.UpdatedAt = time.Now()
	return nil
}

// Complete marks the job as done with a declared value.
func (j *Job) Complete(finalValue int) error {
	if j.Status != JobStatusAccepted && j.Status != JobStatusInProgress {
		return ErrInvalidJobTransition
	}
	j.Status = JobStatusDone
	j.FinalValue = &finalValue
	now := time.Now()
	j.CompletedAt = &now
	j.UpdatedAt = now
	return nil
}

// Cancel cancels the job.
func (j *Job) Cancel() error {
	if j.Status == JobStatusDone || j.Status == JobStatusCancelled {
		return ErrInvalidJobTransition
	}
	j.Status = JobStatusCancelled
	j.UpdatedAt = time.Now()
	return nil
}

// ConfirmByClient marks the job as confirmed by the client.
func (j *Job) ConfirmByClient() error {
	if j.Status != JobStatusDone {
		return ErrInvalidJobTransition
	}
	confirmed := true
	j.ClientConfirmed = &confirmed
	j.UpdatedAt = time.Now()
	return nil
}

// IsExpired checks if the job posting has expired.
func (j *Job) IsExpired() bool {
	if j.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*j.ExpiresAt)
}
