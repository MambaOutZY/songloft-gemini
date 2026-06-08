package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"songloft/internal/services"
)

// TestAutoScanSetting_GetDefault GET 在配置缺失时返回业务默认值 {enabled: false, interval_seconds: 3600}。
func TestAutoScanSetting_GetDefault(t *testing.T) {
	h := newTestScanHandlerWithConfig(t)

	rr := httptest.NewRecorder()
	h.GetAutoScanSetting(rr, httptest.NewRequest("GET", "/api/v1/settings/auto-scan", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp AutoScanSetting
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Enabled != false {
		t.Errorf("default enabled: got %v want false", resp.Enabled)
	}
	if resp.IntervalSeconds != 3600 {
		t.Errorf("default interval: got %d want 3600", resp.IntervalSeconds)
	}
}

// TestAutoScanSetting_UpdateThenRead PUT 写入后 GET 读到最新值。
func TestAutoScanSetting_UpdateThenRead(t *testing.T) {
	h := newTestScanHandlerWithConfig(t)

	rr := httptest.NewRecorder()
	h.UpdateAutoScanSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/auto-scan",
		strings.NewReader(`{"enabled":true,"interval_seconds":600}`)))
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}

	rr2 := httptest.NewRecorder()
	h.GetAutoScanSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/auto-scan", nil))
	var resp AutoScanSetting
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Enabled {
		t.Error("enabled should be true after PUT")
	}
	if resp.IntervalSeconds != 600 {
		t.Errorf("interval: got %d want 600", resp.IntervalSeconds)
	}
}

// TestAutoScanSetting_InvalidInterval interval_seconds 超出范围时返回 400。
func TestAutoScanSetting_InvalidInterval(t *testing.T) {
	h := newTestScanHandlerWithConfig(t)

	for _, tc := range []struct {
		name string
		body string
	}{
		{"too_small", `{"enabled":true,"interval_seconds":10}`},
		{"too_large", `{"enabled":true,"interval_seconds":100000}`},
		{"zero", `{"enabled":true,"interval_seconds":0}`},
		{"negative", `{"enabled":true,"interval_seconds":-1}`},
	} {
		rr := httptest.NewRecorder()
		h.UpdateAutoScanSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/auto-scan",
			strings.NewReader(tc.body)))
		if rr.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d want 400", tc.name, rr.Code)
		}
	}
}

// TestAutoScanSetting_BadJSON 请求体非法时返回 400。
func TestAutoScanSetting_BadJSON(t *testing.T) {
	h := newTestScanHandlerWithConfig(t)

	rr := httptest.NewRecorder()
	h.UpdateAutoScanSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/auto-scan",
		strings.NewReader(`not json`)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("bad JSON: got %d want 400", rr.Code)
	}
}

// TestAutoScanSetting_CallbackFired PUT 成功后异步触发 onAutoScanChanged 回调。
func TestAutoScanSetting_CallbackFired(t *testing.T) {
	h := newTestScanHandlerWithConfig(t)

	called := make(chan struct{}, 1)
	var callCount atomic.Int32
	h.SetOnAutoScanChanged(func(cfg services.AutoScanConfig) {
		callCount.Add(1)
		select {
		case called <- struct{}{}:
		default:
		}
	})

	rr := httptest.NewRecorder()
	h.UpdateAutoScanSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/auto-scan",
		strings.NewReader(`{"enabled":true,"interval_seconds":3600}`)))
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT status: got %d", rr.Code)
	}

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("onAutoScanChanged callback should fire after PUT")
	}
	if callCount.Load() != 1 {
		t.Errorf("callback should fire exactly once, got %d", callCount.Load())
	}
}
