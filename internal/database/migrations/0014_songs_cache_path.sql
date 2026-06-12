-- +goose Up
ALTER TABLE songs ADD COLUMN cache_path TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE songs DROP COLUMN cache_path;
