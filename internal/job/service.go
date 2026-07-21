package job

import (
	"context"
	"fmt"

	"fixapp/internal/auth"
	"fixapp/internal/auth/permission"
	"fixapp/internal/domain"
	"fixapp/internal/notification"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Dispatcher is the interface for lead dispatch (avoids circular dependency).
type Dispatcher interface {
	DispatchJob(ctx context.Context, job *domain.Job, handymanIDs []uuid.UUID) error
}

// WalletRepository defines interface for deducting credits (avoids circular dependency).
type WalletRepository interface {
	DebitAtomic(ctx context.Context, userID uuid.UUID, amount int, reason domain.TransactionReason, referenceID *uuid.UUID, description string) error
}

// LeadRepository defines interface for querying job leads (avoids circular dependency).
type LeadRepository interface {
	ListByJob(ctx context.Context, jobID uuid.UUID) ([]*domain.Lead, error)
}

// Service handles job business logic.
type Service struct {
	repo         Repository
	dispatcher   Dispatcher
	walletRepo   WalletRepository
	leadRepo     LeadRepository
	notifService *notification.Service
	logger       *zap.Logger
}

// NewService creates a new job service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// SetDispatcher sets the dispatcher (called after initialization to break circular dep).
func (s *Service) SetDispatcher(d Dispatcher) {
	s.dispatcher = d
}

func (s *Service) SetWalletRepository(w WalletRepository) {
	s.walletRepo = w
}

func (s *Service) SetLeadRepository(l LeadRepository) {
	s.leadRepo = l
}

// SetNotificationService sets the notification service instance.
func (s *Service) SetNotificationService(ns *notification.Service) {
	s.notifService = ns
}

// Create creates a new job for the authenticated client.
func (s *Service) Create(ctx context.Context, req CreateJobRequest) (*domain.Job, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	clientID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	districtID, err := uuid.Parse(req.DistrictID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	job := domain.NewJob(clientID, categoryID, districtID, req.Title, req.Description)

	// Set optional fields
	if req.Urgency != "" {
		job.Urgency = domain.JobUrgency(req.Urgency)
	}
	if req.Address != "" {
		job.Address = req.Address
	}
	if req.BuildingType != "" {
		job.BuildingType = req.BuildingType
	}
	job.Floor = req.Floor
	job.HasElevator = req.HasElevator
	job.PreferredDate1 = req.PreferredDate1
	job.PreferredDate2 = req.PreferredDate2
	if req.PreferredTime != "" {
		job.PreferredTime = req.PreferredTime
	}
	job.Budget = req.Budget
	job.WantsInvoice = req.WantsInvoice
	if req.ContactMethod != "" {
		job.ContactMethod = req.ContactMethod
	}
	if req.PhotoURLs != nil {
		job.PhotoURLs = req.PhotoURLs
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	s.logger.Info("job created",
		zap.String("job_id", job.ID.String()),
		zap.String("client_id", clientID.String()),
	)

	return job, nil
}

// GetByID retrieves a job by ID with access control.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Clients can see their own jobs, handymen can see active jobs, admins can see all
	userID, _ := uuid.Parse(authUser.ID)
	if job.ClientID != userID && !auth.HasPermission(ctx, permission.JobRead) {
		// Handymen can see active/accepted jobs (they might have a lead for it)
		if authUser.IsHandyman() && (job.Status == domain.JobStatusActive || job.Status == domain.JobStatusAccepted) {
			return job, nil
		}
		return nil, domain.ErrForbidden
	}

	return job, nil
}

// Publish moves a job from draft to active (visible to handymen).
// If handymanIDs is provided and not empty, it will only dispatch leads to those specific handymen.
func (s *Service) Publish(ctx context.Context, id uuid.UUID, handymanIDs []uuid.UUID) (*domain.Job, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only the job owner can publish
	userID, _ := uuid.Parse(authUser.ID)
	if job.ClientID != userID {
		return nil, domain.ErrForbidden
	}

	if err := job.Publish(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, job); err != nil {
		return nil, err
	}

	s.logger.Info("job published",
		zap.String("job_id", job.ID.String()),
	)

	// Dispatch leads to matching handymen (async-safe, non-blocking)
	if s.dispatcher != nil {
		if err := s.dispatcher.DispatchJob(ctx, job, handymanIDs); err != nil {
			s.logger.Error("failed to dispatch job leads",
				zap.String("job_id", job.ID.String()),
				zap.Error(err),
			)
			// Don't fail the publish — job is active, leads can be retried
		}
	}

	return job, nil
}

// StartWork marks the job proposal as accepted by the client and moves status to in_progress.
func (s *Service) StartWork(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only the job owner (client) can confirm proposal and start work
	userID, _ := uuid.Parse(authUser.ID)
	if job.ClientID != userID {
		return nil, domain.ErrForbidden
	}

	if err := job.StartWork(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, job); err != nil {
		return nil, err
	}

	// Deduct credits from handyman ONLY NOW after proposal is accepted by client
	if job.CompletedByID != nil && s.walletRepo != nil && s.leadRepo != nil {
		leads, err := s.leadRepo.ListByJob(ctx, job.ID)
		if err == nil {
			for _, l := range leads {
				if l.HandymanID == *job.CompletedByID && l.Status == domain.LeadStatusAccepted {
					err = s.walletRepo.DebitAtomic(ctx, *job.CompletedByID, l.Price, domain.ReasonLeadAccepted, &l.ID, "Opłata za przyjętą propozycję wyceny zlecenia")
					if err != nil {
						s.logger.Warn("failed to debit credits for accepted lead", zap.Error(err))
					}
					break
				}
			}
		}
	}

	if s.notifService != nil {
		if job.CompletedByID != nil {
			_, _ = s.notifService.CreateNotification(
				ctx,
				*job.CompletedByID,
				domain.NotificationTypeJobStatus,
				"Rozpoczęto prace nad zleceniem",
				fmt.Sprintf("Klient zaakceptował wycenę i rozpoczął prace nad zleceniem '%s'.", job.Title),
				"/pro/requests",
			)
		}
		_, _ = s.notifService.CreateNotification(
			ctx,
			job.ClientID,
			domain.NotificationTypeJobStatus,
			"Status zlecenia zmieniony",
			fmt.Sprintf("Zlecenie '%s' zmieniło status na: W trakcie realizacji.", job.Title),
			"/client/orders",
		)
	}

	s.logger.Info("job proposal accepted, work started",
		zap.String("job_id", job.ID.String()),
		zap.String("client_id", userID.String()),
	)

	return job, nil
}

// Complete marks a job as done with a declared final value.
func (s *Service) Complete(ctx context.Context, id uuid.UUID, finalValue int) (*domain.Job, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only the assigned handyman can complete the job
	handymanID, _ := uuid.Parse(authUser.ID)
	if job.CompletedByID == nil || *job.CompletedByID != handymanID {
		return nil, domain.ErrForbidden
	}

	if err := job.Complete(finalValue); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, job); err != nil {
		return nil, err
	}

	if s.notifService != nil {
		_, _ = s.notifService.CreateNotification(
			ctx,
			job.ClientID,
			domain.NotificationTypeJobStatus,
			"Zlecenie wykonane przez fachowca",
			fmt.Sprintf("Fachowiec oznaczył zlecenie '%s' jako wykonane. Potwierdź wykonanie usugi w aplikacji.", job.Title),
			"/client/orders",
		)
	}

	s.logger.Info("job completed",
		zap.String("job_id", job.ID.String()),
		zap.Int("final_value", finalValue),
	)

	return job, nil
}

// Confirm marks a job as confirmed by the client.
func (s *Service) Confirm(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only the job owner (client) can confirm
	userID, _ := uuid.Parse(authUser.ID)
	if job.ClientID != userID {
		return nil, domain.ErrForbidden
	}

	if err := job.ConfirmByClient(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, job); err != nil {
		return nil, err
	}

	if s.notifService != nil && job.CompletedByID != nil {
		_, _ = s.notifService.CreateNotification(
			ctx,
			*job.CompletedByID,
			domain.NotificationTypeJobStatus,
			"Zlecenie potwierdzone",
			fmt.Sprintf("Klient potwierdził pomyślne wykonanie zlecenia '%s'.", job.Title),
			"/pro/requests",
		)
	}

	s.logger.Info("job confirmed by client",
		zap.String("job_id", job.ID.String()),
		zap.String("client_id", userID.String()),
	)

	return job, nil
}

// Cancel cancels a job.
func (s *Service) Cancel(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Client can cancel their own job, admin can cancel any
	userID, _ := uuid.Parse(authUser.ID)
	if job.ClientID != userID && !auth.HasPermission(ctx, permission.JobDelete) {
		return nil, domain.ErrForbidden
	}

	if err := job.Cancel(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, job); err != nil {
		return nil, err
	}

	if s.notifService != nil {
		_, _ = s.notifService.CreateNotification(
			ctx,
			job.ClientID,
			domain.NotificationTypeJobStatus,
			"Anulowano zlecenie",
			fmt.Sprintf("Zlecenie '%s' zostało anulowane.", job.Title),
			"/client/orders",
		)
		if job.CompletedByID != nil {
			_, _ = s.notifService.CreateNotification(
				ctx,
				*job.CompletedByID,
				domain.NotificationTypeJobStatus,
				"Anulowano zlecenie",
				fmt.Sprintf("Zlecenie '%s' zostało anulowane.", job.Title),
				"/pro/requests",
			)
		}
	}

	s.logger.Info("job cancelled",
		zap.String("job_id", job.ID.String()),
		zap.String("cancelled_by", authUser.ID),
	)

	return job, nil
}

// ListMyJobs lists jobs for the authenticated client.
func (s *Service) ListMyJobs(ctx context.Context, status *domain.JobStatus, limit, offset int) ([]*domain.Job, int64, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, 0, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, 0, domain.ErrInvalidInput
	}

	filter := DefaultListFilter()
	filter.ClientID = &userID
	filter.Status = status
	if limit > 0 && limit <= 100 {
		filter.Limit = limit
	}
	if offset >= 0 {
		filter.Offset = offset
	}

	return s.repo.List(ctx, filter)
}

// ListAll lists all jobs (admin only).
func (s *Service) ListAll(ctx context.Context, filter ListFilter) ([]*domain.Job, int64, error) {
	if !auth.HasPermission(ctx, permission.JobList) {
		return nil, 0, domain.ErrForbidden
	}
	return s.repo.List(ctx, filter)
}
