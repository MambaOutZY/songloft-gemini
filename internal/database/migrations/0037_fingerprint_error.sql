-- +goose Up
ALTER TABLE songs ADD COLUMN fingerprint_error TEXT NOT NULL DEFAULT '';

-- +goose Down
-- SQLite does not support DROP COLUMN in older versions; recreating is out of scope.
