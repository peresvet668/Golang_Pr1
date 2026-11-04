package domain

import "errors"

type OperationType string

const (
	Deposit  OperationType = "DEPOSIT"
	Withdraw OperationType = "WITHDRAW"

	// DefaultWalletID is created when the application starts. The assignment
	// does not define a wallet-creation endpoint, so this is the wallet used by
	// the example requests and the load test.
	DefaultWalletID = "11111111-1111-4111-8111-111111111111"
)

var (
	ErrInvalidWalletID   = errors.New("invalid wallet id")
	ErrInvalidAmount     = errors.New("amount must be greater than zero")
	ErrInvalidOperation  = errors.New("operation type must be DEPOSIT or WITHDRAW")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrBalanceOverflow   = errors.New("balance would overflow")
)
