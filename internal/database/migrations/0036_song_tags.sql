-- +goose Up
-- +goose StatementBegin
CREATE TABLE song_tags (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    color TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE song_tag_links (
    song_id INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    tag_id  INTEGER NOT NULL REFERENCES song_tags(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (song_id, tag_id)
);
CREATE INDEX idx_song_tag_links_tag ON song_tag_links(tag_id);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS song_tag_links;
DROP TABLE IF EXISTS song_tags;
