-- +goose Up
CREATE TABLE accounts(
  id SERIAL PRIMARY KEY,
  github_user_id SERIAL NOT NULL,
  access_token VARCHAR(100) NOT NULL
);

-- +goose Down
DROP TABLE accounts;
