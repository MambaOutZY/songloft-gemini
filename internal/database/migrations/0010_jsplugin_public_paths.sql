-- +goose Up
ALTER TABLE js_plugins ADD COLUMN public_paths TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE js_plugins DROP COLUMN public_paths;
