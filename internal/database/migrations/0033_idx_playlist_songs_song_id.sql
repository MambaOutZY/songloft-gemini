-- +goose Up
CREATE INDEX idx_playlist_songs_song_id ON playlist_songs(song_id);

-- +goose Down
DROP INDEX idx_playlist_songs_song_id;
