package database

import (
	"context"
	"testing"

	"songloft/internal/models"
)

// TestPlaylistSongOrderTieBreaker 验证当主排序列存在等值行时，ASC 与 DESC 仍能产生
// 可见的不同顺序。批量加入歌单的歌曲 ps.added_at 全相同（同一秒内），若无二级排序键，
// SQLite 对等值行的返回顺序不确定，ASC/DESC 视觉无差别（songloft-org/songloft#431）。
func TestPlaylistSongOrderTieBreaker(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// 创建 5 首歌曲
	songs := make([]*models.Song, 5)
	for i := range songs {
		songs[i] = &models.Song{
			Type:     models.TypeLocal,
			Title:    string(rune('A' + i)),
			FilePath: "/music/song" + string(rune('0'+i)) + ".mp3",
		}
	}
	if err := db.SongRepository().BatchCreate(ctx, songs); err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}

	// 创建歌单并按 position 1..5 依次加入（同一事务/同一秒，ps.added_at 全相同）
	playlist := &models.Playlist{Type: "normal", Name: "test"}
	if err := db.PlaylistRepository().Create(ctx, playlist); err != nil {
		t.Fatalf("create playlist: %v", err)
	}
	psRepo := db.PlaylistSongRepository()
	for i, s := range songs {
		if err := psRepo.AddSong(ctx, playlist.ID, s.ID, i+1); err != nil {
			t.Fatalf("AddSong %d: %v", i, err)
		}
	}

	// added_at ASC
	gotAsc, err := psRepo.GetSongsFiltered(ctx, playlist.ID, PlaylistSongFilter{
		OrderBy: "added_at", Order: "asc",
	})
	if err != nil {
		t.Fatalf("GetSongsFiltered asc: %v", err)
	}

	// added_at DESC
	gotDesc, err := psRepo.GetSongsFiltered(ctx, playlist.ID, PlaylistSongFilter{
		OrderBy: "added_at", Order: "desc",
	})
	if err != nil {
		t.Fatalf("GetSongsFiltered desc: %v", err)
	}

	// 验证两次返回的歌曲数一致
	if len(gotAsc) != len(gotDesc) {
		t.Fatalf("row count mismatch: asc=%d desc=%d", len(gotAsc), len(gotDesc))
	}

	// 验证 ASC 与 DESC 顺序不同（至少不是完全相同）
	sameOrder := true
	for i := range gotAsc {
		if gotAsc[i].ID != gotDesc[i].ID {
			sameOrder = false
			break
		}
	}
	if sameOrder {
		t.Errorf("ASC and DESC produced identical ordering; tie-breaker not working")
	}

	// 验证 DESC 是 ASC 的逆序（全等值时 tie-breaker ps.position 同方向兜底，
	// DESC 应为 ASC 的严格反转）
	for i := range gotAsc {
		j := len(gotAsc) - 1 - i
		if gotAsc[i].ID != gotDesc[j].ID {
			t.Errorf("position %d: asc id=%d but desc[%d] id=%d (expected reverse)",
				i, gotAsc[i].ID, j, gotDesc[j].ID)
		}
	}
}
