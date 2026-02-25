package transfer

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luponetn/paycore/internal/db"
)

type CreateTransactionRequest struct {
    SenderWalletID   pgtype.UUID           `json:"sender_wallet_id"`
    ReceiverWalletID pgtype.UUID           `json:"receiver_wallet_id"`
    TransactionType  db.TransactionTypeEnum   `json:"transaction_type"`
    Amount           pgtype.Numeric        `json:"amount"`
    Description      pgtype.Text           `json:"description"`
    Status           db.TransactionStatusEnum `json:"status"`
    Currency         string                `json:"currency"`
    IdempotencyKey   string                `json:"idempotency_key"`
}

