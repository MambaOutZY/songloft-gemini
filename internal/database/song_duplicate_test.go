package database

import (
	"context"
	"testing"

	"songloft/internal/models"
)

// seedDuplicateSong 插入一首带指定指纹 / 指纹时长的本地歌曲。
func seedDuplicateSong(t *testing.T, repo *SongRepository, path, fingerprint string, fpDuration float64) {
	t.Helper()
	song := &models.Song{
		Type:     models.TypeLocal,
		Title:    path,
		FilePath: path,
		Duration: fpDuration,
	}
	ctx := context.Background()
	if err := repo.Create(ctx, song); err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	if err := repo.UpdateFingerprint(ctx, song.ID, fingerprint, fpDuration, 1); err != nil {
		t.Fatalf("update fingerprint %s: %v", path, err)
	}
}

// TestListDuplicateGroups_SameDuration 指纹相同且全片时长接近 → 判为重复。
func TestListDuplicateGroups_SameDuration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := db.SongRepository()

	// 同一首歌的 mp3 / flac 两个版本，时长差 0.4s（转码 padding 造成）
	seedDuplicateSong(t, repo, "/m/song.mp3", "AQAAsomefp", 210.0)
	seedDuplicateSong(t, repo, "/m/song.flac", "AQAAsomefp", 210.4)

	groups, err := repo.ListDuplicateGroups(context.Background())
	if err != nil {
		t.Fatalf("ListDuplicateGroups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("groups: got %d want 1 (%+v)", len(groups), groups)
	}
	if len(groups[0].Songs) != 2 {
		t.Errorf("group size: got %d want 2", len(groups[0].Songs))
	}
}

// TestListDuplicateGroups_DurationGuard 指纹只采样前 120 秒，
// 「统一片头的有声书」这类前若干秒完全一致的文件会撞上同一个指纹；
// 全片时长差得远时必须靠时长护栏排除，不能误判为重复。
func TestListDuplicateGroups_DurationGuard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := db.SongRepository()

	seedDuplicateSong(t, repo, "/m/audiobook/ep001.m4a", "AQAAintro", 1820.0)
	seedDuplicateSong(t, repo, "/m/audiobook/ep002.m4a", "AQAAintro", 2140.0)

	groups, err := repo.ListDuplicateGroups(context.Background())
	if err != nil {
		t.Fatalf("ListDuplicateGroups: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("same fingerprint but far-apart durations must not group, got %+v", groups)
	}
}

// TestListDuplicateGroups_MixedClusters 同指纹里既有真重复也有时长离群项时，
// 只输出成簇的那一组，离群项被剔除。
func TestListDuplicateGroups_MixedClusters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := db.SongRepository()

	seedDuplicateSong(t, repo, "/m/a.mp3", "AQAAmixed", 300.0)
	seedDuplicateSong(t, repo, "/m/b.mp3", "AQAAmixed", 301.0)
	seedDuplicateSong(t, repo, "/m/c.mp3", "AQAAmixed", 900.0)

	groups, err := repo.ListDuplicateGroups(context.Background())
	if err != nil {
		t.Fatalf("ListDuplicateGroups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("groups: got %d want 1 (%+v)", len(groups), groups)
	}
	if len(groups[0].Songs) != 2 {
		t.Fatalf("group size: got %d want 2", len(groups[0].Songs))
	}
	for _, s := range groups[0].Songs {
		if s.FilePath == "/m/c.mp3" {
			t.Errorf("outlier /m/c.mp3 should not be in the duplicate group")
		}
	}
}
