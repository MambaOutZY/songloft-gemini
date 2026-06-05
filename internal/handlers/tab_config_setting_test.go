package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"songloft/internal/database/testutil"
	"songloft/internal/services"
)

func newTestConfigHandler(t *testing.T) *ConfigHandler {
	t.Helper()
	mdb := testutil.OpenMemoryDB(t)
	return NewConfigHandler(services.NewConfigService(mdb.ConfigRepository()))
}

func TestTabConfigSetting_Default(t *testing.T) {
	h := newTestConfigHandler(t)

	rr := httptest.NewRecorder()
	h.GetTabConfigSetting(rr, httptest.NewRequest("GET", "/api/v1/settings/tab-config", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp tabConfigSetting
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if !resp.ShowLibrary {
		t.Error("default show_library should be true")
	}
	if !resp.ShowPlaylists {
		t.Error("default show_playlists should be true")
	}
	if len(resp.PluginTabs) != 0 {
		t.Errorf("default plugin_tabs should be empty, got %d", len(resp.PluginTabs))
	}
}

func TestTabConfigSetting_UpdateThenRead(t *testing.T) {
	h := newTestConfigHandler(t)

	body := `{"show_library":false,"show_playlists":true,"plugin_tabs":[{"plugin_id":1,"entry_path":"subsonic","name":"Subsonic"}]}`
	rr1 := httptest.NewRecorder()
	h.UpdateTabConfigSetting(rr1, httptest.NewRequest("PUT", "/api/v1/settings/tab-config",
		strings.NewReader(body)))
	if rr1.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d want 200, body=%s", rr1.Code, rr1.Body.String())
	}

	rr2 := httptest.NewRecorder()
	h.GetTabConfigSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/tab-config", nil))
	var resp tabConfigSetting
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ShowLibrary {
		t.Error("show_library should be false after update")
	}
	if !resp.ShowPlaylists {
		t.Error("show_playlists should be true")
	}
	if len(resp.PluginTabs) != 1 {
		t.Fatalf("plugin_tabs length: got %d want 1", len(resp.PluginTabs))
	}
	if resp.PluginTabs[0].EntryPath != "subsonic" {
		t.Errorf("entry_path: got %q want subsonic", resp.PluginTabs[0].EntryPath)
	}
}

func TestTabConfigSetting_ExceedLimit(t *testing.T) {
	h := newTestConfigHandler(t)

	// 4 个可选项 = 超过上限 3
	body := `{"show_library":true,"show_playlists":true,"plugin_tabs":[{"plugin_id":1,"entry_path":"a","name":"A"},{"plugin_id":2,"entry_path":"b","name":"B"}]}`
	rr := httptest.NewRecorder()
	h.UpdateTabConfigSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/tab-config",
		strings.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("exceed limit: got %d want 400, body=%s", rr.Code, rr.Body.String())
	}
}

func TestTabConfigSetting_DuplicateEntryPath(t *testing.T) {
	h := newTestConfigHandler(t)

	body := `{"show_library":false,"show_playlists":false,"plugin_tabs":[{"plugin_id":1,"entry_path":"same","name":"A"},{"plugin_id":2,"entry_path":"same","name":"B"}]}`
	rr := httptest.NewRecorder()
	h.UpdateTabConfigSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/tab-config",
		strings.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("duplicate entry_path: got %d want 400", rr.Code)
	}
}

func TestTabConfigSetting_EmptyEntryPath(t *testing.T) {
	h := newTestConfigHandler(t)

	body := `{"show_library":false,"show_playlists":false,"plugin_tabs":[{"plugin_id":1,"entry_path":"","name":"A"}]}`
	rr := httptest.NewRecorder()
	h.UpdateTabConfigSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/tab-config",
		strings.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("empty entry_path: got %d want 400", rr.Code)
	}
}

func TestTabConfigSetting_BadJSON(t *testing.T) {
	h := newTestConfigHandler(t)

	rr := httptest.NewRecorder()
	h.UpdateTabConfigSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/tab-config",
		strings.NewReader(`not json`)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("bad JSON: got %d want 400", rr.Code)
	}
}
