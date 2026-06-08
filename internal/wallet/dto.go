package wallet

import (
	"time"

	"fixapp/internal/domain"
)

// ===== Response DTOs =====

// WalletResponse is the public representation of a wallet.
// @Description Wallet balance information
type WalletResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string    `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Balance   int       `json:"balance" example:"87"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-10T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-15T14:30:00Z"`
}

// TransactionResponse is the public representation of a wallet transaction.
// @Description Wallet transaction record
type TransactionResponse struct {
	ID           string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type         string    `json:"type" example:"debit"`
	Amount       int       `json:"amount" example:"22"`
	Reason       string    `json:"reason" example:"lead_accepted"`
	ReferenceID  *string   `json:"reference_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	Description  string    `json:"description,omitempty" example:"Lead accepted: Hydraulik - Krowodrza"`
	BalanceAfter int       `json:"balance_after" example:"65"`
	CreatedAt    time.Time `json:"created_at" example:"2025-01-15T14:30:00Z"`
}

// TransactionListResponse is the paginated list of transactions.
// @Description Paginated transaction list
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total" example:"25"`
	Limit        int                   `json:"limit" example:"20"`
	Offset       int                   `json:"offset" example:"0"`
	HasMore      bool                  `json:"has_more" example:"true"`
}

// ===== Request DTOs =====

// AdminTopUpRequest is the payload for admin credit top-up.
// @Description Admin top-up request
type AdminTopUpRequest struct {
	UserID      string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Amount      int    `json:"amount" example:"100"`
	Description string `json:"description,omitempty" example:"Package purchase: Starter 100 PLN"`
}

// Validate checks if the request is valid.
func (r *AdminTopUpRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.UserID == "" {
		errs["user_id"] = "user_id is required"
	}
	if r.Amount <= 0 {
		errs["amount"] = "amount must be a positive integer"
	}
	if r.Amount > 10000 {
		errs["amount"] = "amount cannot exceed 10000 credits"
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ===== Mappers =====

// ToWalletResponse converts a domain wallet to API response.
func ToWalletResponse(w *domain.Wallet) WalletResponse {
	return WalletResponse{
		ID:        w.ID.String(),
		UserID:    w.UserID.String(),
		Balance:   w.Balance,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

// ToTransactionResponse converts a domain transaction to API response.
func ToTransactionResponse(tx *domain.WalletTransaction) TransactionResponse {
	resp := TransactionResponse{
		ID:           tx.ID.String(),
		Type:         tx.Type.String(),
		Amount:       tx.Amount,
		Reason:       tx.Reason.String(),
		Description:  tx.Description,
		BalanceAfter: tx.BalanceAfter,
		CreatedAt:    tx.CreatedAt,
	}
	if tx.ReferenceID != nil {
		s := tx.ReferenceID.String()
		resp.ReferenceID = &s
	}
	return resp
}

// ToTransactionListResponse converts a list of transactions to paginated response.
func ToTransactionListResponse(txs []*domain.WalletTransaction, total int64, limit, offset int) TransactionListResponse {
	responses := make([]TransactionResponse, len(txs))
	for i, tx := range txs {
		responses[i] = ToTransactionResponse(tx)
	}
	return TransactionListResponse{
		Transactions: responses,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
		HasMore:      int64(offset+len(txs)) < total,
	}
}
