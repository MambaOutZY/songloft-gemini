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

// newTestHLSHandlerWithConfig 构造带 ConfigService 的 HLSHandler，覆盖业务封装层。
func newTestHLSHandlerWithConfig(t *testing.T) *HLSHandler {
	t.Helper()
	mdb := testutil.OpenMemoryDB(t)
	configService := services.NewConfigService(mdb.ConfigRepository())
	// songService 传 nil：本测试只覆盖 /settings/hls-proxy，不调 ServeProxy
	return NewHLSHandler(nil, configService)
}

// TestHLSProxySetting_DefaultDisabled 没有任何写入时 GET 返回 enabled=false（业务默认值）。
// 这是方向 A 的关键：业务封装承担默认值，不依赖通用 /configs 表预置。
func TestHLSProxySetting_DefaultDisabled(t *testing.T) {
	h := newTestHLSHandlerWithConfig(t)

	rr := httptest.NewRecorder()
	h.GetProxySetting(rr, httptest.NewRequest("GET", "/api/v1/settings/hls-proxy", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if resp["enabled"] != false {
		t.Errorf("default enabled: got %v want false", resp["enabled"])
	}
}

// TestHLSProxySetting_UpdateThenRead PUT 写入后 GET 读到最新值（写后读一致）。
func TestHLSProxySetting_UpdateThenRead(t *testing.T) {
	h := newTestHLSHandlerWithConfig(t)

	// PUT enabled=true
	rr1 := httptest.NewRecorder()
	h.UpdateProxySetting(rr1, httptest.NewRequest("PUT", "/api/v1/settings/hls-proxy",
		strings.NewReader(`{"enabled":true}`)))
	if rr1.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d want 200, body=%s", rr1.Code, rr1.Body.String())
	}

	// IsEnabled 业务方法应返回 true
	if !h.IsEnabled() {
		t.Error("IsEnabled() after PUT true should be true")
	}

	// GET 也应返回 enabled=true
	rr2 := httptest.NewRecorder()
	h.GetProxySetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/hls-proxy", nil))
	var resp map[string]bool
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["enabled"] != true {
		t.Errorf("read-after-write: got %v want true", resp["enabled"])
	}

	// 再 PUT false 切回
	rr3 := httptest.NewRecorder()
	h.UpdateProxySetting(rr3, httptest.NewRequest("PUT", "/api/v1/settings/hls-proxy",
		strings.NewReader(`{"enabled":false}`)))
	if rr3.Code != http.StatusOK {
		t.Fatalf("PUT false status: got %d", rr3.Code)
	}
	if h.IsEnabled() {
		t.Error("IsEnabled() after PUT false should be false")
	}
}

// TestHLSProxySetting_BadJSON 请求体不是合法 JSON 时返回 400，且不修改状态。
func TestHLSProxySetting_BadJSON(t *testing.T) {
	h := newTestHLSHandlerWithConfig(t)

	rr := httptest.NewRecorder()
	h.UpdateProxySetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/hls-proxy",
		strings.NewReader(`not json`)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("bad JSON: got %d want 400", rr.Code)
	}
	if h.IsEnabled() {
		t.Error("state should remain false after bad PUT")
	}
}

// TestHLSProxySetting_NilConfigService IsEnabled 在 configService=nil 时安全返回 false，
// 兼容只用底层反代逻辑的旧测试栈（newTestHLSStack 传 nil）。
func TestHLSProxySetting_NilConfigService(t *testing.T) {
	h := NewHLSHandler(nil, nil)
	if h.IsEnabled() {
		t.Error("nil configService should report disabled")
	}
	if err := h.SetEnabled(true); err == nil {
		t.Error("SetEnabled with nil configService should error")
	}
}
