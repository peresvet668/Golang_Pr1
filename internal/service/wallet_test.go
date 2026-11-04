package service

import (
	"context"
	"testing"

	"github.com/peresvet668/Golang_Pr1/internal/domain"
)

type repositoryStub struct {
	applyCalls int
}

func (r *repositoryStub) ApplyOperation(context.Context, string, domain.OperationType, int64) (int64, error) {
	r.applyCalls++
	return 0, nil
}

func (r *repositoryStub) GetBalance(context.Context, string) (int64, error) {
	return 0, nil
}

func TestApplyOperationRejectsInvalidInputBeforeDatabase(t *testing.T) {
	tests := []struct {
		name      string
		walletID  string
		operation domain.OperationType
		amount    int64
		wantErr   error
	}{
		{name: "invalid wallet id", walletID: "not-a-uuid", operation: domain.Deposit, amount: 1, wantErr: domain.ErrInvalidWalletID},
		{name: "zero amount", walletID: domain.DefaultWalletID, operation: domain.Deposit, amount: 0, wantErr: domain.ErrInvalidAmount},
		{name: "negative amount", walletID: domain.DefaultWalletID, operation: domain.Withdraw, amount: -1, wantErr: domain.ErrInvalidAmount},
		{name: "unknown operation", walletID: domain.DefaultWalletID, operation: "TRANSFER", amount: 1, wantErr: domain.ErrInvalidOperation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &repositoryStub{}
			walletService := New(repository)

			_, err := walletService.ApplyOperation(context.Background(), test.walletID, test.operation, test.amount)
			if err == nil || err.Error() != test.wantErr.Error() {
				t.Fatalf("ApplyOperation() error = %v, want %v", err, test.wantErr)
			}
			if repository.applyCalls != 0 {
				t.Fatalf("repository called %d times for invalid input", repository.applyCalls)
			}
		})
	}
}
