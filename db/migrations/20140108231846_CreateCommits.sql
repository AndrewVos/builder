-- +goose Up
CREATE TABLE commits(
  id       SERIAL PRIMARY KEY NOT NULL,
  build_id SERIAL,
  sha      VARCHAR(50),
  message  TEXT,
  url      VARCHAR(200)
);

-- +goose Down
DROP TABLE commits;
