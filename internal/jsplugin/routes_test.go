package jsplugin

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestInjectHTMLHeadAssetVersioning 验证注入的每个公共资源 URL 都带内容哈希版本号（#278）。
// 若丢失版本号，jsplugin-assets 的 immutable 长缓存会让老浏览器永远拿不到资源更新。
func TestInjectHTMLHeadAssetVersioning(t *testing.T) {
	out := string(injectHTMLHead([]byte("<head></head><body></body>"), "demo", ""))

	// 五个公共资源（theme.css / components.css / common.js / webf-shims.css /
	// webf-shims.js）均应带 ?v=<8位hex>，且版本号与嵌入内容哈希一致。
	for _, name := range []string{"theme.css", "components.css", "common.js", "webf-shims.css", "webf-shims.js"} {
		re := regexp.MustCompile(regexp.QuoteMeta(name) + `\?v=[0-9a-f]{8}"`)
		if !re.MatchString(out) {
			t.Errorf("注入的 %s 缺少版本号 (?v=hash)，实际输出:\n%s", name, out)
		}
		if v := assetVersions[name]; v == "" || !strings.Contains(out, name+"?v="+v) {
			t.Errorf("%s 版本号与嵌入内容哈希不一致: got version %q", name, v)
		}
	}
}

// TestAssetURLFallback 无对应资源版本时回退到无版本 URL，不产生裸 "?v=" 尾巴。
func TestAssetURLFallback(t *testing.T) {
	got := assetURL("/base/", "does-not-exist.css")
	if got != "/base/does-not-exist.css" {
		t.Errorf("未知资源应回退无版本 URL，got %q", got)
	}
}

// TestTryServeStaticFileCOEPHeader 验证插件 HTML 响应携带 COEP: credentialless 头。
//
// Lynx Web 宿主页必须开 COEP: require-corp（web-core 的 SharedArrayBuffer 依赖），
// 而 COEP 页面里的跨源 iframe 要求其文档响应自身也带 COEP 头——缺失时
// Chrome/Firefox 直接 ERR_BLOCKED_BY_RESPONSE，插件 tab 呈空白且无应用层报错
// （corpMiddleware 的 CORP 头只覆盖子资源，盖不住 iframe 文档这一层）。
// 同时验证非 HTML 资源不受影响（强缓存路径保持原样）。
func TestTryServeStaticFileCOEPHeader(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)

	staticRoot := filepath.Join(dataDir, "demo", "static")
	if err := os.MkdirAll(staticRoot, 0o755); err != nil {
		t.Fatalf("create static dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staticRoot, "index.html"), []byte("<head></head><body>hi</body>"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staticRoot, "style.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatalf("write style.css: %v", err)
	}

	manager := NewManager(repo, pluginsDir, dataDir, "", nil, nil)
	t.Cleanup(func() { manager.Close() })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jsplugin/demo/", nil)
	if !manager.tryServeStaticFile(rec, req, staticRoot, "index.html", "demo") {
		t.Fatal("expected index.html to be served")
	}
	if got := rec.Header().Get("Cross-Origin-Embedder-Policy"); got != "credentialless" {
		t.Errorf("插件 HTML 响应应带 COEP: credentialless（Web 宿主页 iframe 嵌入依赖），got %q", got)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/jsplugin/demo/static/style.css", nil)
	if !manager.tryServeStaticFile(rec2, req2, staticRoot, "style.css", "demo") {
		t.Fatal("expected style.css to be served")
	}
	if got := rec2.Header().Get("Cross-Origin-Embedder-Policy"); got != "" {
		t.Errorf("非 HTML 资源不应带 COEP 头，got %q", got)
	}
}
