package store

import (
	"context"

	"github.com/luponetn/paycore/internal/db"
)

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Store interface {
	Begin(ctx context.Context) (Transaction, error)
	WithTx(tx Transaction) db.Querier
	Queries() db.Querier
}
