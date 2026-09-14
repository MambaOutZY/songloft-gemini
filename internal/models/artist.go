package models

import "time"

// 歌手参与角色：主唱（artist）/ 专辑歌手（album_artist）。
const (
	ArtistRoleArtist      = "artist"
	ArtistRoleAlbumArtist = "album_artist"
)

// Artist 歌手实体（规范化多值歌手关联的一方）。
type Artist struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	NormalizedKey string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

// SongArtist 一首歌的某个参与歌手及其角色与顺序。
type SongArtist struct {
	Artist   Artist `json:"artist"`
	Role     string `json:"role"`     // artist | album_artist
	Position int    `json:"position"` // 同角色内的展示顺序
}

// ArtistInput 编辑歌手时的输入项。
type ArtistInput struct {
	Name     string `json:"name" example:"周杰伦"`
	Role     string `json:"role" example:"artist"` // artist | album_artist，缺省 artist
	Position int    `json:"position,omitempty" example:"0"`
}
