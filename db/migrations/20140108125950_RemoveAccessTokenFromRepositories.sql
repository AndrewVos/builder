-- +goose Up
ALTER TABLE repositories DROP COLUMN access_token;
