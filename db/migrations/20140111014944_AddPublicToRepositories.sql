-- +goose Up
ALTER TABLE repositories ADD COLUMN public BOOLEAN;

-- +goose Down
ALTER TABLE repositories DROP COLUMN public;
