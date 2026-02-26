package transfer

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luponetn/paycore/internal/db"
)

type CreateTransactionRequest struct {
    SenderWalletID   pgtype.UUID           `json:"sender_wallet_id" binding:"required"`
    ReceiverWalletID pgtype.UUID           `json:"receiver_wallet_id" binding:"required"`
    TransactionType  db.TransactionTypeEnum   `json:"transaction_type" binding:"required"`
    Amount           pgtype.Numeric        `json:"amount" binding:"required"`
    Description      pgtype.Text           `json:"description"`
    Status           db.TransactionStatusEnum `json:"status"`
    Currency         string                `json:"currency" binding:"required"`
    IdempotencyKey   string                `json:"idempotency_key" binding:"required"`
}

