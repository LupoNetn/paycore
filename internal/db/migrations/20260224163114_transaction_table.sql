-- +goose Up
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_wallet_id UUID,
    receiver_wallet_id UUID,
    transaction_type transaction_type_enum NOT NULL,
    amount NUMERIC(18,2) NOT NULL,
    description TEXT,
    status transaction_status_enum NOT NULL,
    currency VARCHAR(3) NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (sender_wallet_id) REFERENCES wallets(id),
    FOREIGN KEY (receiver_wallet_id) REFERENCES wallets(id)
);

-- +goose Down
DROP TABLE IF EXISTS transactions;
