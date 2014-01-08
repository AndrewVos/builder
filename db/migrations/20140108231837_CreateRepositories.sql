-- +goose Up
CREATE TABLE repositories(
  id SERIAL PRIMARY KEY NOT NULL,
  account_id SERIAL NOT NULL,
  owner VARCHAR(100) NOT NULL,
  repository VARCHAR(100) NOT NULL
);

-- +goose Down
DROP TABLE repositories;
