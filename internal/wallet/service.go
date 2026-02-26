package wallet

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/db"
)

type Service interface {
	GetWalletService(ctx context.Context, walletID uuid.UUID) (db.Wallet, error)
	GetWalletTransactionsService(ctx context.Context, walletID uuid.UUID) ([]db.Transaction, error)
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
	return s.queries.GetWalletById(ctx, walletID)
}

// GetWalletTransactionsService returns all transactions for a wallet (paginated, default limit 50, offset 0)
func (s *Svc) GetWalletTransactionsService(ctx context.Context, walletID uuid.UUID) ([]db.Transaction, error) {
	params := db.GetTransactionsByWalletIdParams{
		SenderWalletID: walletID,
		Limit:          50,
		Offset:         0,
	}
	return s.queries.GetTransactionsByWalletId(ctx, params)
}
