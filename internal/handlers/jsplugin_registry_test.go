package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApplyGithubProxy(t *testing.T) {
	cases := []struct {
		name  string
		url   string
		proxy string
		want  string
	}{
		{
			name:  "无代理原样返回",
			url:   "https://api.github.com/repos/owner/repo/releases/tags/v1",
			proxy: "",
			want:  "https://api.github.com/repos/owner/repo/releases/tags/v1",
		},
		{
			name:  "非 GitHub URL 不加代理前缀",
			url:   "https://example.com/foo",
			proxy: "https://proxy.example.com",
			want:  "https://example.com/foo",
		},
		{
			name:  "api.github.com 加代理前缀",
			url:   "https://api.github.com/repos/owner/repo/releases/tags/v1",
			proxy: "https://proxy.example.com",
			want:  "https://proxy.example.com/https://api.github.com/repos/owner/repo/releases/tags/v1",
		},
		{
			name:  "代理前缀已带斜杠不重复拼接",
			url:   "https://raw.githubusercontent.com/owner/repo/main/registry.json",
			proxy: "https://proxy.example.com/",
			want:  "https://proxy.example.com/https://raw.githubusercontent.com/owner/repo/main/registry.json",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := applyGithubProxy(c.url, c.proxy)
			if got != c.want {
				t.Fatalf("applyGithubProxy(%q, %q) = %q, want %q", c.url, c.proxy, got, c.want)
			}
		})
	}
}

func TestDownloadGitHubReleaseAsset_ThroughProxy(t *testing.T) {
	const token = "secret-pat"
	var gotAuthHeaders []string

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeaders = append(gotAuthHeaders, r.Header.Get("Authorization"))

		// 代理路径形如 /https://api.github.com/repos/owner/repo/releases/tags/v1
		path := strings.TrimPrefix(r.URL.Path, "/")
		switch {
		case strings.Contains(path, "/releases/tags/v1"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"assets": []map[string]any{
					{"id": 42, "name": "plugin.jsplugin.zip"},
				},
			})
		case strings.Contains(path, "/releases/assets/42"):
			w.Write([]byte("ZIPDATA"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer proxy.Close()

	data, err := downloadGitHubReleaseAsset(context.Background(), "owner", "repo", "v1", "plugin.jsplugin.zip", token, proxy.URL)
	if err != nil {
		t.Fatalf("downloadGitHubReleaseAsset() error = %v", err)
	}
	if string(data) != "ZIPDATA" {
		t.Fatalf("data = %q, want ZIPDATA", data)
	}
	if len(gotAuthHeaders) != 2 {
		t.Fatalf("expected 2 upstream requests through proxy, got %d", len(gotAuthHeaders))
	}
	for _, h := range gotAuthHeaders {
		if h != "Bearer "+token {
			t.Fatalf("Authorization header = %q, want %q", h, "Bearer "+token)
		}
	}
}

func TestDownloadGitHubReleaseAsset_AssetNotFound(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"assets": []map[string]any{}})
	}))
	defer proxy.Close()

	_, err := downloadGitHubReleaseAsset(context.Background(), "owner", "repo", "v1", "missing.zip", "tok", proxy.URL)
	if err == nil {
		t.Fatal("expected error for missing asset, got nil")
	}
}
