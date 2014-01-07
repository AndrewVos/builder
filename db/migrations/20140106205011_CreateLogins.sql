-- +goose Up
CREATE TABLE logins(
  id SERIAL PRIMARY KEY,
  account_id SERIAL NOT NULL,
  token VARCHAR(100) NOT NULL
);

-- +goose Down
DROP TABLE logins;
