-- +goose Up
ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_user_id_key;
ALTER TABLE wallets ADD CONSTRAINT wallets_user_id_wallet_type_unique UNIQUE (user_id, wallet_type);

-- +goose Down
ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_user_id_wallet_type_unique;
ALTER TABLE wallets ADD CONSTRAINT wallets_user_id_key UNIQUE (user_id);
