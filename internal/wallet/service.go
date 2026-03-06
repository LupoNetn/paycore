package wallet

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/internal/store"
	"github.com/luponetn/paycore/pkg/utils"
)

type Service interface {
	GetWalletService(ctx context.Context, walletID uuid.UUID) (db.Wallet, error)
	GetWalletsByUserService(ctx context.Context, userID uuid.UUID) ([]db.Wallet, error)
	GetWalletTransactionsService(ctx context.Context, walletID uuid.UUID, limit int32, offset int32) ([]db.Transaction, error)
	ResolveAccountNumberService(ctx context.Context, accountNo string) (db.GetWalletByAccountNoRow, error)
}

type Svc struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &Svc{store: store}
}

// implement services for all wallet operations
func (s *Svc) GetWalletService(ctx context.Context, walletID uuid.UUID) (db.Wallet, error) {
	// Single timeout for entire operation (all retries + backoff + actual work)
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	return utils.Retry(3, 100, func() (db.Wallet, error) {
		wallet, err := s.store.Queries().GetWalletById(ctx, walletID)
		if err != nil {
			return db.Wallet{}, &utils.RetryableError{Err: err}
		}
		return wallet, nil
	})
}

// GetWalletsByUserService returns all wallets for a user
func (s *Svc) GetWalletsByUserService(ctx context.Context, userID uuid.UUID) ([]db.Wallet, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	return utils.Retry(3, 100, func() ([]db.Wallet, error) {
		wallets, err := s.store.Queries().GetWalletsByUserId(ctx, utils.ToPgUUID(userID))
		if err != nil {
			return nil, &utils.RetryableError{Err: err}
		}
		return wallets, nil
	})
}

// GetWalletTransactionsService returns all transactions for a wallet (paginated, default limit 50, offset 0)
func (s *Svc) GetWalletTransactionsService(ctx context.Context, walletID uuid.UUID, limit int32, offset int32) ([]db.Transaction, error) {
	// Single timeout for entire operation (all retries + backoff + actual work)
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	return utils.Retry(3, 100, func() ([]db.Transaction, error) {
		params := db.GetTransactionsByWalletIdParams{
			SenderWalletID: utils.ToPgUUID(walletID),
			Limit:          limit,
			Offset:         offset,
		}
		transactions, err := s.store.Queries().GetTransactionsByWalletId(ctx, params)
		if err != nil {
			return nil, &utils.RetryableError{Err: err}
		}
		return transactions, nil
	})
}

// ResolveAccountNumberService resolves a wallet and user by account number
func (s *Svc) ResolveAccountNumberService(ctx context.Context, accountNo string) (db.GetWalletByAccountNoRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	return utils.Retry(3, 100, func() (db.GetWalletByAccountNoRow, error) {
		row, err := s.store.Queries().GetWalletByAccountNo(ctx, accountNo)
		if err != nil {
			return db.GetWalletByAccountNoRow{}, &utils.RetryableError{Err: err}
		}
		return row, nil
	})
}
