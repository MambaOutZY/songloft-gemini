-- name: UpsertArtist :exec
INSERT INTO artists (name, normalized_key) VALUES (?, ?)
ON CONFLICT(normalized_key) DO UPDATE SET name = excluded.name;

-- name: GetArtistByNormalizedKey :one
SELECT id, name, normalized_key, created_at FROM artists WHERE normalized_key = ?;

-- name: GetArtistByName :one
SELECT id, name, normalized_key, created_at FROM artists WHERE name = ? COLLATE NOCASE;

-- name: DeleteSongArtistsBySong :exec
DELETE FROM song_artists WHERE song_id = ?;

-- name: LinkSongArtist :exec
INSERT INTO song_artists (song_id, artist_id, role, position) VALUES (?, ?, ?, ?)
ON CONFLICT(song_id, artist_id, role) DO UPDATE SET position = excluded.position;

-- name: GetArtistsBySongID :many
SELECT a.id AS artist_id, a.name AS artist_name, a.normalized_key AS artist_normalized_key,
       a.created_at AS artist_created_at,
       sa.role, sa.position
FROM song_artists sa
JOIN artists a ON a.id = sa.artist_id
WHERE sa.song_id = ?
ORDER BY sa.position, a.name;

-- name: CountArtistSongs :one
SELECT COUNT(*) FROM song_artists WHERE artist_id = ?;
