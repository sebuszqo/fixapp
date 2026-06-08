package wallet

import (
	"context"

	"fixapp/internal/auth"
	"fixapp/internal/auth/permission"
	"fixapp/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles wallet business logic.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService creates a new wallet service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateWallet creates a new wallet for a user.
func (s *Service) CreateWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	wallet := domain.NewWallet(userID)
	if err := s.repo.CreateWallet(ctx, wallet); err != nil {
		return nil, err
	}

	s.logger.Info("wallet created",
		zap.String("wallet_id", wallet.ID.String()),
		zap.String("user_id", userID.String()),
	)

	return wallet, nil
}

// CreateWalletWithBonus creates a wallet and gives initial bonus credits.
func (s *Service) CreateWalletWithBonus(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	wallet, err := s.CreateWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Credit initial bonus
	err = s.repo.CreditAtomic(ctx, userID, domain.InitialBonusCredits, domain.ReasonInitialBonus, nil,
		"Welcome bonus - 50 free credits")
	if err != nil {
		s.logger.Error("failed to credit initial bonus",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return wallet, nil // wallet created but bonus failed, non-fatal
	}

	// Refresh wallet to get updated balance
	wallet, err = s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("initial bonus credited",
		zap.String("user_id", userID.String()),
		zap.Int("amount", domain.InitialBonusCredits),
	)

	return wallet, nil
}

// GetMyWallet retrieves the authenticated user's wallet.
func (s *Service) GetMyWallet(ctx context.Context) (*domain.Wallet, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.GetByUserID(ctx, userID)
}

// GetWalletByUserID retrieves a wallet by user ID (admin).
func (s *Service) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	if !auth.HasPermission(ctx, permission.AdminUserMgmt) {
		return nil, domain.ErrForbidden
	}
	return s.repo.GetByUserID(ctx, userID)
}

// GetMyTransactions retrieves the authenticated user's transactions.
func (s *Service) GetMyTransactions(ctx context.Context, limit, offset int) ([]*domain.WalletTransaction, int64, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, 0, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, 0, domain.ErrInvalidInput
	}

	wallet, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListTransactions(ctx, wallet.ID, limit, offset)
}

// AdminTopUp adds credits to a user's wallet (admin only).
func (s *Service) AdminTopUp(ctx context.Context, userID uuid.UUID, amount int, description string) (*domain.Wallet, error) {
	if !auth.HasPermission(ctx, permission.AdminUserMgmt) {
		return nil, domain.ErrForbidden
	}

	if description == "" {
		description = "Admin top-up"
	}

	err := s.repo.CreditAtomic(ctx, userID, amount, domain.ReasonAdminTopUp, nil, description)
	if err != nil {
		return nil, err
	}

	adminUser := auth.FromContext(ctx)
	s.logger.Info("admin top-up",
		zap.String("user_id", userID.String()),
		zap.Int("amount", amount),
		zap.String("admin_id", adminUser.ID),
		zap.String("description", description),
	)

	return s.repo.GetByUserID(ctx, userID)
}

// DebitForLead deducts credits for lead acceptance.
func (s *Service) DebitForLead(ctx context.Context, userID uuid.UUID, amount int, leadID uuid.UUID) error {
	err := s.repo.DebitAtomic(ctx, userID, amount, domain.ReasonLeadAccepted, &leadID,
		"Lead acceptance fee")
	if err != nil {
		return err
	}

	s.logger.Info("lead fee debited",
		zap.String("user_id", userID.String()),
		zap.Int("amount", amount),
		zap.String("lead_id", leadID.String()),
	)

	return nil
}

// RefundForLead refunds credits for a lead (e.g., no-show by client).
func (s *Service) RefundForLead(ctx context.Context, userID uuid.UUID, amount int, leadID uuid.UUID) error {
	err := s.repo.CreditAtomic(ctx, userID, amount, domain.ReasonLeadRefund, &leadID,
		"Lead refund - client no-show")
	if err != nil {
		return err
	}

	s.logger.Info("lead fee refunded",
		zap.String("user_id", userID.String()),
		zap.Int("amount", amount),
		zap.String("lead_id", leadID.String()),
	)

	return nil
}

// RewardClient credits client for confirming job completion.
func (s *Service) RewardClient(ctx context.Context, userID uuid.UUID, jobID uuid.UUID) error {
	err := s.repo.CreditAtomic(ctx, userID, domain.ClientConfirmReward, domain.ReasonClientReward, &jobID,
		"Reward for confirming job completion")
	if err != nil {
		return err
	}

	s.logger.Info("client reward credited",
		zap.String("user_id", userID.String()),
		zap.Int("amount", domain.ClientConfirmReward),
		zap.String("job_id", jobID.String()),
	)

	return nil
}
