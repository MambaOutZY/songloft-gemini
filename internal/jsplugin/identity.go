package jsplugin

import (
	"net/url"
	"regexp"
	"strings"
)

// 插件身份（identity）解决的问题：entry_path 是插件的唯一标识，但它同时是磁盘目录名、
// 路由前缀和 DB 唯一键，因此两个不同作者的插件完全可能撞上同一个 entry_path
// （songloft-org/songloft#339）。此时若仅按 entry_path 判等，商店里同名的两个插件会被
// 去重掉只剩一个，且安装状态互相串台。identity 给出 entry_path 之外的第二维度，
// 让「同名但实际不同」的插件可以在商店里各自成行、各自计算安装状态。
//
// identity 为空表示「无法判定身份」，此时退化为仅按 entry_path 判定，与引入本机制
// 之前的行为一致。

var (
	// authorEmailRegexp 匹配 author 里的 <email> 与 (备注) 片段。
	authorNoiseRegexp = regexp.MustCompile(`<[^>]*>|\([^)]*\)`)
	// whitespaceRegexp 用于把内部连续空白压成单个空格。
	whitespaceRegexp = regexp.MustCompile(`\s+`)
)

// githubLikeHosts 是可以从 URL 路径前两段可靠推断出 owner/repo 的 host。
// 自托管源的路径布局任意，无法据此推断仓库归属，故不在此列——识别不出来时
// identity 退化为空比猜错更安全（猜错会把同一插件分裂成两条商店条目）。
var githubLikeHosts = map[string]bool{
	"github.com":                    true,
	"raw.githubusercontent.com":     true,
	"objects.githubusercontent.com": true,
}

// PluginIdentity 返回插件在 entry_path 之外的稳定身份。
//
// 优先使用规范化后的 author；author 为空时退化为 updateURL 指向的 GitHub 仓库
// （形如 "repo:owner/name"）；两者都无法得出时返回 ""（此时仅按 entry_path 判定）。
func PluginIdentity(author, updateURL string) string {
	if a := NormalizeAuthor(author); a != "" {
		return a
	}
	if repo := repoFromURL(updateURL); repo != "" {
		return "repo:" + repo
	}
	return ""
}

// NormalizeAuthor 规范化 author 字段，使同一作者的不同写法归一。
// 剥掉 <email> 与 (备注) → 压缩空白 → trim → 转小写。
// 这样 "hanxi"、"Hanxi <a@b.com>"、" HANXI " 都归一为 "hanxi"，
// 避免同一插件在不同源里因 author 写法不同而被分裂成两条商店条目。
func NormalizeAuthor(author string) string {
	s := authorNoiseRegexp.ReplaceAllString(author, " ")
	s = whitespaceRegexp.ReplaceAllString(s, " ")
	return strings.ToLower(strings.TrimSpace(s))
}

// SameIdentity 判断两个身份是否指向同一个插件作者/仓库。
// 任一方为空表示「无法判定」，保守地视为同一插件——宁可漏报冲突，
// 也不要因为对方缺 author 就误报冲突、拦住用户的正常更新。
func SameIdentity(a, b string) bool {
	if a == "" || b == "" {
		return true
	}
	return a == b
}

// repoFromURL 从 GitHub 系 URL 中提取 "owner/repo"。非 GitHub 系 host
// 或路径不足两段时返回 ""。
func repoFromURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil || !githubLikeHosts[strings.ToLower(u.Host)] {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	return strings.ToLower(parts[0] + "/" + parts[1])
}

// registryDedupKey 返回商店去重用的复合键：entry_path + 身份。
//
// 刻意不含订阅源 URL：官方源与社区聚合源经常同时收录同一个插件，若把源并入键，
// 同一插件会在「全部」模式里重复显示多遍。identity 已足够区分真正不同的插件。
func registryDedupKey(entryPath, author, updateURL string) string {
	return entryPath + "\x00" + PluginIdentity(author, updateURL)
}
