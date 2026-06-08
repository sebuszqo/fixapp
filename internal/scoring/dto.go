package scoring

import (
	"time"

	"fixapp/internal/domain"
)

// ===== Response DTOs =====

// CommitScoreResponse is the public representation of a Commit Score.
// @Description Client Commit Score
type CommitScoreResponse struct {
	Score           int       `json:"score" example:"75"`
	Level           string    `json:"level" example:"standard"`
	ClientMultiplier float64  `json:"client_multiplier" example:"1.0"`

	// Breakdown
	PhoneVerified   bool `json:"phone_verified" example:"true"`
	ProfileComplete bool `json:"profile_complete" example:"true"`
	HasAvatar       bool `json:"has_avatar" example:"false"`
	HasJobHistory   bool `json:"has_job_history" example:"true"`
	NoNoShows       bool `json:"no_no_shows" example:"true"`
	NoExcessCancels bool `json:"no_excess_cancels" example:"true"`

	// Stats
	JobsCompleted int `json:"jobs_completed" example:"3"`
	JobsCancelled int `json:"jobs_cancelled" example:"0"`
	NoShowCount   int `json:"no_show_count" example:"0"`

	UpdatedAt time.Time `json:"updated_at"`
}

// ProScoreResponse is the public representation of a ProScore.
// @Description Handyman ProScore
type ProScoreResponse struct {
	Score             int     `json:"score" example:"847"`
	Level             string  `json:"level" example:"pro_partner"`
	Badge             string  `json:"badge,omitempty" example:"Pro Partner"`
	HandymanMultiplier float64 `json:"handyman_multiplier" example:"0.8"`

	// Positive factors
	JobsCompleted   int  `json:"jobs_completed" example:"12"`
	FiveStarReviews int  `json:"five_star_reviews" example:"8"`
	AvgResponseMins int  `json:"avg_response_mins" example:"45"`
	ProfileComplete bool `json:"profile_complete" example:"true"`
	ActiveLast7Days bool `json:"active_last_7_days" example:"true"`
	PortfolioCount  int  `json:"portfolio_count" example:"6"`

	// Penalties
	NoShowCount          int `json:"no_show_count" example:"0"`
	CancelledAfterAccept int `json:"cancelled_after_accept" example:"0"`
	SlowResponseCount    int `json:"slow_response_count" example:"1"`
	LowRatingCount       int `json:"low_rating_count" example:"0"`

	UpdatedAt time.Time `json:"updated_at"`
}

// LeadPriceResponse shows the calculated lead price.
// @Description Dynamic lead price calculation
type LeadPriceResponse struct {
	BasePrice          int     `json:"base_price" example:"22"`
	ClientMultiplier   float64 `json:"client_multiplier" example:"0.8"`
	HandymanMultiplier float64 `json:"handyman_multiplier" example:"0.8"`
	FinalPrice         int     `json:"final_price" example:"14"`
}

// ===== Mappers =====

// ToCommitScoreResponse converts a domain CommitScore to API response.
func ToCommitScoreResponse(cs *domain.CommitScore) CommitScoreResponse {
	return CommitScoreResponse{
		Score:            cs.Score,
		Level:            string(cs.Level()),
		ClientMultiplier: cs.ClientMultiplier(),
		PhoneVerified:    cs.PhoneVerified,
		ProfileComplete:  cs.ProfileComplete,
		HasAvatar:        cs.HasAvatar,
		HasJobHistory:    cs.HasJobHistory,
		NoNoShows:        cs.NoNoShows,
		NoExcessCancels:  cs.NoExcessCancels,
		JobsCompleted:    cs.JobsCompleted,
		JobsCancelled:    cs.JobsCancelled,
		NoShowCount:      cs.NoShowCount,
		UpdatedAt:        cs.UpdatedAt,
	}
}

// ToProScoreResponse converts a domain ProScore to API response.
func ToProScoreResponse(ps *domain.ProScore) ProScoreResponse {
	return ProScoreResponse{
		Score:              ps.Score,
		Level:              string(ps.Level()),
		Badge:              ps.Badge(),
		HandymanMultiplier: ps.HandymanMultiplier(),
		JobsCompleted:      ps.JobsCompleted,
		FiveStarReviews:    ps.FiveStarReviews,
		AvgResponseMins:    ps.AvgResponseMins,
		ProfileComplete:    ps.ProfileComplete,
		ActiveLast7Days:    ps.ActiveLast7Days,
		PortfolioCount:     ps.PortfolioCount,
		NoShowCount:        ps.NoShowCount,
		CancelledAfterAccept: ps.CancelledAfterAccept,
		SlowResponseCount:  ps.SlowResponseCount,
		LowRatingCount:     ps.LowRatingCount,
		UpdatedAt:          ps.UpdatedAt,
	}
}
