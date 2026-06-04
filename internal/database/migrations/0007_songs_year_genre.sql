-- +goose Up
-- +goose StatementBegin
ALTER TABLE songs ADD COLUMN year INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE songs ADD COLUMN genre TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE songs DROP COLUMN year;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE songs DROP COLUMN genre;
-- +goose StatementEnd
