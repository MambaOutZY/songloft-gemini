package database

import (
	"context"
	"testing"

	"songloft/internal/models"
)

// TestSongArtistRepositoryDuet 验证多值歌手关联的核心链路：
// 对唱歌曲写入两位主唱 + 一位专辑歌手后，按任一搭档检索都能命中，
// facet artist 把对唱拆成两个独立歌手，显示串为 "A & B"。
func TestSongArtistRepositoryDuet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// 建一首对唱歌
	song := &models.Song{Type: models.TypeLocal, Title: "对唱", Artist: "Alice & Bob", FilePath: "/m/duet.mp3"}
	if err := db.SongRepository().Create(ctx, song); err != nil {
		t.Fatalf("create song: %v", err)
	}

	ar := db.SongArtistRepository()
	inputs := []models.ArtistInput{
		{Name: "Alice", Role: models.ArtistRoleArtist, Position: 0},
		{Name: "Bob", Role: models.ArtistRoleArtist, Position: 1},
		{Name: "Various", Role: models.ArtistRoleAlbumArtist, Position: 0},
	}
	if err := ar.SetSongArtists(ctx, song.ID, inputs); err != nil {
		t.Fatalf("SetSongArtists: %v", err)
	}

	// 读回：三位关联，主唱两位在前
	got, err := ar.GetBySongID(ctx, song.ID)
	if err != nil {
		t.Fatalf("GetBySongID: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 song artists, got %d (%+v)", len(got), got)
	}

	// 按任一搭档检索歌曲都命中（role=artist）
	for _, name := range []string{"Alice", "Bob"} {
		ids, err := ar.SongIDsByArtistName(ctx, name, false)
		if err != nil {
			t.Fatalf("SongIDsByArtistName %q: %v", name, err)
		}
		if len(ids) != 1 || ids[0] != song.ID {
			t.Fatalf("artist %q should match the duet song, got %v", name, ids)
		}
	}

	// 专辑歌手不被算作主唱（includeAlbumArtist=false 命中 0，true 命中 1）
	if ids, _ := ar.SongIDsByArtistName(ctx, "Various", false); len(ids) != 0 {
		t.Fatalf("album_artist should not match role=artist filter, got %v", ids)
	}
	if ids, _ := ar.SongIDsByArtistName(ctx, "Various", true); len(ids) != 1 {
		t.Fatalf("album_artist should match with includeAlbumArtist, got %v", ids)
	}

	// facet artist：对唱拆成 Alice / Bob 两个独立取值（不合并）
	repo := db.SongRepository()
	facets, err := repo.ListFacet(ctx, "artist", nil)
	if err != nil {
		t.Fatalf("facet artist: %v", err)
	}
	names := map[string]int64{}
	for _, f := range facets {
		names[f.Value] = f.Count
	}
	if names["Alice"] != 1 || names["Bob"] != 1 {
		t.Fatalf("expected Alice=1 and Bob=1 as separate artists, got %+v", names)
	}
	// "Alice & Bob" 整串不应作为单独歌手出现
	if _, ok := names["Alice & Bob"]; ok {
		t.Fatalf("duet string should not appear as a merged artist")
	}

	// CountFacet artist = 2（Alice/Bob，不含专辑歌手 Various）
	total, err := repo.CountFacet(ctx, "artist", "")
	if err != nil {
		t.Fatalf("count facet artist: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 distinct artists, got %d", total)
	}

	// ListDistinctNames artist 返回拆开的两位
	distinct, err := repo.ListDistinctNames(ctx, "artist")
	if err != nil {
		t.Fatalf("distinct artist: %v", err)
	}
	if len(distinct) != 2 {
		t.Fatalf("expected 2 distinct artist names, got %v", distinct)
	}

	// 归一：大小写/首尾空白应命中同一实体
	ids2, _ := ar.SongIDsByArtistName(ctx, "  alice  ", false)
	if len(ids2) != 1 {
		t.Fatalf("normalized key should match case/whitespace variants, got %v", ids2)
	}

	// 全量替换：清掉 Alice，只剩 Bob + Various → facet artist 只剩 Bob
	if err := ar.SetSongArtists(ctx, song.ID, []models.ArtistInput{
		{Name: "Bob", Role: models.ArtistRoleArtist},
	}); err != nil {
		t.Fatalf("SetSongArtists replace: %v", err)
	}
	facets2, _ := repo.ListFacet(ctx, "artist", nil)
	if len(facets2) != 1 || facets2[0].Value != "Bob" {
		t.Fatalf("after replace expected only Bob, got %+v", facets2)
	}
}

// TestSongArtistRoleValidation 非法 role 被拒。
func TestSongArtistRoleValidation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()
	song := &models.Song{Type: models.TypeLocal, Title: "X", FilePath: "/m/x.mp3"}
	if err := db.SongRepository().Create(ctx, song); err != nil {
		t.Fatalf("create: %v", err)
	}
	err := db.SongArtistRepository().SetSongArtists(ctx, song.ID, []models.ArtistInput{
		{Name: "A", Role: "contributing"},
	})
	if err == nil {
		t.Fatalf("expected error for invalid role, got nil")
	}
}

// TestNormalizeArtistKey 与迁移 SQL 的 lower(trim()) 一致。
func TestNormalizeArtistKey(t *testing.T) {
	cases := map[string]string{
		"周杰伦":         "周杰伦",
		"  AC/DC  ":   "ac/dc",
		"The Beatles": "the beatles",
		"":            "",
		"   ":         "",
	}
	for in, want := range cases {
		if got := NormalizeArtistKey(in); got != want {
			t.Errorf("NormalizeArtistKey(%q) = %q, want %q", in, got, want)
		}
	}
}
