package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"songloft/internal/services"
)

// TestNewScanHandler 测试创建扫描处理器
func TestNewScanHandler(t *testing.T) {
	repo := newTestSongRepo(t)
	songService := services.NewSongService(repo, nil, nil, nil, nil, nil)
	handler := NewScanHandler(songService, nil, nil)

	if handler == nil {
		t.Error("NewScanHandler() returned nil")
	}

	if handler.songService == nil {
		t.Error("NewScanHandler() songService should not be nil")
	}
}

// TestScanHandlerStructure 测试扫描处理器结构
func TestScanHandlerStructure(t *testing.T) {
	repo := newTestSongRepo(t)
	songService := services.NewSongService(repo, nil, nil, nil, nil, nil)
	handler := NewScanHandler(songService, nil, nil)

	// 验证处理器结构正确
	if handler.songService != songService {
		t.Error("ScanHandler songService should match the provided service")
	}
}

// TestScanAndImportSuccess 测试成功的扫描导入
func TestScanAndImportSuccess(t *testing.T) {
	repo := newTestSongRepo(t)

	// 创建临时测试目录
	tempDir := t.TempDir()

	// 创建 scanner 和 extractor
	scanner := services.NewScanner(&services.ScanConfig{
		MusicPath:        tempDir,
		ExcludeDirs:      []string{},
		SupportedFormats: []string{"mp3", "flac"},
	})
	extractor := services.NewMetadataExtractor(&services.MetadataConfig{
		FFProbePath: "ffprobe",
	})

	songService := services.NewSongService(repo, nil, extractor, scanner, nil, nil)
	handler := NewScanHandler(songService, scanner, nil)

	req := httptest.NewRequest("POST", "/api/v1/scan", nil)
	rr := httptest.NewRecorder()

	handler.ScanAndImport(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 异步扫描立即返回"扫描任务已启动"
	if response["message"] != "扫描任务已启动" {
		t.Errorf("handler returned wrong message: got %v want 扫描任务已启动", response["message"])
	}
}

// TestScanAndImportError 测试扫描导入失败
// 注意：异步扫描即使路径不存在也会立即返回 200 OK，错误在异步处理中
func TestScanAndImportError(t *testing.T) {
	repo := newTestSongRepo(t)

	// 创建会返回错误的 scanner（传入不存在的路径）
	scanner := services.NewScanner(&services.ScanConfig{
		MusicPath:        "/nonexistent/path/that/does/not/exist",
		ExcludeDirs:      []string{},
		SupportedFormats: []string{"mp3"},
	})
	extractor := services.NewMetadataExtractor(&services.MetadataConfig{
		FFProbePath: "ffprobe",
	})

	songService := services.NewSongService(repo, nil, extractor, scanner, nil, nil)
	handler := NewScanHandler(songService, scanner, nil)

	req := httptest.NewRequest("POST", "/api/v1/scan", nil)
	rr := httptest.NewRecorder()

	handler.ScanAndImport(rr, req)

	// 异步扫描即使路径不存在也会返回 200，错误在异步任务中处理
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
}

// TestScanAndImportRejectsPathOutsideMusicRoot 定向扫描目录越界应返回 400（防目录遍历）。
func TestScanAndImportRejectsPathOutsideMusicRoot(t *testing.T) {
	repo := newTestSongRepo(t)
	tempDir := t.TempDir()
	scanner := services.NewScanner(&services.ScanConfig{
		MusicPath:        tempDir,
		SupportedFormats: []string{"mp3"},
	})
	songService := services.NewSongService(repo, nil, nil, scanner, nil, nil)
	handler := NewScanHandler(songService, scanner, nil)

	req := httptest.NewRequest("POST", "/api/v1/scan",
		strings.NewReader(`{"paths":["/etc/passwd"]}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ScanAndImport(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("out-of-root path: got status %d, want 400", rr.Code)
	}
}

// TestSanitizeScanPaths 校验去重、剔除冗余子目录、越界拒绝。
func TestSanitizeScanPaths(t *testing.T) {
	root := t.TempDir()
	scanner := services.NewScanner(&services.ScanConfig{MusicPath: root})
	handler := NewScanHandler(nil, scanner, nil)

	a := filepath.Join(root, "A")
	aSub := filepath.Join(root, "A", "sub")
	b := filepath.Join(root, "B")

	// 空列表 → nil（全库扫描）。
	if got, err := handler.sanitizeScanPaths(nil); err != nil || got != nil {
		t.Errorf("empty paths => (%v, %v), want (nil, nil)", got, err)
	}

	// 去重 + 剔除被祖先覆盖的子目录：{A, A, A/sub, B} → {A, B}。
	got, err := handler.sanitizeScanPaths([]string{a, a, aSub, b})
	if err != nil {
		t.Fatalf("sanitizeScanPaths error: %v", err)
	}
	set := map[string]bool{}
	for _, p := range got {
		set[p] = true
	}
	if len(got) != 2 || !set[a] || !set[b] {
		t.Errorf("sanitizeScanPaths = %v, want [%s %s]", got, a, b)
	}

	// 越界路径 → error。
	if _, err := handler.sanitizeScanPaths([]string{filepath.Join(root, "..", "evil")}); err == nil {
		t.Error("out-of-root path should return error")
	}

	// 明确传了 paths 但全是空白 → error（不静默退化为全库扫描/清理）。
	if _, err := handler.sanitizeScanPaths([]string{"", "  "}); err == nil {
		t.Error("all-blank paths should return error, not silently degrade to full scan")
	}

	// 真正的空列表（未传 paths）→ 全库扫描（nil, nil），不报错。
	if got, err := handler.sanitizeScanPaths([]string{}); err != nil || got != nil {
		t.Errorf("empty slice => (%v, %v), want (nil, nil)", got, err)
	}
}
