package transfer

import (
	"context"

	"github.com/luponetn/paycore/internal/db"
)

type Service interface {
	CreateTransaction(ctx context.Context) (bool, error)
}

type Svc struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) Service {
	return &Svc{queries: queries}
}

// implement Services
func (s *Svc) CreateTransaction(ctx context.Context) (bool, error) {
	return true, nil
}
