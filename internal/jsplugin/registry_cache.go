package jsplugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"
)

// 插件商店的分页与搜索都在**完整插件列表**上做（服务端切片 / 子串过滤），所以每次
// 翻页、每次改搜索词，以前都会重新递归拉取整棵注册表树——最多 500 个 plugin.json、
// 8 并发、单请求 15s 超时。用户翻 5 页就是 5 次全量拉取。
//
// 这里给拉取结果加一层进程内 TTL 缓存：同一组输入（源、token、GitHub 代理）在 TTL
// 内只拉一次，翻页与搜索直接命中缓存。用户点「刷新」时走 force 路径绕过缓存。
//
// 刻意不缓存**安装状态**：installed / has_update / conflict 每次请求都从数据库实时
// 计算（buildInstalledMap），所以装完插件立刻翻页也能看到状态更新，无需失效缓存。

const (
	// registryCacheTTL 缓存有效期。注册表是远端易变数据，取值需要平衡
	//「翻页不重拉」与「新发布的插件多久可见」——用户想立刻看到新插件时有刷新按钮。
	registryCacheTTL = 5 * time.Minute
	// registryCacheMaxEntries 缓存条目上限。键含源列表与代理地址，用户反复改配置
	// 会不断产生新键，需要上限兜底防止无界增长。
	registryCacheMaxEntries = 16
)

// registryCacheEntry 一次拉取的完整结果。
type registryCacheEntry struct {
	entries   []RegistryEntry
	warnings  []string
	fetchedAt time.Time
}

// FetchAndMergeCached 同 FetchAndMerge，但结果在 TTL 内复用。
// force 为 true 时跳过缓存强制重拉（并回填缓存）。
func (s *RegistryService) FetchAndMergeCached(ctx context.Context, registryURL string, githubProxy string, token string, force bool) ([]RegistryEntry, []string, error) {
	key := singleSourceCacheKey(registryURL, token, githubProxy)
	if !force {
		if e := s.cacheGet(key); e != nil {
			return e.entries, e.warnings, nil
		}
	}
	entries, warnings, err := s.FetchAndMerge(ctx, registryURL, githubProxy, token)
	if err != nil {
		// 失败不写缓存，也不动既有缓存：下次请求重试，而不是把错误状态粘住 TTL 之久。
		return nil, warnings, err
	}
	s.cachePut(key, entries, warnings)
	return entries, warnings, nil
}

// FetchAndMergeMultiCached 同 FetchAndMergeMulti，但结果在 TTL 内复用。
// force 为 true 时跳过缓存强制重拉（并回填缓存）。
func (s *RegistryService) FetchAndMergeMultiCached(ctx context.Context, sources []RegistryConfig, githubProxy string, force bool) ([]RegistryEntry, []string) {
	key := multiSourceCacheKey(sources, githubProxy)
	if !force {
		if e := s.cacheGet(key); e != nil {
			return e.entries, e.warnings
		}
	}
	entries, warnings := s.FetchAndMergeMulti(ctx, sources, githubProxy)
	// FetchAndMergeMulti 不返回 error：单源失败只进 warnings。全部源都失败时
	// entries 为空，此时不缓存空结果，否则用户修好网络后还要等一个 TTL。
	if len(entries) == 0 {
		return entries, warnings
	}
	s.cachePut(key, entries, warnings)
	return entries, warnings
}

// InvalidateCache 清空全部缓存。订阅源配置变更后调用。
//
// 严格说源列表变更会让缓存键自然改变（键含源指纹），不清也不会读到脏数据；
// 但显式清理能立刻释放旧配置占用的条目，避免它们占满上限把有用的条目挤掉。
func (s *RegistryService) InvalidateCache() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cache = nil
}

func (s *RegistryService) cacheGet(key string) *registryCacheEntry {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	e, ok := s.cache[key]
	if !ok {
		return nil
	}
	if time.Since(e.fetchedAt) > registryCacheTTL {
		delete(s.cache, key)
		return nil
	}
	slog.Debug("plugin registry cache hit", "plugins", len(e.entries), "age_s", int(time.Since(e.fetchedAt).Seconds()))
	return e
}

func (s *RegistryService) cachePut(key string, entries []RegistryEntry, warnings []string) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if s.cache == nil {
		s.cache = make(map[string]*registryCacheEntry)
	}
	// 先清掉过期条目；仍然超限则淘汰最旧的一条。
	if len(s.cache) >= registryCacheMaxEntries {
		for k, v := range s.cache {
			if time.Since(v.fetchedAt) > registryCacheTTL {
				delete(s.cache, k)
			}
		}
	}
	if len(s.cache) >= registryCacheMaxEntries {
		var oldestKey string
		var oldest time.Time
		for k, v := range s.cache {
			if oldestKey == "" || v.fetchedAt.Before(oldest) {
				oldestKey, oldest = k, v.fetchedAt
			}
		}
		delete(s.cache, oldestKey)
	}
	s.cache[key] = &registryCacheEntry{
		entries:   entries,
		warnings:  warnings,
		fetchedAt: time.Now(),
	}
}

// singleSourceCacheKey 单源模式的缓存键。
func singleSourceCacheKey(registryURL, token, githubProxy string) string {
	return "one\x00" + registryURL + "\x00" + hashToken(token) + "\x00" + githubProxy
}

// multiSourceCacheKey 「全部」聚合模式的缓存键。
// 顺序敏感：FetchAndMergeMulti 按源顺序决定同版本插件由哪个源胜出，
// 所以源顺序变化必须产生不同的键。
func multiSourceCacheKey(sources []RegistryConfig, githubProxy string) string {
	var b strings.Builder
	b.WriteString("all\x00")
	for _, src := range sources {
		b.WriteString(src.URL)
		b.WriteString("\x00")
		b.WriteString(hashToken(src.Token))
		b.WriteString("\x00")
	}
	b.WriteString(githubProxy)
	return b.String()
}

// hashToken 把 token 折叠成短摘要：token 变化必须让缓存失效（换 token 可能看到
// 不同的插件集），但缓存键会进日志与调试输出，不能带明文凭据。
func hashToken(token string) string {
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:8])
}
