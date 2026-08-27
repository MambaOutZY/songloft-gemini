package models

import "time"

type SongTag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	SongCount int       `json:"song_count"`
	CoverURL  string    `json:"cover_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
