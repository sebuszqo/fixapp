package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"fixapp/internal/domain"

	"github.com/google/uuid"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL wallet repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateWallet creates a new wallet for a user.
func (r *PostgresRepository) CreateWallet(ctx context.Context, wallet *domain.Wallet) error {
	query := `
		INSERT INTO wallets (id, user_id, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		wallet.ID, wallet.UserID, wallet.Balance, wallet.CreatedAt, wallet.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			return domain.ErrWalletAlreadyExists
		}
		return err
	}
	return nil
}

// GetByUserID retrieves a wallet by user ID.
func (r *PostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	query := `SELECT id, user_id, balance, created_at, updated_at FROM wallets WHERE user_id = $1`

	w := &domain.Wallet{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&w.ID, &w.UserID, &w.Balance, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}
		return nil, err
	}
	return w, nil
}

// UpdateBalance updates the wallet balance.
func (r *PostgresRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, newBalance int) error {
	query := `UPDATE wallets SET balance = $2 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, walletID, newBalance)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrWalletNotFound
	}
	return nil
}

// CreateTransaction records a wallet transaction.
func (r *PostgresRepository) CreateTransaction(ctx context.Context, tx *domain.WalletTransaction) error {
	query := `
		INSERT INTO wallet_transactions (id, wallet_id, type, amount, reason, reference_id, description, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		tx.ID, tx.WalletID, tx.Type, tx.Amount, tx.Reason,
		tx.ReferenceID, nullString(tx.Description), tx.BalanceAfter, tx.CreatedAt,
	)
	return err
}

// ListTransactions retrieves transactions for a wallet with pagination.
func (r *PostgresRepository) ListTransactions(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*domain.WalletTransaction, int64, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM wallet_transactions WHERE wallet_id = $1`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, walletID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch transactions
	query := `
		SELECT id, wallet_id, type, amount, reason, reference_id, description, balance_after, created_at
		FROM wallet_transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, walletID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []*domain.WalletTransaction
	for rows.Next() {
		tx := &domain.WalletTransaction{}
		var description sql.NullString
		if err := rows.Scan(
			&tx.ID, &tx.WalletID, &tx.Type, &tx.Amount, &tx.Reason,
			&tx.ReferenceID, &description, &tx.BalanceAfter, &tx.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		tx.Description = description.String
		transactions = append(transactions, tx)
	}

	return transactions, total, rows.Err()
}

// DebitAtomic atomically debits the wallet and records the transaction.
func (r *PostgresRepository) DebitAtomic(ctx context.Context, userID uuid.UUID, amount int, reason domain.TransactionReason, referenceID *uuid.UUID, description string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock and get wallet
	var walletID uuid.UUID
	var balance int
	err = tx.QueryRowContext(ctx,
		`SELECT id, balance FROM wallets WHERE user_id = $1 FOR UPDATE`, userID,
	).Scan(&walletID, &balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrWalletNotFound
		}
		return err
	}

	// Check sufficient balance
	if balance < amount {
		return domain.ErrInsufficientCredits
	}

	newBalance := balance - amount

	// Update balance
	_, err = tx.ExecContext(ctx,
		`UPDATE wallets SET balance = $2 WHERE id = $1`, walletID, newBalance,
	)
	if err != nil {
		return err
	}

	// Record transaction
	txID := uuid.New()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, type, amount, reason, reference_id, description, balance_after, created_at)
		 VALUES ($1, $2, 'debit', $3, $4, $5, $6, $7, NOW())`,
		txID, walletID, amount, reason, referenceID, nullString(description), newBalance,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CreditAtomic atomically credits the wallet and records the transaction.
func (r *PostgresRepository) CreditAtomic(ctx context.Context, userID uuid.UUID, amount int, reason domain.TransactionReason, referenceID *uuid.UUID, description string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock and get wallet
	var walletID uuid.UUID
	var balance int
	err = tx.QueryRowContext(ctx,
		`SELECT id, balance FROM wallets WHERE user_id = $1 FOR UPDATE`, userID,
	).Scan(&walletID, &balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrWalletNotFound
		}
		return err
	}

	newBalance := balance + amount

	// Update balance
	_, err = tx.ExecContext(ctx,
		`UPDATE wallets SET balance = $2 WHERE id = $1`, walletID, newBalance,
	)
	if err != nil {
		return err
	}

	// Record transaction
	txID := uuid.New()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, type, amount, reason, reference_id, description, balance_after, created_at)
		 VALUES ($1, $2, 'credit', $3, $4, $5, $6, $7, NOW())`,
		txID, walletID, amount, reason, referenceID, nullString(description), newBalance,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// nullString converts an empty string to sql.NullString.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	return err != nil && fmt.Sprintf("%v", err) != "" &&
		(errors.As(err, new(interface{ Code() string })) ||
			containsString(err.Error(), "duplicate key") ||
			containsString(err.Error(), "unique constraint"))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
