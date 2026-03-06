package transfer

import "errors"

var (
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrSameWallet         = errors.New("sender and receiver wallet cannot be the same")
	ErrCurrencyMismatch   = errors.New("wallet currencies must be the same")
	ErrInvalidAmount       = errors.New("amount must be greater than 0")
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrUnauthorizedWallet  = errors.New("you do not own this wallet")
	ErrTransactionFailed   = errors.New("transaction failed")
)
