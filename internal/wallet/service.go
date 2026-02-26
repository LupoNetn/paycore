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
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return s.queries.GetWalletById(ctx, walletID)
}

// GetWalletTransactionsService returns all transactions for a wallet (paginated, default limit 50, offset 0)
func (s *Svc) GetWalletTransactionsService(ctx context.Context, walletID uuid.UUID, limit int32, offset int32) ([]db.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	params := db.GetTransactionsByWalletIdParams{
		SenderWalletID: utils.ToPgUUID(walletID),
		Limit:          limit,
		Offset:         offset,
	}
	return s.queries.GetTransactionsByWalletId(ctx, params)
}
