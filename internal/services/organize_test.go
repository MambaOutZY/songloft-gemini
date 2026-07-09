package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"songloft/internal/database"
	"songloft/internal/database/testutil"
	"songloft/internal/models"
)

// newOrganizeService 构造一个以 musicDir 为 music_path 的 SongService（真实 :memory: DB）。
func newOrganizeService(t *testing.T, musicDir string) (*SongService, *database.SongRepository) {
	t.Helper()
	repo := testutil.OpenMemoryDB(t).SongRepository()
	scanner := NewScanner(&ScanConfig{MusicPath: musicDir})
	return NewSongService(repo, nil, nil, scanner, nil, nil), repo
}

// makeLocalSong 在 musicDir 下写一个真实文件并落库为 local 歌曲。relPath 相对 musicDir。
// file_path 按扫描器实际格式存储为「以 music_path 为根的完整路径」（= Join(musicDir, relPath)）。
func makeLocalSong(t *testing.T, repo *database.SongRepository, musicDir, relPath string) *models.Song {
	t.Helper()
	abs := filepath.Join(musicDir, relPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(abs, []byte("audio-"+relPath), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	song := &models.Song{Type: models.TypeLocal, Title: "T", Artist: "A", Album: "B", FilePath: abs}
	if err := repo.Create(context.Background(), song); err != nil {
		t.Fatalf("create song: %v", err)
	}
	return song
}

func TestOrganizeExecute_MovesFileAndUpdatesDB(t *testing.T) {
	dir := t.TempDir()
	svc, repo := newOrganizeService(t, dir)
	ctx := context.Background()
	song := makeLocalSong(t, repo, dir, "flat.mp3")

	results := svc.OrganizeSongs(ctx, []OrganizeItem{{ID: song.ID, TargetPath: "A/B/T.mp3"}})
	if len(results) != 1 || results[0].Status != "ok" {
		t.Fatalf("want ok, got %+v", results)
	}

	// 源文件已移走，目标文件到位。
	if _, err := os.Stat(filepath.Join(dir, "flat.mp3")); !os.IsNotExist(err) {
		t.Errorf("source should be gone, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "A/B/T.mp3")); err != nil {
		t.Errorf("target should exist: %v", err)
	}
	// DB file_path 已更新为以 music_path 为根的完整路径。
	wantPath := filepath.Join(dir, "A/B/T.mp3")
	got, _ := repo.GetByID(ctx, song.ID)
	if got.FilePath != wantPath {
		t.Errorf("db file_path = %q, want %q", got.FilePath, wantPath)
	}
}

func TestOrganizePreview_NoSideEffects(t *testing.T) {
	dir := t.TempDir()
	svc, repo := newOrganizeService(t, dir)
	ctx := context.Background()
	song := makeLocalSong(t, repo, dir, "flat.mp3")

	results := svc.PreviewOrganize(ctx, []OrganizeItem{{ID: song.ID, TargetPath: "A/B/T.mp3"}})
	if len(results) != 1 || results[0].Status != "ok" {
		t.Fatalf("want ok preview, got %+v", results)
	}
	if results[0].OldPath != filepath.Join(dir, "flat.mp3") || results[0].NewPath != filepath.Join(dir, "A/B/T.mp3") {
		t.Errorf("old/new = %q/%q", results[0].OldPath, results[0].NewPath)
	}
	// dry-run：文件系统与 DB 无变化。
	if _, err := os.Stat(filepath.Join(dir, "flat.mp3")); err != nil {
		t.Errorf("source must remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "A/B/T.mp3")); !os.IsNotExist(err) {
		t.Errorf("target must not be created, err=%v", err)
	}
	got, _ := repo.GetByID(ctx, song.ID)
	if got.FilePath != filepath.Join(dir, "flat.mp3") {
		t.Errorf("db must be unchanged, got %q", got.FilePath)
	}
}

func TestOrganizePreview_Conflict(t *testing.T) {
	dir := t.TempDir()
	svc, repo := newOrganizeService(t, dir)
	ctx := context.Background()

	// 目标已存在的冲突。
	s1 := makeLocalSong(t, repo, dir, "s1.mp3")
	makeLocalSong(t, repo, dir, "occupied.mp3") // 占位目标
	res := svc.PreviewOrganize(ctx, []OrganizeItem{{ID: s1.ID, TargetPath: "occupied.mp3"}})
	if res[0].Status != "conflict" {
		t.Errorf("want conflict (target exists), got %+v", res[0])
	}

	// 批内两项撞同一目标。
	a := makeLocalSong(t, repo, dir, "a.mp3")
	b := makeLocalSong(t, repo, dir, "b.mp3")
	res = svc.PreviewOrganize(ctx, []OrganizeItem{
		{ID: a.ID, TargetPath: "dst/x.mp3"},
		{ID: b.ID, TargetPath: "dst/x.mp3"},
	})
	if res[0].Status != "ok" {
		t.Errorf("first should be ok, got %+v", res[0])
	}
	if res[1].Status != "conflict" {
		t.Errorf("second should be conflict, got %+v", res[1])
	}
}

func TestOrganizeExecute_ConflictRejectNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	svc, repo := newOrganizeService(t, dir)
	ctx := context.Background()
	song := makeLocalSong(t, repo, dir, "src.mp3") // 内容 "audio-src.mp3"

	// 手写一个已存在的目标文件，内容不同。
	tgt := filepath.Join(dir, "dst.mp3")
	if err := os.WriteFile(tgt, []byte("EXISTING"), 0644); err != nil {
		t.Fatal(err)
	}

	results := svc.OrganizeSongs(ctx, []OrganizeItem{{ID: song.ID, TargetPath: "dst.mp3"}})
	if results[0].Status != "error" {
		t.Fatalf("want error on conflict, got %+v", results[0])
	}
	// 目标文件未被覆盖，源文件仍在。
	if data, _ := os.ReadFile(tgt); string(data) != "EXISTING" {
		t.Errorf("target overwritten! content=%q", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "src.mp3")); err != nil {
		t.Errorf("source should remain: %v", err)
	}
	got, _ := repo.GetByID(ctx, song.ID)
	if got.FilePath != filepath.Join(dir, "src.mp3") {
		t.Errorf("db must be unchanged, got %q", got.FilePath)
	}
}

func TestOrganizeCueSkipped(t *testing.T) {
	dir := t.TempDir()
	svc, repo := newOrganizeService(t, dir)
	ctx := context.Background()

	// CUE 拆分歌曲：cue_source_path 非空。
	abs := filepath.Join(dir, "album.flac")
	if err := os.WriteFile(abs, []byte("whole-album"), 0644); err != nil {
		t.Fatal(err)
	}
	song := &models.Song{
		Type: models.TypeLocal, Title: "Track1", FilePath: "album.flac",
		CueSourcePath: filepath.Join(dir, "album.cue"), CueTrackIndex: 1,
	}
	if err := repo.Create(ctx, song); err != nil {
		t.Fatalf("create cue song: %v", err)
	}

	prev := svc.PreviewOrganize(ctx, []OrganizeItem{{ID: song.ID, TargetPath: "X/Y.flac"}})
	if prev[0].Status != "skip" {
		t.Errorf("preview want skip, got %+v", prev[0])
	}
	exec := svc.OrganizeSongs(ctx, []OrganizeItem{{ID: song.ID, TargetPath: "X/Y.flac"}})
	if exec[0].Status != "skip" {
		t.Errorf("execute want skip, got %+v", exec[0])
	}
	// 共享音频未被搬动。
	if _, err := os.Stat(abs); err != nil {
		t.Errorf("cue audio must not move: %v", err)
	}
}

func TestOrganizeValidation(t *testing.T) {
	dir := t.TempDir()
	svc, repo := newOrganizeService(t, dir)
	ctx := context.Background()

	// 非 local 歌曲。
	remote := &models.Song{Type: models.TypeRemote, Title: "R", URL: "http://x/y"}
	if err := repo.Create(ctx, remote); err != nil {
		t.Fatal(err)
	}
	// 扩展名不一致 & 路径穿越。
	local := makeLocalSong(t, repo, dir, "song.mp3")

	cases := []struct {
		name   string
		item   OrganizeItem
		status string
	}{
		{"non-local", OrganizeItem{ID: remote.ID, TargetPath: "a/b.mp3"}, "error"},
		{"ext-mismatch", OrganizeItem{ID: local.ID, TargetPath: "a/b.flac"}, "error"},
		{"traversal", OrganizeItem{ID: local.ID, TargetPath: "../evil.mp3"}, "error"},
		{"not-found", OrganizeItem{ID: 999999, TargetPath: "a/b.mp3"}, "error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := svc.OrganizeSongs(ctx, []OrganizeItem{c.item})
			if res[0].Status != c.status {
				t.Errorf("%s: want %s, got %+v", c.name, c.status, res[0])
			}
		})
	}
}

func TestOrganizeMusicPathUnset(t *testing.T) {
	// scanner 有但 MusicPath 为空。
	repo := testutil.OpenMemoryDB(t).SongRepository()
	svc := NewSongService(repo, nil, nil, NewScanner(&ScanConfig{MusicPath: ""}), nil, nil)
	ctx := context.Background()

	exec := svc.OrganizeSongs(ctx, []OrganizeItem{{ID: 1, TargetPath: "a.mp3"}})
	if exec[0].Status != "error" {
		t.Errorf("execute want error when music_path unset, got %+v", exec[0])
	}
	prev := svc.PreviewOrganize(ctx, []OrganizeItem{{ID: 1, TargetPath: "a.mp3"}})
	if prev[0].Status != "error" {
		t.Errorf("preview want error when music_path unset, got %+v", prev[0])
	}
}
