-- +goose Up

CREATE TYPE wallet_type_enum AS ENUM (
    'savings',
    'fixed',
    'misc'
);

CREATE TYPE transaction_status_enum AS ENUM (
    'pending',
    'completed',
    'failed'
);

CREATE TYPE transaction_type_enum AS ENUM (
    'credit',
    'debit',
    'transfer'
);

-- +goose Down

DROP TYPE IF EXISTS transaction_type_enum;
DROP TYPE IF EXISTS transaction_status_enum;
DROP TYPE IF EXISTS wallet_type_enum;