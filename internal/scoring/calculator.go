package scoring

import "fixapp/internal/domain"

// CalculateCommitScore computes the Commit Score (0-100) from factors.
//
// Positive factors:
//   +20 Verified phone
//   +15 Complete profile (name + district)
//   +10 Has avatar photo
//   +10 Has job history (completed > 0)
//   +10 No no-shows in history
//   +10 No excess cancellations (< 2 after accepted)
//
// Negative factors:
//   -20 Has been a no-show
//   -15 Cancelled 2+ jobs after accepted
//   -10 No avatar and no phone
//
// Per-job bonuses (not stored, computed at lead creation time):
//   +20 Description > 50 words
//   +15 At least 1 photo attached
//   +10 Specific time window selected
func CalculateCommitScore(cs *domain.CommitScore) int {
	score := 0

	// Positive factors
	if cs.PhoneVerified {
		score += 20
	}
	if cs.ProfileComplete {
		score += 15
	}
	if cs.HasAvatar {
		score += 10
	}
	if cs.HasJobHistory {
		score += 10
	}
	if cs.NoNoShows {
		score += 10
	}
	if cs.NoExcessCancels {
		score += 10
	}

	// Negative adjustments
	if cs.NoShowCount > 0 {
		score -= 20
	}
	if cs.JobsCancelled >= 2 {
		score -= 15
	}

	// Clamp to 0-100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// CalculatePerJobBonus computes extra Commit Score points for a specific job.
// This is used at lead creation time to give handymen a better picture.
func CalculatePerJobBonus(descriptionWords int, photoCount int, hasSpecificTime bool) int {
	bonus := 0
	if descriptionWords > 50 {
		bonus += 20
	}
	if photoCount >= 1 {
		bonus += 15
	}
	if hasSpecificTime {
		bonus += 10
	}
	return bonus
}

// CalculateProScore computes the ProScore (0-1000) from factors.
//
// Positive factors:
//   +50 per job completed (via platform) - max contribution 500
//   +30 per 5-star review - max contribution 300
//   +20 response time < 1h (based on avg)
//   +15 profile 100% complete
//   +10 active in last 7 days
//   +5  per portfolio photo - max contribution 50
//
// Penalties:
//   -100 per no-show
//   -50  per cancelled after accepted
//   -30  per slow response (> 24h count)
//   -20  per 1-2 star review
func CalculateProScore(ps *domain.ProScore) int {
	score := 0

	// Positive: completed jobs (max 500 pts, i.e. 10 jobs to max)
	jobPts := ps.JobsCompleted * 50
	if jobPts > 500 {
		jobPts = 500
	}
	score += jobPts

	// Positive: 5-star reviews (max 300 pts, i.e. 10 reviews to max)
	reviewPts := ps.FiveStarReviews * 30
	if reviewPts > 300 {
		reviewPts = 300
	}
	score += reviewPts

	// Positive: fast response time
	if ps.AvgResponseMins > 0 && ps.AvgResponseMins < 60 {
		score += 20
	}

	// Positive: profile complete
	if ps.ProfileComplete {
		score += 15
	}

	// Positive: recent activity
	if ps.ActiveLast7Days {
		score += 10
	}

	// Positive: portfolio (max 50 pts, i.e. 10 photos)
	portfolioPts := ps.PortfolioCount * 5
	if portfolioPts > 50 {
		portfolioPts = 50
	}
	score += portfolioPts

	// Penalties
	score -= ps.NoShowCount * 100
	score -= ps.CancelledAfterAccept * 50
	score -= ps.SlowResponseCount * 30
	score -= ps.LowRatingCount * 20

	// Clamp to 0-1000
	if score < 0 {
		score = 0
	}
	if score > 1000 {
		score = 1000
	}

	return score
}
