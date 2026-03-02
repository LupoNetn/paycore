package transfer

import (
	"context"
	"log/slog"
	"testing"
	"time"
	"sync"

	"github.com/google/uuid"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

//helper function to convert decimal.Decimal to pgtype.Numeric for testing
func decimalToPgtypeNumeric(d decimal.Decimal, err error) pgtype.Numeric {
    // Convert shopspring/decimal's internal representation to pgtype.Numeric fields
    // shopspring/decimal stores value in Int (big.Int) and Exp (int32, negative for scale)
    // pgtype.Numeric uses Int (big.Int) and Exp (int32, positive for exponent)
    // Note: The specific implementation for manual conversion is complex due to different internal representations of exponent/scale.

    // A more practical approach is to use the .Scan() method if possible, or string conversion.
    var pn pgtype.Numeric
    if err := pn.Scan(d.String()); err != nil {
        slog.Error("failed to convert decimal to pgtype.Numeric", "error", err)
		return pgtype.Numeric{} // return zero value on error
    }
    return pn
}


//happy testing
func TestCreateTransaction(t *testing.T) {
  //create a fake store and initialize the service with it
  f := store.NewFakeStore()
  svc := NewService(f)

  //create two wallets with initial balance
  senderWalletID := uuid.New()
  receiverWalletID := uuid.New()

  senderWallet := db.GetWalletsAndLockByWalletIdsRow{
	ID: senderWalletID,
	Balance: decimalToPgtypeNumeric(decimal.NewFromString("100")),
	Currency: "NGN",
  }
  receiverWallet := db.GetWalletsAndLockByWalletIdsRow{
	ID: receiverWalletID,
	Balance: decimalToPgtypeNumeric(decimal.NewFromString("50")),
	Currency: "NGN",
  }

  f.AddFakeWallet(senderWallet)
  f.AddFakeWallet(receiverWallet)

  

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  wallets,err := f.GetWalletsAndLockByWalletIds(ctx, db.GetWalletsAndLockByWalletIdsParams{
	ID: senderWalletID,
	ID_2: receiverWalletID,
  })

  require.NoError(t, err," expected to fetch wallets without error")
  require.Len(t, wallets, 2, "expected to fetch 2 wallets")

  //create a transfer request
  req := CreateTransactionRequest{
	SenderWalletID:  pgtype.UUID{
        Bytes: senderWalletID,
    },
	ReceiverWalletID:  pgtype.UUID{
        Bytes: receiverWalletID,
    },
	Amount: decimalToPgtypeNumeric(decimal.NewFromInt(20), nil),
	Description: pgtype.Text{String: "Test transfer", Valid: true},
	Currency: "NGN",
	IdempotencyKey: uuid.New().String(),
  }

  _, txErr := svc.CreateTransaction(ctx, req)
 require.NoError(t, txErr, "expected transaction to be created successfully")

}

func TestCreateTransaction_SameWalletShouldFail(t *testing.T) {

    f := store.NewFakeStore()
    svc := NewService(f)

    senderWalletID := uuid.New()

    senderWallet := db.GetWalletsAndLockByWalletIdsRow{
        ID: senderWalletID,
        Balance: decimalToPgtypeNumeric(decimal.NewFromString("100")),
        Currency: "NGN",
    }

    f.AddFakeWallet(senderWallet)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // SAME sender & receiver
    req := CreateTransactionRequest{
        SenderWalletID: pgtype.UUID{
            Bytes: senderWalletID,
            Valid: true,
        },
        ReceiverWalletID: pgtype.UUID{
            Bytes: senderWalletID,
            Valid: true,
        },
        Amount: decimalToPgtypeNumeric(decimal.NewFromInt(20), nil),
        Description: pgtype.Text{String: "Test transfer", Valid: true},
        Currency: "NGN",
        IdempotencyKey: uuid.New().String(),
    }

    _, txErr := svc.CreateTransaction(ctx, req)

    require.Error(t, txErr, "expected error when sender and receiver wallets are the same")
}

func TestCreateTransaction_SameIdempotencyKeyShouldFail(t *testing.T) {

	f := store.NewFakeStore()
	svc := NewService(f)

	//create two wallets with initial balance
	
	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()
    senderWallet := db.GetWalletsAndLockByWalletIdsRow{
		ID: senderWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("100")),
		Currency: "NGN",
	}
	receiverWallet := db.GetWalletsAndLockByWalletIdsRow{
	ID: receiverWalletID,
	Balance: decimalToPgtypeNumeric(decimal.NewFromString("50")),
	Currency: "NGN",
	}


	f.AddFakeWallet(senderWallet)
	f.AddFakeWallet(receiverWallet)

	//create a new transaction request with same idempotency key
	newReq := CreateTransactionRequest{
	SenderWalletID: pgtype.UUID{
		Bytes: uuid.New(),
		Valid: true,
	},
	ReceiverWalletID: pgtype.UUID{
		Bytes: uuid.New(),
		Valid: true,
	},
	Amount: decimalToPgtypeNumeric(decimal.NewFromInt(20), nil),
	Description: pgtype.Text{String: "Test transfer with same idempotency key", Valid: true},
	Currency: "NGN",
	IdempotencyKey: "same-idempotency-key",
	}

	//create first transaction with the idempotency key
	ctx,cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	f.CreateTransaction(ctx, db.CreateTransactionParams{
	SenderWalletID:   newReq.SenderWalletID,
	ReceiverWalletID: newReq.ReceiverWalletID,
	Amount:           newReq.Amount,
	Description:      newReq.Description,
	Status:           "pending",
	Currency:         newReq.Currency,
	IdempotencyKey:   newReq.IdempotencyKey,
	})


	//attempt to create second transaction with the same idempotency key
	_, err := svc.CreateTransaction(ctx, newReq)

	require.Error(t, err, "expected error when creating transaction with duplicate idempotency key")



}

func TestCreateTransaction_CurrencyMismatch(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)
    
	//create 2 wallets with different currencies
	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()
	senderWallet := db.GetWalletsAndLockByWalletIdsRow{
	ID: senderWalletID,
    Balance: decimalToPgtypeNumeric(decimal.NewFromString("100")),
    Currency: "USD",

}
receiverWallet := db.GetWalletsAndLockByWalletIdsRow{
	ID: receiverWalletID,
	Balance: decimalToPgtypeNumeric(decimal.NewFromString("50")),
	Currency: "EUR",
}

f.AddFakeWallet(senderWallet)
f.AddFakeWallet(receiverWallet)

	//create a transfer request with mismatched currencies
	req := CreateTransactionRequest{
		SenderWalletID: pgtype.UUID{
			Bytes: senderWalletID,
			Valid: true,
		},
		ReceiverWalletID: pgtype.UUID{
			Bytes: receiverWalletID,
			Valid: true,
		},
		Amount: decimalToPgtypeNumeric(decimal.NewFromInt(20), nil),
		Description: pgtype.Text{String: "Test transfer with different currencies", Valid: true},
		Currency: "USD",
		IdempotencyKey: uuid.New().String(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, txErr := svc.CreateTransaction(ctx, req)

	require.Error(t, txErr, "expected error when sender and receiver wallets have different currencies")
}

func TestCreateTransaction_InsufficientFundsShouldFail(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()

	senderWallet := db.GetWalletsAndLockByWalletIdsRow{
		ID: senderWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("10")),
		Currency: "NGN",
	}

	receiverWallet := db.GetWalletsAndLockByWalletIdsRow{
		ID: receiverWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("50")),
		Currency: "NGN",
	}

	f.AddFakeWallet(senderWallet)
	f.AddFakeWallet(receiverWallet)

	req := CreateTransactionRequest{
		SenderWalletID: pgtype.UUID{Bytes: senderWalletID, Valid: true},
		ReceiverWalletID: pgtype.UUID{Bytes: receiverWalletID, Valid: true},
		Amount: decimalToPgtypeNumeric(decimal.NewFromInt(100), nil),
		Description: pgtype.Text{String: "Insufficient funds test", Valid: true},
		Currency: "NGN",
		IdempotencyKey: uuid.New().String(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := svc.CreateTransaction(ctx, req)

	require.Error(t, err, "expected insufficient funds error")
}

func TestCreateTransaction_WalletNotFoundShouldFail(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	req := CreateTransactionRequest{
		SenderWalletID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		ReceiverWalletID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Amount: decimalToPgtypeNumeric(decimal.NewFromInt(20), nil),
		Description: pgtype.Text{String: "Wallet not found", Valid: true},
		Currency: "NGN",
		IdempotencyKey: uuid.New().String(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := svc.CreateTransaction(ctx, req)

	require.Error(t, err, "expected wallet not found error")
}

func TestCreateTransaction_ZeroAmountShouldFail(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()

	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{
		ID: senderWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("100")),
		Currency: "NGN",
	})

	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{
		ID: receiverWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("50")),
		Currency: "NGN",
	})

	req := CreateTransactionRequest{
		SenderWalletID: pgtype.UUID{Bytes: senderWalletID, Valid: true},
		ReceiverWalletID: pgtype.UUID{Bytes: receiverWalletID, Valid: true},
		Amount: decimalToPgtypeNumeric(decimal.Zero, nil),
		Description: pgtype.Text{String: "Zero transfer", Valid: true},
		Currency: "NGN",
		IdempotencyKey: uuid.New().String(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := svc.CreateTransaction(ctx, req)

	require.Error(t, err, "expected zero amount error")
}

func TestCreateTransaction_NegativeAmountShouldFail(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()

	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{
		ID: senderWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("100")),
		Currency: "NGN",
	})

	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{
		ID: receiverWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("50")),
		Currency: "NGN",
	})

	req := CreateTransactionRequest{
		SenderWalletID: pgtype.UUID{Bytes: senderWalletID, Valid: true},
		ReceiverWalletID: pgtype.UUID{Bytes: receiverWalletID, Valid: true},
		Amount: decimalToPgtypeNumeric(decimal.NewFromInt(-10), nil),
		Description: pgtype.Text{String: "Negative transfer", Valid: true},
		Currency: "NGN",
		IdempotencyKey: uuid.New().String(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := svc.CreateTransaction(ctx, req)

	require.Error(t, err, "expected negative amount error")
}

func TestCreateTransaction_ExactBalanceShouldSucceed(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()

	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{
		ID: senderWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("100")),
		Currency: "NGN",
	})

	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{
		ID: receiverWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("50")),
		Currency: "NGN",
	})

	req := CreateTransactionRequest{
		SenderWalletID: pgtype.UUID{Bytes: senderWalletID, Valid: true},
		ReceiverWalletID: pgtype.UUID{Bytes: receiverWalletID, Valid: true},
		Amount: decimalToPgtypeNumeric(decimal.NewFromInt(100), nil),
		Description: pgtype.Text{String: "Exact balance transfer", Valid: true},
		Currency: "NGN",
		IdempotencyKey: uuid.New().String(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := svc.CreateTransaction(ctx, req)

	require.NoError(t, err, "expected exact balance transfer to succeed")
}

func TestCreateTransaction_DoubleSpendingOnlyOneShouldWork(t *testing.T) {
	f := store.NewFakeStore()
	svc := NewService(f)

	senderWalletID := uuid.New()
	receiverWalletID := uuid.New()

	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{
		ID: senderWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("100")),
		Currency: "NGN",
	})
	f.AddFakeWallet(db.GetWalletsAndLockByWalletIdsRow{
		ID: receiverWalletID,
		Balance: decimalToPgtypeNumeric(decimal.NewFromString("50")),
		Currency: "NGN",
	})
    
    var wg sync.WaitGroup
	wg.Add(2)

	result := make(chan error, 2)

	makeTransaction := func() {
		defer wg.Done()
		req := CreateTransactionRequest{
			SenderWalletID: pgtype.UUID{Bytes: senderWalletID, Valid: true},
			ReceiverWalletID: pgtype.UUID{Bytes: receiverWalletID, Valid: true},
			Amount: decimalToPgtypeNumeric(decimal.NewFromInt(100), nil),
			Description: pgtype.Text{String: "Double spending test", Valid: true},
			Currency: "NGN",
			IdempotencyKey: uuid.New().String(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := svc.CreateTransaction(ctx, req)
		result <- err
	}
	
	go makeTransaction()
	go makeTransaction()

	wg.Wait()
	close(result)

	var successCount, errorCount int
	for err := range result {
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	require.Equal(t,1,successCount,"only one can succed")
	require.Equal(t,1,errorCount,"one should fail due to insufficient funds")
}