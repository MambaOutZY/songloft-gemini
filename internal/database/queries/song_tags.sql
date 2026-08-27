-- name: CreateSongTag :execlastid
INSERT INTO song_tags (name, color) VALUES (?, ?);

-- name: GetSongTagByID :one
SELECT id, name, color, created_at FROM song_tags WHERE id = ?;

-- name: GetSongTagByName :one
SELECT id, name, color, created_at FROM song_tags WHERE name = ?;

-- name: UpdateSongTag :exec
UPDATE song_tags SET name = ?, color = ? WHERE id = ?;

-- name: DeleteSongTag :exec
DELETE FROM song_tags WHERE id = ?;

-- name: LinkSongTag :exec
INSERT OR IGNORE INTO song_tag_links (song_id, tag_id) VALUES (?, ?);

-- name: UnlinkSongTag :exec
DELETE FROM song_tag_links WHERE song_id = ? AND tag_id = ?;

-- name: UnlinkAllBySong :exec
DELETE FROM song_tag_links WHERE song_id = ?;

-- name: GetTagsBySongID :many
SELECT t.id, t.name, t.color, t.created_at
FROM song_tags t
JOIN song_tag_links l ON t.id = l.tag_id
WHERE l.song_id = ?
ORDER BY t.name;

-- name: CountSongsByTag :one
SELECT COUNT(*) FROM song_tag_links WHERE tag_id = ?;
