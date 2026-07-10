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

	wantPath := filepath.Join(musicDir, "新标题.mp3")
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
	song := makeLocalSong(t, repo, musicDir, "keep.mp3")
	song.Title = "已改标题" // 标题变了但文件名清理后与原名相同

	changed, err := svc.RenameLocalSongFile(context.Background(), song, "keep")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if changed {
		t.Fatalf("expected changed=false for same name")
	}
	if !fileExists(filepath.Join(musicDir, "keep.mp3")) {
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
	makeLocalSong(t, repo, musicDir, "b.mp3") // 目标已被占用

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
	song := makeLocalSong(t, repo, musicDir, "keep.mp3")

	changed, err := svc.RenameLocalSongFile(context.Background(), song, "Keep")
	if err != nil {
		t.Fatalf("case-only rename should not error: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true for case-only rename")
	}
	if song.FilePath != filepath.Join(musicDir, "Keep.mp3") {
		t.Fatalf("song.FilePath = %q, want %q", song.FilePath, filepath.Join(musicDir, "Keep.mp3"))
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
