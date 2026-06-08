package domain

import (
	"time"

	"github.com/google/uuid"
)

// CommitScore represents a client's reliability score (0-100).
type CommitScore struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Score  int // 0-100

	// Score breakdown (stored for transparency)
	PhoneVerified    bool // +20
	ProfileComplete  bool // +15 (name + district)
	HasAvatar        bool // +10
	HasJobHistory    bool // +10 (completed jobs > 0)
	NoNoShows        bool // +10 (no no_show history)
	NoExcessCancels  bool // +10 (< 2 cancellations after accepted)
	// Dynamic factors computed per-job:
	// +20 description > 50 words
	// +15 at least 1 photo
	// +10 specific time window

	// Stats
	JobsCompleted  int
	JobsCancelled  int
	NoShowCount    int

	UpdatedAt time.Time
}

// CommitScoreLevel represents the tier based on score.
type CommitScoreLevel string

const (
	CommitScoreLevelVerified    CommitScoreLevel = "verified"    // 80-100
	CommitScoreLevelStandard   CommitScoreLevel = "standard"    // 50-79
	CommitScoreLevelUnverified CommitScoreLevel = "unverified"  // 0-49
)

// Level returns the CommitScore tier.
func (cs *CommitScore) Level() CommitScoreLevel {
	switch {
	case cs.Score >= 80:
		return CommitScoreLevelVerified
	case cs.Score >= 50:
		return CommitScoreLevelStandard
	default:
		return CommitScoreLevelUnverified
	}
}

// ClientMultiplier returns the dynamic pricing multiplier for this score.
func (cs *CommitScore) ClientMultiplier() float64 {
	switch cs.Level() {
	case CommitScoreLevelVerified:
		return 0.8
	case CommitScoreLevelStandard:
		return 1.0
	default:
		return 1.2
	}
}

// ProScore represents a handyman's reputation score (0-1000).
type ProScore struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Score  int // 0-1000

	// Score breakdown
	JobsCompleted     int
	FiveStarReviews   int
	AvgResponseMins   int  // average response time in minutes
	ProfileComplete   bool // 100% profile completion
	ActiveLast7Days   bool
	PortfolioCount    int

	// Penalties
	NoShowCount       int
	CancelledAfterAccept int
	SlowResponseCount int // > 24h responses
	LowRatingCount    int // 1-2 star reviews

	UpdatedAt time.Time
}

// ProScoreLevel represents the tier based on score.
type ProScoreLevel string

const (
	ProScoreLevelPartner  ProScoreLevel = "pro_partner" // 800+
	ProScoreLevelStandard ProScoreLevel = "standard"    // 300-799
	ProScoreLevelLow      ProScoreLevel = "low"         // 0-299
)

// Level returns the ProScore tier.
func (ps *ProScore) Level() ProScoreLevel {
	switch {
	case ps.Score >= 800:
		return ProScoreLevelPartner
	case ps.Score >= 300:
		return ProScoreLevelStandard
	default:
		return ProScoreLevelLow
	}
}

// HandymanMultiplier returns the dynamic pricing multiplier for this score.
func (ps *ProScore) HandymanMultiplier() float64 {
	switch ps.Level() {
	case ProScoreLevelPartner:
		return 0.8
	case ProScoreLevelStandard:
		return 1.0
	default:
		return 1.1
	}
}

// Badge returns the badge name for this score level.
func (ps *ProScore) Badge() string {
	if ps.Level() == ProScoreLevelPartner {
		return "Pro Partner"
	}
	return ""
}
