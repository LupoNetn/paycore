package transfer

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/internal/store"
	"github.com/stretchr/testify/require"
)

func TestCreateTransaction(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	userID := uuid.New()
	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()

	senderWallet := db.GetWalletsAndLockByWalletIdsRow{
		ID:       senderWalletID,
		UserID:   pgtype.UUID{Bytes: userID, Valid: true},
		Balance:  pgtype.Numeric{},
		Currency: "NGN",
	}
	_ = senderWallet.Balance.Scan("100")

	receiverWallet := db.GetWalletsAndLockByWalletIdsRow{
		ID:       receiverWalletID,
		UserID:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Balance:  pgtype.Numeric{},
		Currency: "NGN",
	}
	_ = receiverWallet.Balance.Scan("50")

	f.AddFakeWallet(senderWallet)
	f.AddFakeWallet(receiverWallet)

	req := CreateTransactionRequest{
		SenderWalletID:   senderWalletID.String(),
		ReceiverWalletID: receiverWalletID.String(),
		TransactionType:  "transfer",
		Amount:           "20.00",
		Description:      "Test transfer",
		Currency:         "NGN",
		IdempotencyKey:   uuid.New().String(),
	}

	ctx := context.Background()
	_, err := svc.CreateTransaction(ctx, userID, req)
	require.NoError(t, err)

	// Verify balance update in fake store
	wallets, _ := f.GetWalletsAndLockByWalletIds(ctx, db.GetWalletsAndLockByWalletIdsParams{
		ID:  senderWalletID,
		ID2: receiverWalletID,
	})

	for _, w := range wallets {
		if w.ID == senderWalletID {
			require.Equal(t, "80.00", w.Balance.Int.String()) // This depends on how pgtype.Numeric works in fake store
		}
	}
}

func TestCreateTransaction_Unauthorized(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	userID := uuid.New()
	wrongUserID := uuid.New()
	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()

	senderWallet := db.GetWalletsAndLockByWalletIdsRow{
		ID:       senderWalletID,
		UserID:   pgtype.UUID{Bytes: wrongUserID, Valid: true},
		Currency: "NGN",
	}
	_ = senderWallet.Balance.Scan("100")

	f.AddFakeWallet(senderWallet)
	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{ID: receiverWalletID, Currency: "NGN"})

	req := CreateTransactionRequest{
		SenderWalletID:   senderWalletID.String(),
		ReceiverWalletID: receiverWalletID.String(),
		TransactionType:  "transfer",
		Amount:           "20.00",
		Currency:         "NGN",
		IdempotencyKey:   uuid.New().String(),
	}

	_, err := svc.CreateTransaction(context.Background(), userID, req)
	require.ErrorIs(t, err, ErrUnauthorizedWallet)
}

func TestCreateTransaction_InsufficientFunds(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	userID := uuid.New()
	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()

	senderWallet := db.GetWalletsAndLockByWalletIdsRow{
		ID:       senderWalletID,
		UserID:   pgtype.UUID{Bytes: userID, Valid: true},
		Currency: "NGN",
	}
	_ = senderWallet.Balance.Scan("10")

	f.AddFakeWallet(senderWallet)
	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{ID: receiverWalletID, Currency: "NGN"})

	req := CreateTransactionRequest{
		SenderWalletID:   senderWalletID.String(),
		ReceiverWalletID: receiverWalletID.String(),
		TransactionType:  "transfer",
		Amount:           "100.00",
		Currency:         "NGN",
		IdempotencyKey:   uuid.New().String(),
	}

	_, err := svc.CreateTransaction(context.Background(), userID, req)
	require.ErrorIs(t, err, ErrInsufficientFunds)
}

func TestCreateTransaction_Concurrency(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	userID := uuid.New()
	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()

	senderWallet := db.GetWalletsAndLockByWalletIdsRow{
		ID:       senderWalletID,
		UserID:   pgtype.UUID{Bytes: userID, Valid: true},
		Currency: "NGN",
	}
	_ = senderWallet.Balance.Scan("100")

	f.AddFakeWallet(senderWallet)
	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{ID: receiverWalletID, Currency: "NGN", UserID: pgtype.UUID{Bytes: uuid.New(), Valid: true}})

	const numReqs = 2
	var wg sync.WaitGroup
	wg.Add(numReqs)

	errs := make(chan error, numReqs)

	for i := 0; i < numReqs; i++ {
		go func() {
			defer wg.Done()
			req := CreateTransactionRequest{
				SenderWalletID:   senderWalletID.String(),
				ReceiverWalletID: receiverWalletID.String(),
				TransactionType:  "transfer",
				Amount:           "100.00",
				Currency:         "NGN",
				IdempotencyKey:   uuid.New().String(),
			}
			_, err := svc.CreateTransaction(context.Background(), userID, req)
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	var success, insufficient int
	for err := range errs {
		if err == nil {
			success++
		} else if errors.Is(err, ErrInsufficientFunds) {
			insufficient++
		}
	}

	require.Equal(t, 1, success)
	require.Equal(t, 1, insufficient)
}
