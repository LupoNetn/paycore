-- name: CreateTransaction :one
INSERT INTO transactions (sender_wallet_id, receiver_wallet_id, transaction_type, amount, description, status, currency,idempotency_key) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: GetTransactionById :one
SELECT * FROM transactions WHERE id = $1;

-- name: GetTransactionsByWalletId :many
SELECT * FROM transactions WHERE sender_wallet_id = $1 OR recipient_wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateTransactionStatus :exec
UPDATE transactions SET status = $1 WHERE id = $2;

-- name: GetPendingTransactionsByWalletId :many
SELECT * FROM transactions WHERE (sender_wallet_id = $1 OR recipient_wallet_id = $1) AND status = 'pending' ORDER BY created_at DESC;

-- name: GetTransactionByIdempotencyKey :one
SELECT * FROM transactions WHERE idempotency_key = $1;