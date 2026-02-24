-- +goose Up

DROP TABLE IF EXISTS ledgers;

CREATE TYPE ledger_entry_type AS ENUM ('debit', 'credit');

CREATE TABLE ledgers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
  transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
  amount NUMERIC(18,2) NOT NULL,
  entry_type ledger_entry_type NOT NULL,
  currency VARCHAR(6) NOT NULL,
  balance_before NUMERIC(18,2) NOT NULL,
  balance_after NUMERIC(18,2) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE wallets
ADD COLUMN IF NOT EXISTS currency VARCHAR(6) NOT NULL DEFAULT 'USD';

-- +goose Down

DROP TABLE IF EXISTS ledgers;

ALTER TABLE wallets
DROP COLUMN IF EXISTS currency;