-- +goose Up
-- +goose StatementBegin
-- 规范化多值歌手：artists 是歌手实体（按 normalized_key 去重），
-- song_artists 是歌曲与歌手的多对多关联（role 区分主唱 artist / 专辑歌手 album_artist）。
-- 检索与 facet 以 song_artists 为真相来源；songs.artist 列降级为显示串缓存。
CREATE TABLE artists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    normalized_key TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE song_artists (
    song_id INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    artist_id INTEGER NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK(role IN ('artist', 'album_artist')),
    position INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (song_id, artist_id, role)
);
CREATE INDEX idx_song_artists_artist ON song_artists(artist_id);
CREATE INDEX idx_song_artists_song ON song_artists(song_id);
-- +goose StatementEnd

-- 回填：把现有 songs.artist（整串）作为单行 artist 关联回填，
-- 保持当前"按整串精确匹配"的检索行为不退化。多值拆分只发生在重新扫描/刷新重读 tag 时
-- （用结构化 Artists()，不拆字符串）。normalized_key 与 Go 侧 normalizeArtistKey
-- 完全一致（lower(trim(x))），避免回填键与重扫键不一致导致重复实体。
-- +goose StatementBegin
INSERT INTO artists (name, normalized_key)
SELECT artist, lower(trim(artist))
FROM songs
WHERE artist != ''
GROUP BY lower(trim(artist))
ON CONFLICT(normalized_key) DO NOTHING;

INSERT INTO song_artists (song_id, artist_id, role, position)
SELECT s.id, a.id, 'artist', 0
FROM songs s
JOIN artists a ON a.normalized_key = lower(trim(s.artist))
WHERE s.artist != ''
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS song_artists;
DROP TABLE IF EXISTS artists;
-- +goose StatementEnd
