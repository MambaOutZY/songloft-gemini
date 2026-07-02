-- +goose Up
ALTER TABLE songs ADD COLUMN file_modified_at DATETIME;

-- +goose Down
ALTER TABLE songs DROP COLUMN file_modified_at;
