package transfer

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/db"
	"github.com/shopspring/decimal"
)

type Service interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (db.Transaction, error)
}

type Svc struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewService(queries *db.Queries, db *pgxpool.Pool) Service {
	return &Svc{queries: queries, db: db}
}

// implement Services
func (s *Svc) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (db.Transaction, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		slog.Error("could not initialized db transaction for create transaction service", "error", err)
		return db.Transaction{}, nil
	}

	qtx := s.queries.WithTx(tx)

	//check for idempotency key, if there is return cached result
	idempotencyKeyResult, err := qtx.GetTransactionByIdempotencyKey(ctx, req.IdempotencyKey)
	if err == nil {
		return idempotencyKeyResult, nil
	} else if err != nil {
		slog.Error("error occurred when trying to get transfer with the same idempotency key", "error", err)
		return db.Transaction{}, err
	}

	//check which user has the smallest id and lock first
	fetchWalletsParams := db.GetWalletsAndLockByWalletIdsParams{
		ID:   req.SenderWalletID.Bytes,
		ID_2: req.ReceiverWalletID.Bytes,
	}
	var senderWallet, receiverWallet db.GetWalletsAndLockByWalletIdsRow
	wallets, err := qtx.GetWalletsAndLockByWalletIds(ctx, fetchWalletsParams)

	//get who owns the wallet
	if wallets[0].ID == req.SenderWalletID.Bytes {
		senderWallet = wallets[0]
		receiverWallet = wallets[1]
	} else {
		senderWallet = wallets[1]
		receiverWallet = wallets[0]
	}

	//confirm sender balance is accurate
	if senderWallet.Balance < req.Amount {
		slog.Error("insufficient balance, fix up", "error", err)
	}

	//TODO:
	/*
		1. confirm sender ballance
		2. update ledger
		3. recalculate balances for sender and receiver wallet based on ledger
		4. create transaction and send success
		-- if error rollback, else commit
	*/

	return db.Transaction{}, nil
}
