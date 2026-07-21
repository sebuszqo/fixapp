package lead

import (
	"context"
	"fmt"
	"time"

	"fixapp/internal/auth"
	"fixapp/internal/domain"
	"fixapp/internal/job"
	"fixapp/internal/notification"
	"fixapp/internal/wallet"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// DefaultLeadExpiry is how long a handyman has to respond to a lead.
	DefaultLeadExpiry = 24 * time.Hour
)

// Service handles lead business logic.
type Service struct {
	repo         Repository
	jobRepo      job.Repository
	walletRepo   wallet.Repository
	notifService *notification.Service
	logger       *zap.Logger
}

// NewService creates a new lead service.
func NewService(repo Repository, jobRepo job.Repository, walletRepo wallet.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:       repo,
		jobRepo:    jobRepo,
		walletRepo: walletRepo,
		logger:     logger,
	}
}

// SetNotificationService sets the notification service instance.
func (s *Service) SetNotificationService(ns *notification.Service) {
	s.notifService = ns
}

// GetByID retrieves a lead by ID with access control.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	lead, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only the handyman who received the lead or admin can view it
	userID, _ := uuid.Parse(authUser.ID)
	if lead.HandymanID != userID && !authUser.IsAdmin() {
		return nil, domain.ErrForbidden
	}

	return lead, nil
}

// GetDetail retrieves a lead with job details.
func (s *Service) GetDetail(ctx context.Context, id uuid.UUID) (*domain.Lead, *domain.Job, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, nil, domain.ErrUnauthorized
	}

	lead, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// Only the handyman who received the lead or admin can view it
	userID, _ := uuid.Parse(authUser.ID)
	if lead.HandymanID != userID && !authUser.IsAdmin() {
		return nil, nil, domain.ErrForbidden
	}

	j, err := s.jobRepo.GetByID(ctx, lead.JobID)
	if err != nil {
		return nil, nil, err
	}

	return lead, j, nil
}

// Accept marks a lead as accepted. This deducts credits, stores handyman proposal, and updates status.
func (s *Service) Accept(ctx context.Context, id uuid.UUID, req *AcceptLeadRequest) (*domain.Lead, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	lead, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only the assigned handyman can accept
	handymanID, _ := uuid.Parse(authUser.ID)
	if lead.HandymanID != handymanID {
		return nil, domain.ErrForbidden
	}

	// Check if job is still active
	j, err := s.jobRepo.GetByID(ctx, lead.JobID)
	if err != nil {
		return nil, err
	}
	if j.Status != domain.JobStatusActive {
		return nil, domain.ErrJobAlreadyAccepted
	}

	// Accept the lead and attach proposal info if provided (credits deducted only when client accepts proposal)
	if err := lead.Accept(); err != nil {
		return nil, err
	}

	if req != nil {
		lead.EstimatedPrice = req.EstimatedPrice
		lead.ArrivalTime = req.ArrivalTime
		lead.ProposalMessage = req.ProposalMessage
	}

	// Update lead status and proposal
	if err := s.repo.Update(ctx, lead); err != nil {
		// Refund credits if DB update fails
		_ = s.walletRepo.CreditAtomic(ctx, handymanID, lead.Price, domain.ReasonLeadRefund, &lead.ID, "Refund: lead update failed")
		return nil, err
	}

	// Update job status to accepted
	if err := j.Accept(handymanID); err != nil {
		s.logger.Error("failed to accept job after lead acceptance",
			zap.String("lead_id", lead.ID.String()),
			zap.String("job_id", j.ID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	if err := s.jobRepo.Update(ctx, j); err != nil {
		return nil, err
	}

	// Send notifications to both handyman and client on proposal submission/update
	if s.notifService != nil {
		// Handyman notification
		handymanMsg := fmt.Sprintf("Pobrano prowizję %d PLN. Twoja propozycja dla zlecenia '%s' została wysłana do klienta.", lead.Price, j.Title)
		_, _ = s.notifService.CreateNotification(
			ctx,
			handymanID,
			domain.NotificationTypePayment,
			"Propozycja wysłana do klienta",
			handymanMsg,
			"/pro/requests",
		)

		// Client notification
		clientMsg := fmt.Sprintf("Fachowiec przesłał propozycję dla zlecenia '%s'.", j.Title)
		if lead.EstimatedPrice != nil {
			clientMsg += fmt.Sprintf(" Wycena: %d PLN.", *lead.EstimatedPrice)
		}
		if lead.ArrivalTime != "" {
			clientMsg += fmt.Sprintf(" Termin: %s.", lead.ArrivalTime)
		}
		if lead.ProposalMessage != "" {
			clientMsg += fmt.Sprintf(" Wiadomość: %s", lead.ProposalMessage)
		}

		_, _ = s.notifService.CreateNotification(
			ctx,
			j.ClientID,
			domain.NotificationTypeJobStatus,
			"Aktualizacja propozycji fachowca",
			clientMsg,
			"/client/orders",
		)
	}

	s.logger.Info("lead accepted",
		zap.String("lead_id", lead.ID.String()),
		zap.String("handyman_id", handymanID.String()),
		zap.String("job_id", lead.JobID.String()),
		zap.Int("price", lead.Price),
	)

	return lead, nil
}

// Reject marks a lead as rejected by the handyman.
func (s *Service) Reject(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	lead, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only the assigned handyman can reject
	handymanID, _ := uuid.Parse(authUser.ID)
	if lead.HandymanID != handymanID {
		return nil, domain.ErrForbidden
	}

	if err := lead.Reject(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, lead); err != nil {
		return nil, err
	}

	s.logger.Info("lead rejected",
		zap.String("lead_id", lead.ID.String()),
		zap.String("handyman_id", handymanID.String()),
	)

	return lead, nil
}

// ListMyLeads lists leads for the authenticated handyman.
func (s *Service) ListMyLeads(ctx context.Context, status *domain.LeadStatus, limit, offset int) ([]*domain.Lead, int64, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, 0, domain.ErrUnauthorized
	}

	// Expire any pending leads that have passed their expiration time
	_, _ = s.ExpireOldLeads(ctx)

	handymanID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, 0, domain.ErrInvalidInput
	}

	filter := DefaultListFilter()
	filter.Status = status
	if limit > 0 && limit <= 100 {
		filter.Limit = limit
	}
	if offset >= 0 {
		filter.Offset = offset
	}

	return s.repo.ListByHandyman(ctx, handymanID, filter)
}

// CreateLeadForJob creates a lead for a specific handyman for a given job.
// This is called by the matching/dispatch system when a job is published.
func (s *Service) CreateLeadForJob(ctx context.Context, jobID, handymanID uuid.UUID, price, clientCommitScore int) (*domain.Lead, error) {
	lead := domain.NewLead(jobID, handymanID, price, clientCommitScore, DefaultLeadExpiry)

	if err := s.repo.Create(ctx, lead); err != nil {
		return nil, err
	}

	s.logger.Info("lead created",
		zap.String("lead_id", lead.ID.String()),
		zap.String("job_id", jobID.String()),
		zap.String("handyman_id", handymanID.String()),
		zap.Int("price", price),
	)

	return lead, nil
}

// ExpireOldLeads marks expired pending leads. Should be called periodically.
func (s *Service) ExpireOldLeads(ctx context.Context) (int64, error) {
	count, err := s.repo.ExpirePendingLeads(ctx)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		s.logger.Info("expired pending leads", zap.Int64("count", count))
	}
	return count, nil
}
