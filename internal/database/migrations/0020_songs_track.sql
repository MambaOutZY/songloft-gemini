-- +goose Up
ALTER TABLE songs ADD COLUMN track TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE songs DROP COLUMN track;
