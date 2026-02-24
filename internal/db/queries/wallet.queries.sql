-- name: CreateWallet :one
INSERT INTO wallets (
    user_id,
    wallet_type,
    currency
) VALUES ($1,$2,$3)
RETURNING *;