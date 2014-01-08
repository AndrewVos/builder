-- +goose Up
CREATE TABLE accounts(
  id SERIAL PRIMARY KEY NOT NULL,
  access_token VARCHAR(100) NOT NULL
);

-- +goose Down
DROP TABLE accounts;
