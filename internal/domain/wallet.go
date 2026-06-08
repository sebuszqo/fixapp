package domain

import (
	"time"

	"github.com/google/uuid"
)

// TransactionType represents credit or debit.
type TransactionType string

const (
	TransactionTypeCredit TransactionType = "credit"
	TransactionTypeDebit  TransactionType = "debit"
)

func (t TransactionType) String() string {
	return string(t)
}

// TransactionReason describes why the transaction occurred.
type TransactionReason string

const (
	ReasonInitialBonus   TransactionReason = "initial_bonus"
	ReasonAdminTopUp     TransactionReason = "admin_top_up"
	ReasonPackagePurchase TransactionReason = "package_purchase"
	ReasonLeadAccepted   TransactionReason = "lead_accepted"
	ReasonLeadRefund     TransactionReason = "lead_refund"
	ReasonClientReward   TransactionReason = "client_reward"
)

func (r TransactionReason) String() string {
	return string(r)
}

const (
	// InitialBonusCredits given to new handymen on registration.
	InitialBonusCredits = 50
	// ClientConfirmReward given to client for confirming job completion.
	ClientConfirmReward = 5
)

// Wallet represents a user's credit balance.
type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Balance   int // in credits (1 credit = 1 PLN)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewWallet creates a new wallet with zero balance.
func NewWallet(userID uuid.UUID) *Wallet {
	now := time.Now()
	return &Wallet{
		ID:        uuid.New(),
		UserID:    userID,
		Balance:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// CanAfford checks if the wallet has enough balance for a debit.
func (w *Wallet) CanAfford(amount int) bool {
	return w.Balance >= amount
}

// Credit adds credits to the wallet.
func (w *Wallet) Credit(amount int) {
	w.Balance += amount
	w.UpdatedAt = time.Now()
}

// Debit removes credits from the wallet. Returns error if insufficient.
func (w *Wallet) Debit(amount int) error {
	if !w.CanAfford(amount) {
		return ErrInsufficientCredits
	}
	w.Balance -= amount
	w.UpdatedAt = time.Now()
	return nil
}

// WalletTransaction records a single credit/debit event.
type WalletTransaction struct {
	ID          uuid.UUID
	WalletID    uuid.UUID
	Type        TransactionType
	Amount      int
	Reason      TransactionReason
	ReferenceID *uuid.UUID // optional: lead_id, job_id, etc.
	Description string
	BalanceAfter int
	CreatedAt   time.Time
}

// NewWalletTransaction creates a new transaction record.
func NewWalletTransaction(walletID uuid.UUID, txType TransactionType, amount int, reason TransactionReason, balanceAfter int) *WalletTransaction {
	return &WalletTransaction{
		ID:           uuid.New(),
		WalletID:     walletID,
		Type:         txType,
		Amount:       amount,
		Reason:       reason,
		BalanceAfter: balanceAfter,
		CreatedAt:    time.Now(),
	}
}
