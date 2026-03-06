package transfer

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/internal/store"
	"github.com/luponetn/paycore/pkg/utils"
	"github.com/shopspring/decimal"
)

type Service interface {
	CreateTransaction(ctx context.Context, userID uuid.UUID, req CreateTransactionRequest) (db.Transaction, error)
	GetTransactionByID(ctx context.Context, transactionID uuid.UUID) (db.Transaction, error)
}

type Svc struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &Svc{store: store}
}

// CreateTransaction - creates an atomic wallet-to-wallet transfer
func (s *Svc) CreateTransaction(ctx context.Context, userID uuid.UUID, req CreateTransactionRequest) (db.Transaction, error) {
	// Single timeout for entire operation
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Parse amounts and IDs outside the retry loop if possible,
	// but here we deal with strings so it's safer inside or just before.
	amountDecimal, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return db.Transaction{}, ErrInvalidAmount
	}

	if amountDecimal.LessThanOrEqual(decimal.Zero) {
		return db.Transaction{}, ErrInvalidAmount
	}

	senderID, err := uuid.Parse(req.SenderWalletID)
	if err != nil {
		return db.Transaction{}, errors.New("invalid sender wallet id")
	}

	var receiverID uuid.UUID
	if req.ReceiverAccountNo != "" {
		// Resolve wallet ID from account number
		w, err := s.store.Queries().GetWalletByAccountNo(ctx, req.ReceiverAccountNo)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return db.Transaction{}, errors.New("receiver account not found")
			}
			return db.Transaction{}, err
		}
		receiverID = w.WalletID
	} else if req.ReceiverWalletID != "" {
		receiverID, err = uuid.Parse(req.ReceiverWalletID)
		if err != nil {
			return db.Transaction{}, errors.New("invalid receiver wallet id")
		}
	} else {
		return db.Transaction{}, errors.New("either receiver wallet id or account number is required")
	}

	if senderID == receiverID {
		return db.Transaction{}, ErrSameWallet
	}

	return utils.Retry(3, 100, func() (db.Transaction, error) {
		tx, err := s.store.Begin(ctx)
		if err != nil {
			return db.Transaction{}, &utils.RetryableError{Err: err}
		}

		defer func() {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				slog.Error("failed to rollback tx", "error", rbErr)
			}
		}()

		qtx := s.store.WithTx(tx)

		// 1. Fetch wallets and lock in deterministic order
		fetchWalletsParams := db.GetWalletsAndLockByWalletIdsParams{
			ID:  senderID,
			ID2: receiverID,
		}

		wallets, err := qtx.GetWalletsAndLockByWalletIds(ctx, fetchWalletsParams)
		if err != nil {
			return db.Transaction{}, &utils.RetryableError{Err: err}
		}

		if len(wallets) != 2 {
			return db.Transaction{}, ErrWalletNotFound
		}

		var senderWallet, receiverWallet db.GetWalletsAndLockByWalletIdsRow
		if wallets[0].ID == senderID {
			senderWallet = wallets[0]
			receiverWallet = wallets[1]
		} else {
			senderWallet = wallets[1]
			receiverWallet = wallets[0]
		}

		// 2. Security Check: Authenticated user must own the sender wallet
		if !senderWallet.UserID.Valid || uuid.UUID(senderWallet.UserID.Bytes) != userID {
			return db.Transaction{}, ErrUnauthorizedWallet
		}

		// 3. Business Validation
		senderBalance := decimal.NewFromBigInt(senderWallet.Balance.Int, senderWallet.Balance.Exp)
		if senderBalance.LessThan(amountDecimal) {
			return db.Transaction{}, ErrInsufficientFunds
		}

		if senderWallet.Currency != req.Currency || receiverWallet.Currency != req.Currency {
			return db.Transaction{}, ErrCurrencyMismatch
		}

		// 4. Create Transaction record (Idempotency)
		transactionParams := db.CreateTransactionParams{
			SenderWalletID:   pgtype.UUID{Bytes: senderID, Valid: true},
			ReceiverWalletID: pgtype.UUID{Bytes: receiverID, Valid: true},
			TransactionType:  db.TransactionTypeEnum(req.TransactionType),
			Amount:           utils.DecimalToNumeric(amountDecimal),
			Description:      pgtype.Text{String: req.Description, Valid: req.Description != ""},
			Status:           db.TransactionStatusEnumPending,
			Currency:         req.Currency,
			IdempotencyKey:   req.IdempotencyKey,
		}

		createdTransaction, transactionErr := qtx.CreateTransaction(ctx, transactionParams)
		if transactionErr != nil {
			if pgErr, ok := transactionErr.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				existingTx, getErr := qtx.GetTransactionByIdempotencyKey(ctx, req.IdempotencyKey)
				if getErr != nil {
					return db.Transaction{}, &utils.RetryableError{Err: getErr}
				}
				return existingTx, nil
			}
			return db.Transaction{}, &utils.RetryableError{Err: transactionErr}
		}

		// 5. Update Balances and Ledger
		newSenderBalance := senderBalance.Sub(amountDecimal)
		receiverBalance := decimal.NewFromBigInt(receiverWallet.Balance.Int, receiverWallet.Balance.Exp)
		newReceiverBalance := receiverBalance.Add(amountDecimal)

		// Create Ledger Entries
		if _, err := qtx.CreateLedger(ctx, db.CreateLedgerParams{
			WalletID:      senderWallet.ID,
			TransactionID: createdTransaction.ID,
			Amount:        utils.DecimalToNumeric(amountDecimal),
			EntryType:     db.LedgerEntryTypeDebit,
			Currency:      senderWallet.Currency,
			BalanceBefore: senderWallet.Balance,
			BalanceAfter:  utils.DecimalToNumeric(newSenderBalance),
		}); err != nil {
			return db.Transaction{}, &utils.RetryableError{Err: err}
		}

		if _, err := qtx.CreateLedger(ctx, db.CreateLedgerParams{
			WalletID:      receiverWallet.ID,
			TransactionID: createdTransaction.ID,
			Amount:        utils.DecimalToNumeric(amountDecimal),
			EntryType:     db.LedgerEntryTypeCredit,
			Currency:      receiverWallet.Currency,
			BalanceBefore: receiverWallet.Balance,
			BalanceAfter:  utils.DecimalToNumeric(newReceiverBalance),
		}); err != nil {
			return db.Transaction{}, &utils.RetryableError{Err: err}
		}

		// Update Wallet Balances
		if err := qtx.UpdateWalletBalance(ctx, db.UpdateWalletBalanceParams{
			Balance: utils.DecimalToNumeric(newSenderBalance),
			ID:      senderWallet.ID,
		}); err != nil {
			return db.Transaction{}, &utils.RetryableError{Err: err}
		}

		if err := qtx.UpdateWalletBalance(ctx, db.UpdateWalletBalanceParams{
			Balance: utils.DecimalToNumeric(newReceiverBalance),
			ID:      receiverWallet.ID,
		}); err != nil {
			return db.Transaction{}, &utils.RetryableError{Err: err}
		}

		// 6. Complete Transaction
		if err := qtx.UpdateTransactionStatus(ctx, db.UpdateTransactionStatusParams{
			Status: db.TransactionStatusEnumCompleted,
			ID:     createdTransaction.ID,
		}); err != nil {
			return db.Transaction{}, &utils.RetryableError{Err: err}
		}

		if err := tx.Commit(ctx); err != nil {
			return db.Transaction{}, &utils.RetryableError{Err: err}
		}

		return createdTransaction, nil
	})
}

func (s *Svc) GetTransactionByID(ctx context.Context, transactionID uuid.UUID) (db.Transaction, error) {
	return s.store.Queries().GetTransactionById(ctx, transactionID)
}
