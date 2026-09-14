package database

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"songloft/internal/database/sqlc"
	"songloft/internal/models"
)

// NormalizeArtistKey 把歌手名归一为去重键。与迁移 0038 的 SQL 回填
// (lower(trim(x))) 完全一致，确保回填键与重扫键不会错位产生重复实体。
// 不做标点/前缀剥离：那会误合并 "AC/DC" 与 "ACDC"，且与 SQL 回填不一致；
// 仅做大小写与首尾空白折叠，覆盖绝大多数去重需求。
func NormalizeArtistKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// JoinDisplayArtists 把多值歌手列表拼成 songs.artist 显示串：
// 多人用 " & " 连接，单人即该名，空列表返回空串。这是结构化多值的显示兜底，
// 取代 tag.Artist()（ID3v2.4 多值会被合并成无分隔串）。
func JoinDisplayArtists(names []string) string {
	out := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" {
			out = append(out, n)
		}
	}
	switch len(out) {
	case 0:
		return ""
	case 1:
		return out[0]
	default:
		return strings.Join(out, " & ")
	}
}

// SongArtistRepository 歌曲多值歌手关联仓储（artists + song_artists）。
type SongArtistRepository struct {
	db      sqlc.DBTX
	queries *sqlc.Queries
	sb      sq.StatementBuilderType
}

func NewSongArtistRepository(db sqlc.DBTX) *SongArtistRepository {
	return &SongArtistRepository{db: db, queries: sqlc.New(db), sb: sq.StatementBuilder.PlaceholderFormat(sq.Question)}
}

// GetBySongID 返回一首歌的全部参与歌手（按 position 排序）。
func (r *SongArtistRepository) GetBySongID(ctx context.Context, songID int64) ([]models.SongArtist, error) {
	rows, err := r.queries.GetArtistsBySongID(ctx, songID)
	if err != nil {
		return nil, err
	}
	out := make([]models.SongArtist, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.SongArtist{
			Artist: models.Artist{
				ID:            row.ArtistID,
				Name:          row.ArtistName,
				NormalizedKey: row.ArtistNormalizedKey,
				CreatedAt:     row.ArtistCreatedAt,
			},
			Role:     row.Role,
			Position: int(row.Position),
		})
	}
	return out, nil
}

// SetSongArtists 全量替换一首歌的参与歌手（删除旧关联，按 inputs 重建）。
// role 校验：仅 artist / album_artist；name 为空跳过。需在调用方事务内执行。
func (r *SongArtistRepository) SetSongArtists(ctx context.Context, songID int64, inputs []models.ArtistInput) error {
	if err := r.queries.DeleteSongArtistsBySong(ctx, songID); err != nil {
		return fmt.Errorf("delete song_artists: %w", err)
	}
	for _, in := range inputs {
		name := strings.TrimSpace(in.Name)
		if name == "" {
			continue
		}
		role := in.Role
		if role == "" {
			role = models.ArtistRoleArtist
		}
		if role != models.ArtistRoleArtist && role != models.ArtistRoleAlbumArtist {
			return fmt.Errorf("invalid artist role %q", role)
		}
		key := NormalizeArtistKey(name)
		if err := r.queries.UpsertArtist(ctx, sqlc.UpsertArtistParams{Name: name, NormalizedKey: key}); err != nil {
			return fmt.Errorf("upsert artist %q: %w", name, err)
		}
		artist, err := r.queries.GetArtistByNormalizedKey(ctx, key)
		if err != nil {
			return fmt.Errorf("lookup artist %q: %w", name, err)
		}
		if err := r.queries.LinkSongArtist(ctx, sqlc.LinkSongArtistParams{
			SongID:   songID,
			ArtistID: artist.ID,
			Role:     role,
			Position: int64(in.Position),
		}); err != nil {
			return fmt.Errorf("link song artist: %w", err)
		}
	}
	return nil
}

// SongIDsByArtistName 返回某歌手（按归一键匹配）参与的全部歌曲 id，用于过滤下钻。
// includeAlbumArtist 为 true 时把专辑歌手也算作该歌手的歌曲。
func (r *SongArtistRepository) SongIDsByArtistName(ctx context.Context, name string, includeAlbumArtist bool) ([]int64, error) {
	key := NormalizeArtistKey(name)
	if key == "" {
		return nil, nil
	}
	sb := r.sb.Select("DISTINCT sa.song_id").
		From("song_artists sa").
		Join("artists a ON a.id = sa.artist_id").
		Where(sq.Eq{"a.normalized_key": key})
	if !includeAlbumArtist {
		sb = sb.Where(sq.Eq{"sa.role": models.ArtistRoleArtist})
	}
	query, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
