package jsplugin

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestLoadPlugin_IdempotentAcrossStartupAndLazyLoad 复现并锁定用户现场的启动期竞态：
//
// 插件管理器是异步启动的（app.Init 里 go manager.Start），而 HTTP 路由 / 音源调用早已就绪。
// 于是启动循环 loadPlugins 还没轮到某个插件时，一次请求触发的 EnsureLoaded 可能先把它加载出来；
// 随后启动循环再 LoadPlugin 一次，撞上 jsruntime「envID 全局唯一」守卫报
// "env jsplugin-xxx already exists"，并把 DB 状态刷成 error——之后 health 自愈每轮都再撞
// 同一个错、永不成功，插件在管理界面永久显示异常。
func TestLoadPlugin_IdempotentAcrossStartupAndLazyLoad(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	ctx := context.Background()
	const entryPath = "test-idempotent-load"

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

	// 第一次加载代表「懒加载先跑完」。QuickJS 不可用时按现有测试模式跳过。
	if err := manager.LoadPlugin(ctx, plugin); err != nil {
		t.Skipf("LoadPlugin failed (may need QuickJS runtime): %v", err)
	}

	t.Run("second_load_is_noop_not_error", func(t *testing.T) {
		// 这一次代表启动循环随后到达。修复前会返回 "env ... already exists"。
		if err := manager.LoadPlugin(ctx, plugin); err != nil {
			t.Fatalf("重复 LoadPlugin 应幂等成功，却返回错误: %v", err)
		}
		if _, ok := manager.GetService(entryPath); !ok {
			t.Fatal("幂等返回后 service 应仍在内存中")
		}
	})

	t.Run("startup_loop_does_not_flip_status_to_error", func(t *testing.T) {
		// loadPlugins 是启动循环的真实入口：它在失败时会把 DB 刷成 error。
		manager.loadPlugins(ctx, []*JSPlugin{plugin})

		dbPlugin, err := repo.GetByEntryPath(ctx, entryPath)
		if err != nil {
			t.Fatalf("GetByEntryPath: %v", err)
		}
		if dbPlugin.Status != JSPluginStatusActive {
			t.Fatalf("良性竞态不应把插件标成异常，DB status = %s，期望 active", dbPlugin.Status)
		}
	})

	t.Run("concurrent_loads_are_serialized", func(t *testing.T) {
		// 先卸载，让并发加载真的要各自建环境，验证 loadGroup 把它们合并成一次。
		if err := manager.UnloadPlugin(ctx, entryPath); err != nil {
			t.Fatalf("UnloadPlugin: %v", err)
		}

		const goroutines = 8
		errs := make([]error, goroutines)
		var wg sync.WaitGroup
		for i := range goroutines {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				errs[idx] = manager.LoadPlugin(ctx, plugin)
			}(i)
		}
		wg.Wait()

		for i, err := range errs {
			if err != nil {
				t.Fatalf("并发 LoadPlugin[%d] 应全部成功，却返回: %v", i, err)
			}
		}
		if _, ok := manager.GetService(entryPath); !ok {
			t.Fatal("并发加载后 service 应在内存中")
		}
	})

	t.Run("orphan_env_is_recoverable", func(t *testing.T) {
		// 模拟「上一次加载建好 JS 环境之后才失败」留下的孤儿 env：services map 里没有
		// service，但 jsManager 里还占着同一个 envID。修复前这会让插件此后永远加载不了
		// （CreateEnv 遇同名只报错不覆盖，UnloadPlugin 又找不到它），只能重启进程。
		if err := manager.UnloadPlugin(ctx, entryPath); err != nil {
			t.Fatalf("UnloadPlugin: %v", err)
		}
		envID := "jsplugin-" + entryPath
		if err := manager.jsManager.CreateEnv(envID, "", plugin.ID); err != nil {
			t.Fatalf("构造孤儿 env 失败: %v", err)
		}

		if err := manager.LoadPlugin(ctx, plugin); err != nil {
			t.Fatalf("存在孤儿 env 时 LoadPlugin 应能自愈，却返回: %v", err)
		}
		if _, ok := manager.GetService(entryPath); !ok {
			t.Fatal("自愈后 service 应在内存中")
		}
	})

	t.Run("ensure_loaded_does_not_deadlock", func(t *testing.T) {
		// EnsureLoaded 在自己的 loadGroup.Do 内部调 LoadPlugin。只要 LoadPlugin 不套
		// singleflight 就没问题；若给它加上同 group 同 key 的 singleflight，
		// 这里会等自己完成从而死锁，本用例超时。
		if err := manager.UnloadPlugin(ctx, entryPath); err != nil {
			t.Fatalf("UnloadPlugin: %v", err)
		}
		svc, err := manager.EnsureLoaded(ctx, entryPath)
		if err != nil {
			t.Fatalf("EnsureLoaded failed: %v", err)
		}
		if svc == nil {
			t.Fatal("EnsureLoaded 应返回 service")
		}
	})
}
