-- +goose Up
ALTER TABLE users
ADD COLUMN country_code VARCHAR(6) NOT NULL;

-- +goose Down
ALTER TABLE users
DROP COLUMN country_code;
