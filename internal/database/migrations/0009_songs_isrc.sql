-- +goose Up
ALTER TABLE songs ADD COLUMN isrc TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE songs DROP COLUMN isrc;
