-- +goose Up
-- +goose StatementBegin
ALTER TABLE playlists ADD COLUMN sort_by TEXT NOT NULL DEFAULT 'position';
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE playlists ADD COLUMN sort_order TEXT NOT NULL DEFAULT 'asc';
-- +goose StatementEnd

-- +goose Down
-- SQLite 不支持 DROP COLUMN（3.35.0 之前），但 goose down 场景极少，此处仅做标记。
