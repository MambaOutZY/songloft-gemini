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

// newTestSongHandlerForVolumeNormalize 构造带同一 configService 的 SongHandler + CacheService，
// 保证 handler 写入的响度配置能被 cacheService 读到（与生产装配一致）。
func newTestSongHandlerForVolumeNormalize(t *testing.T) *SongHandler {
	t.Helper()
	mdb := testutil.OpenMemoryDB(t)
	configService := services.NewConfigService(mdb.ConfigRepository())
	cacheService := services.NewCacheService(t.TempDir(), configService)
	h := NewSongHandler(nil, cacheService, nil, nil, nil, nil)
	h.SetConfigService(configService)
	return h
}

// TestVolumeNormalizeSetting_Default 未配置时 GET 返回 enabled=false、loudness=-16。
func TestVolumeNormalizeSetting_Default(t *testing.T) {
	h := newTestSongHandlerForVolumeNormalize(t)

	rr := httptest.NewRecorder()
	h.GetVolumeNormalizeSetting(rr, httptest.NewRequest("GET", "/api/v1/settings/volume-normalize", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp volumeNormalizeRequest
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if resp.Enabled {
		t.Errorf("default enabled: got true want false")
	}
	if resp.Loudness == nil || *resp.Loudness != -16 {
		got := "<nil>"
		if resp.Loudness != nil {
			got = ""
		}
		t.Errorf("default loudness: got %v want -16", got)
	}
}

// TestVolumeNormalizeSetting_UpdateLoudness PUT 同时设开关与响度，GET 读到最新值。
func TestVolumeNormalizeSetting_UpdateLoudness(t *testing.T) {
	h := newTestSongHandlerForVolumeNormalize(t)

	rr := httptest.NewRecorder()
	h.UpdateVolumeNormalizeSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/volume-normalize",
		strings.NewReader(`{"enabled":true,"loudness":-14}`)))
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp volumeNormalizeRequest
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if !resp.Enabled {
		t.Errorf("enabled: got false want true")
	}
	if resp.Loudness == nil || *resp.Loudness != -14 {
		t.Errorf("loudness: got %v want -14", resp.Loudness)
	}

	// GET 读回一致
	rr2 := httptest.NewRecorder()
	h.GetVolumeNormalizeSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/volume-normalize", nil))
	var got volumeNormalizeRequest
	if err := json.Unmarshal(rr2.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode GET: %v", err)
	}
	if !got.Enabled {
		t.Errorf("GET enabled: got false want true")
	}
	if got.Loudness == nil || *got.Loudness != -14 {
		t.Errorf("GET loudness: got %v want -14", got.Loudness)
	}
}

// TestVolumeNormalizeSetting_UpdateEnabledOnly 省略 loudness 时只切开关，不改响度（向后兼容旧前端）。
func TestVolumeNormalizeSetting_UpdateEnabledOnly(t *testing.T) {
	h := newTestSongHandlerForVolumeNormalize(t)
	// 先把响度设成 -14
	rr0 := httptest.NewRecorder()
	h.UpdateVolumeNormalizeSetting(rr0, httptest.NewRequest("PUT", "/api/v1/settings/volume-normalize",
		strings.NewReader(`{"enabled":true,"loudness":-14}`)))
	if rr0.Code != http.StatusOK {
		t.Fatalf("setup PUT: got %d want 200", rr0.Code)
	}

	// 只发 enabled（模拟旧前端），响度应保持 -14 不动
	rr := httptest.NewRecorder()
	h.UpdateVolumeNormalizeSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/volume-normalize",
		strings.NewReader(`{"enabled":false}`)))
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT enabled-only status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp volumeNormalizeRequest
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Enabled {
		t.Errorf("enabled: got true want false")
	}
	if resp.Loudness == nil || *resp.Loudness != -14 {
		t.Errorf("loudness should be preserved -14, got %v", resp.Loudness)
	}
}

// TestVolumeNormalizeSetting_LoudnessOutOfRange PUT 越界响度返回 400，且不改动既有配置。
func TestVolumeNormalizeSetting_LoudnessOutOfRange(t *testing.T) {
	h := newTestSongHandlerForVolumeNormalize(t)
	// 先置一个合法值
	rr0 := httptest.NewRecorder()
	h.UpdateVolumeNormalizeSetting(rr0, httptest.NewRequest("PUT", "/api/v1/settings/volume-normalize",
		strings.NewReader(`{"enabled":true,"loudness":-14}`)))
	if rr0.Code != http.StatusOK {
		t.Fatalf("setup PUT: got %d", rr0.Code)
	}

	for _, body := range []string{
		`{"enabled":true,"loudness":-50}`,
		`{"enabled":true,"loudness":0}`,
		`{"enabled":true,"loudness":5}`,
	} {
		rr := httptest.NewRecorder()
		h.UpdateVolumeNormalizeSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/volume-normalize",
			strings.NewReader(body)))
		if rr.Code != http.StatusBadRequest {
			t.Errorf("PUT %s: got %d want 400, body=%s", body, rr.Code, rr.Body.String())
		}
	}

	// 越界请求不应改动响度：仍为 -14
	rr2 := httptest.NewRecorder()
	h.GetVolumeNormalizeSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/volume-normalize", nil))
	var got volumeNormalizeRequest
	if err := json.Unmarshal(rr2.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Loudness == nil || *got.Loudness != -14 {
		t.Errorf("越界 PUT 后 loudness 应保持 -14, got %v", got.Loudness)
	}
}

// TestVolumeNormalizeSetting_BadJSON 请求体非法返回 400，不改状态。
func TestVolumeNormalizeSetting_BadJSON(t *testing.T) {
	h := newTestSongHandlerForVolumeNormalize(t)
	rr := httptest.NewRecorder()
	h.UpdateVolumeNormalizeSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/volume-normalize",
		strings.NewReader(`{not json`)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad JSON status: got %d want 400, body=%s", rr.Code, rr.Body.String())
	}
}

// TestVolumeNormalizeSetting_NilServices configService 未注入时 PUT 返回 500（不 panic）；
// GET 在 cacheService 也为 nil 时仍回落默认 -16（NormalizeLoudness 是 nil-safety 方法）。
func TestVolumeNormalizeSetting_NilServices(t *testing.T) {
	h := NewSongHandler(nil, nil, nil, nil, nil, nil)

	rr := httptest.NewRecorder()
	h.GetVolumeNormalizeSetting(rr, httptest.NewRequest("GET", "/api/v1/settings/volume-normalize", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("GET nil services status: got %d want 200", rr.Code)
	}
	var resp volumeNormalizeRequest
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if resp.Loudness == nil || *resp.Loudness != -16 {
		t.Errorf("nil services GET loudness: got %v want -16", resp.Loudness)
	}

	rr2 := httptest.NewRecorder()
	h.UpdateVolumeNormalizeSetting(rr2, httptest.NewRequest("PUT", "/api/v1/settings/volume-normalize",
		strings.NewReader(`{"enabled":true}`)))
	if rr2.Code != http.StatusInternalServerError {
		t.Errorf("PUT nil configService: got %d want 500", rr2.Code)
	}
}
