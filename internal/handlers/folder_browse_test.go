package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"songloft/internal/database"
	"songloft/internal/models"
	"songloft/internal/services"
)

func newTestFolderHandler(t *testing.T, musicPath string) (*SongHandler, *database.SongRepository) {
	t.Helper()
	repo := newTestSongRepo(t)
	songService := services.NewSongService(repo, nil, nil, nil, nil, nil)
	handler := NewSongHandler(songService, nil, nil, nil, nil, nil)
	handler.SetGetMusicPath(func() string { return musicPath })
	return handler, repo
}

func seedFolderSongs(t *testing.T, repo *database.SongRepository) {
	t.Helper()
	songs := []*models.Song{
		{Type: models.TypeLocal, Title: "001", FilePath: "/music/评书/三国演义/001.mp3"},
		{Type: models.TypeLocal, Title: "002", FilePath: "/music/评书/三国演义/002.mp3"},
		{Type: models.TypeLocal, Title: "001", FilePath: "/music/评书/水浒传/001.mp3"},
		{Type: models.TypeLocal, Title: "intro", FilePath: "/music/评书/intro.mp3"},
		{Type: models.TypeLocal, Title: "song1", FilePath: "/music/英语/新概念/lesson1.mp3"},
		{Type: models.TypeLocal, Title: "root", FilePath: "/music/root.mp3"},
	}
	for _, s := range songs {
		seedSong(t, repo, s)
	}
}

func TestListFolders_Root(t *testing.T) {
	h, repo := newTestFolderHandler(t, "/music")
	seedFolderSongs(t, repo)

	rr := httptest.NewRecorder()
	h.ListFolders(rr, httptest.NewRequest("GET", "/api/v1/songs/folders", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp["path"] != "" {
		t.Errorf("path: got %q want empty", resp["path"])
	}
	if resp["parent_path"] != "" {
		t.Errorf("parent_path: got %q want empty", resp["parent_path"])
	}
	if resp["music_path"] != "/music" {
		t.Errorf("music_path: got %q want /music", resp["music_path"])
	}

	folders := resp["folders"].([]any)
	if len(folders) != 2 {
		t.Fatalf("folders: got %d want 2 (评书, 英语)", len(folders))
	}

	songs := resp["songs"].([]any)
	if len(songs) != 1 {
		t.Errorf("direct songs at root: got %d want 1 (root.mp3)", len(songs))
	}
}

func TestListFolders_Nested(t *testing.T) {
	h, repo := newTestFolderHandler(t, "/music")
	seedFolderSongs(t, repo)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/songs/folders?path=评书", nil)
	h.ListFolders(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)

	if resp["path"] != "评书" {
		t.Errorf("path: got %q want 评书", resp["path"])
	}
	if resp["parent_path"] != "" {
		t.Errorf("parent_path: got %q want empty", resp["parent_path"])
	}

	folders := resp["folders"].([]any)
	if len(folders) != 2 {
		t.Fatalf("folders: got %d want 2 (三国演义, 水浒传)", len(folders))
	}

	songs := resp["songs"].([]any)
	if len(songs) != 1 {
		t.Errorf("direct songs in 评书/: got %d want 1 (intro.mp3)", len(songs))
	}
}

func TestListFolders_LeafFolder(t *testing.T) {
	h, repo := newTestFolderHandler(t, "/music")
	seedFolderSongs(t, repo)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/songs/folders?path=评书/三国演义", nil)
	h.ListFolders(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rr.Code)
	}

	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)

	folders := resp["folders"].([]any)
	if len(folders) != 0 {
		t.Errorf("folders: got %d want 0 (leaf folder)", len(folders))
	}

	songs := resp["songs"].([]any)
	if len(songs) != 2 {
		t.Errorf("songs: got %d want 2", len(songs))
	}

	if resp["parent_path"] != "评书" {
		t.Errorf("parent_path: got %q want 评书", resp["parent_path"])
	}
}

func TestListFolders_Keyword(t *testing.T) {
	h, repo := newTestFolderHandler(t, "/music")
	seedFolderSongs(t, repo)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/songs/folders?path=评书&keyword=三国", nil)
	h.ListFolders(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rr.Code)
	}

	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)

	folders := resp["folders"].([]any)
	if len(folders) != 1 {
		t.Fatalf("folders with keyword 三国: got %d want 1", len(folders))
	}
	f := folders[0].(map[string]any)
	if f["name"] != "三国演义" {
		t.Errorf("folder name: got %q want 三国演义", f["name"])
	}
}

func TestListFolders_EmptyMusicPath(t *testing.T) {
	h, _ := newTestFolderHandler(t, "")

	rr := httptest.NewRecorder()
	h.ListFolders(rr, httptest.NewRequest("GET", "/api/v1/songs/folders", nil))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("empty music_path: got %d want 400", rr.Code)
	}
}

func TestListFolders_PathTraversal(t *testing.T) {
	h, _ := newTestFolderHandler(t, "/music")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/songs/folders?path=../etc", nil)
	h.ListFolders(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("path traversal: got %d want 400", rr.Code)
	}
}
