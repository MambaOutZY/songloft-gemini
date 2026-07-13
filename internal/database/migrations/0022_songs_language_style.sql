-- +goose Up
-- +goose StatementBegin
ALTER TABLE songs ADD COLUMN language TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE songs ADD COLUMN style TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE songs DROP COLUMN language;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE songs DROP COLUMN style;
-- +goose StatementEnd
