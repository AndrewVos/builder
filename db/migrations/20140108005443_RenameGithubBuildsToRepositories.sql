-- +goose Up
ALTER TABLE github_builds RENAME TO repositories;
ALTER TABLE github_builds_id_seq RENAME TO repositories_id_seq;

-- +goose Down
ALTER TABLE repositories RENAME TO github_builds;
ALTER TABLE repositories_id_seq RENAME TO github_builds_id_seq;
