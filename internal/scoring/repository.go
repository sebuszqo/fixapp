// Package scoring provides Commit Score and ProScore calculation and storage.
package scoring

import (
	"context"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for score data access.
type Repository interface {
	// CommitScore
	GetCommitScore(ctx context.Context, userID uuid.UUID) (*domain.CommitScore, error)
	UpsertCommitScore(ctx context.Context, score *domain.CommitScore) error

	// ProScore
	GetProScore(ctx context.Context, userID uuid.UUID) (*domain.ProScore, error)
	UpsertProScore(ctx context.Context, score *domain.ProScore) error
}
