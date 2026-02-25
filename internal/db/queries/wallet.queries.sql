-- name: CreateWallet :one
INSERT INTO wallets (
    user_id,
    wallet_type,
    currency
) VALUES ($1,$2,$3)
RETURNING *;

-- name: GetWalletById :one
SELECT * FROM wallets WHERE id = $1 FOR UPDATE;

-- name: GetWalletsAndLockByWalletIds :many
SELECT id, balance, currency
FROM wallets
WHERE id IN ($1, $2)
ORDER BY id
FOR UPDATE;

-- name: UpdateWalletBalance :exec
UPDATE wallets
SET balance = $1
WHERE id = $2;