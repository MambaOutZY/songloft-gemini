package services

import (
	"context"
	"path/filepath"
	"testing"

	"songloft/internal/models"
)

func TestRenameLocalSongFile_MovesFileAndUpdatesDB(t *testing.T) {
	musicDir := t.TempDir()
	svc, repo := newOrganizeService(t, musicDir)
	song := makeLocalSong(t, repo, musicDir, "old name.mp3")
	song.Title = "新标题"

	changed, err := svc.RenameLocalSongFile(context.Background(), song, "新标题")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true")
	}

	// makeLocalSong 默认 Artist="A"，命名格式为 "{artist} - {title}"。
	wantPath := filepath.Join(musicDir, "A - 新标题.mp3")
	if song.FilePath != wantPath {
		t.Fatalf("song.FilePath = %q, want %q", song.FilePath, wantPath)
	}
	if !fileExists(wantPath) {
		t.Fatalf("new file not found: %s", wantPath)
	}
	if fileExists(filepath.Join(musicDir, "old name.mp3")) {
		t.Fatalf("old file still exists")
	}

	// DB 中 file_path 与 title 应已更新。
	got, err := repo.GetByID(context.Background(), song.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.FilePath != wantPath {
		t.Fatalf("db file_path = %q, want %q", got.FilePath, wantPath)
	}
	if got.Title != "新标题" {
		t.Fatalf("db title = %q, want 新标题", got.Title)
	}
}

func TestRenameLocalSongFile_SameNameNoop(t *testing.T) {
	musicDir := t.TempDir()
	svc, repo := newOrganizeService(t, musicDir)
	// makeLocalSong 默认 Artist="A"，磁盘文件建为拼接后同名，触发同名 noop 路径。
	song := makeLocalSong(t, repo, musicDir, "A - keep.mp3")
	song.Title = "已改标题" // 标题变了但文件名清理后与原名相同

	changed, err := svc.RenameLocalSongFile(context.Background(), song, "keep")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if changed {
		t.Fatalf("expected changed=false for same name")
	}
	if !fileExists(filepath.Join(musicDir, "A - keep.mp3")) {
		t.Fatalf("file should remain")
	}
	// 仍应写回 DB 的 title。
	got, _ := repo.GetByID(context.Background(), song.ID)
	if got.Title != "已改标题" {
		t.Fatalf("db title = %q, want 已改标题", got.Title)
	}
}

func TestRenameLocalSongFile_TargetExists(t *testing.T) {
	musicDir := t.TempDir()
	svc, repo := newOrganizeService(t, musicDir)
	song := makeLocalSong(t, repo, musicDir, "a.mp3")
	makeLocalSong(t, repo, musicDir, "A - b.mp3") // 目标已被占用（含歌手前缀）

	_, err := svc.RenameLocalSongFile(context.Background(), song, "b")
	if err == nil {
		t.Fatalf("expected error for existing target")
	}
	// 源文件应原封不动。
	if !fileExists(filepath.Join(musicDir, "a.mp3")) {
		t.Fatalf("source file should remain untouched")
	}
}

// 仅改标题大小写时不应被误判为「目标已存在」冲突。
// 在大小写不敏感 FS（macOS/Windows）上此前会命中原文件自身而报错；这里在
// 大小写敏感的 Linux 上验证正常改名路径不被 SameFile 例外逻辑破坏。
func TestRenameLocalSongFile_CaseOnlyChange(t *testing.T) {
	musicDir := t.TempDir()
	svc, repo := newOrganizeService(t, musicDir)
	song := makeLocalSong(t, repo, musicDir, "A - keep.mp3")

	changed, err := svc.RenameLocalSongFile(context.Background(), song, "Keep")
	if err != nil {
		t.Fatalf("case-only rename should not error: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true for case-only rename")
	}
	want := filepath.Join(musicDir, "A - Keep.mp3")
	if song.FilePath != want {
		t.Fatalf("song.FilePath = %q, want %q", song.FilePath, want)
	}
}

func TestRenameLocalSongFile_EmptyTitle(t *testing.T) {
	musicDir := t.TempDir()
	svc, repo := newOrganizeService(t, musicDir)
	song := makeLocalSong(t, repo, musicDir, "c.mp3")

	if _, err := svc.RenameLocalSongFile(context.Background(), song, "   "); err == nil {
		t.Fatalf("expected error for empty sanitized title")
	}
}

func TestRenameLocalSongFile_CueRejected(t *testing.T) {
	musicDir := t.TempDir()
	svc, repo := newOrganizeService(t, musicDir)
	song := makeLocalSong(t, repo, musicDir, "cue.flac")
	song.CueSourcePath = filepath.Join(musicDir, "album.cue")

	if _, err := svc.RenameLocalSongFile(context.Background(), song, "track1"); err == nil {
		t.Fatalf("expected error for cue song")
	}
}

func TestRenameLocalSongFile_NonLocalRejected(t *testing.T) {
	musicDir := t.TempDir()
	svc, _ := newOrganizeService(t, musicDir)
	song := &models.Song{Type: models.TypeRemote, Title: "远程", FilePath: ""}

	if _, err := svc.RenameLocalSongFile(context.Background(), song, "x"); err == nil {
		t.Fatalf("expected error for non-local song")
	}
}

// 多歌手（前端按 " & " 拼接）应完整保留在文件名中，避免与同名歌曲冲突。
func TestRenameLocalSongFile_MultipleArtists(t *testing.T) {
	musicDir := t.TempDir()
	svc, repo := newOrganizeService(t, musicDir)
	song := makeLocalSong(t, repo, musicDir, "raw.mp3")
	song.Artist = "A & B"
	song.Title = "对唱"

	changed, err := svc.RenameLocalSongFile(context.Background(), song, "对唱")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true")
	}
	want := filepath.Join(musicDir, "A & B - 对唱.mp3")
	if song.FilePath != want {
		t.Fatalf("song.FilePath = %q, want %q", song.FilePath, want)
	}
	if !fileExists(want) {
		t.Fatalf("new file not found: %s", want)
	}
}

// artist 为空时回退到仅标题，保持旧行为，避免脏数据触发 " - 标题" 前导。
func TestRenameLocalSongFile_EmptyArtistFallback(t *testing.T) {
	musicDir := t.TempDir()
	svc, repo := newOrganizeService(t, musicDir)
	song := makeLocalSong(t, repo, musicDir, "x.mp3")
	song.Artist = "   " // 清理后为空

	changed, err := svc.RenameLocalSongFile(context.Background(), song, "标题")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true")
	}
	want := filepath.Join(musicDir, "标题.mp3")
	if song.FilePath != want {
		t.Fatalf("song.FilePath = %q, want %q", song.FilePath, want)
	}
}

// artist 或 title 中的 "/" 会被替换为 "_"，避免误建子目录（如 "AC/DC"）。
func TestRenameLocalSongFile_ArtistWithSlash(t *testing.T) {
	musicDir := t.TempDir()
	svc, repo := newOrganizeService(t, musicDir)
	song := makeLocalSong(t, repo, musicDir, "y.mp3")
	song.Artist = "AC/DC"

	changed, err := svc.RenameLocalSongFile(context.Background(), song, "Thunder")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true")
	}
	want := filepath.Join(musicDir, "AC_DC - Thunder.mp3")
	if song.FilePath != want {
		t.Fatalf("song.FilePath = %q, want %q", song.FilePath, want)
	}
}
