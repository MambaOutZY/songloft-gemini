-- +goose Up

-- 播放历史（songloft-org/songloft#333）：每个「播放上下文」独立记住最近播放过的歌曲，
-- 使用户切换歌单 / 歌手 / 专辑后能回到原上下文接着往下播。
-- context_type 复用既有的分面维度枚举（见 internal/database/filters.go 的 songFacetColumn）
-- 再加上 'playlist'，因此不引入第四份维度清单。
-- context_key 是 TEXT（playlist 存 id 字符串，分面存 value 原文），故无法对 playlists 建外键，
-- 删歌单时由 PlaylistService 显式调 ClearPlayHistoryByPlaylist 清理。

-- +goose StatementBegin
CREATE TABLE play_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    context_type TEXT NOT NULL,
    context_key TEXT NOT NULL,
    song_id INTEGER NOT NULL,
    played_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    play_count INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE,
    UNIQUE(context_type, context_key, song_id)
);
-- +goose StatementEnd

-- played_at 是秒级精度会撞，故读取与裁剪都以 id DESC 兜底 tie-break，保证顺序确定
CREATE INDEX idx_play_history_ctx ON play_history(context_type, context_key, played_at DESC, id DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_play_history_ctx;
DROP TABLE IF EXISTS play_history;
