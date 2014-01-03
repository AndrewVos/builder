-- +goose Up
CREATE TABLE builds(
  build_id   VARCHAR(100) PRIMARY KEY,
  url        VARCHAR(200),
  owner      VARCHAR(100),
  repository VARCHAR(100),
  ref        VARCHAR(100),
  sha        VARCHAR(50),
  complete   BOOLEAN,
  success    BOOLEAN,
  result     VARCHAR(30),
  github_url VARCHAR(200)
);

-- +goose Down
DROP TABLE builds;

