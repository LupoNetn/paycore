package transfer

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/pkg/utils"
	"github.com/shopspring/decimal"
)

type Service interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (db.Transaction, error)
	GetTransactionByID(ctx context.Context, transactionID uuid.UUID) (db.Transaction, error)
}

type Svc struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewService(queries *db.Queries, db *pgxpool.Pool) Service {
	return &Svc{queries: queries, db: db}
}

// implement Services

// 1. CreateTransaction - creates a transaction record with pending status, creates double ledger entries for both sender and receiver, updates wallet balance for both sender and receiver, and finally updates transaction record with completed status. All of these operations are done in a single db transaction to ensure atomicity. Also implements idempotency key to prevent duplicate transactions.
func (s *Svc) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (db.Transaction, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		slog.Error("could not initialized db transaction for create transaction service", "error", err)
		return db.Transaction{}, err
	}

	// Ensure rollback on error/panic. Ignore ErrTxClosed which is returned when
	// the tx has already been committed.
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			slog.Error("failed to rollback tx", "error", rbErr)
		}
	}()

	qtx := s.queries.WithTx(tx)

	// Note: idempotency is handled atomically by relying on a DB unique
	// constraint on `idempotency_key` and handling duplicate-key errors below.

	//check which user has the smallest id and lock first
	fetchWalletsParams := db.GetWalletsAndLockByWalletIdsParams{
		ID:   req.SenderWalletID.Bytes,
		ID_2: req.ReceiverWalletID.Bytes,
	}
	var senderWallet, receiverWallet db.GetWalletsAndLockByWalletIdsRow
	wallets, err := qtx.GetWalletsAndLockByWalletIds(ctx, fetchWalletsParams)
	if err != nil {
		slog.Error("error fetching wallets and acquiring lock for transfer transaction", "error", err)
		return db.Transaction{}, err
	}

	//get who owns the wallet
	if wallets[0].ID == req.SenderWalletID.Bytes {
		senderWallet = wallets[0]
		receiverWallet = wallets[1]
	} else {
		senderWallet = wallets[1]
		receiverWallet = wallets[0]
	}

	//confirm sender balance is accurate
	//convert pgType.Numeric struct to decimal for comparison
	senderBalanceDecimal := decimal.NewFromBigInt(senderWallet.Balance.Int, senderWallet.Balance.Exp)
	reqAmountDecimal := decimal.NewFromBigInt(req.Amount.Int, req.Amount.Exp)

	if senderBalanceDecimal.LessThan(reqAmountDecimal) {
		slog.Error("insufficient balance, fix up", "error", err)
		return db.Transaction{}, errors.New("Insufficient funds")
	}

	//check for same wallet transfer
	if senderWallet.ID == receiverWallet.ID {
		slog.Error("sender and receiver wallet cannot be the same", "error", err)
		return db.Transaction{}, err
	}

	//check for currency match
	if senderWallet.Currency != receiverWallet.Currency {
		slog.Error("sender and receiver wallet must have the same currency", "error", err)
		return db.Transaction{}, err
	}

	//create transaction record with pending status
	transactionParams := db.CreateTransactionParams{
		SenderWalletID:   req.SenderWalletID,
		ReceiverWalletID: req.ReceiverWalletID,
		TransactionType:  req.TransactionType,
		Amount:           req.Amount,
		Description:      req.Description,
		Status:           "pending",
		Currency:         senderWallet.Currency,
		IdempotencyKey:   req.IdempotencyKey,
	}

	createdTransaction, transactionErr := qtx.CreateTransaction(ctx, transactionParams)
	if transactionErr != nil {
		// If a concurrent request already created a transaction with this
		// idempotency key, the DB should raise unique_violation (23505).
		// Fetch and return the existing transaction instead of failing.
		if pgErr, ok := transactionErr.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			existingTx, getErr := qtx.GetTransactionByIdempotencyKey(ctx, req.IdempotencyKey)
			if getErr != nil {
				slog.Error("failed to fetch existing transaction after unique violation", "error", getErr)
				return db.Transaction{}, getErr
			}
			return existingTx, nil
		}

		slog.Error("error occurred when trying to create transaction record with pending status", "error", transactionErr)
		return db.Transaction{}, transactionErr
	}

	//get sender and receiver balance before transfer for ledger entry
	senderBalanceBefore := senderWallet.Balance
	receiverBalanceBefore := receiverWallet.Balance

	//get sender and receiver balance after transfer for ledger entry
	receiverBalanceDecimal := decimal.NewFromBigInt(receiverWallet.Balance.Int, receiverWallet.Balance.Exp)
	amountDecimal := decimal.NewFromBigInt(req.Amount.Int, req.Amount.Exp)

	newSenderBalance := senderBalanceDecimal.Sub(amountDecimal)
	newReceiverBalance := receiverBalanceDecimal.Add(amountDecimal)

	//create double-entry ledger for both debit and credit

	//1. create debit ledger entry for sender
	debitLedgerParams := db.CreateLedgerParams{
		WalletID:      senderWallet.ID,
		TransactionID: createdTransaction.ID,
		Amount:        req.Amount,
		EntryType:     db.LedgerEntryTypeDebit,
		Currency:      senderWallet.Currency,
		BalanceBefore: senderBalanceBefore,
		BalanceAfter:  utils.DecimalToNumeric(newSenderBalance),
	}

	//2. create credit ledger entry for receiver
	creditLedgerParams := db.CreateLedgerParams{
		WalletID:      receiverWallet.ID,
		TransactionID: createdTransaction.ID,
		Amount:        req.Amount,
		EntryType:     db.LedgerEntryTypeCredit,
		Currency:      receiverWallet.Currency,
		BalanceBefore: receiverBalanceBefore,
		BalanceAfter:  utils.DecimalToNumeric(newReceiverBalance),
	}

	if _, err := qtx.CreateLedger(ctx, debitLedgerParams); err != nil {
		slog.Error("error occurred when trying to create debit ledger entry", "error", err)
		return db.Transaction{}, err
	}

	if _, err := qtx.CreateLedger(ctx, creditLedgerParams); err != nil {
		slog.Error("error occurred when trying to create credit ledger entry", "error", err)
		return db.Transaction{}, err
	}

	//update wallet balance for sender and receiver
	//1. update sender wallet balance
	UpdateSenderBalanceErr := qtx.UpdateWalletBalance(ctx, db.UpdateWalletBalanceParams{
		Balance: utils.DecimalToNumeric(newSenderBalance),
		ID:      senderWallet.ID,
	})

	if UpdateSenderBalanceErr != nil {
		slog.Error("error occurred when trying to update sender wallet balance", "error", UpdateSenderBalanceErr)
		return db.Transaction{}, UpdateSenderBalanceErr
	}

	//2. update receiver wallet balance
	UpdateReceiverBalanceErr := qtx.UpdateWalletBalance(ctx, db.UpdateWalletBalanceParams{
		Balance: utils.DecimalToNumeric(newReceiverBalance),
		ID:      receiverWallet.ID,
	})

	if UpdateReceiverBalanceErr != nil {
		slog.Error("error occurred when trying to update receiver wallet balance", "error", UpdateReceiverBalanceErr)
		return db.Transaction{}, UpdateReceiverBalanceErr
	}

	//update transaction record with success status
	updateTransactionStatusErr := qtx.UpdateTransactionStatus(ctx, db.UpdateTransactionStatusParams{
		Status: db.TransactionStatusEnumCompleted,
		ID:     createdTransaction.ID,
	})
	if updateTransactionStatusErr != nil {
		slog.Error("error occurred when trying to update transaction status to success", "error", updateTransactionStatusErr)
		return db.Transaction{}, updateTransactionStatusErr
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("error occurred when trying to commit db transaction", "error", err)
		return db.Transaction{}, err
	}

	return createdTransaction, nil
}

func (s *Svc) GetTransactionByID(ctx context.Context, transactionID uuid.UUID) (db.Transaction, error) {
	return s.queries.GetTransactionById(ctx, transactionID)
}