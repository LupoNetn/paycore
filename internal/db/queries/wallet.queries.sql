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
SELECT id, user_id, balance, currency
FROM wallets
WHERE id IN (sqlc.arg(id)::uuid, sqlc.arg(id_2)::uuid)
ORDER BY id
FOR UPDATE;

-- name: UpdateWalletBalance :exec
UPDATE wallets
SET balance = $1
WHERE id = $2;

-- name: GetWalletsByUserId :many
SELECT * FROM wallets WHERE user_id = $1 ORDER BY created_at;

-- name: GetWalletByAccountNo :one
SELECT 
    w.id as wallet_id, 
    w.user_id, 
    w.balance, 
    w.currency, 
    u.full_name, 
    u.email, 
    u.account_no 
FROM wallets w
JOIN users u ON w.user_id = u.id
WHERE u.account_no = $1
LIMIT 1;