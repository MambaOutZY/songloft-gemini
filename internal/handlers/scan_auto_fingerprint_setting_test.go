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

// newTestScanHandlerForFingerprint 构造只带 ConfigService 的 ScanHandler。
// songService / scanner 传 nil：本测试只覆盖 /settings/scan-auto-fingerprint。
func newTestScanHandlerForFingerprint(t *testing.T) *ScanHandler {
	t.Helper()
	mdb := testutil.OpenMemoryDB(t)
	configService := services.NewConfigService(mdb.ConfigRepository())
	return NewScanHandler(nil, nil, configService)
}

// TestScanAutoFingerprintSetting_DefaultDisabled 未做任何修改时 GET 返回 false。
// 默认关闭是 songloft-org/songloft#323 的核心修复：扫描完成后不再自动跑全库 ffmpeg。
func TestScanAutoFingerprintSetting_DefaultDisabled(t *testing.T) {
	h := newTestScanHandlerForFingerprint(t)

	rr := httptest.NewRecorder()
	h.GetScanAutoFingerprintSetting(rr, httptest.NewRequest("GET", "/api/v1/settings/scan-auto-fingerprint", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if resp["enabled"] {
		t.Errorf("default enabled: got true want false")
	}
}

// TestScanAutoFingerprintSetting_UpdateThenRead PUT 写入后 GET 读到最新值。
func TestScanAutoFingerprintSetting_UpdateThenRead(t *testing.T) {
	h := newTestScanHandlerForFingerprint(t)

	rr1 := httptest.NewRecorder()
	h.UpdateScanAutoFingerprintSetting(rr1, httptest.NewRequest("PUT", "/api/v1/settings/scan-auto-fingerprint",
		strings.NewReader(`{"enabled":true}`)))
	if rr1.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d want 200, body=%s", rr1.Code, rr1.Body.String())
	}

	rr2 := httptest.NewRecorder()
	h.GetScanAutoFingerprintSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/scan-auto-fingerprint", nil))
	var resp map[string]bool
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr2.Body.String())
	}
	if !resp["enabled"] {
		t.Errorf("enabled after PUT true: got false want true")
	}

	// 再关回去，确认双向可写
	rr3 := httptest.NewRecorder()
	h.UpdateScanAutoFingerprintSetting(rr3, httptest.NewRequest("PUT", "/api/v1/settings/scan-auto-fingerprint",
		strings.NewReader(`{"enabled":false}`)))
	if rr3.Code != http.StatusOK {
		t.Fatalf("PUT false status: got %d want 200", rr3.Code)
	}
	rr4 := httptest.NewRecorder()
	h.GetScanAutoFingerprintSetting(rr4, httptest.NewRequest("GET", "/api/v1/settings/scan-auto-fingerprint", nil))
	if err := json.Unmarshal(rr4.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["enabled"] {
		t.Errorf("enabled after PUT false: got true want false")
	}
}

// TestScanAutoFingerprintSetting_BadJSON 请求体非法时返回 400，且不改动已有状态。
func TestScanAutoFingerprintSetting_BadJSON(t *testing.T) {
	h := newTestScanHandlerForFingerprint(t)

	// 先置为 true
	rr0 := httptest.NewRecorder()
	h.UpdateScanAutoFingerprintSetting(rr0, httptest.NewRequest("PUT", "/api/v1/settings/scan-auto-fingerprint",
		strings.NewReader(`{"enabled":true}`)))
	if rr0.Code != http.StatusOK {
		t.Fatalf("setup PUT status: got %d want 200", rr0.Code)
	}

	rr := httptest.NewRecorder()
	h.UpdateScanAutoFingerprintSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/scan-auto-fingerprint",
		strings.NewReader(`{not json`)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad JSON status: got %d want 400, body=%s", rr.Code, rr.Body.String())
	}

	rr2 := httptest.NewRecorder()
	h.GetScanAutoFingerprintSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/scan-auto-fingerprint", nil))
	var resp map[string]bool
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp["enabled"] {
		t.Errorf("bad JSON should not change state: got enabled=false want true")
	}
}

// TestScanAutoFingerprintSetting_NilConfigService configService 未注入时
// GET 返回默认值、PUT 返回 500（不 panic）。
func TestScanAutoFingerprintSetting_NilConfigService(t *testing.T) {
	h := NewScanHandler(nil, nil, nil)

	rr := httptest.NewRecorder()
	h.GetScanAutoFingerprintSetting(rr, httptest.NewRequest("GET", "/api/v1/settings/scan-auto-fingerprint", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("GET status: got %d want 200", rr.Code)
	}
	var resp map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["enabled"] {
		t.Errorf("nil configService default: got true want false")
	}

	rr2 := httptest.NewRecorder()
	h.UpdateScanAutoFingerprintSetting(rr2, httptest.NewRequest("PUT", "/api/v1/settings/scan-auto-fingerprint",
		strings.NewReader(`{"enabled":true}`)))
	if rr2.Code != http.StatusInternalServerError {
		t.Fatalf("PUT with nil configService: got %d want 500", rr2.Code)
	}
}

// TestGetFingerprintStatus_NilSongService songService 未注入时状态端点不 panic，
// 且能正确回报 auto_enabled。
func TestGetFingerprintStatus_NilSongService(t *testing.T) {
	h := newTestScanHandlerForFingerprint(t)

	rr0 := httptest.NewRecorder()
	h.UpdateScanAutoFingerprintSetting(rr0, httptest.NewRequest("PUT", "/api/v1/settings/scan-auto-fingerprint",
		strings.NewReader(`{"enabled":true}`)))
	if rr0.Code != http.StatusOK {
		t.Fatalf("setup PUT status: got %d want 200", rr0.Code)
	}

	rr := httptest.NewRecorder()
	h.GetFingerprintStatus(rr, httptest.NewRequest("GET", "/api/v1/scan/fingerprints/status", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var got FingerprintStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if !got.AutoEnabled {
		t.Error("auto_enabled: got false want true")
	}
	if got.Total != 0 || got.Computed != 0 || got.Failed != 0 || got.Missing != 0 {
		t.Errorf("nil songService should leave counters zero, got %+v", got)
	}
}

// TestCancelFingerprintCompute_NoTask 没有任务在跑时返回 cancelled=false，不报错。
func TestCancelFingerprintCompute_NoTask(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	h := NewScanHandler(nil, nil, services.NewConfigService(mdb.ConfigRepository()))
	h.SetFingerprintService(services.NewFingerprintService(mdb.SongRepository()))

	rr := httptest.NewRecorder()
	h.CancelFingerprintCompute(rr, httptest.NewRequest("POST", "/api/v1/scan/fingerprints/cancel", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["cancelled"] {
		t.Error("cancelled with no running task: got true want false")
	}
}
