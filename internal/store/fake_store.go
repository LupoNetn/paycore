package store

import (
	"bytes"
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luponetn/paycore/internal/db"
)

// FakeTx implements store.Transaction
type FakeTx struct{}

func (t *FakeTx) Commit(ctx context.Context) error {
	return nil
}

func (t *FakeTx) Rollback(ctx context.Context) error {
	return nil
}

// FakeStore implements Store interface for testing
type FakeStore struct {
	mu           sync.Mutex
	transactions map[uuid.UUID]db.Transaction
	wallets      map[uuid.UUID]db.GetWalletsAndLockByWalletIdsRow
}

// constructor
func NewFakeStore() *FakeStore {
	return &FakeStore{
		transactions: make(map[uuid.UUID]db.Transaction),
		wallets:      make(map[uuid.UUID]db.GetWalletsAndLockByWalletIdsRow),
	}
}

// Begin returns a fake transaction
func (f *FakeStore) Begin(ctx context.Context) (Transaction, error) {
	return &FakeTx{}, nil
}

// WithTx just returns the fake store (for query calls)
func (f *FakeStore) WithTx(tx Transaction) db.Querier {
	return f
}

// Queries just returns itself
func (f *FakeStore) Queries() db.Querier {
	return f
}

// ----- Fake DB operations for testing -----

func (f *FakeStore) GetWalletsAndLockByWalletIds(ctx context.Context, params db.GetWalletsAndLockByWalletIdsParams) ([]db.GetWalletsAndLockByWalletIdsRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	sender, ok1 := f.wallets[uuid.UUID(params.ID)]
	receiver, ok2 := f.wallets[uuid.UUID(params.ID_2)]

	if !ok1 || !ok2 {
		return nil, errors.New("wallet not found")
	}

	// Compare UUIDs as byte slices to determine order
	if bytes.Compare(sender.ID[:], receiver.ID[:]) < 0 {
		return []db.GetWalletsAndLockByWalletIdsRow{sender, receiver}, nil
	}
	return []db.GetWalletsAndLockByWalletIdsRow{receiver, sender}, nil
}

func (f *FakeStore) CreateTransaction(ctx context.Context, params db.CreateTransactionParams) (db.Transaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// simulate idempotency key conflict
	for _, t := range f.transactions {
		if t.IdempotencyKey == params.IdempotencyKey {
			return t, errors.New("transaction with this idempotency key already exists")
		}
	}

	// create transaction
	txID := uuid.New()
	newTx := db.Transaction{
		ID:               txID,
		SenderWalletID:   params.SenderWalletID,
		ReceiverWalletID: params.ReceiverWalletID,
		Amount:           params.Amount,
		Status:           params.Status,
		Currency:         params.Currency,
		IdempotencyKey:   params.IdempotencyKey,
	}

	f.transactions[txID] = newTx
	return newTx, nil
}

func (f *FakeStore) UpdateTransactionStatus(ctx context.Context, params db.UpdateTransactionStatusParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	tx, ok := f.transactions[params.ID]
	if !ok {
		return errors.New("transaction not found")
	}

	tx.Status = db.TransactionStatusEnumCompleted
	f.transactions[params.ID] = tx
	return nil
}

func (f *FakeStore) UpdateWalletBalance(ctx context.Context, params db.UpdateWalletBalanceParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	wallet, ok := f.wallets[params.ID]
	if !ok {
		return errors.New("wallet not found")
	}

	wallet.Balance = params.Balance
	f.wallets[params.ID] = wallet
	return nil
}

func (f *FakeStore) CreateLedger(ctx context.Context, params db.CreateLedgerParams) (db.Ledger, error) {
	// return fake ledger entry
	return db.Ledger{
		ID:            uuid.New(),
		WalletID:      params.WalletID,
		TransactionID: params.TransactionID,
		Amount:        params.Amount,
		EntryType:     params.EntryType,
		BalanceBefore: params.BalanceBefore,
		BalanceAfter:  params.BalanceAfter,
		Currency:      params.Currency,
	}, nil
}

func (f *FakeStore) GetTransactionById(ctx context.Context, id uuid.UUID) (db.Transaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	tx, ok := f.transactions[id]
	if !ok {
		return db.Transaction{}, errors.New("transaction not found")
	}
	return tx, nil
}

func (f *FakeStore) GetTransactionByIdempotencyKey(ctx context.Context, key string) (db.Transaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, tx := range f.transactions {
		if tx.IdempotencyKey == key {
			return tx, nil
		}
	}
	return db.Transaction{}, errors.New("transaction not found")
}

func (f *FakeStore) CreateOTP(ctx context.Context, arg db.CreateOTPParams) (db.Otp, error) {
	return db.Otp{}, errors.New("not implemented")
}

func (f *FakeStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return db.User{}, errors.New("not implemented")
}

func (f *FakeStore) CreateWallet(ctx context.Context, arg db.CreateWalletParams) (db.Wallet, error) {
	return db.Wallet{}, errors.New("not implemented")
}

func (f *FakeStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *FakeStore) GetPendingTransactionsByWalletId(ctx context.Context, senderWalletID pgtype.UUID) ([]db.Transaction, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeStore) GetTransactionsByWalletId(ctx context.Context, arg db.GetTransactionsByWalletIdParams) ([]db.Transaction, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeStore) GetUserBalance(ctx context.Context, walletID uuid.UUID) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeStore) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return db.User{}, errors.New("not implemented")
}

func (f *FakeStore) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	return db.User{}, errors.New("not implemented")
}

func (f *FakeStore) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	return db.User{}, errors.New("not implemented")
}

func (f *FakeStore) GetWalletById(ctx context.Context, id uuid.UUID) (db.Wallet, error) {
	return db.Wallet{}, errors.New("not implemented")
}

func (f *FakeStore) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	return db.User{}, errors.New("not implemented")
}

// helper: populate fake wallets for testing
func (f *FakeStore) AddFakeWallet(wallet db.GetWalletsAndLockByWalletIdsRow) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.wallets[wallet.ID] = wallet
}
