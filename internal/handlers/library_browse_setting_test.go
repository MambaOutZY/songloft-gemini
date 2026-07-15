package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLibraryBrowseSetting_Default(t *testing.T) {
	h := newTestConfigHandler(t)

	rr := httptest.NewRecorder()
	h.GetLibraryBrowseSetting(rr, httptest.NewRequest("GET", "/api/v1/settings/library-browse", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp libraryBrowseSetting
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if len(resp.Views) != len(libraryViewKeys) {
		t.Fatalf("default views length: got %d want %d", len(resp.Views), len(libraryViewKeys))
	}
	// 默认顺序与 libraryViewKeys 一致、全部可见
	for i, v := range resp.Views {
		if v.Key != libraryViewKeys[i] {
			t.Errorf("view[%d] key: got %q want %q", i, v.Key, libraryViewKeys[i])
		}
		if !v.Visible {
			t.Errorf("view[%d] %q should be visible by default", i, v.Key)
		}
	}
}

func TestLibraryBrowseSetting_UpdateThenRead(t *testing.T) {
	h := newTestConfigHandler(t)

	// 只提交部分 key + 自定义顺序/显隐；其余应被自动补到末尾
	body := `{"views":[{"key":"artist","visible":true},{"key":"all","visible":false}]}`
	rr1 := httptest.NewRecorder()
	h.UpdateLibraryBrowseSetting(rr1, httptest.NewRequest("PUT", "/api/v1/settings/library-browse",
		strings.NewReader(body)))
	if rr1.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d want 200, body=%s", rr1.Code, rr1.Body.String())
	}

	rr2 := httptest.NewRecorder()
	h.GetLibraryBrowseSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/library-browse", nil))
	var resp libraryBrowseSetting
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Views) != len(libraryViewKeys) {
		t.Fatalf("views length after partial update: got %d want %d", len(resp.Views), len(libraryViewKeys))
	}
	if resp.Views[0].Key != "artist" || !resp.Views[0].Visible {
		t.Errorf("view[0] should be artist/visible, got %q/%v", resp.Views[0].Key, resp.Views[0].Visible)
	}
	if resp.Views[1].Key != "all" || resp.Views[1].Visible {
		t.Errorf("view[1] should be all/hidden, got %q/%v", resp.Views[1].Key, resp.Views[1].Visible)
	}
}

func TestLibraryBrowseSetting_InvalidKey(t *testing.T) {
	h := newTestConfigHandler(t)

	body := `{"views":[{"key":"bogus","visible":true}]}`
	rr := httptest.NewRecorder()
	h.UpdateLibraryBrowseSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/library-browse",
		strings.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("invalid key: got %d want 400, body=%s", rr.Code, rr.Body.String())
	}
}

func TestLibraryBrowseSetting_DuplicateKey(t *testing.T) {
	h := newTestConfigHandler(t)

	body := `{"views":[{"key":"album","visible":true},{"key":"album","visible":false}]}`
	rr := httptest.NewRecorder()
	h.UpdateLibraryBrowseSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/library-browse",
		strings.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("duplicate key: got %d want 400", rr.Code)
	}
}

func TestLibraryBrowseSetting_BadJSON(t *testing.T) {
	h := newTestConfigHandler(t)

	rr := httptest.NewRecorder()
	h.UpdateLibraryBrowseSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/library-browse",
		strings.NewReader(`not json`)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("bad JSON: got %d want 400", rr.Code)
	}
}
