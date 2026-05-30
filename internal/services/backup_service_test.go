package services

import (
	"context"
	"testing"

	"songloft/internal/database/testutil"
	"songloft/internal/models"
)

func TestExportEmpty(t *testing.T) {
	db := testutil.OpenMemoryDB(t)
	svc := NewBackupService(db)
	ctx := context.Background()

	data, err := svc.Export(ctx)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if data.Version != models.BackupVersion {
		t.Errorf("Version = %d, want %d", data.Version, models.BackupVersion)
	}
	// 内置歌单（收藏 + 电台收藏）应该被导出
	if len(data.Playlists) != 2 {
		t.Errorf("Playlists count = %d, want 2 (built-in)", len(data.Playlists))
	}
}

func TestExportWithData(t *testing.T) {
	db := testutil.OpenMemoryDB(t)
	ctx := context.Background()

	// 创建一首本地歌曲
	song := &models.Song{
		Type:     models.TypeLocal,
		Title:    "测试歌曲",
		Artist:   "测试艺术家",
		Album:    "测试专辑",
		Duration: 180.0,
		FilePath: "test/song.mp3",
		Format:   "mp3",
		BitRate:  320,
	}
	if err := db.SongRepository().Create(ctx, song); err != nil {
		t.Fatalf("create song: %v", err)
	}

	// 创建歌单并添加歌曲
	playlist := &models.Playlist{
		Type: models.PlaylistTypeNormal,
		Name: "测试歌单",
	}
	if err := db.PlaylistRepository().Create(ctx, playlist); err != nil {
		t.Fatalf("create playlist: %v", err)
	}
	if err := db.PlaylistSongRepository().AddSong(ctx, playlist.ID, song.ID, 1); err != nil {
		t.Fatalf("add song: %v", err)
	}

	svc := NewBackupService(db)
	data, err := svc.Export(ctx)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// 2 个内置歌单 + 1 个自建歌单
	if len(data.Playlists) != 3 {
		t.Fatalf("Playlists count = %d, want 3", len(data.Playlists))
	}

	// 找到我们的测试歌单
	var found *models.BackupPlaylist
	for i := range data.Playlists {
		if data.Playlists[i].Name == "测试歌单" {
			found = &data.Playlists[i]
			break
		}
	}
	if found == nil {
		t.Fatal("测试歌单 not found in export")
	}
	if len(found.Songs) != 1 {
		t.Fatalf("songs count = %d, want 1", len(found.Songs))
	}
	if found.Songs[0].Title != "测试歌曲" {
		t.Errorf("song title = %q, want %q", found.Songs[0].Title, "测试歌曲")
	}
	if found.Songs[0].FilePath != "test/song.mp3" {
		t.Errorf("song file_path = %q, want %q", found.Songs[0].FilePath, "test/song.mp3")
	}
}

func TestImportNewPlaylists(t *testing.T) {
	db := testutil.OpenMemoryDB(t)
	svc := NewBackupService(db)
	ctx := context.Background()

	// 先创建一首本地歌曲供匹配
	song := &models.Song{
		Type:     models.TypeLocal,
		Title:    "已有歌曲",
		FilePath: "existing/song.mp3",
	}
	if err := db.SongRepository().Create(ctx, song); err != nil {
		t.Fatalf("create song: %v", err)
	}

	data := &models.BackupData{
		Version: models.BackupVersion,
		Playlists: []models.BackupPlaylist{
			{
				Name: "新歌单",
				Type: models.PlaylistTypeNormal,
				Songs: []models.BackupSong{
					{
						Type:     models.TypeLocal,
						Title:    "已有歌曲",
						FilePath: "existing/song.mp3",
					},
					{
						Type:            models.TypeRemote,
						Title:           "网络歌曲",
						Artist:          "某艺术家",
						PluginEntryPath: "testplugin",
						DedupKey:        "remote:001",
					},
				},
			},
		},
	}

	result, err := svc.Import(ctx, data)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if result.PlaylistsCreated != 1 {
		t.Errorf("PlaylistsCreated = %d, want 1", result.PlaylistsCreated)
	}
	if result.SongsMatched != 1 {
		t.Errorf("SongsMatched = %d, want 1", result.SongsMatched)
	}
	if result.SongsCreated != 1 {
		t.Errorf("SongsCreated = %d, want 1", result.SongsCreated)
	}

	// 验证歌单实际被创建
	pl, err := db.PlaylistRepository().FindByName(ctx, "新歌单")
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	songs, err := db.PlaylistSongRepository().GetSongs(ctx, pl.ID)
	if err != nil {
		t.Fatalf("GetSongs: %v", err)
	}
	if len(songs) != 2 {
		t.Errorf("playlist songs count = %d, want 2", len(songs))
	}
}

func TestImportMerge(t *testing.T) {
	db := testutil.OpenMemoryDB(t)
	svc := NewBackupService(db)
	ctx := context.Background()

	// 创建一个已有歌单，含一首远程歌曲
	existingSong := &models.Song{
		Type:            models.TypeRemote,
		Title:           "已有远程歌曲",
		PluginEntryPath: "testplugin",
		DedupKey:        "remote:existing",
	}
	if err := db.SongRepository().Create(ctx, existingSong); err != nil {
		t.Fatalf("create song: %v", err)
	}
	playlist := &models.Playlist{
		Type: models.PlaylistTypeNormal,
		Name: "合并歌单",
	}
	if err := db.PlaylistRepository().Create(ctx, playlist); err != nil {
		t.Fatalf("create playlist: %v", err)
	}
	if err := db.PlaylistSongRepository().AddSong(ctx, playlist.ID, existingSong.ID, 1); err != nil {
		t.Fatalf("add song: %v", err)
	}

	// 导入同名歌单，含已有歌曲和新歌曲
	data := &models.BackupData{
		Version: models.BackupVersion,
		Playlists: []models.BackupPlaylist{
			{
				Name: "合并歌单",
				Type: models.PlaylistTypeNormal,
				Songs: []models.BackupSong{
					{
						Type:            models.TypeRemote,
						Title:           "已有远程歌曲",
						PluginEntryPath: "testplugin",
						DedupKey:        "remote:existing",
					},
					{
						Type:            models.TypeRemote,
						Title:           "新远程歌曲",
						PluginEntryPath: "testplugin",
						DedupKey:        "remote:new",
					},
				},
			},
		},
	}

	result, err := svc.Import(ctx, data)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if result.PlaylistsMerged != 1 {
		t.Errorf("PlaylistsMerged = %d, want 1", result.PlaylistsMerged)
	}
	if result.PlaylistsCreated != 0 {
		t.Errorf("PlaylistsCreated = %d, want 0", result.PlaylistsCreated)
	}
	if result.SongsSkipped != 1 {
		t.Errorf("SongsSkipped = %d, want 1 (duplicate)", result.SongsSkipped)
	}
	if result.SongsCreated != 1 {
		t.Errorf("SongsCreated = %d, want 1", result.SongsCreated)
	}

	// 验证歌单中现在有 2 首歌曲
	songs, err := db.PlaylistSongRepository().GetSongs(ctx, playlist.ID)
	if err != nil {
		t.Fatalf("GetSongs: %v", err)
	}
	if len(songs) != 2 {
		t.Errorf("playlist songs count = %d, want 2", len(songs))
	}
}

func TestImportLocalSongNoFile(t *testing.T) {
	db := testutil.OpenMemoryDB(t)
	svc := NewBackupService(db)
	ctx := context.Background()

	data := &models.BackupData{
		Version: models.BackupVersion,
		Playlists: []models.BackupPlaylist{
			{
				Name: "本地歌单",
				Type: models.PlaylistTypeNormal,
				Songs: []models.BackupSong{
					{
						Type:     models.TypeLocal,
						Title:    "不存在的歌曲",
						FilePath: "nonexistent/song.mp3",
					},
				},
			},
		},
	}

	result, err := svc.Import(ctx, data)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if result.PlaylistsCreated != 1 {
		t.Errorf("PlaylistsCreated = %d, want 1", result.PlaylistsCreated)
	}
	// 歌曲未匹配到，不应被添加到歌单
	pl, err := db.PlaylistRepository().FindByName(ctx, "本地歌单")
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	songs, err := db.PlaylistSongRepository().GetSongs(ctx, pl.ID)
	if err != nil {
		t.Fatalf("GetSongs: %v", err)
	}
	if len(songs) != 0 {
		t.Errorf("playlist songs count = %d, want 0", len(songs))
	}
}

func TestImportUnsupportedVersion(t *testing.T) {
	db := testutil.OpenMemoryDB(t)
	svc := NewBackupService(db)
	ctx := context.Background()

	data := &models.BackupData{
		Version: 99,
	}

	_, err := svc.Import(ctx, data)
	if err == nil {
		t.Fatal("Import() should return error for unsupported version")
	}
}

func TestImportRoundTrip(t *testing.T) {
	db1 := testutil.OpenMemoryDB(t)
	ctx := context.Background()

	// 在 db1 中创建一些数据
	song := &models.Song{
		Type:            models.TypeRemote,
		Title:           "远程歌曲",
		Artist:          "艺术家",
		Album:           "专辑",
		Duration:        200.0,
		PluginEntryPath: "testplugin",
		DedupKey:        "remote:rt001",
		URL:             "jsplugin://testplugin/music/url/xxx",
		CoverURL:        "https://cdn.example.com/cover.jpg",
	}
	if err := db1.SongRepository().Create(ctx, song); err != nil {
		t.Fatalf("create song: %v", err)
	}
	playlist := &models.Playlist{
		Type:        models.PlaylistTypeNormal,
		Name:        "往返歌单",
		Description: "round trip test",
	}
	if err := db1.PlaylistRepository().Create(ctx, playlist); err != nil {
		t.Fatalf("create playlist: %v", err)
	}
	if err := db1.PlaylistSongRepository().AddSong(ctx, playlist.ID, song.ID, 1); err != nil {
		t.Fatalf("add song: %v", err)
	}

	// 从 db1 导出
	svc1 := NewBackupService(db1)
	exported, err := svc1.Export(ctx)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// 导入到 db2
	db2 := testutil.OpenMemoryDB(t)
	svc2 := NewBackupService(db2)
	result, err := svc2.Import(ctx, exported)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	// 内置歌单合并 + 自定义歌单创建
	if result.PlaylistsMerged != 2 {
		t.Errorf("PlaylistsMerged = %d, want 2 (built-in)", result.PlaylistsMerged)
	}
	if result.PlaylistsCreated != 1 {
		t.Errorf("PlaylistsCreated = %d, want 1", result.PlaylistsCreated)
	}
	if result.SongsCreated != 1 {
		t.Errorf("SongsCreated = %d, want 1", result.SongsCreated)
	}

	// 再从 db2 导出并比较
	exported2, err := svc2.Export(ctx)
	if err != nil {
		t.Fatalf("Export() from db2 error = %v", err)
	}

	// 找到往返歌单比较
	var bp1, bp2 *models.BackupPlaylist
	for i := range exported.Playlists {
		if exported.Playlists[i].Name == "往返歌单" {
			bp1 = &exported.Playlists[i]
			break
		}
	}
	for i := range exported2.Playlists {
		if exported2.Playlists[i].Name == "往返歌单" {
			bp2 = &exported2.Playlists[i]
			break
		}
	}
	if bp1 == nil || bp2 == nil {
		t.Fatal("往返歌单 not found in one of the exports")
	}
	if len(bp1.Songs) != len(bp2.Songs) {
		t.Fatalf("songs count mismatch: %d vs %d", len(bp1.Songs), len(bp2.Songs))
	}
	if bp1.Songs[0].Title != bp2.Songs[0].Title {
		t.Errorf("song title mismatch: %q vs %q", bp1.Songs[0].Title, bp2.Songs[0].Title)
	}
	if bp1.Songs[0].DedupKey != bp2.Songs[0].DedupKey {
		t.Errorf("song dedup_key mismatch: %q vs %q", bp1.Songs[0].DedupKey, bp2.Songs[0].DedupKey)
	}
}
