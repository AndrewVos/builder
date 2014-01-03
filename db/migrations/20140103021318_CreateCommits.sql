-- +goose Up
CREATE TABLE commits(
  build_id VARCHAR(100),
  sha      VARCHAR(50),
  message  TEXT,
  url      VARCHAR(200)
);

-- +goose Down
DROP TABLE commits;
