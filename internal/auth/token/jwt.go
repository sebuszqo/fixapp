// Package token provides JWT token generation and validation.
package token

import (
	"errors"
	"fmt"
	"time"

	"fixapp/internal/auth/role"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired   = errors.New("token has expired")
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrTokenMalformed = errors.New("token is malformed")
)

// Claims represents the JWT claims for user authentication.
type Claims struct {
	jwt.RegisteredClaims
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     role.Role `json:"role"`
	Provider string    `json:"provider"`
}

// Service handles JWT token operations.
type Service struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
}

// Config holds JWT service configuration.
type Config struct {
	SecretKey     string        // Secret key for signing tokens
	AccessExpiry  time.Duration // Access token expiration (e.g., 15 minutes)
	RefreshExpiry time.Duration // Refresh token expiration (e.g., 7 days)
	Issuer        string        // Token issuer (e.g., "fixapp")
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig(secretKey string) Config {
	return Config{
		SecretKey:     secretKey,
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "fixapp",
	}
}

// NewService creates a new JWT service.
func NewService(cfg Config) *Service {
	return &Service{
		secretKey:     []byte(cfg.SecretKey),
		accessExpiry:  cfg.AccessExpiry,
		refreshExpiry: cfg.RefreshExpiry,
		issuer:        cfg.Issuer,
	}
}

// TokenPair represents an access and refresh token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// GenerateTokenPair creates both access and refresh tokens for a user.
func (s *Service) GenerateTokenPair(userID, email, name string, r role.Role, provider string) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessToken, err := s.generateToken(userID, email, name, r, provider, s.accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// Generate refresh token (longer lived, fewer claims)
	refreshToken, err := s.generateRefreshToken(userID, s.refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(s.accessExpiry),
		TokenType:    "Bearer",
	}, nil
}

// GenerateAccessToken creates an access token for a user.
func (s *Service) GenerateAccessToken(userID, email, name string, r role.Role, provider string) (string, error) {
	return s.generateToken(userID, email, name, r, provider, s.accessExpiry)
}

// ValidateAccessToken validates an access token and returns the claims.
func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID.
func (s *Service) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &refreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrTokenExpired
		}
		return "", ErrTokenInvalid
	}

	claims, ok := token.Claims.(*refreshClaims)
	if !ok || !token.Valid {
		return "", ErrTokenInvalid
	}

	return claims.UserID, nil
}

// generateToken creates a signed JWT with user claims.
func (s *Service) generateToken(userID, email, name string, r role.Role, provider string, expiry time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:   userID,
		Email:    email,
		Name:     name,
		Role:     r,
		Provider: provider,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// refreshClaims are minimal claims for refresh tokens.
type refreshClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// generateRefreshToken creates a refresh token with minimal claims.
func (s *Service) generateRefreshToken(userID string, expiry time.Duration) (string, error) {
	now := time.Now()

	claims := refreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

