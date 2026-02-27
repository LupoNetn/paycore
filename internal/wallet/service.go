package wallet

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/pkg/utils"
)

type Service interface {
	GetWalletService(ctx context.Context, walletID uuid.UUID) (db.Wallet, error)
	GetWalletTransactionsService(ctx context.Context, walletID uuid.UUID, limit int32, offset int32) ([]db.Transaction, error)
}

type Svc struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewService(queries *db.Queries, db *pgxpool.Pool) Service {
	return &Svc{queries: queries, db: db}
}

// implement services for all wallet operations
func (s *Svc) GetWalletService(ctx context.Context, walletID uuid.UUID) (db.Wallet, error) {
	// Single timeout for entire operation (all retries + backoff + actual work)
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	return utils.Retry(3, 100, func() (db.Wallet, error) {
		wallet, err := s.queries.GetWalletById(ctx, walletID)
		if err != nil {
			return db.Wallet{}, &utils.RetryableError{Err: err}
		}
		return wallet, nil
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
		transactions, err := s.queries.GetTransactionsByWalletId(ctx, params)
		if err != nil {
			return nil, &utils.RetryableError{Err: err}
		}
		return transactions, nil
	})
}
