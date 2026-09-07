package services

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

// 这些用例锁住「遍历目录项时不再逐条 os.Stat」（resolveEntry）之后仍需保持的语义：
// 软链接目录照样递归、软链接文件照样收录、断链条目跳过。

func writeAudio(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte("audio"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func newWalkScanner(musicPath string) *Scanner {
	return NewScanner(&ScanConfig{
		MusicPath:        musicPath,
		SupportedFormats: []string{"mp3", "flac"},
	})
}

func TestScanFilesWithCue_FollowsSymlinkedDir(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	writeAudio(t, filepath.Join(root, "local", "a.mp3"))
	writeAudio(t, filepath.Join(outside, "b.mp3"))
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Skipf("当前环境不支持创建软链接: %v", err)
	}

	result, err := newWalkScanner(root).ScanFilesWithCueInDirs(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ScanFilesWithCueInDirs() error = %v", err)
	}
	if len(result.AudioFiles) != 2 {
		t.Fatalf("软链接目录应被递归，扫到 %d 个文件: %v", len(result.AudioFiles), result.AudioFiles)
	}
}

func TestScanFilesWithCue_FollowsSymlinkedFile(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	target := filepath.Join(outside, "real.mp3")
	writeAudio(t, target)
	if err := os.Symlink(target, filepath.Join(root, "link.mp3")); err != nil {
		t.Skipf("当前环境不支持创建软链接: %v", err)
	}

	result, err := newWalkScanner(root).ScanFilesWithCueInDirs(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ScanFilesWithCueInDirs() error = %v", err)
	}
	if len(result.AudioFiles) != 1 {
		t.Fatalf("软链接音频文件应被收录，扫到 %d 个: %v", len(result.AudioFiles), result.AudioFiles)
	}
}

func TestScanFilesWithCue_SkipsBrokenSymlink(t *testing.T) {
	root := t.TempDir()

	writeAudio(t, filepath.Join(root, "good.mp3"))
	if err := os.Symlink(filepath.Join(root, "gone.mp3"), filepath.Join(root, "dangling.mp3")); err != nil {
		t.Skipf("当前环境不支持创建软链接: %v", err)
	}

	result, err := newWalkScanner(root).ScanFilesWithCueInDirs(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ScanFilesWithCueInDirs() error = %v", err)
	}
	if len(result.AudioFiles) != 1 {
		t.Fatalf("断链条目应被跳过，扫到 %d 个: %v", len(result.AudioFiles), result.AudioFiles)
	}
	if filepath.Base(result.AudioFiles[0]) != "good.mp3" {
		t.Errorf("扫到的文件 = %s, want good.mp3", result.AudioFiles[0])
	}
}

func TestCollectAllDirNames_FollowsSymlinkedDir(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	if err := os.MkdirAll(filepath.Join(outside, "远端专辑"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "软链")); err != nil {
		t.Skipf("当前环境不支持创建软链接: %v", err)
	}

	names, err := newWalkScanner(root).CollectAllDirNames(context.Background())
	if err != nil {
		t.Fatalf("CollectAllDirNames() error = %v", err)
	}
	if !slices.Contains(names, "软链") || !slices.Contains(names, "远端专辑") {
		t.Errorf("软链接目录及其子目录名都应收集到，得到 %v", names)
	}
}

func TestCollectAllDirNames_SkipsBrokenSymlink(t *testing.T) {
	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "实目录"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(filepath.Join(root, "不存在"), filepath.Join(root, "断链")); err != nil {
		t.Skipf("当前环境不支持创建软链接: %v", err)
	}

	names, err := newWalkScanner(root).CollectAllDirNames(context.Background())
	if err != nil {
		t.Fatalf("CollectAllDirNames() error = %v", err)
	}
	if slices.Contains(names, "断链") {
		t.Errorf("断链条目不应计入目录名，得到 %v", names)
	}
	if !slices.Contains(names, "实目录") {
		t.Errorf("真实目录应计入，得到 %v", names)
	}
}

// TestCollectAllDirNames_CachesResult 锁住 TTL 缓存：客户端超时后的重试要命中缓存，
// 而不是又触发一次全树遍历（songloft-org/songloft#447）。
func TestCollectAllDirNames_CachesResult(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "first"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	scanner := newWalkScanner(root)
	ctx := context.Background()

	first, err := scanner.CollectAllDirNames(ctx)
	if err != nil {
		t.Fatalf("CollectAllDirNames() error = %v", err)
	}
	if !slices.Contains(first, "first") {
		t.Fatalf("首次调用应收集到 first，得到 %v", first)
	}

	// TTL 内新建目录不应出现（证明走的是缓存而非重新遍历）。
	if err := os.MkdirAll(filepath.Join(root, "second"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cached, err := scanner.CollectAllDirNames(ctx)
	if err != nil {
		t.Fatalf("CollectAllDirNames() error = %v", err)
	}
	if slices.Contains(cached, "second") {
		t.Errorf("TTL 内应返回缓存，却看到新目录 second: %v", cached)
	}

	// 让缓存过期后应重新遍历。
	scanner.dirNamesAt = time.Now().Add(-dirNamesCacheTTL - time.Second)
	refreshed, err := scanner.CollectAllDirNames(ctx)
	if err != nil {
		t.Fatalf("CollectAllDirNames() error = %v", err)
	}
	if !slices.Contains(refreshed, "second") {
		t.Errorf("缓存过期后应重新遍历并看到 second，得到 %v", refreshed)
	}
}

// TestCollectAllDirNames_ErrorNotCached 校验失败不写缓存：
// music_path 暂时不可用时不能把错误状态粘住整个 TTL。
func TestCollectAllDirNames_ErrorNotCached(t *testing.T) {
	root := filepath.Join(t.TempDir(), "later")
	scanner := newWalkScanner(root)
	ctx := context.Background()

	if _, err := scanner.CollectAllDirNames(ctx); err == nil {
		t.Fatal("目录不存在时应返回错误")
	}

	if err := os.MkdirAll(filepath.Join(root, "已就绪"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	names, err := scanner.CollectAllDirNames(ctx)
	if err != nil {
		t.Fatalf("目录就绪后应成功，error = %v", err)
	}
	if !slices.Contains(names, "已就绪") {
		t.Errorf("目录就绪后应收集到已就绪，得到 %v", names)
	}
}
