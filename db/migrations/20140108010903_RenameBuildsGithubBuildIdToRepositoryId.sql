-- +goose Up
ALTER TABLE builds RENAME COLUMN github_build_id TO repository_id;
ALTER TABLE builds_github_build_id_seq RENAME TO builds_repository_id_seq;

-- +goose Down
ALTER TABLE builds RENAME COLUMN repository_id TO github_build_id;
ALTER TABLE builds_repository_id_seq RENAME TO builds_github_build_id_seq;
