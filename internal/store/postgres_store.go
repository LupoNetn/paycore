package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/db"
)

type PostgresStore struct {
	db *pgxpool.Pool
	queries *db.Queries
}

func NewPostgresStore(db *pgxpool.Pool, queries *db.Queries) *PostgresStore {
	return &PostgresStore{
		db: db,
		queries: queries,
	}
}

//make postgres store implement the Store interface
func (s *PostgresStore) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.db.Begin(ctx)
}

func (s *PostgresStore) WithTx(tx pgx.Tx) *db.Queries {
	return s.queries.WithTx(tx)
}

func (s *PostgresStore) Queries() *db.Queries {
	return s.queries
}