package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"songloft/internal/httputil"
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
			got := httputil.ApplyGithubProxy(c.url, c.proxy)
			if got != c.want {
				t.Fatalf("ApplyGithubProxy(%q, %q) = %q, want %q", c.url, c.proxy, got, c.want)
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

// --- 商店安装状态判定（#339） ---

func installedAlice() map[string]installedPlugin {
	return map[string]installedPlugin{
		"demo": {Name: "Demo by Alice", Author: "Alice", Version: "1.0.0", Identity: "alice"},
	}
}

func TestResolveInstallState_NotInstalled(t *testing.T) {
	p := registryPluginEntry{EntryPath: "other", Version: "1.0.0", Identity: "bob"}
	resolveInstallState(&p, installedAlice())
	if p.Installed || p.Conflict || p.HasUpdate {
		t.Errorf("expected clean state, got %+v", p)
	}
}

func TestResolveInstallState_SameIdentityUpToDate(t *testing.T) {
	p := registryPluginEntry{EntryPath: "demo", Version: "1.0.0", Identity: "alice"}
	resolveInstallState(&p, installedAlice())
	if !p.Installed {
		t.Error("expected installed=true")
	}
	if p.HasUpdate {
		t.Error("expected has_update=false for identical version")
	}
	if p.Conflict {
		t.Error("expected conflict=false for same identity")
	}
	if p.InstalledVersion != "1.0.0" {
		t.Errorf("expected installed_version 1.0.0, got %q", p.InstalledVersion)
	}
}

func TestResolveInstallState_SameIdentityHasUpdate(t *testing.T) {
	p := registryPluginEntry{EntryPath: "demo", Version: "2.0.0", Identity: "alice"}
	resolveInstallState(&p, installedAlice())
	if !p.Installed || !p.HasUpdate {
		t.Errorf("expected installed with update, got %+v", p)
	}
}

// TestResolveInstallState_OlderRemoteIsNotUpdate 回归：以前用字符串 != 判定，
// 商店版本比本地旧时也会显示「可更新」，点下去会把本地降级。
func TestResolveInstallState_OlderRemoteIsNotUpdate(t *testing.T) {
	p := registryPluginEntry{EntryPath: "demo", Version: "0.9.0", Identity: "alice"}
	resolveInstallState(&p, installedAlice())
	if !p.Installed {
		t.Error("expected installed=true")
	}
	if p.HasUpdate {
		t.Error("expected has_update=false when registry version is older")
	}
}

// TestResolveInstallState_DifferentIdentityConflicts 是 #339 的核心断言：
// 装了 Alice 的 demo 之后，Bob 的同名 demo 必须显示为冲突，而不是「已安装」。
func TestResolveInstallState_DifferentIdentityConflicts(t *testing.T) {
	p := registryPluginEntry{EntryPath: "demo", Version: "3.0.0", Identity: "bob"}
	resolveInstallState(&p, installedAlice())
	if p.Installed {
		t.Error("expected installed=false for a different plugin sharing the entry_path")
	}
	if p.HasUpdate {
		t.Error("expected has_update=false: it is not an update, it is a replacement")
	}
	if !p.Conflict {
		t.Fatal("expected conflict=true")
	}
	if !strings.Contains(p.ConflictWith, "Alice") || !strings.Contains(p.ConflictWith, "1.0.0") {
		t.Errorf("conflict_with should describe the occupying plugin, got %q", p.ConflictWith)
	}
}

// TestResolveInstallState_UnknownIdentityTreatedAsSame 验证身份无法判定时
// 保守视为同一插件（宁可漏报冲突，也不误报拦住正常更新）。
func TestResolveInstallState_UnknownIdentityTreatedAsSame(t *testing.T) {
	installed := map[string]installedPlugin{
		"demo": {Name: "Demo", Version: "1.0.0", Identity: ""},
	}
	p := registryPluginEntry{EntryPath: "demo", Version: "2.0.0", Identity: "alice"}
	resolveInstallState(&p, installed)
	if p.Conflict {
		t.Error("expected no conflict when identity is undecidable")
	}
	if !p.Installed || !p.HasUpdate {
		t.Errorf("expected installed with update, got %+v", p)
	}
}

func TestInstalledPluginDescribe(t *testing.T) {
	withAuthor := installedPlugin{Name: "Demo", Author: "Alice", Version: "1.0.0"}
	if got := withAuthor.describe(); !strings.Contains(got, "Alice") || !strings.Contains(got, "Demo") {
		t.Errorf("unexpected describe output: %q", got)
	}
	noAuthor := installedPlugin{Name: "Demo", Version: "1.0.0"}
	if got := noAuthor.describe(); strings.Contains(got, "作者") {
		t.Errorf("should omit author section when empty, got %q", got)
	}
}
