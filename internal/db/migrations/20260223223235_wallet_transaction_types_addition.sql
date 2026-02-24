-- +goose Up

-- Wallet type
ALTER TABLE wallets
ADD COLUMN wallet_type wallet_type_enum NOT NULL;

-- Ledger status enum conversion
ALTER TABLE ledgers ADD COLUMN status_new transaction_status_enum;
UPDATE ledgers SET status_new = status::transaction_status_enum;
ALTER TABLE ledgers DROP COLUMN status;
ALTER TABLE ledgers RENAME COLUMN status_new TO status;
ALTER TABLE ledgers ALTER COLUMN status SET NOT NULL;

-- Ledger type enum conversion
ALTER TABLE ledgers ADD COLUMN type_new transaction_type_enum;
UPDATE ledgers SET type_new = type::transaction_type_enum;
ALTER TABLE ledgers DROP COLUMN type;
ALTER TABLE ledgers RENAME COLUMN type_new TO type;
ALTER TABLE ledgers ALTER COLUMN type SET NOT NULL;

-- +goose Down

ALTER TABLE wallets DROP COLUMN IF EXISTS wallet_type;

ALTER TABLE ledgers ALTER COLUMN status DROP NOT NULL;
ALTER TABLE ledgers ALTER COLUMN type DROP NOT NULL;

ALTER TABLE ledgers DROP COLUMN IF EXISTS status;
ALTER TABLE ledgers DROP COLUMN IF EXISTS type;