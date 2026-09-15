package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"songloft/internal/database/testutil"
	"songloft/internal/models"
	"songloft/internal/services"
)

func TestPluginOrderSetting_Default(t *testing.T) {
	h := newTestConfigHandler(t)

	rr := httptest.NewRecorder()
	h.GetPluginOrderSetting(rr, httptest.NewRequest("GET", "/api/v1/settings/plugin-order", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp pluginOrderSetting
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if len(resp.Order) != 0 {
		t.Errorf("default order should be empty, got %d", len(resp.Order))
	}
	// 未配置时必须序列化为 "order":[]，而不是 null —— 否则前端 (a ?? []) 之类的
	// 兜底不会触发，反而在 map/forEach 里立刻炸掉。
	if !strings.Contains(rr.Body.String(), `"order":[]`) {
		t.Errorf("default order should serialize as [], got body=%s", rr.Body.String())
	}
}

func TestPluginOrderSetting_UpdateThenRead(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	seedTestPlugin(t, mdb, "subsonic", models.JSPluginStatusActive)
	seedTestPlugin(t, mdb, "netease", models.JSPluginStatusActive)
	h := NewConfigHandler(services.NewConfigService(mdb.ConfigRepository()), mdb.JSPluginRepository())

	body := `{"order":["netease","subsonic"]}`
	rr1 := httptest.NewRecorder()
	h.UpdatePluginOrderSetting(rr1, httptest.NewRequest("PUT", "/api/v1/settings/plugin-order",
		strings.NewReader(body)))
	if rr1.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d want 200, body=%s", rr1.Code, rr1.Body.String())
	}

	rr2 := httptest.NewRecorder()
	h.GetPluginOrderSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/plugin-order", nil))
	var resp pluginOrderSetting
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Order) != 2 || resp.Order[0] != "netease" || resp.Order[1] != "subsonic" {
		t.Errorf("order after update: got %+v want [netease subsonic]", resp.Order)
	}
}

// TestPluginOrderSetting_PruneOrphanEntries 卸载后遗留的 entry_path 应在 PUT 时被静默清理。
// 与 tab_config 保持一致：客户端应以响应为准，不需要再自己校对已卸载插件。
func TestPluginOrderSetting_PruneOrphanEntries(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	seedTestPlugin(t, mdb, "alive", models.JSPluginStatusActive)
	h := NewConfigHandler(services.NewConfigService(mdb.ConfigRepository()), mdb.JSPluginRepository())

	body := `{"order":["alive","ghost"]}`
	rr := httptest.NewRecorder()
	h.UpdatePluginOrderSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/plugin-order",
		strings.NewReader(body)))
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp pluginOrderSetting
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Order) != 1 || resp.Order[0] != "alive" {
		t.Errorf("orphan should be pruned: got %+v want [alive]", resp.Order)
	}

	// 持久化的也应是清理后的配置
	rr2 := httptest.NewRecorder()
	h.GetPluginOrderSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/plugin-order", nil))
	var persisted pluginOrderSetting
	if err := json.Unmarshal(rr2.Body.Bytes(), &persisted); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(persisted.Order) != 1 || persisted.Order[0] != "alive" {
		t.Errorf("persisted order: got %+v want [alive]", persisted.Order)
	}
}

// TestPluginOrderSetting_DisabledPluginKept 禁用插件的 entry_path 保留，
// 重新启用后位置恢复。语义与 tab_config 对齐。
func TestPluginOrderSetting_DisabledPluginKept(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	seedTestPlugin(t, mdb, "active", models.JSPluginStatusActive)
	seedTestPlugin(t, mdb, "paused", models.JSPluginStatusInactive)
	h := NewConfigHandler(services.NewConfigService(mdb.ConfigRepository()), mdb.JSPluginRepository())

	body := `{"order":["paused","active"]}`
	rr := httptest.NewRecorder()
	h.UpdatePluginOrderSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/plugin-order",
		strings.NewReader(body)))
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp pluginOrderSetting
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Order) != 2 || resp.Order[0] != "paused" || resp.Order[1] != "active" {
		t.Errorf("disabled plugin entry should be kept in place: got %+v want [paused active]", resp.Order)
	}
}

func TestPluginOrderSetting_DuplicateEntryPath(t *testing.T) {
	h := newTestConfigHandler(t)

	body := `{"order":["same","same"]}`
	rr := httptest.NewRecorder()
	h.UpdatePluginOrderSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/plugin-order",
		strings.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("duplicate entry_path: got %d want 400", rr.Code)
	}
}

func TestPluginOrderSetting_EmptyEntryPath(t *testing.T) {
	h := newTestConfigHandler(t)

	body := `{"order":[""]}`
	rr := httptest.NewRecorder()
	h.UpdatePluginOrderSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/plugin-order",
		strings.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("empty entry_path: got %d want 400", rr.Code)
	}
}

func TestPluginOrderSetting_BadJSON(t *testing.T) {
	h := newTestConfigHandler(t)

	rr := httptest.NewRecorder()
	h.UpdatePluginOrderSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/plugin-order",
		strings.NewReader(`not json`)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("bad JSON: got %d want 400", rr.Code)
	}
}

// TestPluginOrderSetting_EmptyOrder 空数组是合法输入（用户清空所有排序，回退到默认顺序）。
func TestPluginOrderSetting_EmptyOrder(t *testing.T) {
	h := newTestConfigHandler(t)

	body := `{"order":[]}`
	rr := httptest.NewRecorder()
	h.UpdatePluginOrderSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/plugin-order",
		strings.NewReader(body)))
	if rr.Code != http.StatusOK {
		t.Errorf("empty order: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
}

// TestRemovePluginOrderEntry 卸载插件时应从 plugin_order 中移除对应条目。
// 与 tab_config 对称：不清理会让下次 PUT 前配置里持续挂着孤儿 entry_path。
func TestRemovePluginOrderEntry(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	cfgSvc := services.NewConfigService(mdb.ConfigRepository())

	cfg := pluginOrderSetting{Order: []string{"keep", "gone"}}
	if err := cfgSvc.SetJSON(pluginOrderKey, cfg); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	h := NewJSPluginHandler(nil, nil, nil, nil, cfgSvc, nil)
	h.removePluginOrderEntry("gone")

	var got pluginOrderSetting
	if err := cfgSvc.GetJSON(pluginOrderKey, &got); err != nil {
		t.Fatalf("read config: %v", err)
	}
	if len(got.Order) != 1 || got.Order[0] != "keep" {
		t.Errorf("order after removal: got %+v want [keep]", got.Order)
	}

	// 幂等：再次移除已不存在的条目不改动配置
	h.removePluginOrderEntry("not-there")
	var got2 pluginOrderSetting
	if err := cfgSvc.GetJSON(pluginOrderKey, &got2); err != nil {
		t.Fatalf("read config: %v", err)
	}
	if len(got2.Order) != 1 {
		t.Errorf("idempotent removal changed entries: got %d want 1", len(got2.Order))
	}
}
