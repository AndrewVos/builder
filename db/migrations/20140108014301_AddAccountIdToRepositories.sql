-- +goose Up
ALTER TABLE repositories ADD COLUMN account_id SERIAL;

-- +goose Down
ALTER TABLE repositories DROP COLUMN account_id;
