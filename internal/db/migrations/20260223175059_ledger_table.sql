-- +goose Up
CREATE TABLE IF NOT EXISTS ledgers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_wallet_id UUID REFERENCES wallets(id) ON DELETE CASCADE,
    receiver_wallet_id UUID REFERENCES wallets(id) ON DELETE CASCADE,
    amount NUMERIC(18,2) NOT NULL,
    type TEXT NOT NULL,        -- 'debit', 'credit', 'refund'
    status TEXT NOT NULL CHECK (status IN ('pending','completed','failed')),
    reference TEXT,            -- optional human-readable reference
    reference_id UUID,         -- optional link to another ledger entry
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS ledgers;