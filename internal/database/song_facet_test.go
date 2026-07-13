package database

import (
	"context"
	"testing"

	"songloft/internal/models"
)

// seedFacetSongs 造几首带标签维度的本地歌曲，供 facet / 过滤测试复用。
func seedFacetSongs(t *testing.T, db DB) {
	t.Helper()
	ctx := context.Background()
	songs := []*models.Song{
		{Type: models.TypeLocal, Title: "A", Artist: "周杰伦", Album: "范特西", Genre: "Pop", Language: "国语", Style: "R&B", Year: 2001, FilePath: "/m/a.mp3"},
		{Type: models.TypeLocal, Title: "B", Artist: "周杰伦", Album: "范特西", Genre: "Pop", Language: "国语", Style: "抒情", Year: 2001, FilePath: "/m/b.mp3"},
		{Type: models.TypeLocal, Title: "C", Artist: "Beyond", Album: "海阔天空", Genre: "Rock", Language: "粤语", Year: 1993, FilePath: "/m/c.mp3"},
		{Type: models.TypeLocal, Title: "D", Artist: "Adele", Album: "21", Genre: "Pop", Language: "英语", Year: 2011, FilePath: "/m/d.mp3"},
		{Type: models.TypeLocal, Title: "E", Artist: "无标签", FilePath: "/m/e.mp3"}, // genre/year 空，不进 facet
	}
	if err := db.SongRepository().BatchCreate(ctx, songs); err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}
}

func TestListFacet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedFacetSongs(t, db)
	repo := db.SongRepository()
	ctx := context.Background()

	// genre：Pop=3, Rock=1，空值不计入；按计数降序。
	genres, err := repo.ListFacet(ctx, "genre")
	if err != nil {
		t.Fatalf("facet genre: %v", err)
	}
	if len(genres) != 2 {
		t.Fatalf("expected 2 genres, got %d (%+v)", len(genres), genres)
	}
	if genres[0].Value != "Pop" || genres[0].Count != 3 {
		t.Fatalf("expected top genre Pop=3, got %+v", genres[0])
	}

	// language：国语=2, 粤语=1, 英语=1
	langs, err := repo.ListFacet(ctx, "language")
	if err != nil {
		t.Fatalf("facet language: %v", err)
	}
	if len(langs) != 3 || langs[0].Value != "国语" || langs[0].Count != 2 {
		t.Fatalf("unexpected language facet: %+v", langs)
	}

	// style：R&B=1, 抒情=1（空值不计）
	styles, err := repo.ListFacet(ctx, "style")
	if err != nil {
		t.Fatalf("facet style: %v", err)
	}
	if len(styles) != 2 {
		t.Fatalf("expected 2 styles, got %d (%+v)", len(styles), styles)
	}

	// year：2001=2, 1993=1, 2011=1（0 不计）
	years, err := repo.ListFacet(ctx, "year")
	if err != nil {
		t.Fatalf("facet year: %v", err)
	}
	if len(years) != 3 {
		t.Fatalf("expected 3 years, got %d (%+v)", len(years), years)
	}

	// decade：2000=2, 1990=1, 2010=1
	decades, err := repo.ListFacet(ctx, "decade")
	if err != nil {
		t.Fatalf("facet decade: %v", err)
	}
	if len(decades) != 3 {
		t.Fatalf("expected 3 decades, got %d (%+v)", len(decades), decades)
	}
	// 断言含 "2000" 且计数 2
	found := false
	for _, d := range decades {
		if d.Value == "2000" {
			found = true
			if d.Count != 2 {
				t.Fatalf("expected decade 2000 count 2, got %d", d.Count)
			}
		}
	}
	if !found {
		t.Fatalf("decade 2000 not found in %+v", decades)
	}

	// 未知维度返回 ErrNotFound
	if _, err := repo.ListFacet(ctx, "bogus"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for unknown field, got %v", err)
	}
}

func TestSongFilterByTag(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedFacetSongs(t, db)
	repo := db.SongRepository()
	ctx := context.Background()

	// genre=Pop → 3 首
	got, err := repo.List(ctx, &SongFilter{Genre: "Pop"})
	if err != nil {
		t.Fatalf("filter genre: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 Pop songs, got %d", len(got))
	}

	// artist=周杰伦 + language=国语 → 2 首（组合过滤）
	got, err = repo.List(ctx, &SongFilter{Artist: "周杰伦", Language: "国语"})
	if err != nil {
		t.Fatalf("filter artist+language: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 songs, got %d", len(got))
	}

	// decade=2000 → 2001 年的 2 首，不含 1993/2011
	got, err = repo.List(ctx, &SongFilter{DecadeStart: 2000})
	if err != nil {
		t.Fatalf("filter decade: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 songs in 2000s, got %d", len(got))
	}
	for _, s := range got {
		if s.Year < 2000 || s.Year >= 2010 {
			t.Fatalf("decade filter leaked year %d", s.Year)
		}
	}

	// year=1993 精确 → 1 首
	got, err = repo.List(ctx, &SongFilter{Year: 1993})
	if err != nil {
		t.Fatalf("filter year: %v", err)
	}
	if len(got) != 1 || got[0].Title != "C" {
		t.Fatalf("expected only song C for year 1993, got %+v", got)
	}

	// Count 与 List 共享过滤
	cnt, err := repo.Count(ctx, &SongFilter{Genre: "Pop"})
	if err != nil {
		t.Fatalf("count genre: %v", err)
	}
	if cnt != 3 {
		t.Fatalf("expected count 3, got %d", cnt)
	}
}
