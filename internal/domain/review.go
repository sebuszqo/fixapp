package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReviewType defines the direction of a review.
type ReviewType string

const (
	ReviewTypeClientToHandyman ReviewType = "client_to_handyman"
	ReviewTypeHandymanToClient ReviewType = "handyman_to_client"
)

func (rt ReviewType) String() string {
	return string(rt)
}

func (rt ReviewType) IsValid() bool {
	switch rt {
	case ReviewTypeClientToHandyman, ReviewTypeHandymanToClient:
		return true
	default:
		return false
	}
}

// Review represents a rating/review left after a job is completed.
type Review struct {
	ID         uuid.UUID
	JobID      uuid.UUID
	ReviewerID uuid.UUID // who writes the review
	RevieweeID uuid.UUID // who is being reviewed
	Type       ReviewType
	Rating     int    // 1-5
	Comment    string // optional text

	CreatedAt time.Time
}

// NewReview creates a new review.
func NewReview(jobID, reviewerID, revieweeID uuid.UUID, reviewType ReviewType, rating int, comment string) *Review {
	return &Review{
		ID:         uuid.New(),
		JobID:      jobID,
		ReviewerID: reviewerID,
		RevieweeID: revieweeID,
		Type:       reviewType,
		Rating:     rating,
		Comment:    comment,
		CreatedAt:  time.Now(),
	}
}

// Validate checks that the review has valid data.
func (r *Review) Validate() error {
	if r.Rating < 1 || r.Rating > 5 {
		return ErrInvalidRating
	}
	if !r.Type.IsValid() {
		return ErrInvalidReviewType
	}
	if r.ReviewerID == r.RevieweeID {
		return ErrCannotReviewSelf
	}
	return nil
}
