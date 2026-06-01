-- +goose Up
-- +goose StatementBegin
-- 为 songs.file_path 建索引，加速「按文件夹选歌」的 LIKE 'prefix%' 查询。
-- SQLite 默认 BINARY collation 下，前缀 LIKE 可走该索引。
CREATE INDEX IF NOT EXISTS idx_songs_file_path ON songs(file_path);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_songs_file_path;
-- +goose StatementEnd
