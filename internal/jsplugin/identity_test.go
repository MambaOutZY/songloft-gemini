package jsplugin

import "testing"

func TestNormalizeAuthor(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hanxi", "hanxi"},
		{"case", " HANXI ", "hanxi"},
		{"email", "Hanxi <a@b.com>", "hanxi"},
		{"note", "Hanxi (Songloft)", "hanxi"},
		{"email and note", "Hanxi <a@b.com> (Songloft)", "hanxi"},
		{"inner whitespace", "Zhang   San", "zhang san"},
		{"empty", "", ""},
		{"only email", "<a@b.com>", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeAuthor(tt.in); got != tt.want {
				t.Errorf("NormalizeAuthor(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPluginIdentity(t *testing.T) {
	tests := []struct {
		name      string
		author    string
		updateURL string
		want      string
	}{
		{
			name:   "author wins over updateURL",
			author: "Hanxi <a@b.com>",
			// 即便有 updateURL 也优先用 author，保证同作者跨仓库迁移时身份不变
			updateURL: "https://raw.githubusercontent.com/songloft-org/plugin-x/main/manifest.json",
			want:      "hanxi",
		},
		{
			name:      "fallback to github repo",
			updateURL: "https://raw.githubusercontent.com/songloft-org/plugin-x/main/manifest.json",
			want:      "repo:songloft-org/plugin-x",
		},
		{
			name:      "fallback to github.com repo",
			updateURL: "https://github.com/Alice/My-Plugin/releases/latest/manifest.json",
			want:      "repo:alice/my-plugin",
		},
		{
			name: "both empty",
			want: "",
		},
		{
			// 自托管源的路径布局任意，推断 owner/repo 会把同一插件的不同镜像路径
			// 误判成两个插件，故不推断。
			name:      "self-hosted host is not inferred",
			updateURL: "https://my.server/plugins/a/manifest.json",
			want:      "",
		},
		{
			name:      "github path too short",
			updateURL: "https://github.com/onlyowner",
			want:      "",
		},
		{
			name:      "invalid url",
			updateURL: "://nope",
			want:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PluginIdentity(tt.author, tt.updateURL); got != tt.want {
				t.Errorf("PluginIdentity(%q, %q) = %q, want %q", tt.author, tt.updateURL, got, tt.want)
			}
		})
	}
}

func TestSameIdentity(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"equal", "hanxi", "hanxi", true},
		{"different", "hanxi", "alice", false},
		{"a empty is undecidable", "", "alice", true},
		{"b empty is undecidable", "hanxi", "", true},
		{"both empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SameIdentity(tt.a, tt.b); got != tt.want {
				t.Errorf("SameIdentity(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestRegistryDedupKey(t *testing.T) {
	// 同 entry_path 不同作者 → 不同键（商店里各自成行）
	a := registryDedupKey("demo", "Alice", "")
	b := registryDedupKey("demo", "Bob", "")
	if a == b {
		t.Errorf("expected different keys for different authors, both = %q", a)
	}
	// 同 entry_path 同作者不同写法 → 同一键（不该被分裂）
	c := registryDedupKey("demo", "Alice <a@b.com>", "")
	if a != c {
		t.Errorf("expected same key for equivalent authors, got %q vs %q", a, c)
	}
	// 无 author 且自托管源无法推断仓库 → 退化为按 entry_path 合并，与旧行为一致
	d := registryDedupKey("demo", "", "https://my.server/x/manifest.json")
	e := registryDedupKey("demo", "", "https://my.server/y/manifest.json")
	if d != e {
		t.Errorf("expected merge when identity is unknown, got %q vs %q", d, e)
	}
}
