package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"songloft/internal/models"
)

// TestAutoCreate_DirCoverJpg 验证自动创建歌单时，目录下有 cover.jpg 会被用作歌单封面。
func TestAutoCreate_DirCoverJpg(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	musicRoot := t.TempDir()
	coverStorage := t.TempDir()

	albumDir := filepath.Join(musicRoot, "albumA")
	if err := os.MkdirAll(albumDir, 0755); err != nil {
		t.Fatal(err)
	}

	coverData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}
	if err := os.WriteFile(filepath.Join(albumDir, "cover.jpg"), coverData, 0644); err != nil {
		t.Fatal(err)
	}

	songPath1 := filepath.Join(albumDir, "01.mp3")
	songPath2 := filepath.Join(albumDir, "02.mp3")
	if err := os.WriteFile(songPath1, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(songPath2, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	songs := []*models.Song{
		{Type: models.TypeLocal, Title: "Song 1", FilePath: songPath1, Duration: 120},
		{Type: models.TypeLocal, Title: "Song 2", FilePath: songPath2, Duration: 180},
	}
	if err := db.SongRepository().BatchCreate(ctx, songs); err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}

	resp, err := db.PlaylistRepository().AutoCreate(ctx, models.PlaylistModeDirectory, nil, coverStorage, musicRoot)
	if err != nil {
		t.Fatalf("AutoCreate: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected 1 playlist, got %d", resp.Total)
	}

	playlists, err := db.PlaylistRepository().List(ctx, &PlaylistFilter{
		Labels: []string{models.PlaylistLabelAutoCreated},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(playlists) != 1 {
		t.Fatalf("expected 1 auto-created playlist, got %d", len(playlists))
	}

	pl := playlists[0]
	if pl.CoverPath == "" {
		t.Errorf("歌单封面 CoverPath 为空！应该从目录下的 cover.jpg 获取封面")
	}
}

// TestAutoCreate_CueAlbum_DirCover 验证 CUE 专辑歌单也使用目录下的 cover.jpg。
func TestAutoCreate_CueAlbum_DirCover(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	musicRoot := t.TempDir()
	coverStorage := t.TempDir()

	albumDir := filepath.Join(musicRoot, "cue-album")
	if err := os.MkdirAll(albumDir, 0755); err != nil {
		t.Fatal(err)
	}

	coverData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}
	if err := os.WriteFile(filepath.Join(albumDir, "cover.jpg"), coverData, 0644); err != nil {
		t.Fatal(err)
	}

	cuePath := filepath.Join(albumDir, "album.cue")
	audioPath := filepath.Join(albumDir, "album.flac")
	songs := []*models.Song{
		{Type: models.TypeLocal, Title: "Track 1", Album: "My CUE Album",
			FilePath: audioPath, CueSourcePath: cuePath, CueTrackIndex: 1, Duration: 300},
		{Type: models.TypeLocal, Title: "Track 2", Album: "My CUE Album",
			FilePath: audioPath, CueSourcePath: cuePath, CueTrackIndex: 2, Duration: 250},
	}
	if err := db.SongRepository().BatchCreate(ctx, songs); err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}

	resp, err := db.PlaylistRepository().AutoCreate(ctx, models.PlaylistModeDirectory, nil, coverStorage, musicRoot)
	if err != nil {
		t.Fatalf("AutoCreate: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected 1 CUE album playlist, got %d", resp.Total)
	}

	playlists, err := db.PlaylistRepository().List(ctx, &PlaylistFilter{
		Labels: []string{models.PlaylistLabelAutoCreated},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(playlists) != 1 {
		t.Fatalf("expected 1 auto-created playlist, got %d", len(playlists))
	}

	pl := playlists[0]
	if pl.CoverPath == "" {
		t.Errorf("CUE 专辑歌单封面 CoverPath 为空！应该从目录下的 cover.jpg 获取封面")
	}
}
