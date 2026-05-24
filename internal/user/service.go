package user

import (
	"context"
	"errors"

	"fixapp/internal/auth"
	"fixapp/internal/auth/permission"
	"fixapp/internal/auth/role"
	"fixapp/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles user business logic.
// It uses the Repository interface, making it easy to test with mocks.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService creates a new user service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateFromSSO creates or retrieves a user from SSO provider data.
// This is called after successful OAuth authentication.
func (s *Service) CreateFromSSO(ctx context.Context, provider domain.AuthProvider, providerID, email, name, avatarURL string) (*domain.User, error) {
	// Check if user exists by provider
	user, err := s.repo.GetByProvider(ctx, provider, providerID)
	if err == nil {
		// Existing user - update last login
		if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
			s.logger.Warn("failed to update last login", zap.Error(err))
		}
		return user, nil
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	// Check if email is already registered with different provider
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		// Email exists with different provider - link accounts or reject
		s.logger.Info("email already exists with different provider",
			zap.String("email", email),
			zap.String("existing_provider", existingUser.Provider.String()),
			zap.String("new_provider", provider.String()),
		)
		// For now, return the existing user (you might want different behavior)
		return nil, domain.ErrUserAlreadyExists
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	// Create new user
	user = domain.NewUser(email, name, provider, providerID)
	user.AvatarURL = avatarURL
	user.EmailVerified = true // SSO providers verify email

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("created new user from SSO",
		zap.String("user_id", user.ID.String()),
		zap.String("email", email),
		zap.String("provider", provider.String()),
	)

	return user, nil
}

// GetByID retrieves a user by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByEmail retrieves a user by email.
func (s *Service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

// GetProfile retrieves the profile of the authenticated user.
func (s *Service) GetProfile(ctx context.Context) (*domain.User, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	id, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.GetByID(ctx, id)
}

// UpdateProfile updates the authenticated user's profile.
func (s *Service) UpdateProfile(ctx context.Context, name, phone, avatarURL string) (*domain.User, error) {
	authUser := auth.FromContext(ctx)
	if authUser == nil {
		return nil, domain.ErrUnauthorized
	}

	id, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update allowed fields
	if name != "" {
		user.Name = name
	}
	if phone != "" {
		user.Phone = phone
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// List retrieves users with filters (admin only).
func (s *Service) List(ctx context.Context, filter ListFilter) ([]*domain.User, int64, error) {
	if !auth.HasPermission(ctx, permission.UserList) {
		return nil, 0, domain.ErrForbidden
	}

	return s.repo.List(ctx, filter)
}

// UpdateRole changes a user's role (admin only).
func (s *Service) UpdateRole(ctx context.Context, userID uuid.UUID, newRole role.Role) error {
	if !auth.HasPermission(ctx, permission.AdminRoleMgmt) {
		return domain.ErrForbidden
	}

	// Prevent self-demotion from admin
	authUser := auth.FromContext(ctx)
	if authUser != nil && authUser.ID == userID.String() && authUser.Role == role.Admin && newRole != role.Admin {
		return errors.New("cannot demote yourself from admin")
	}

	if !newRole.IsValid() {
		return domain.ErrInvalidInput
	}

	if err := s.repo.UpdateRole(ctx, userID, newRole); err != nil {
		return err
	}

	s.logger.Info("user role updated",
		zap.String("user_id", userID.String()),
		zap.String("new_role", newRole.String()),
		zap.String("updated_by", authUser.ID),
	)

	return nil
}

// Deactivate disables a user account (admin only).
func (s *Service) Deactivate(ctx context.Context, userID uuid.UUID) error {
	if !auth.HasPermission(ctx, permission.UserDelete) {
		return domain.ErrForbidden
	}

	// Prevent self-deactivation
	authUser := auth.FromContext(ctx)
	if authUser != nil && authUser.ID == userID.String() {
		return errors.New("cannot deactivate yourself")
	}

	if err := s.repo.Delete(ctx, userID); err != nil {
		return err
	}

	s.logger.Info("user deactivated",
		zap.String("user_id", userID.String()),
		zap.String("deactivated_by", authUser.ID),
	)

	return nil
}

// GetByIDForAdmin retrieves any user by ID (admin only).
func (s *Service) GetByIDForAdmin(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	if !auth.HasPermission(ctx, permission.UserRead) {
		return nil, domain.ErrForbidden
	}

	return s.repo.GetByID(ctx, userID)
}

