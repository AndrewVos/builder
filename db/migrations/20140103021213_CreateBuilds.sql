-- +goose Up
CREATE TABLE builds(
  id              SERIAL PRIMARY KEY,
  github_build_id SERIAL NOT NULL,
  url             VARCHAR(200),
  owner           VARCHAR(100),
  repository      VARCHAR(100),
  ref             VARCHAR(100),
  sha             VARCHAR(50),
  complete        BOOLEAN,
  success         BOOLEAN,
  result          VARCHAR(30),
  github_url      VARCHAR(200)
);

-- +goose Down
DROP TABLE builds;

