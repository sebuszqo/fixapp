// Package dispatch handles lead matching and distribution.
// When a job is published, this service finds matching handymen,
// calculates dynamic pricing, and creates leads.
package dispatch

import (
	"context"
	"time"

	"fixapp/internal/catalog"
	"fixapp/internal/domain"
	"fixapp/internal/handyman"
	"fixapp/internal/lead"
	"fixapp/internal/scoring"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// LeadExpiry is how long handymen have to respond to a lead.
	LeadExpiry = 24 * time.Hour
)

// Service handles job-to-lead matching and dispatch.
type Service struct {
	handymanRepo handyman.Repository
	catalogRepo  catalog.Repository
	leadRepo     lead.Repository
	scoringRepo  scoring.Repository
	logger       *zap.Logger
}

// NewService creates a new dispatch service.
func NewService(
	handymanRepo handyman.Repository,
	catalogRepo catalog.Repository,
	leadRepo lead.Repository,
	scoringRepo scoring.Repository,
	logger *zap.Logger,
) *Service {
	return &Service{
		handymanRepo: handymanRepo,
		catalogRepo:  catalogRepo,
		leadRepo:     leadRepo,
		scoringRepo:  scoringRepo,
		logger:       logger,
	}
}

// DispatchResult holds the result of dispatching a job.
type DispatchResult struct {
	LeadsCreated int
	HandymanIDs  []uuid.UUID
}

// DispatchJob finds matching handymen for a published job and creates leads.
// This should be called after a job transitions to "active" status.
func (s *Service) DispatchJob(ctx context.Context, job *domain.Job) (*DispatchResult, error) {
	if job.Status != domain.JobStatusActive {
		return nil, domain.ErrInvalidJobTransition
	}

	// Get the category's base price
	category, err := s.catalogRepo.GetCategoryByID(ctx, job.CategoryID)
	if err != nil {
		return nil, err
	}

	// Get client's commit score for pricing and display
	commitScore, err := s.scoringRepo.GetCommitScore(ctx, job.ClientID)
	if err != nil {
		return nil, err
	}

	// Find matching handymen (available + serve this category + district)
	isEmergency := job.Urgency == domain.JobUrgencyEmergency
	profiles, err := s.handymanRepo.FindMatchingForJob(ctx, job.CategoryID, job.DistrictID, isEmergency)
	if err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		s.logger.Info("no matching handymen found for job",
			zap.String("job_id", job.ID.String()),
			zap.String("category_id", job.CategoryID.String()),
			zap.String("district_id", job.DistrictID.String()),
		)
		return &DispatchResult{LeadsCreated: 0}, nil
	}

	// Calculate the effective client commit score (base + per-job bonus)
	effectiveCommitScore := commitScore.Score
	perJobBonus := calculatePerJobBonus(job)
	effectiveCommitScore += perJobBonus
	if effectiveCommitScore > 100 {
		effectiveCommitScore = 100
	}

	result := &DispatchResult{}

	for _, profile := range profiles {
		// Calculate dynamic price for this handyman
		proScore, err := s.scoringRepo.GetProScore(ctx, profile.UserID)
		if err != nil {
			s.logger.Warn("failed to get pro score for handyman, using default price",
				zap.String("handyman_id", profile.UserID.String()),
				zap.Error(err),
			)
			proScore = &domain.ProScore{Score: 500} // default to standard
		}

		price := calculateLeadPrice(category.BasePrice, commitScore, proScore)

		// Create lead
		lead := domain.NewLead(job.ID, profile.UserID, price, effectiveCommitScore, LeadExpiry)
		if err := s.leadRepo.Create(ctx, lead); err != nil {
			s.logger.Error("failed to create lead for handyman",
				zap.String("job_id", job.ID.String()),
				zap.String("handyman_id", profile.UserID.String()),
				zap.Error(err),
			)
			continue // skip this handyman, try others
		}

		result.LeadsCreated++
		result.HandymanIDs = append(result.HandymanIDs, profile.UserID)
	}

	s.logger.Info("job dispatched",
		zap.String("job_id", job.ID.String()),
		zap.Int("matching_handymen", len(profiles)),
		zap.Int("leads_created", result.LeadsCreated),
		zap.Int("base_price", category.BasePrice),
		zap.Int("client_commit_score", effectiveCommitScore),
	)

	return result, nil
}

// calculateLeadPrice computes: BasePrice × ClientMultiplier × HandymanMultiplier
func calculateLeadPrice(basePrice int, cs *domain.CommitScore, ps *domain.ProScore) int {
	price := float64(basePrice) * cs.ClientMultiplier() * ps.HandymanMultiplier()
	result := int(price + 0.5)
	if result < 1 {
		result = 1
	}
	return result
}

// calculatePerJobBonus computes extra commit score points based on job quality.
func calculatePerJobBonus(job *domain.Job) int {
	bonus := 0

	// +20 for detailed description (> 50 words approx = > 250 chars)
	if len(job.Description) > 250 {
		bonus += 20
	}

	// +15 for at least 1 photo
	if len(job.PhotoURLs) >= 1 {
		bonus += 15
	}

	// +10 for specific time window
	if job.PreferredTime != "" && job.PreferredTime != "flexible" {
		bonus += 10
	}

	return bonus
}
