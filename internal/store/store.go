package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/luponetn/paycore/internal/db"
)

type Store interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	WithTx(tx pgx.Tx) *db.Queries
	Queries() *db.Queries
}