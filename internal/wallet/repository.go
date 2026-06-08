// Package wallet provides credit wallet management functionality.
package wallet

import (
	"context"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for wallet data access.
type Repository interface {
	// CreateWallet creates a new wallet for a user.
	CreateWallet(ctx context.Context, wallet *domain.Wallet) error

	// GetByUserID retrieves a wallet by user ID.
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)

	// UpdateBalance updates the wallet balance (use within a transaction).
	UpdateBalance(ctx context.Context, walletID uuid.UUID, newBalance int) error

	// CreateTransaction records a wallet transaction.
	CreateTransaction(ctx context.Context, tx *domain.WalletTransaction) error

	// ListTransactions retrieves transactions for a wallet with pagination.
	ListTransactions(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*domain.WalletTransaction, int64, error)

	// DebitAtomic atomically debits the wallet and records the transaction.
	// Returns ErrInsufficientCredits if balance is too low.
	DebitAtomic(ctx context.Context, userID uuid.UUID, amount int, reason domain.TransactionReason, referenceID *uuid.UUID, description string) error

	// CreditAtomic atomically credits the wallet and records the transaction.
	CreditAtomic(ctx context.Context, userID uuid.UUID, amount int, reason domain.TransactionReason, referenceID *uuid.UUID, description string) error
}
