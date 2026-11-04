package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/peresvet668/Golang_Pr1/internal/domain"
)

type WalletRepository interface {
	ApplyOperation(ctx context.Context, walletID string, operation domain.OperationType, amount int64) (int64, error)
	GetBalance(ctx context.Context, walletID string) (int64, error)
}

type WalletService struct {
	repository WalletRepository
}

func New(repository WalletRepository) *WalletService {
	return &WalletService{repository: repository}
}

func (s *WalletService) ApplyOperation(ctx context.Context, walletID string, operation domain.OperationType, amount int64) (int64, error) {
	parsedID, err := uuid.Parse(walletID)
	if err != nil {
		return 0, domain.ErrInvalidWalletID
	}
	if amount <= 0 {
		return 0, domain.ErrInvalidAmount
	}
	if operation != domain.Deposit && operation != domain.Withdraw {
		return 0, domain.ErrInvalidOperation
	}

	return s.repository.ApplyOperation(ctx, parsedID.String(), operation, amount)
}

func (s *WalletService) GetBalance(ctx context.Context, walletID string) (int64, error) {
	parsedID, err := uuid.Parse(walletID)
	if err != nil {
		return 0, domain.ErrInvalidWalletID
	}

	return s.repository.GetBalance(ctx, parsedID.String())
}
