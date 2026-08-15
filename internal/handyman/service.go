package handyman

import (
	"context"
	"errors"
	"time"

	"fixapp/internal/auth"
	"fixapp/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationService defines methods for sending notifications.
type NotificationService interface {
	CreateNotification(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, content, linkURL string) (*domain.Notification, error)
}

// Service handles handyman profile business logic.
type Service struct {
	repo         Repository
	logger       *zap.Logger
	notifService NotificationService
}

// NewService creates a new handyman service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// SetNotificationService sets the notification service for handyman.
func (s *Service) SetNotificationService(ns NotificationService) {
	s.notifService = ns
}

// CheckAndNotifyIncompleteProfile checks if the profile has missing data and sends a system notification.
func (s *Service) CheckAndNotifyIncompleteProfile(ctx context.Context, userID uuid.UUID, profile *domain.HandymanProfile) {
	if s.notifService == nil {
		return
	}
	isIncomplete := profile == nil || profile.NIP == "" || profile.CompanyName == "" || len(profile.Categories) == 0 || (profile.VerificationStatus != "verified" && !profile.IsVerified)
	if isIncomplete {
		_, _ = s.notifService.CreateNotification(
			ctx,
			userID,
			domain.NotificationTypeSystem,
			"Dokończ rejestrację profilu fachowca",
			"Twój profil nie jest jeszcze w pełni uzupełniony. Kliknij tutaj, aby dokończyć konfigurację i zacząć otrzymywać zlecenia.",
			"/register?type=pro",
		)
	}
}

// CreateProfile creates a new handyman profile.
func (s *Service) CreateProfile(ctx context.Context, userID uuid.UUID) (*domain.HandymanProfile, error) {
	profile := domain.NewHandymanProfile(userID)
	if err := s.repo.CreateProfile(ctx, profile); err != nil {
		return nil, err
	}

	s.logger.Info("handyman profile created",
		zap.String("profile_id", profile.ID.String()),
		zap.String("user_id", userID.String()),
	)

	return profile, nil
}

// GetMyProfile retrieves the authenticated handyman's profile.
func (s *Service) GetMyProfile(ctx context.Context) (*domain.HandymanProfile, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			profile, err = s.CreateProfile(ctx, userID)
			if err != nil {
				s.CheckAndNotifyIncompleteProfile(ctx, userID, nil)
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	s.CheckAndNotifyIncompleteProfile(ctx, userID, profile)
	return profile, nil
}

// GetFullProfile retrieves a profile with pricing and portfolio.
func (s *Service) GetFullProfile(ctx context.Context, profileID uuid.UUID) (*domain.HandymanProfile, []*domain.PricingItem, []*domain.PortfolioItem, error) {
	profile, err := s.repo.GetByID(ctx, profileID)
	if err != nil {
		return nil, nil, nil, err
	}

	pricing, err := s.repo.ListPricing(ctx, profileID)
	if err != nil {
		return nil, nil, nil, err
	}

	portfolio, err := s.repo.ListPortfolio(ctx, profileID)
	if err != nil {
		return nil, nil, nil, err
	}

	return profile, pricing, portfolio, nil
}

// GetMyFullProfile retrieves the authenticated handyman's full profile.
func (s *Service) GetMyFullProfile(ctx context.Context) (*domain.HandymanProfile, []*domain.PricingItem, []*domain.PortfolioItem, error) {
	profile, err := s.GetMyProfile(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	pricing, err := s.repo.ListPricing(ctx, profile.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	portfolio, err := s.repo.ListPortfolio(ctx, profile.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	return profile, pricing, portfolio, nil
}

// UpdateProfile updates the authenticated handyman's profile.
func (s *Service) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*domain.HandymanProfile, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.CompanyName != nil {
		profile.CompanyName = *req.CompanyName
	}
	if req.NIP != nil {
		profile.NIP = *req.NIP
	}
	if req.CompanyAddress != nil {
		profile.CompanyAddress = *req.CompanyAddress
	}
	if req.PKDCode != nil {
		profile.PKDCode = *req.PKDCode
	}
	if req.GUSStatus != nil {
		profile.GUSStatus = *req.GUSStatus
	}
	if req.GUSVerified != nil {
		profile.GUSVerified = *req.GUSVerified
	}
	if req.Phone != nil {
		profile.Phone = *req.Phone
	}
	if req.Email != nil {
		profile.Email = *req.Email
	}
	if req.BusinessType != nil {
		profile.BusinessType = *req.BusinessType
	}
	if req.IsVATPayer != nil {
		profile.IsVATPayer = *req.IsVATPayer
	}
	if req.ConsentIdentityVerification != nil {
		profile.ConsentIdentityVerification = *req.ConsentIdentityVerification
	}
	if req.ConsentMarketing != nil {
		profile.ConsentMarketing = *req.ConsentMarketing
	}
	if req.Bio != nil {
		profile.Bio = *req.Bio
	}
	if req.AvatarURL != nil {
		profile.AvatarURL = *req.AvatarURL
	}
	if req.ExperienceYears != nil {
		profile.ExperienceYears = *req.ExperienceYears
	}
	if req.WorkingHours != nil {
		profile.WorkingHours = *req.WorkingHours
	}
	if req.VerificationStatus != nil {
		profile.VerificationStatus = *req.VerificationStatus
	}
	if req.CEIDGDocumentURL != nil {
		profile.CEIDGDocumentURL = *req.CEIDGDocumentURL
	}
	if req.MicrotransferStatus != nil {
		profile.MicrotransferStatus = *req.MicrotransferStatus
	}
	if req.MicrotransferCode != nil {
		profile.MicrotransferCode = *req.MicrotransferCode
	}
	if req.Categories != nil {
		categories := make([]uuid.UUID, 0, len(req.Categories))
		for _, c := range req.Categories {
			id, err := uuid.Parse(c)
			if err != nil {
				return nil, domain.ErrInvalidInput
			}
			categories = append(categories, id)
		}
		if len(categories) > 3 {
			return nil, domain.ErrTooManyCategories
		}
		profile.Categories = categories
	}
	if req.Districts != nil {
		districts := make([]uuid.UUID, 0, len(req.Districts))
		for _, d := range req.Districts {
			id, err := uuid.Parse(d)
			if err != nil {
				return nil, domain.ErrInvalidInput
			}
			districts = append(districts, id)
		}
		profile.Districts = districts
	}
	if req.IsAvailable != nil {
		profile.IsAvailable = *req.IsAvailable
	}
	if req.EmergencyAvailable != nil {
		profile.EmergencyAvailable = *req.EmergencyAvailable
	}

	profile.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// LookupGUS performs a mock/live GUS database lookup for a Polish NIP number.
func (s *Service) LookupGUS(ctx context.Context, nip string) (*GUSLookupResponse, error) {
	// Clean NIP string
	cleanNip := ""
	for _, r := range nip {
		if r >= '0' && r <= '9' {
			cleanNip += string(r)
		}
	}

	if len(cleanNip) != 10 {
		return &GUSLookupResponse{
			Found:   false,
			NIP:     nip,
			Message: "Podany NIP musi składać się z 10 cyfr",
		}, nil
	}

	// Test case for invalid / suspended NIP
	if cleanNip == "0000000000" || cleanNip == "9999999999" || cleanNip == "1111111111" {
		return &GUSLookupResponse{
			Found:   false,
			NIP:     cleanNip,
			Status:  "Zawieszona",
			Message: "Podany NIP nie istnieje w rejestrze GUS lub działalność jest zawieszona/wykreślona.",
		}, nil
	}

	// Standard response for active business
	return &GUSLookupResponse{
		Found:       true,
		NIP:         cleanNip,
		CompanyName: "Usługi Hydrauliczne i Budowlane Jan Kowalski",
		Address:     "Kraków, ul. Przykładowa 1",
		PKD:         "4322Z - Wykonywanie instalacji wodno-kanalizacyjnych",
		Status:      "Aktywna",
	}, nil
}

// UploadCEIDG processes and attaches a CEIDG PDF document for verification.
func (s *Service) UploadCEIDG(ctx context.Context, docURL string) (*domain.HandymanProfile, error) {
	profile, err := s.GetMyProfile(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	profile.CEIDGDocumentURL = docURL
	profile.CEIDGUploadedAt = &now
	profile.VerificationStatus = "pending_review"
	profile.UpdatedAt = now

	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// ConfirmMicrotransfer records user confirmation of 1 gr microtransfer.
func (s *Service) ConfirmMicrotransfer(ctx context.Context, code string) (*domain.HandymanProfile, error) {
	profile, err := s.GetMyProfile(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	profile.MicrotransferStatus = "pending"
	profile.MicrotransferCode = code
	profile.MicrotransferConfirmedAt = &now
	if profile.VerificationStatus == "unverified" {
		profile.VerificationStatus = "pending_review"
	}
	profile.UpdatedAt = now

	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// Search finds handyman profiles matching criteria (public).
func (s *Service) Search(ctx context.Context, filter SearchFilter) ([]*domain.HandymanProfile, int64, error) {
	// Only show available profiles in public search
	available := true
	filter.Available = &available

	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return s.repo.Search(ctx, filter)
}

// ===== Pricing =====

// AddPricing adds a pricing item to the authenticated handyman's profile.
func (s *Service) AddPricing(ctx context.Context, req CreatePricingRequest) (*domain.PricingItem, error) {
	profile, err := s.GetMyProfile(ctx)
	if err != nil {
		return nil, err
	}

	unit := req.Unit
	if unit == "" {
		unit = "per service"
	}

	item := &domain.PricingItem{
		ID:                uuid.New(),
		ProfileID:         profile.ID,
		ServiceName:       req.ServiceName,
		PriceFrom:         req.PriceFrom,
		PriceTo:           req.PriceTo,
		Unit:              unit,
		EstimatedDuration: req.EstimatedDuration,
		SortOrder:         req.SortOrder,
	}

	if err := s.repo.CreatePricingItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// DeletePricing removes a pricing item (must belong to authenticated handyman).
func (s *Service) DeletePricing(ctx context.Context, itemID uuid.UUID) error {
	// We rely on the cascading profile ownership check via the handler
	return s.repo.DeletePricingItem(ctx, itemID)
}

// ===== Portfolio =====

// AddPortfolio adds a portfolio photo to the authenticated handyman's profile.
func (s *Service) AddPortfolio(ctx context.Context, req AddPortfolioRequest) (*domain.PortfolioItem, error) {
	profile, err := s.GetMyProfile(ctx)
	if err != nil {
		return nil, err
	}

	item := &domain.PortfolioItem{
		ID:        uuid.New(),
		ProfileID: profile.ID,
		ImageURL:  req.ImageURL,
		Caption:   req.Caption,
		SortOrder: req.SortOrder,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreatePortfolioItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// DeletePortfolio removes a portfolio photo.
func (s *Service) DeletePortfolio(ctx context.Context, itemID uuid.UUID) error {
	return s.repo.DeletePortfolioItem(ctx, itemID)
}

