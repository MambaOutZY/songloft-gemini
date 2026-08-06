package jsplugin

import (
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
