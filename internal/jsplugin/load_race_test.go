package jsplugin

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestLoadPluginAndEnsureLoadedConcurrently 锁定 LoadPlugin 与 EnsureLoaded 并发时的契约。
//
// 两者都可能触发同一个 entryPath 的加载，但语义不同：EnsureLoaded 要返回 *JSService、
// 且会按 DB status 拒绝（ErrPluginDisabled / ErrPluginErrorState），LoadPlugin 只报成败、
// 刻意不看 DB status（health 自愈依赖这一点：先把状态推回 active 再加载）。
//
// 因此两者绝不能共用同一个 singleflight group + 同一个 key，否则会互相拿到对方的返回值：
//   - LoadPlugin 当 leader → EnsureLoaded 的 follower 拿到 (nil, nil) → 对 nil 做
//     v.(*JSService) 类型断言 → panic；
//   - EnsureLoaded 当 leader → LoadPlugin 的 follower 拿到 ErrPluginDisabled 之类，
//     于是 EnablePlugin 报「插件已禁用」、health 自愈把好插件回滚成 error。
func TestLoadPluginAndEnsureLoadedConcurrently(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	ctx := context.Background()
	const entryPath = "test-load-race"

	manifest := testManifest(entryPath)
	zipData := createTestPluginZip(t, manifest, simpleJSCode)
	zipFileName := entryPath + ".jsplugin.zip"
	if err := os.WriteFile(filepath.Join(pluginsDir, zipFileName), zipData, 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	plugin := &JSPlugin{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		EntryPath:   manifest.EntryPath,
		Main:        manifest.Main,
		Permissions: manifest.Permissions,
		Status:      JSPluginStatusActive,
		FilePath:    zipFileName,
	}
	if err := repo.Create(ctx, plugin); err != nil {
		t.Fatalf("create plugin record: %v", err)
	}

	manager := NewManager(repo, pluginsDir, dataDir, "", nil, nil)
	t.Cleanup(func() { manager.Close() })

	if err := manager.LoadPlugin(ctx, plugin); err != nil {
		t.Skipf("LoadPlugin failed (may need QuickJS runtime): %v", err)
	}

	// 反复「卸载 → 同时打 LoadPlugin 与一批 EnsureLoaded」，撞出两者的合并窗口。
	const rounds = 60
	const ensurePerRound = 6
	for round := range rounds {
		if err := manager.UnloadPlugin(ctx, entryPath); err != nil {
			t.Fatalf("round %d UnloadPlugin: %v", round, err)
		}

		var wg sync.WaitGroup
		loadErr := make([]error, 1)
		ensureErrs := make([]error, ensurePerRound)
		ensureSvcs := make([]*JSService, ensurePerRound)

		wg.Add(1)
		go func() {
			defer wg.Done()
			loadErr[0] = manager.LoadPlugin(ctx, plugin)
		}()
		// 让 LoadPlugin 先成为 leader：不 stagger 的话 EnsureLoaded 往往先抢到 key，
		// 于是 LoadPlugin 变成 follower（那是另一个方向的问题，不会 panic），
		// 「EnsureLoaded 拿到 LoadPlugin 的 nil 返回值」这条路径就撞不到。
		time.Sleep(time.Duration(round%3) * time.Millisecond)
		for i := range ensurePerRound {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				// 契约要求：要么返回非 nil service，要么返回 error。绝不能 panic。
				svc, err := manager.EnsureLoaded(ctx, entryPath)
				ensureSvcs[idx], ensureErrs[idx] = svc, err
			}(i)
		}
		wg.Wait()

		if loadErr[0] != nil {
			t.Fatalf("round %d LoadPlugin 应成功，却返回: %v", round, loadErr[0])
		}
		for i := range ensurePerRound {
			if ensureErrs[i] != nil {
				t.Fatalf("round %d EnsureLoaded[%d] 应成功，却返回: %v", round, i, ensureErrs[i])
			}
			if ensureSvcs[i] == nil {
				t.Fatalf("round %d EnsureLoaded[%d] 返回了 nil service 但没有 error", round, i)
			}
		}
	}
}
