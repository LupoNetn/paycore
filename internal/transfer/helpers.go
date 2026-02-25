package transfer

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/luponetn/paycore/internal/db"
)

func CreateDoubleLedgerEntries(ctx context.Context, args db.CreateLedgerParams, qtx *db.Queries) error {
	ledgerParams := db.CreateLedgerParams{
		WalletID:      args.WalletID,
		TransactionID: args.TransactionID,
		Amount:        args.Amount,
		EntryType:     args.EntryType,
		Currency:      args.Currency,
		BalanceBefore: args.BalanceBefore,
		BalanceAfter:  args.BalanceAfter,
	}
	//create ledger entry
	_, err := qtx.CreateLedger(ctx, ledgerParams)
	if err != nil {
		slog.Error("error creating ledger entry", "error", err)
		return errors.New("error creating ledger entry")
	}

	return nil
}

func GetUserBalance(ctx context.Context, walletID uuid.UUID, qtx *db.Queries) (interface{}, error) {
	balance, err := qtx.GetUserBalance(ctx, walletID)
	if err != nil {
		slog.Error("error getting user balance", "error", err)
		return nil, errors.New("error getting user balance")
	}
	return balance, nil
}
