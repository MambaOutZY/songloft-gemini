package jsplugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIsValidRenderEngine 覆盖 renderEngine 的合法与非法取值。
// 空串必须合法：语义是「跟随宿主默认」，绝大多数插件不写这个字段。
func TestIsValidRenderEngine(t *testing.T) {
	valid := []string{"", RenderEngineWebView, RenderEngineWebF}
	for _, v := range valid {
		if !IsValidRenderEngine(v) {
			t.Errorf("IsValidRenderEngine(%q) = false, want true", v)
		}
	}

	invalid := []string{"WebView", "WEBF", "web_view", "webkit", "flutter", " webf", "webf "}
	for _, v := range invalid {
		if IsValidRenderEngine(v) {
			t.Errorf("IsValidRenderEngine(%q) = true, want false", v)
		}
	}
}

// TestValidateManifest_RenderEngine 验证 ValidateManifest 对 renderEngine 的取值校验，
// 且非法取值的错误信息里必须带上实际收到的值（否则排查插件打包问题只能猜）。
func TestValidateManifest_RenderEngine(t *testing.T) {
	// 合法取值（含缺省）
	for _, v := range []string{"", RenderEngineWebView, RenderEngineWebF} {
		m := testManifest("demo")
		m.RenderEngine = v
		// entryHash / zipHash 是必填项，这里只关心 renderEngine，故补上占位合法 hash。
		m.EntryHash = strings.Repeat("a", 64)
		m.ZipHash = strings.Repeat("b", 64)
		if err := ValidateManifest(m); err != nil {
			t.Errorf("ValidateManifest with renderEngine=%q: unexpected error %v", v, err)
		}
	}

	// 非法取值
	m := testManifest("demo")
	m.RenderEngine = "Webview"
	m.EntryHash = strings.Repeat("a", 64)
	m.ZipHash = strings.Repeat("b", 64)
	err := ValidateManifest(m)
	if err == nil {
		t.Fatal("expected error for renderEngine=Webview, got nil")
	}
	if !strings.Contains(err.Error(), "Webview") {
		t.Errorf("error must include the offending value, got: %v", err)
	}
	if !strings.Contains(err.Error(), "renderEngine") {
		t.Errorf("error must mention renderEngine, got: %v", err)
	}
}

// TestInstall_PersistsRenderEngine 验证安装带 renderEngine 的插件后，
// 该声明确实落进了 DB 并能原值读回。
func TestInstall_PersistsRenderEngine(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	pm := NewPackageManager(pluginsDir, dataDir, repo)
	ctx := context.Background()

	manifest := testManifest("webf-demo")
	manifest.RenderEngine = RenderEngineWebF
	zipData := createTestPluginZip(t, manifest, simpleJSCode)

	installed, wasUpdate, err := pm.InstallFromUpload(zipData)
	if err != nil {
		t.Fatalf("InstallFromUpload: %v", err)
	}
	if wasUpdate {
		t.Fatal("expected fresh install, got update")
	}
	if installed.RenderEngine != RenderEngineWebF {
		t.Errorf("returned plugin render engine = %q, want %q", installed.RenderEngine, RenderEngineWebF)
	}

	fromDB, err := repo.GetByEntryPath(ctx, "webf-demo")
	if err != nil {
		t.Fatalf("GetByEntryPath: %v", err)
	}
	if fromDB.RenderEngine != RenderEngineWebF {
		t.Errorf("DB render engine = %q, want %q", fromDB.RenderEngine, RenderEngineWebF)
	}
}

// TestInstall_DefaultRenderEngineIsEmpty 验证插件未声明 renderEngine 时库里存空串，
// 而不是被回填成 "webview"。空串保留「跟随宿主默认」语义，将来改默认值无需再迁移。
func TestInstall_DefaultRenderEngineIsEmpty(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	pm := NewPackageManager(pluginsDir, dataDir, repo)
	ctx := context.Background()

	zipData := createTestPluginZip(t, testManifest("plain-demo"), simpleJSCode)
	if _, _, err := pm.InstallFromUpload(zipData); err != nil {
		t.Fatalf("InstallFromUpload: %v", err)
	}

	fromDB, err := repo.GetByEntryPath(ctx, "plain-demo")
	if err != nil {
		t.Fatalf("GetByEntryPath: %v", err)
	}
	if fromDB.RenderEngine != "" {
		t.Errorf("DB render engine = %q, want empty string", fromDB.RenderEngine)
	}
}

// TestUpdate_PersistsRenderEngineChange 验证升级插件（代码变了，zipHash 变了）时
// 新 manifest 的 renderEngine 会覆盖旧值——包括「从 webf 改回 webview」这个方向。
func TestUpdate_PersistsRenderEngineChange(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	pm := NewPackageManager(pluginsDir, dataDir, repo)
	ctx := context.Background()

	first := testManifest("switch-demo")
	first.RenderEngine = RenderEngineWebF
	installed, _, err := pm.InstallFromUpload(createTestPluginZip(t, first, simpleJSCode))
	if err != nil {
		t.Fatalf("InstallFromUpload: %v", err)
	}

	second := testManifest("switch-demo")
	second.Version = "2.0.0"
	second.RenderEngine = RenderEngineWebView
	// 改动 JS 代码，让 zipHash 变化以走真正的 Update 路径
	if _, err := pm.Update(installed.ID, createTestPluginZip(t, second, simpleJSCode+"\n// v2\n")); err != nil {
		t.Fatalf("Update: %v", err)
	}

	fromDB, err := repo.GetByEntryPath(ctx, "switch-demo")
	if err != nil {
		t.Fatalf("GetByEntryPath: %v", err)
	}
	if fromDB.RenderEngine != RenderEngineWebView {
		t.Errorf("DB render engine after update = %q, want %q", fromDB.RenderEngine, RenderEngineWebView)
	}
}

// TestSyncFromDirectory_PersistsRenderEngine 覆盖「从磁盘 zip 重新同步」这条路径的两种情形：
//  1. 全新 zip 被自动安装 → renderEngine 落库
//  2. zipHash 不变但 plugin.json 改了 renderEngine → 必须被 syncManifestMetadata 补偿同步
//
// 情形 2 是真陷阱：ComputeCanonicalZipHash 刻意排除 plugin.json，
// 只改渲染引擎声明的重打包不会触发 Update，漏掉补偿的话声明会静默保持旧值。
func TestSyncFromDirectory_PersistsRenderEngine(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	pm := NewPackageManager(pluginsDir, dataDir, repo)
	ctx := context.Background()

	zipPath := filepath.Join(pluginsDir, "disk-demo.jsplugin.zip")
	first := testManifest("disk-demo")
	first.RenderEngine = RenderEngineWebF
	if err := os.WriteFile(zipPath, createTestPluginZip(t, first, simpleJSCode), 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	if _, err := pm.SyncPluginsFromDirectory(); err != nil {
		t.Fatalf("SyncPluginsFromDirectory: %v", err)
	}
	fromDB, err := repo.GetByEntryPath(ctx, "disk-demo")
	if err != nil {
		t.Fatalf("GetByEntryPath: %v", err)
	}
	if fromDB.RenderEngine != RenderEngineWebF {
		t.Fatalf("after first sync render engine = %q, want %q", fromDB.RenderEngine, RenderEngineWebF)
	}

	// 只改 plugin.json 的 renderEngine，JS 代码与 static 一字不动 → zipHash 不变
	second := testManifest("disk-demo")
	second.RenderEngine = RenderEngineWebView
	if err := os.WriteFile(zipPath, createTestPluginZip(t, second, simpleJSCode), 0o644); err != nil {
		t.Fatalf("rewrite zip: %v", err)
	}
	if _, err := pm.SyncPluginsFromDirectory(); err != nil {
		t.Fatalf("second SyncPluginsFromDirectory: %v", err)
	}

	fromDB, err = repo.GetByEntryPath(ctx, "disk-demo")
	if err != nil {
		t.Fatalf("GetByEntryPath after re-sync: %v", err)
	}
	if fromDB.RenderEngine != RenderEngineWebView {
		t.Errorf("after metadata-only re-sync render engine = %q, want %q",
			fromDB.RenderEngine, RenderEngineWebView)
	}
}
