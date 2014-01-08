-- +goose Up
ALTER TABLE accounts RENAME COLUMN github_user_id TO id;
ALTER TABLE accounts_github_user_id_seq RENAME TO accounts_id_seq;
