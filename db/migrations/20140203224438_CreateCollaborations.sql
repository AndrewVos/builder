-- +goose Up
CREATE TABLE collaborations(
  account_id SERIAL NOT NULL,
  repository_id SERIAL NOT NULL
);

-- +goose Down
DROP TABLE collaborations;
