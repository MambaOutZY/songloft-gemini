package database

import (
	"context"
	"testing"

	"songloft/internal/models"
)

// TestListSongsExcludePlaylistLabels 验证按歌单 label 排除歌曲：
// 隐藏歌单里的歌不出现在主列表；一首歌只要属于任一隐藏歌单就被排除。
func TestListSongsExcludePlaylistLabels(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	songs := []*models.Song{
		{Type: models.TypeLocal, Title: "儿歌A", FilePath: "/music/kids/a.mp3"},
		{Type: models.TypeLocal, Title: "流行B", FilePath: "/music/pop/b.mp3"},
	}
	if err := db.SongRepository().BatchCreate(ctx, songs); err != nil {
		t.Fatalf("BatchCreate songs: %v", err)
	}
	hiddenSong, visibleSong := songs[0], songs[1]

	hiddenPlaylist := &models.Playlist{Type: "normal", Name: "儿歌", Labels: []string{models.PlaylistLabelHidden}}
	visiblePlaylist := &models.Playlist{Type: "normal", Name: "流行", Labels: []string{}}
	if err := db.PlaylistRepository().Create(ctx, hiddenPlaylist); err != nil {
		t.Fatalf("create hidden playlist: %v", err)
	}
	if err := db.PlaylistRepository().Create(ctx, visiblePlaylist); err != nil {
		t.Fatalf("create visible playlist: %v", err)
	}

	psRepo := db.PlaylistSongRepository()
	if err := psRepo.AddSong(ctx, hiddenPlaylist.ID, hiddenSong.ID, 0); err != nil {
		t.Fatalf("add hidden song: %v", err)
	}
	if err := psRepo.AddSong(ctx, visiblePlaylist.ID, visibleSong.ID, 0); err != nil {
		t.Fatalf("add visible song: %v", err)
	}

	repo := db.SongRepository()

	// 默认不排除：两首都在。
	all, err := repo.List(ctx, &SongFilter{})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 songs without exclusion, got %d", len(all))
	}

	// 排除 hidden：只剩可见歌单的歌。
	filter := &SongFilter{ExcludePlaylistLabels: []string{models.PlaylistLabelHidden}}
	got, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("list excluding hidden: %v", err)
	}
	if len(got) != 1 || got[0].ID != visibleSong.ID {
		t.Fatalf("expected only visible song %d, got %+v", visibleSong.ID, got)
	}

	// Count 与 List 共享过滤条件。
	cnt, err := repo.Count(ctx, filter)
	if err != nil {
		t.Fatalf("count excluding hidden: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected count 1, got %d", cnt)
	}

	// 语义：一首歌同时在隐藏+可见歌单，也被排除。
	if err := psRepo.AddSong(ctx, hiddenPlaylist.ID, visibleSong.ID, 1); err != nil {
		t.Fatalf("add visible song to hidden playlist: %v", err)
	}
	got2, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("list after cross-membership: %v", err)
	}
	if len(got2) != 0 {
		t.Fatalf("expected 0 songs after song joins a hidden playlist, got %d", len(got2))
	}
}
