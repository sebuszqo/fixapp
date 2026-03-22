package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"fixapp/internal/auth/provider"
	"fixapp/internal/auth/role"
	"fixapp/internal/auth/token"
	"fixapp/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrProviderNotFound = errors.New("auth provider not found")
	ErrInvalidState     = errors.New("invalid OAuth state")
	ErrUserDisabled     = errors.New("user account is disabled")
)

// UserRepository defines the minimal interface needed by auth service.
// This avoids import cycles with the user package.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByProvider(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateRole(ctx context.Context, id uuid.UUID, role role.Role) error
}

// Service handles authentication operations.
type Service struct {
	providers    *provider.Registry
	tokenService *token.Service
	userRepo     UserRepository
	logger       *zap.Logger

	// State storage for CSRF protection (in production, use Redis)
	states      map[string]stateData
	statesMu    sync.RWMutex
	stateExpiry time.Duration
}

// stateData holds OAuth state information.
type stateData struct {
	Provider    string
	CreatedAt   time.Time
	RedirectURL string    // Where to redirect after successful auth
	RegisterAs  role.Role // Role to assign on registration (user or handyman)
}

// Config holds auth service configuration.
type Config struct {
	StateExpiry time.Duration // How long OAuth states are valid
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		StateExpiry: 10 * time.Minute,
	}
}

// NewService creates a new auth service.
func NewService(
	providers *provider.Registry,
	tokenService *token.Service,
	userRepo UserRepository,
	logger *zap.Logger,
	cfg Config,
) *Service {
	s := &Service{
		providers:    providers,
		tokenService: tokenService,
		userRepo:     userRepo,
		logger:       logger,
		states:       make(map[string]stateData),
		stateExpiry:  cfg.StateExpiry,
	}

	// Start cleanup goroutine for expired states
	go s.cleanupStates()

	return s
}

// GetAuthURL returns the OAuth authorization URL for a provider.
// registerAs can be "user" or "handyman" - admin cannot self-register.
func (s *Service) GetAuthURL(providerName, redirectURL, registerAs string) (string, string, error) {
	p, ok := s.providers.GetOAuth(providerName)
	if !ok {
		return "", "", ErrProviderNotFound
	}

	// Determine role for registration
	registrationRole := role.User // Default
	if registerAs == "handyman" {
		registrationRole = role.Handyman
	}
	// Note: "admin" is intentionally NOT allowed here

	// Generate random state for CSRF protection
	state, err := generateState()
	if err != nil {
		return "", "", fmt.Errorf("generate state: %w", err)
	}

	// Store state for validation
	s.statesMu.Lock()
	s.states[state] = stateData{
		Provider:    providerName,
		CreatedAt:   time.Now(),
		RedirectURL: redirectURL,
		RegisterAs:  registrationRole,
	}
	s.statesMu.Unlock()

	authURL := p.AuthURL(state)
	return authURL, state, nil
}

// HandleCallback processes the OAuth callback and returns tokens.
func (s *Service) HandleCallback(ctx context.Context, providerName, code, state string) (*token.TokenPair, string, error) {
	// Validate state
	s.statesMu.Lock()
	stateInfo, ok := s.states[state]
	if ok {
		delete(s.states, state) // One-time use
	}
	s.statesMu.Unlock()

	if !ok {
		return nil, "", ErrInvalidState
	}

	if stateInfo.Provider != providerName {
		return nil, "", ErrInvalidState
	}

	if time.Since(stateInfo.CreatedAt) > s.stateExpiry {
		return nil, "", ErrInvalidState
	}

	// Get provider
	p, ok := s.providers.GetOAuth(providerName)
	if !ok {
		return nil, "", ErrProviderNotFound
	}

	// Exchange code for user info
	userInfo, err := p.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("exchange code: %w", err)
	}

	// Find or create user (pass the registration role for new users)
	domainUser, err := s.findOrCreateUser(ctx, providerName, userInfo, stateInfo.RegisterAs)
	if err != nil {
		return nil, "", err
	}

	// Check if user can login
	if !domainUser.IsActive {
		return nil, "", ErrUserDisabled
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, domainUser.ID); err != nil {
		s.logger.Warn("failed to update last login", zap.Error(err))
	}

	// Generate tokens
	tokens, err := s.tokenService.GenerateTokenPair(
		domainUser.ID.String(),
		domainUser.Email,
		domainUser.Name,
		domainUser.Role,
		domainUser.Provider.String(),
	)
	if err != nil {
		return nil, "", fmt.Errorf("generate tokens: %w", err)
	}

	s.logger.Info("user authenticated",
		zap.String("user_id", domainUser.ID.String()),
		zap.String("email", domainUser.Email),
		zap.String("provider", providerName),
	)

	return tokens, stateInfo.RedirectURL, nil
}

// RefreshTokens generates new tokens using a refresh token.
func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (*token.TokenPair, error) {
	// Validate refresh token
	userID, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user from database
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	domainUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !domainUser.IsActive {
		return nil, ErrUserDisabled
	}

	// Generate new tokens
	return s.tokenService.GenerateTokenPair(
		domainUser.ID.String(),
		domainUser.Email,
		domainUser.Name,
		domainUser.Role,
		domainUser.Provider.String(),
	)
}

// ValidateToken validates an access token and returns claims.
func (s *Service) ValidateToken(tokenString string) (*token.Claims, error) {
	return s.tokenService.ValidateAccessToken(tokenString)
}

// ListProviders returns available OAuth providers.
func (s *Service) ListProviders() []string {
	return s.providers.ListOAuth()
}

// findOrCreateUser looks up or creates a user from OAuth data.
// registrationRole is only used for NEW users - existing users keep their role.
func (s *Service) findOrCreateUser(ctx context.Context, providerName string, info *provider.UserInfo, registrationRole role.Role) (*domain.User, error) {
	authProvider := domain.AuthProvider(providerName)

	// Try to find by provider ID
	domainUser, err := s.userRepo.GetByProvider(ctx, authProvider, info.ProviderID)
	if err == nil {
		// EXISTING USER - keep their current role, just update profile info
		updated := false
		if domainUser.Name != info.Name {
			domainUser.Name = info.Name
			updated = true
		}
		if domainUser.AvatarURL != info.AvatarURL {
			domainUser.AvatarURL = info.AvatarURL
			updated = true
		}
		if updated {
			if err := s.userRepo.Update(ctx, domainUser); err != nil {
				s.logger.Warn("failed to update user info", zap.Error(err))
			}
		}
		return domainUser, nil
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	// Check if email exists (different provider)
	existingUser, err := s.userRepo.GetByEmail(ctx, info.Email)
	if err == nil {
		// Email exists with different provider
		// You could choose to link accounts here, or reject
		s.logger.Warn("email exists with different provider",
			zap.String("email", info.Email),
			zap.String("existing_provider", existingUser.Provider.String()),
			zap.String("new_provider", providerName),
		)
		return nil, fmt.Errorf("email already registered with %s", existingUser.Provider)
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	// Create NEW user with the requested role
	newUser := domain.NewUser(info.Email, info.Name, authProvider, info.ProviderID)
	newUser.AvatarURL = info.AvatarURL
	newUser.EmailVerified = info.EmailVerified
	
	// Apply the registration role (user or handyman, never admin)
	if registrationRole == role.User || registrationRole == role.Handyman {
		newUser.Role = registrationRole
	}
	// If registrationRole is invalid or admin, keep default (user)

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.logger.Info("created new user",
		zap.String("user_id", newUser.ID.String()),
		zap.String("email", newUser.Email),
		zap.String("provider", providerName),
		zap.String("role", newUser.Role.String()),
	)

	return newUser, nil
}

// cleanupStates periodically removes expired OAuth states.
func (s *Service) cleanupStates() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.statesMu.Lock()
		now := time.Now()
		for state, data := range s.states {
			if now.Sub(data.CreatedAt) > s.stateExpiry {
				delete(s.states, state)
			}
		}
		s.statesMu.Unlock()
	}
}

// generateState creates a cryptographically secure random state string.
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
