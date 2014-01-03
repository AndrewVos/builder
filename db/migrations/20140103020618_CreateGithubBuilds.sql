-- +goose Up
CREATE TABLE github_builds(
  id SERIAL PRIMARY KEY,
  access_token VARCHAR(100) NOT NULL,
  owner VARCHAR(100) NOT NULL,
  repository VARCHAR(100) NOT NULL
);

-- +goose Down
DROP TABLE github_builds;

