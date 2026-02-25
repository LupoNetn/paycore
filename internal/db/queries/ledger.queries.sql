-- name: CreateLedger :one
INSERT INTO ledgers (wallet_id,transaction_id,amount,entry_type,currency,balance_before,balance_after) 
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING *;

-- name: GetUserBalance :one
SELECT COALESCE(SUM(
  CASE 
    WHEN entry_type = 'credit' THEN amount
    WHEN entry_type = 'debit' THEN -amount
  END
), 0)
FROM ledgers
WHERE wallet_id = $1;