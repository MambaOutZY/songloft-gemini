package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"songloft/internal/database/testutil"
	"songloft/internal/services"
)

// newTestLogHandler 构造一个带真实 ConfigService 和独立 LevelVar 的 LogHandler。
// LevelVar 初始为 LevelInfo（slog.LevelVar 零值），便于断言 PUT 后的等级变化。
func newTestLogHandler(t *testing.T) (*LogHandler, *slog.LevelVar) {
	t.Helper()
	mdb := testutil.OpenMemoryDB(t)
	configService := services.NewConfigService(mdb.ConfigRepository())
	levelVar := new(slog.LevelVar)
	return NewLogHandler(configService, levelVar), levelVar
}

// TestLogLevel_DefaultInfo 无任何写入时 GET 返回 level=info（业务默认值）。
func TestLogLevel_DefaultInfo(t *testing.T) {
	h, _ := newTestLogHandler(t)

	rr := httptest.NewRecorder()
	h.GetLevelSetting(rr, httptest.NewRequest("GET", "/api/v1/settings/log-level", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if resp["level"] != "info" {
		t.Errorf("default level: got %q want \"info\"", resp["level"])
	}
}

// TestLogLevel_UpdateAppliesAndPersists PUT 一个合法等级后，
// LevelVar 即时切换 + DB 持久化（后续 GET 读到同值）。
func TestLogLevel_UpdateAppliesAndPersists(t *testing.T) {
	h, levelVar := newTestLogHandler(t)

	cases := []struct {
		name string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"info", slog.LevelInfo},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.UpdateLevelSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/log-level",
				strings.NewReader(`{"level":"`+c.name+`"}`)))
			if rr.Code != http.StatusOK {
				t.Fatalf("PUT %s: status %d want 200, body=%s", c.name, rr.Code, rr.Body.String())
			}
			if levelVar.Level() != c.want {
				t.Errorf("PUT %s: levelVar got %v want %v", c.name, levelVar.Level(), c.want)
			}

			// GET 应返回同值
			rr2 := httptest.NewRecorder()
			h.GetLevelSetting(rr2, httptest.NewRequest("GET", "/api/v1/settings/log-level", nil))
			var resp map[string]string
			if err := json.Unmarshal(rr2.Body.Bytes(), &resp); err != nil {
				t.Fatalf("GET decode: %v", err)
			}
			if resp["level"] != c.name {
				t.Errorf("read-after-write: got %q want %q", resp["level"], c.name)
			}
		})
	}
}

// TestLogLevel_InvalidValue 非法等级值返回 400，且不改 LevelVar / 不写库。
func TestLogLevel_InvalidValue(t *testing.T) {
	h, levelVar := newTestLogHandler(t)

	rr := httptest.NewRecorder()
	h.UpdateLevelSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/log-level",
		strings.NewReader(`{"level":"trace"}`)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("invalid level: got %d want 400", rr.Code)
	}
	if levelVar.Level() != slog.LevelInfo {
		t.Errorf("LevelVar should not change on invalid input: got %v", levelVar.Level())
	}
	if h.GetLevel() != "info" {
		t.Errorf("GetLevel should still return info, got %q", h.GetLevel())
	}
}

// TestLogLevel_BadJSON 请求体非法 JSON 返回 400。
func TestLogLevel_BadJSON(t *testing.T) {
	h, levelVar := newTestLogHandler(t)

	rr := httptest.NewRecorder()
	h.UpdateLevelSetting(rr, httptest.NewRequest("PUT", "/api/v1/settings/log-level",
		strings.NewReader(`not json`)))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("bad JSON: got %d want 400", rr.Code)
	}
	if levelVar.Level() != slog.LevelInfo {
		t.Error("LevelVar should not change on bad JSON")
	}
}

// TestLogLevel_NilConfigService GetLevel 在 configService=nil 时返回默认 info；SetLevel 报错。
func TestLogLevel_NilConfigService(t *testing.T) {
	h := NewLogHandler(nil, nil)
	if got := h.GetLevel(); got != "info" {
		t.Errorf("nil configService GetLevel: got %q want \"info\"", got)
	}
	if err := h.SetLevel("debug"); err == nil {
		t.Error("SetLevel with nil configService should error")
	}
}

// TestParseLogLevel 覆盖所有合法值与典型非法值。
func TestParseLogLevel(t *testing.T) {
	valid := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	for s, want := range valid {
		got, ok := ParseLogLevel(s)
		if !ok || got != want {
			t.Errorf("ParseLogLevel(%q): got (%v, %v), want (%v, true)", s, got, ok, want)
		}
	}
	for _, s := range []string{"", "trace", "DEBUG", "Info", "fatal", "warning"} {
		if _, ok := ParseLogLevel(s); ok {
			t.Errorf("ParseLogLevel(%q) should be invalid", s)
		}
	}
}
