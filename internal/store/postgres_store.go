package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/db"
)

type PostgresStore struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewPostgresStore(db *pgxpool.Pool, queries *db.Queries) *PostgresStore {
	return &PostgresStore{
		db:      db,
		queries: queries,
	}
}

// make postgres store implement the Store interface
func (s *PostgresStore) Begin(ctx context.Context) (Transaction, error) {
	return s.db.Begin(ctx)
}

func (s *PostgresStore) WithTx(tx Transaction) db.Querier {
	return s.queries.WithTx(tx.(pgx.Tx))
}

func (s *PostgresStore) Queries() db.Querier {
	return s.queries
}
