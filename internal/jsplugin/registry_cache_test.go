package jsplugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// countingRegistry 返回一个 registry handler 与它被请求的次数计数器。
func countingRegistry(pluginURLs ...string) (http.HandlerFunc, *atomic.Int32) {
	var hits atomic.Int32
	return func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		json.NewEncoder(w).Encode(RegistryJSON{Plugins: pluginURLs})
	}, &hits
}

// TestFetchAndMergeCached_ReusesResult 验证 TTL 内重复调用只拉取一次远端。
// 这是商店翻页/搜索的场景：以前每翻一页都会重拉整棵注册表树。
func TestFetchAndMergeCached_ReusesResult(t *testing.T) {
	pluginMux := http.NewServeMux()
	pluginMux.HandleFunc("/a/plugin.json", servePluginJSON("Plugin A", "plugin-a", "1.0.0", "https://example.com/a.zip"))
	pluginSrv := httptest.NewServer(pluginMux)
	defer pluginSrv.Close()

	handler, hits := countingRegistry(pluginSrv.URL + "/a/plugin.json")
	srv := httptest.NewServer(handler)
	defer srv.Close()

	svc := NewRegistryService()
	for i := range 5 {
		plugins, _, err := svc.FetchAndMergeCached(context.Background(), srv.URL, "", "", false)
		if err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
		if len(plugins) != 1 {
			t.Fatalf("iter %d: expected 1 plugin, got %d", i, len(plugins))
		}
	}
	if got := hits.Load(); got != 1 {
		t.Errorf("expected registry fetched once across 5 cached calls, got %d", got)
	}
}

// TestFetchAndMergeCached_ForceBypasses 验证 force=true 绕过缓存（刷新按钮）。
func TestFetchAndMergeCached_ForceBypasses(t *testing.T) {
	pluginMux := http.NewServeMux()
	pluginMux.HandleFunc("/a/plugin.json", servePluginJSON("Plugin A", "plugin-a", "1.0.0", "https://example.com/a.zip"))
	pluginSrv := httptest.NewServer(pluginMux)
	defer pluginSrv.Close()

	handler, hits := countingRegistry(pluginSrv.URL + "/a/plugin.json")
	srv := httptest.NewServer(handler)
	defer srv.Close()

	svc := NewRegistryService()
	if _, _, err := svc.FetchAndMergeCached(context.Background(), srv.URL, "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.FetchAndMergeCached(context.Background(), srv.URL, "", "", true); err != nil {
		t.Fatal(err)
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("expected 2 fetches (1 cold + 1 forced), got %d", got)
	}
	// force 之后应回填缓存，下一次非 force 调用不再拉取
	if _, _, err := svc.FetchAndMergeCached(context.Background(), srv.URL, "", "", false); err != nil {
		t.Fatal(err)
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("forced fetch should repopulate cache, got %d fetches", got)
	}
}

// TestFetchAndMergeCached_TokenIsolatesCache 验证不同 token 不共享缓存条目：
// 换 token 可能看到不同的插件集合，串用会泄露另一个身份可见的列表。
func TestFetchAndMergeCached_TokenIsolatesCache(t *testing.T) {
	pluginMux := http.NewServeMux()
	pluginMux.HandleFunc("/a/plugin.json", servePluginJSON("Plugin A", "plugin-a", "1.0.0", "https://example.com/a.zip"))
	pluginSrv := httptest.NewServer(pluginMux)
	defer pluginSrv.Close()

	handler, hits := countingRegistry(pluginSrv.URL + "/a/plugin.json")
	srv := httptest.NewServer(handler)
	defer srv.Close()

	svc := NewRegistryService()
	for _, token := range []string{"", "token-a", "token-b", "token-a"} {
		if _, _, err := svc.FetchAndMergeCached(context.Background(), srv.URL, "", token, false); err != nil {
			t.Fatal(err)
		}
	}
	// 三个不同 token（含空）各拉一次，重复的 token-a 命中缓存
	if got := hits.Load(); got != 3 {
		t.Errorf("expected 3 fetches for 3 distinct tokens, got %d", got)
	}
}

// TestFetchAndMergeCached_ErrorNotCached 验证失败不写缓存，下次请求会重试
// （而不是把错误状态粘住一个 TTL）。
func TestFetchAndMergeCached_ErrorNotCached(t *testing.T) {
	var hits atomic.Int32
	var failing atomic.Bool
	failing.Store(true)

	pluginMux := http.NewServeMux()
	pluginMux.HandleFunc("/a/plugin.json", servePluginJSON("Plugin A", "plugin-a", "1.0.0", "https://example.com/a.zip"))
	pluginSrv := httptest.NewServer(pluginMux)
	defer pluginSrv.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if failing.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(RegistryJSON{Plugins: []string{pluginSrv.URL + "/a/plugin.json"}})
	}))
	defer srv.Close()

	svc := NewRegistryService()
	if _, _, err := svc.FetchAndMergeCached(context.Background(), srv.URL, "", "", false); err == nil {
		t.Fatal("expected error on first fetch")
	}
	failing.Store(false)
	plugins, _, err := svc.FetchAndMergeCached(context.Background(), srv.URL, "", "", false)
	if err != nil {
		t.Fatalf("retry should succeed: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin after retry, got %d", len(plugins))
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("expected the failed fetch to be retried, got %d fetches", got)
	}
}

// TestFetchAndMergeMultiCached_ReusesResult 验证「全部」聚合模式同样走缓存。
func TestFetchAndMergeMultiCached_ReusesResult(t *testing.T) {
	pluginMux := http.NewServeMux()
	pluginMux.HandleFunc("/a/plugin.json", servePluginJSONFull(
		"Demo A", "demo-a", "1.0.0", "https://example.com/a.zip", "Alice", ""))
	pluginMux.HandleFunc("/b/plugin.json", servePluginJSONFull(
		"Demo B", "demo-b", "1.0.0", "https://example.com/b.zip", "Bob", ""))
	pluginSrv := httptest.NewServer(pluginMux)
	defer pluginSrv.Close()

	handlerA, hitsA := countingRegistry(pluginSrv.URL + "/a/plugin.json")
	srvA := httptest.NewServer(handlerA)
	defer srvA.Close()
	handlerB, hitsB := countingRegistry(pluginSrv.URL + "/b/plugin.json")
	srvB := httptest.NewServer(handlerB)
	defer srvB.Close()

	sources := []RegistryConfig{
		{URL: srvA.URL, Enabled: true},
		{URL: srvB.URL, Enabled: true},
	}
	svc := NewRegistryService()
	for i := range 4 {
		plugins, _ := svc.FetchAndMergeMultiCached(context.Background(), sources, "", false)
		if len(plugins) != 2 {
			t.Fatalf("iter %d: expected 2 plugins, got %d", i, len(plugins))
		}
	}
	if hitsA.Load() != 1 || hitsB.Load() != 1 {
		t.Errorf("expected each source fetched once, got A=%d B=%d", hitsA.Load(), hitsB.Load())
	}
}

// TestFetchAndMergeMultiCached_SourceListIsolatesCache 验证源列表变化会换键：
// 增删源、改源顺序（顺序决定同版本插件由哪个源胜出）都必须重新拉取。
func TestFetchAndMergeMultiCached_SourceListIsolatesCache(t *testing.T) {
	pluginMux := http.NewServeMux()
	pluginMux.HandleFunc("/a/plugin.json", servePluginJSONFull(
		"Demo A", "demo-a", "1.0.0", "https://example.com/a.zip", "Alice", ""))
	pluginMux.HandleFunc("/b/plugin.json", servePluginJSONFull(
		"Demo B", "demo-b", "1.0.0", "https://example.com/b.zip", "Bob", ""))
	pluginSrv := httptest.NewServer(pluginMux)
	defer pluginSrv.Close()

	handlerA, hitsA := countingRegistry(pluginSrv.URL + "/a/plugin.json")
	srvA := httptest.NewServer(handlerA)
	defer srvA.Close()
	handlerB, _ := countingRegistry(pluginSrv.URL + "/b/plugin.json")
	srvB := httptest.NewServer(handlerB)
	defer srvB.Close()

	a := RegistryConfig{URL: srvA.URL, Enabled: true}
	b := RegistryConfig{URL: srvB.URL, Enabled: true}
	svc := NewRegistryService()

	svc.FetchAndMergeMultiCached(context.Background(), []RegistryConfig{a}, "", false)
	if hitsA.Load() != 1 {
		t.Fatalf("cold fetch: A=%d", hitsA.Load())
	}
	// 加了一个源 → 不同键 → 重新拉取
	svc.FetchAndMergeMultiCached(context.Background(), []RegistryConfig{a, b}, "", false)
	if hitsA.Load() != 2 {
		t.Errorf("adding a source should re-fetch, A=%d", hitsA.Load())
	}
	// 换顺序 → 不同键 → 重新拉取
	svc.FetchAndMergeMultiCached(context.Background(), []RegistryConfig{b, a}, "", false)
	if hitsA.Load() != 3 {
		t.Errorf("reordering sources should re-fetch, A=%d", hitsA.Load())
	}
}

// TestInvalidateCache 验证显式失效后重新拉取。
func TestInvalidateCache(t *testing.T) {
	pluginMux := http.NewServeMux()
	pluginMux.HandleFunc("/a/plugin.json", servePluginJSON("Plugin A", "plugin-a", "1.0.0", "https://example.com/a.zip"))
	pluginSrv := httptest.NewServer(pluginMux)
	defer pluginSrv.Close()

	handler, hits := countingRegistry(pluginSrv.URL + "/a/plugin.json")
	srv := httptest.NewServer(handler)
	defer srv.Close()

	svc := NewRegistryService()
	svc.FetchAndMergeCached(context.Background(), srv.URL, "", "", false)
	svc.InvalidateCache()
	svc.FetchAndMergeCached(context.Background(), srv.URL, "", "", false)
	if got := hits.Load(); got != 2 {
		t.Errorf("expected re-fetch after invalidation, got %d", got)
	}
}

// TestRegistryCacheEvictsWhenFull 验证缓存条目数不会无界增长。
func TestRegistryCacheEvictsWhenFull(t *testing.T) {
	svc := NewRegistryService()
	for i := range registryCacheMaxEntries * 2 {
		svc.cachePut(string(rune('a'+i%26))+string(rune(i)), []RegistryEntry{{EntryPath: "x"}}, nil)
	}
	svc.cacheMu.Lock()
	n := len(svc.cache)
	svc.cacheMu.Unlock()
	if n > registryCacheMaxEntries {
		t.Errorf("cache grew to %d entries, cap is %d", n, registryCacheMaxEntries)
	}
}

// TestRegistryCacheExpires 验证过期条目不会被返回。
func TestRegistryCacheExpires(t *testing.T) {
	svc := NewRegistryService()
	key := "k"
	svc.cachePut(key, []RegistryEntry{{EntryPath: "x"}}, nil)
	if svc.cacheGet(key) == nil {
		t.Fatal("expected fresh entry to be readable")
	}
	// 手动把时间戳推到 TTL 之外
	svc.cacheMu.Lock()
	svc.cache[key].fetchedAt = time.Now().Add(-registryCacheTTL - time.Second)
	svc.cacheMu.Unlock()
	if svc.cacheGet(key) != nil {
		t.Error("expected expired entry to be treated as a miss")
	}
}

// TestFetchAndMergeCached_ConcurrentReuse 验证 RegistryService 作为长生命周期
// 单例被并发复用时安全（配合 -race）。这是加缓存后新引入的前提：以前每个请求
// 新建一个实例，现在 handler 全程持有一个。
func TestFetchAndMergeCached_ConcurrentReuse(t *testing.T) {
	pluginMux := http.NewServeMux()
	for _, n := range []string{"a", "b", "c"} {
		pluginMux.HandleFunc("/"+n+"/plugin.json",
			servePluginJSON("Plugin "+n, "plugin-"+n, "1.0.0", "https://example.com/"+n+".zip"))
	}
	pluginSrv := httptest.NewServer(pluginMux)
	defer pluginSrv.Close()

	handler, _ := countingRegistry(
		pluginSrv.URL+"/a/plugin.json",
		pluginSrv.URL+"/b/plugin.json",
		pluginSrv.URL+"/c/plugin.json",
	)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	svc := NewRegistryService()
	const workers = 16
	errs := make(chan error, workers)
	for i := range workers {
		go func(i int) {
			// 混合 force 与非 force，让读写缓存交错
			plugins, _, err := svc.FetchAndMergeCached(context.Background(), srv.URL, "", "", i%4 == 0)
			if err != nil {
				errs <- err
				return
			}
			if len(plugins) != 3 {
				errs <- errUnexpectedCount(len(plugins))
				return
			}
			errs <- nil
		}(i)
	}
	for range workers {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent fetch: %v", err)
		}
	}
}

type errUnexpectedCount int

func (e errUnexpectedCount) Error() string {
	return "unexpected plugin count"
}
