-- +goose Up
-- +goose StatementBegin
ALTER TABLE songs ADD COLUMN fingerprint TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE songs ADD COLUMN fingerprint_duration REAL NOT NULL DEFAULT 0;
-- +goose StatementEnd

CREATE INDEX idx_songs_fingerprint ON songs(fingerprint) WHERE fingerprint != '';

-- +goose Down
DROP INDEX IF EXISTS idx_songs_fingerprint;
