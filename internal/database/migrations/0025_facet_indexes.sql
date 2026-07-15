-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_songs_genre ON songs(genre);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_songs_album ON songs(album);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_songs_language ON songs(language);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_songs_style ON songs(style);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_songs_year ON songs(year);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_songs_genre;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_songs_album;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_songs_language;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_songs_style;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_songs_year;
-- +goose StatementEnd
