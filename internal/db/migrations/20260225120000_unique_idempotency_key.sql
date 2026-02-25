-- +goose Up
ALTER TABLE transactions
ADD CONSTRAINT unique_transactions_idempotency_key UNIQUE (idempotency_key);

-- +goose Down
ALTER TABLE transactions
DROP CONSTRAINT IF EXISTS unique_transactions_idempotency_key;
