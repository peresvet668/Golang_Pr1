package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/peresvet668/Golang_Pr1/internal/domain"
)

const depositQuery = `
	UPDATE wallets
	SET balance = balance + $2, updated_at = NOW()
	WHERE id = $1::uuid
	  AND balance <= 9223372036854775807 - $2
	RETURNING balance
`

const withdrawQuery = `
	UPDATE wallets
	SET balance = balance - $2, updated_at = NOW()
	WHERE id = $1::uuid
	  AND balance >= $2
	RETURNING balance
`

func (r *Repository) ApplyOperation(ctx context.Context, walletID string, operation domain.OperationType, amount int64) (int64, error) {
	query := ""
	switch operation {
	case domain.Deposit:
		query = depositQuery
	case domain.Withdraw:
		query = withdrawQuery
	default:
		return 0, domain.ErrInvalidOperation
	}

	var balance int64
	err := r.db.QueryRowContext(ctx, query, walletID, amount).Scan(&balance)
	if err == nil {
		return balance, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("update wallet balance: %w", err)
	}

	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM wallets WHERE id = $1::uuid)`, walletID).Scan(&exists); err != nil {
		return 0, fmt.Errorf("check wallet existence: %w", err)
	}
	if !exists {
		return 0, domain.ErrWalletNotFound
	}
	if operation == domain.Withdraw {
		return 0, domain.ErrInsufficientFunds
	}
	return 0, domain.ErrBalanceOverflow
}

func (r *Repository) GetBalance(ctx context.Context, walletID string) (int64, error) {
	var balance int64
	err := r.db.QueryRowContext(ctx, `SELECT balance FROM wallets WHERE id = $1::uuid`, walletID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrWalletNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get wallet balance: %w", err)
	}
	return balance, nil
}
