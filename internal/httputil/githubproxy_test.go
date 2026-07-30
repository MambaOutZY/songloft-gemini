package httputil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestApplyGithubProxy(t *testing.T) {
	cases := []struct {
		name, rawURL, proxy, want string
	}{
		{"空代理", "https://github.com/a/b", "", "https://github.com/a/b"},
		{"非GitHub不套", "https://example.com/a", "https://p.com/", "https://example.com/a"},
		{"自动补斜杠", "https://github.com/a/b", "https://p.com", "https://p.com/https://github.com/a/b"},
		{"raw域名", "https://raw.githubusercontent.com/a/b", "https://p.com/", "https://p.com/https://raw.githubusercontent.com/a/b"},
		{"github.io", "https://foo.github.io/x", "https://p.com/", "https://p.com/https://foo.github.io/x"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ApplyGithubProxy(c.rawURL, c.proxy); got != c.want {
				t.Errorf("ApplyGithubProxy(%q, %q) = %q, want %q", c.rawURL, c.proxy, got, c.want)
			}
		})
	}
}

func TestIsGitHubURL(t *testing.T) {
	if !IsGitHubURL("https://objects.githubusercontent.com/x") {
		t.Error("objects.githubusercontent.com should be GitHub URL")
	}
	if IsGitHubURL("https://evil.com/https://github.com/") {
		t.Error("non-GitHub host should not match")
	}
}

func mustRead(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

func TestFallbackOnProxyConnError(t *testing.T) {
	var directHits atomic.Int32
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		directHits.Add(1)
		_, _ = w.Write([]byte("ok"))
	}))
	defer direct.Close()

	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := dead.URL
	dead.Close()

	resp, err := getWithFallback(context.Background(), direct.Client(), direct.URL, deadURL, "proxy", GithubGetOptions{})
	if err != nil {
		t.Fatalf("expected fallback success, got error: %v", err)
	}
	if got := mustRead(t, resp); got != "ok" {
		t.Errorf("body = %q, want ok", got)
	}
	if directHits.Load() != 1 {
		t.Errorf("direct hits = %d, want 1", directHits.Load())
	}
}

func TestFallbackOnProxyNon2xx(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer proxy.Close()
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer direct.Close()

	resp, err := getWithFallback(context.Background(), direct.Client(), direct.URL, proxy.URL, "proxy", GithubGetOptions{})
	if err != nil {
		t.Fatalf("expected fallback success, got error: %v", err)
	}
	if got := mustRead(t, resp); got != "ok" {
		t.Errorf("body = %q, want ok", got)
	}
}

func TestNoFallbackWithoutProxy(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	resp, err := getWithFallback(context.Background(), srv.Client(), srv.URL, srv.URL, "", GithubGetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
	if hits.Load() != 1 {
		t.Errorf("hits = %d, want 1", hits.Load())
	}
}

func TestDirectResultReturnedAsIs(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := dead.URL
	dead.Close()

	var directHits atomic.Int32
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		directHits.Add(1)
		http.NotFound(w, r)
	}))
	defer direct.Close()

	resp, err := getWithFallback(context.Background(), direct.Client(), direct.URL, deadURL, "proxy", GithubGetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	if directHits.Load() != 1 {
		t.Errorf("direct hits = %d, want 1", directHits.Load())
	}
}

func TestProxyDownMemo(t *testing.T) {
	var proxyHits atomic.Int32
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := dead.URL
	dead.Close()

	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer direct.Close()

	var down atomic.Bool
	opts := GithubGetOptions{ProxyDown: &down}

	resp, err := getWithFallback(context.Background(), direct.Client(), direct.URL, deadURL, "proxy", opts)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	mustRead(t, resp)
	if !down.Load() {
		t.Fatal("ProxyDown should be set after transport error")
	}

	// 第二次调用：代理侧换成可计数的活 server，验证被跳过
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyHits.Add(1)
		_, _ = w.Write([]byte("proxied"))
	}))
	defer proxy.Close()
	resp, err = getWithFallback(context.Background(), direct.Client(), direct.URL, proxy.URL, "proxy", opts)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if got := mustRead(t, resp); got != "ok" {
		t.Errorf("body = %q, want ok (direct)", got)
	}
	if proxyHits.Load() != 0 {
		t.Errorf("proxy hits = %d, want 0", proxyHits.Load())
	}
}

func TestProxyDownNotSetOn4xx(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer proxy.Close()
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer direct.Close()

	var down atomic.Bool
	resp, err := getWithFallback(context.Background(), direct.Client(), direct.URL, proxy.URL, "proxy", GithubGetOptions{ProxyDown: &down})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustRead(t, resp)
	if down.Load() {
		t.Error("ProxyDown should not be set on 4xx")
	}
}

func TestProxyDownSetOn5xx(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer proxy.Close()
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer direct.Close()

	var down atomic.Bool
	resp, err := getWithFallback(context.Background(), direct.Client(), direct.URL, proxy.URL, "proxy", GithubGetOptions{ProxyDown: &down})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustRead(t, resp)
	if !down.Load() {
		t.Error("ProxyDown should be set on proxy 5xx")
	}
}

func TestAttemptTimeout(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		_, _ = w.Write([]byte("late"))
	}))
	defer proxy.Close()
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok-direct-body"))
	}))
	defer direct.Close()

	resp, err := getWithFallback(context.Background(), direct.Client(), direct.URL, proxy.URL, "proxy",
		GithubGetOptions{AttemptTimeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatalf("expected fallback success, got error: %v", err)
	}
	if got := mustRead(t, resp); got != "ok-direct-body" {
		t.Errorf("body = %q, want ok-direct-body", got)
	}
}
