package services

import (
	"context"
	"errors"
	"testing"

	"songloft/internal/database"
	"songloft/internal/database/testutil"
	"songloft/internal/models"
)

// newConvertServiceForTest 构造一个仅注入 db 的最小 ConvertService,
// 用于直接测试 applyConvertInPlace 这种纯库内方法。
// 其他依赖(cacheService / orchestrator / lyricFetcher 等)在本测试不会被触发。
func newConvertServiceForTest(t *testing.T) (*ConvertService, *database.SQLiteDB) {
	t.Helper()
	mdb := testutil.OpenMemoryDB(t)
	var db database.DB = mdb
	return &ConvertService{db: &db}, mdb
}

// TestApplyConvertInPlace_KeepsIDAndPreservesDedupFields 验证原地 UPDATE 保留 id 与 dedup 相关字段。
func TestApplyConvertInPlace_KeepsIDAndPreservesDedupFields(t *testing.T) {
	cs, mdb := newConvertServiceForTest(t)
	ctx := context.Background()

	song := &models.Song{
		Type:            models.TypeRemote,
		Title:           "枪火",
		Artist:          "宝石Gem",
		URL:             "/api/v1/jsplugin/lxmusic/music/url/abc123",
		PluginEntryPath: "lxmusic",
		DedupKey:        "kg:F479CC9AEF",
		SourceData:      `{"hash":"F479CC9AEF"}`,
		LyricRemoteURL:  "/api/v1/jsplugin/lxmusic/api/lyric?hash=abc",
		LyricSource:     models.LyricSourceURL,
	}
	if err := mdb.SongRepository().Create(ctx, song); err != nil {
		t.Fatalf("create remote song: %v", err)
	}
	originalID := song.ID

	updated, err := cs.applyConvertInPlace(ctx, originalID,
		"music/zz/宝石Gem - 枪火.mp3", "",
		".mp3", 2048,
		`{"lyric":"[00:00]test"}`, models.LyricSourceCached,
	)
	if err != nil {
		t.Fatalf("applyConvertInPlace error = %v", err)
	}
	if updated.ID != originalID {
		t.Fatalf("id changed: got %d want %d", updated.ID, originalID)
	}

	got, err := mdb.SongRepository().GetByID(ctx, originalID)
	if err != nil {
		t.Fatalf("GetByID after convert: %v", err)
	}
	if got.Type != models.TypeLocal {
		t.Errorf("Type = %q, want local", got.Type)
	}
	if got.FilePath != "music/zz/宝石Gem - 枪火.mp3" {
		t.Errorf("FilePath = %q, unexpected", got.FilePath)
	}
	if got.LyricSource != models.LyricSourceCached {
		t.Errorf("LyricSource = %q, want cached", got.LyricSource)
	}
	// 关键:保留的字段
	if got.URL != song.URL {
		t.Errorf("URL not preserved: got %q want %q", got.URL, song.URL)
	}
	if got.PluginEntryPath != song.PluginEntryPath {
		t.Errorf("PluginEntryPath not preserved: got %q want %q", got.PluginEntryPath, song.PluginEntryPath)
	}
	if got.DedupKey != song.DedupKey {
		t.Errorf("DedupKey not preserved: got %q want %q", got.DedupKey, song.DedupKey)
	}
	if got.SourceData != song.SourceData {
		t.Errorf("SourceData not preserved: got %q want %q", got.SourceData, song.SourceData)
	}
	if got.LyricRemoteURL != song.LyricRemoteURL {
		t.Errorf("LyricRemoteURL not preserved: got %q want %q", got.LyricRemoteURL, song.LyricRemoteURL)
	}
}

// TestApplyConvertInPlace_SkipsWhenAlreadyLocal 验证重读 song 时 type 已不是 remote 直接 skip。
func TestApplyConvertInPlace_SkipsWhenAlreadyLocal(t *testing.T) {
	cs, mdb := newConvertServiceForTest(t)
	ctx := context.Background()

	song := &models.Song{
		Type:     models.TypeLocal,
		Title:    "already local",
		FilePath: "music/foo.mp3",
	}
	if err := mdb.SongRepository().Create(ctx, song); err != nil {
		t.Fatalf("create song: %v", err)
	}

	_, err := cs.applyConvertInPlace(ctx, song.ID,
		"music/bar.mp3", "", ".mp3", 1024, "", "")
	if !errors.Is(err, errSkipAlreadyLocal) {
		t.Fatalf("expected errSkipAlreadyLocal, got %v", err)
	}

	// 字段不应被覆盖
	got, _ := mdb.SongRepository().GetByID(ctx, song.ID)
	if got.FilePath != "music/foo.mp3" {
		t.Errorf("FilePath was overwritten: %q", got.FilePath)
	}
}

// TestUpsertRemote_LocalHitReusesID 验证再次添加同首远程歌时,
// 如果对应 (plugin, dedup) 已被转为 local,则仅复用 id,不覆盖本地化字段。
func TestUpsertRemote_LocalHitReusesID(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	repo := mdb.SongRepository()
	ctx := context.Background()

	// 1. 先 UpsertRemote 一首 remote song
	first := &models.Song{
		Type:            models.TypeRemote,
		Title:           "原始远程标题",
		Artist:          "Artist",
		URL:             "/jsplugin/x/y",
		CoverURL:        "https://cdn/cover-remote.jpg",
		PluginEntryPath: "lxmusic",
		DedupKey:        "kg:HASHX",
	}
	if err := repo.UpsertRemote(ctx, first); err != nil {
		t.Fatalf("first UpsertRemote: %v", err)
	}
	originalID := first.ID

	// 2. 模拟"转 local":手工 Update 把它翻成 local + 改本地化字段(保留 dedup)
	first.Type = models.TypeLocal
	first.FilePath = "music/local/path.mp3"
	first.Title = "本地清理后的标题"
	first.CoverURL = ""
	first.CoverPath = "/covers/local.jpg"
	if err := repo.Update(ctx, first); err != nil {
		t.Fatalf("manual convert update: %v", err)
	}

	// 3. 用同样的 (plugin, dedup) 再 UpsertRemote 一次 —— 模拟"用户又把这首远程歌加进新歌单"
	second := &models.Song{
		Type:            models.TypeRemote,
		Title:           "远程返回的标题(不应覆盖)",
		Artist:          "Artist Remote",
		URL:             "/jsplugin/x/y/refresh",
		CoverURL:        "https://cdn/cover-NEW.jpg",
		PluginEntryPath: "lxmusic",
		DedupKey:        "kg:HASHX",
	}
	if err := repo.UpsertRemote(ctx, second); err != nil {
		t.Fatalf("second UpsertRemote: %v", err)
	}
	if second.ID != originalID {
		t.Fatalf("id changed: got %d want %d", second.ID, originalID)
	}

	// 4. DB 里的行应仍是 local + 本地化字段未被覆盖
	got, err := repo.GetByID(ctx, originalID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Type != models.TypeLocal {
		t.Errorf("Type = %q, want local", got.Type)
	}
	if got.Title != "本地清理后的标题" {
		t.Errorf("Title got overwritten: %q", got.Title)
	}
	if got.CoverPath != "/covers/local.jpg" {
		t.Errorf("CoverPath got overwritten: %q", got.CoverPath)
	}
	if got.CoverURL != "" {
		t.Errorf("CoverURL got overwritten: %q", got.CoverURL)
	}
	if got.FilePath != "music/local/path.mp3" {
		t.Errorf("FilePath got overwritten: %q", got.FilePath)
	}
}

// TestLyricURLPath_AfterConvertToLocal 验证 lyric_remote_url 残留时 LyricURLPath 不再误报。
func TestLyricURLPath_AfterConvertToLocal(t *testing.T) {
	tests := []struct {
		name string
		song models.Song
		want string // "" 表示无歌词,"endpoint" 表示返回 /api/v1/songs/{id}/lyric
	}{
		{
			name: "lyric_source=url + lyric_remote_url 非空 -> 返回端点",
			song: models.Song{
				ID:             1,
				LyricSource:    models.LyricSourceURL,
				LyricRemoteURL: "/api/lyric?id=1",
			},
			want: "endpoint",
		},
		{
			name: "lyric_source=cached + lyric 非空 -> 返回端点",
			song: models.Song{
				ID:          1,
				LyricSource: models.LyricSourceCached,
				Lyric:       `{"lyric":"[00:00]test"}`,
			},
			want: "endpoint",
		},
		{
			name: "已转 local: lyric_source=空 + lyric 空 + lyric_remote_url 残留 -> 不返回端点(关键修复)",
			song: models.Song{
				ID:             1,
				LyricSource:    "",
				Lyric:          "",
				LyricRemoteURL: "/api/lyric?id=1",
			},
			want: "",
		},
		{
			name: "已转 local: lyric_source=cached + lyric 空 -> 不返回端点",
			song: models.Song{
				ID:          1,
				LyricSource: models.LyricSourceCached,
				Lyric:       "",
			},
			want: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.song.LyricURLPath()
			if tc.want == "" && got != "" {
				t.Errorf("want empty, got %q", got)
			}
			if tc.want == "endpoint" && got == "" {
				t.Errorf("want endpoint, got empty")
			}
		})
	}
}
