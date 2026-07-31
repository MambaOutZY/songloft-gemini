package jsplugin

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// authoredManifest 返回一个指定作者的合法 manifest。
func authoredManifest(entryPath, name, author, version string) *PluginManifest {
	m := testManifest(entryPath)
	m.Name = name
	m.Author = author
	m.Version = version
	return m
}

// TestInstallFromUpload_RejectIdentityConflict 验证商店安装路径遇到「entryPath 已被
// 另一个作者的插件占用」时拒绝安装，且原插件的 ZIP 与 DB 记录一字未动（#339）。
// 以前这里会静默走 Update：覆盖 ZIP、清空 static 目录、原地改写 DB 记录，
// 导致原插件被无声销毁、它的 plugin_storage 数据被新插件继承。
func TestInstallFromUpload_RejectIdentityConflict(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	pm := NewPackageManager(pluginsDir, dataDir, repo)
	ctx := context.Background()

	aliceZip := createTestPluginZip(t, authoredManifest("demo", "Demo by Alice", "Alice", "1.0.0"), simpleJSCode)
	if _, _, err := pm.InstallFromUpload(aliceZip); err != nil {
		t.Fatalf("install alice: %v", err)
	}

	zipPath := filepath.Join(pluginsDir, "demo.jsplugin.zip")
	before, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read installed zip: %v", err)
	}

	bobZip := createTestPluginZip(t, authoredManifest("demo", "Demo by Bob", "Bob", "2.0.0"), simpleJSCode)
	_, _, err = pm.InstallFromUploadWithOptions(bobZip, InstallOptions{RejectIdentityConflict: true})

	var conflict *EntryPathConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected *EntryPathConflictError, got %v", err)
	}
	if conflict.EntryPath != "demo" || conflict.ExistingAuthor != "Alice" || conflict.IncomingAuthor != "Bob" {
		t.Errorf("unexpected conflict details: %+v", conflict)
	}
	if conflict.ExistingVersion != "1.0.0" {
		t.Errorf("expected existing version 1.0.0, got %q", conflict.ExistingVersion)
	}

	// 磁盘未被覆盖
	after, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read zip after rejected install: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("zip file was modified despite rejected install")
	}

	// DB 记录未被改写
	dbPlugin, err := repo.GetByEntryPath(ctx, "demo")
	if err != nil {
		t.Fatalf("GetByEntryPath: %v", err)
	}
	if dbPlugin.Author != "Alice" || dbPlugin.Name != "Demo by Alice" || dbPlugin.Version != "1.0.0" {
		t.Errorf("DB record was modified: name=%q author=%q version=%q",
			dbPlugin.Name, dbPlugin.Author, dbPlugin.Version)
	}
}

// TestInstallFromUpload_OverwriteAcrossIdentities 验证用户二次确认后（不开
// RejectIdentityConflict）仍可覆盖安装另一个作者的同名插件。
func TestInstallFromUpload_OverwriteAcrossIdentities(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	pm := NewPackageManager(pluginsDir, dataDir, repo)
	ctx := context.Background()

	aliceZip := createTestPluginZip(t, authoredManifest("demo", "Demo by Alice", "Alice", "1.0.0"), simpleJSCode)
	if _, _, err := pm.InstallFromUpload(aliceZip); err != nil {
		t.Fatalf("install alice: %v", err)
	}

	bobZip := createTestPluginZip(t, authoredManifest("demo", "Demo by Bob", "Bob", "2.0.0"), simpleJSCode)
	plugin, wasUpdate, err := pm.InstallFromUploadWithOptions(bobZip, InstallOptions{RejectIdentityConflict: false})
	if err != nil {
		t.Fatalf("overwrite install: %v", err)
	}
	if !wasUpdate {
		t.Error("expected wasUpdate=true when overwriting an existing entryPath")
	}
	if plugin.Author != "Bob" {
		t.Errorf("expected author Bob after overwrite, got %q", plugin.Author)
	}

	dbPlugin, err := repo.GetByEntryPath(ctx, "demo")
	if err != nil {
		t.Fatalf("GetByEntryPath: %v", err)
	}
	if dbPlugin.Author != "Bob" || dbPlugin.Version != "2.0.0" {
		t.Errorf("expected Bob v2.0.0 in DB, got %q v%s", dbPlugin.Author, dbPlugin.Version)
	}
}

// TestInstallFromUpload_SameIdentityStillUpdates 验证同一作者的正常升级不会被
// 冲突检测拦住，且 author 写法变动（加邮箱）也不误报。
func TestInstallFromUpload_SameIdentityStillUpdates(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	pm := NewPackageManager(pluginsDir, dataDir, repo)

	v1 := createTestPluginZip(t, authoredManifest("demo", "Demo", "hanxi", "1.0.0"), simpleJSCode)
	if _, _, err := pm.InstallFromUpload(v1); err != nil {
		t.Fatalf("install v1: %v", err)
	}

	v2 := createTestPluginZip(t, authoredManifest("demo", "Demo", "Hanxi <a@b.com>", "2.0.0"), simpleJSCode)
	plugin, wasUpdate, err := pm.InstallFromUploadWithOptions(v2, InstallOptions{RejectIdentityConflict: true})
	if err != nil {
		t.Fatalf("expected same-identity update to succeed, got %v", err)
	}
	if !wasUpdate {
		t.Error("expected wasUpdate=true")
	}
	if plugin.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %q", plugin.Version)
	}
}

// TestInstallFromUpload_UnknownIdentityStillUpdates 验证已装插件缺 author 时
// 身份无法判定，保守放行更新（宁可漏报冲突，也不拦住正常升级）。
func TestInstallFromUpload_UnknownIdentityStillUpdates(t *testing.T) {
	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	pm := NewPackageManager(pluginsDir, dataDir, repo)

	anon := createTestPluginZip(t, authoredManifest("demo", "Demo", "", "1.0.0"), simpleJSCode)
	if _, _, err := pm.InstallFromUpload(anon); err != nil {
		t.Fatalf("install anonymous: %v", err)
	}

	named := createTestPluginZip(t, authoredManifest("demo", "Demo", "Alice", "2.0.0"), simpleJSCode)
	if _, wasUpdate, err := pm.InstallFromUploadWithOptions(named, InstallOptions{RejectIdentityConflict: true}); err != nil {
		t.Fatalf("expected undecidable identity to be allowed, got %v", err)
	} else if !wasUpdate {
		t.Error("expected wasUpdate=true")
	}
}
