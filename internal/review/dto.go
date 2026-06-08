package review

import (
	"time"

	"fixapp/internal/domain"
)

// ===== Request DTOs =====

// CreateReviewRequest is the request body for creating a review.
// @Description Create review request
type CreateReviewRequest struct {
	JobID   string `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Rating  int    `json:"rating" example:"5"`
	Comment string `json:"comment" example:"Great work, very professional!"`
}

// ===== Response DTOs =====

// ReviewResponse is the API representation of a review.
// @Description Review information
type ReviewResponse struct {
	ID         string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	JobID      string    `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ReviewerID string    `json:"reviewer_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	RevieweeID string    `json:"reviewee_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	Type       string    `json:"type" example:"client_to_handyman"`
	Rating     int       `json:"rating" example:"5"`
	Comment    string    `json:"comment" example:"Great work, very professional!"`
	CreatedAt  time.Time `json:"created_at" example:"2025-01-15T14:30:00Z"`
}

// ReviewListResponse is the paginated list of reviews.
// @Description Paginated review list
type ReviewListResponse struct {
	Reviews []ReviewResponse `json:"reviews"`
	Total   int64            `json:"total" example:"12"`
	Limit   int              `json:"limit" example:"20"`
	Offset  int              `json:"offset" example:"0"`
	HasMore bool             `json:"has_more" example:"false"`
}

// RatingSummaryResponse is the rating summary for a user.
// @Description Rating summary
type RatingSummaryResponse struct {
	AverageRating float64 `json:"average_rating" example:"4.7"`
	TotalReviews  int     `json:"total_reviews" example:"12"`
}

// ===== Mappers =====

// ToReviewResponse converts a domain review to API response.
func ToReviewResponse(review *domain.Review) ReviewResponse {
	return ReviewResponse{
		ID:         review.ID.String(),
		JobID:      review.JobID.String(),
		ReviewerID: review.ReviewerID.String(),
		RevieweeID: review.RevieweeID.String(),
		Type:       review.Type.String(),
		Rating:     review.Rating,
		Comment:    review.Comment,
		CreatedAt:  review.CreatedAt,
	}
}

// ToReviewListResponse converts a list of reviews to paginated response.
func ToReviewListResponse(reviews []*domain.Review, total int64, limit, offset int) ReviewListResponse {
	responses := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		responses[i] = ToReviewResponse(r)
	}
	return ReviewListResponse{
		Reviews: responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: int64(offset+len(reviews)) < total,
	}
}
