package database

import (
	"context"
	"testing"

	"songloft/internal/models"
)

// TestListSongsNeedingMetadata 批量元数据刷新只处理本地有文件的歌曲：
// 本地歌曲（file_path 非空、非 CUE 虚拟轨）与已落地缓存的网络歌曲（cache_path 非空）。
// 未缓存的网络歌曲不入选——刷新已改为纯本地提取，远程解析会整批挂死在插件超时上。
func TestListSongsNeedingMetadata(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := db.SongRepository()
	ctx := context.Background()

	create := func(s *models.Song) *models.Song {
		t.Helper()
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("create %s: %v", s.Title, err)
		}
		return s
	}

	localMissing := create(&models.Song{Type: models.TypeLocal, Title: "local-missing", FilePath: "/m/a.mp3"})
	create(&models.Song{Type: models.TypeLocal, Title: "local-complete", FilePath: "/m/b.mp3",
		Duration: 200, BitRate: 320, SampleRate: 44100, Format: "mp3"})
	create(&models.Song{Type: models.TypeLocal, Title: "cue-track", FilePath: "/m/image.flac",
		CueSourcePath: "/m/image.cue"})

	remoteCached := create(&models.Song{Type: models.TypeRemote, Title: "remote-cached", URL: "https://x/1"})
	if err := repo.UpdateCachePath(ctx, remoteCached.ID, "/cache/1.mp3"); err != nil {
		t.Fatalf("update cache path: %v", err)
	}
	create(&models.Song{Type: models.TypeRemote, Title: "remote-uncached", URL: "https://x/2"})

	rows, err := repo.ListSongsNeedingMetadata(ctx)
	if err != nil {
		t.Fatalf("ListSongsNeedingMetadata: %v", err)
	}

	got := map[int64]bool{}
	for _, r := range rows {
		got[r.ID] = true
	}
	if len(rows) != 2 {
		t.Fatalf("rows: got %d want 2 (%+v)", len(rows), rows)
	}
	if !got[localMissing.ID] {
		t.Errorf("local song with file_path should be included")
	}
	if !got[remoteCached.ID] {
		t.Errorf("remote song with cache_path should be included")
	}
	for _, r := range rows {
		if r.ID == localMissing.ID && r.FilePath != "/m/a.mp3" {
			t.Errorf("local row file_path: got %q", r.FilePath)
		}
		if r.ID == remoteCached.ID && r.CachePath != "/cache/1.mp3" {
			t.Errorf("remote row cache_path: got %q", r.CachePath)
		}
	}
}
