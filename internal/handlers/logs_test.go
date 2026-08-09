package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogExportHandler_ExportsRecentCompleteLines(t *testing.T) {
	dir := t.TempDir()
	older := "drop-one\ndrop-two password=secret\nkeep-tail\n"
	latest := "latest-one\nlatest-two\n"
	if err := os.WriteFile(filepath.Join(dir, "songloft-2026-08-08.log"), []byte(older), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "songloft-2026-08-09.log"), []byte(latest), 0o644); err != nil {
		t.Fatal(err)
	}

	h := NewLogExportHandler(dir)
	// 从旧文件的第二行中间开始，处理器应丢弃残行，并保留后续完整行。
	h.maxExportSize = int64(len(latest) + len("keep-tail\n") + 8)
	recorder := httptest.NewRecorder()
	h.ExportLogs(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/logs/export", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, want := range []string{"日志已截断", "keep-tail", "latest-one", "latest-two"} {
		if !strings.Contains(body, want) {
			t.Errorf("响应缺少 %q：%s", want, body)
		}
	}
	for _, unwanted := range []string{"drop-one", "drop-two", "secret"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("响应不应包含 %q：%s", unwanted, body)
		}
	}
}

func TestLogExportHandler_RedactsUntruncatedLogs(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "songloft-2026-08-09.log"),
		[]byte("token=secret normal=value\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	NewLogExportHandler(dir).ExportLogs(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/v1/logs/export", nil),
	)
	body := recorder.Body.String()
	if strings.Contains(body, "secret") || !strings.Contains(body, "token=***") {
		t.Fatalf("日志未正确脱敏：%s", body)
	}
	if strings.Contains(body, "日志已截断") {
		t.Fatalf("小日志不应标记截断：%s", body)
	}
}
