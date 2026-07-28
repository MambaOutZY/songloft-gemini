package services

import (
	"context"
	"runtime"
	"testing"
	"time"

	"songloft/internal/database"
	"songloft/internal/database/testutil"
	"songloft/internal/models"
)

// seedFingerprintSong 插入一首本地歌曲，返回其 ID。
func seedFingerprintSong(t *testing.T, repo *database.SongRepository, filePath string) int64 {
	t.Helper()
	song := &models.Song{
		Type:     models.TypeLocal,
		Title:    filePath,
		FilePath: filePath,
		Duration: 300,
	}
	if err := repo.Create(context.Background(), song); err != nil {
		t.Fatalf("create song: %v", err)
	}
	return song.ID
}

// TestMarkFingerprintAttempted_NotRetried 是 songloft-org/songloft#323 的核心回归：
// 指纹计算失败后必须落库「已尝试」标记，否则每轮扫描都会把同一批注定失败的文件
// 重新捞出来跑 ffmpeg 全解码，形成永久 CPU 占用。
func TestMarkFingerprintAttempted_NotRetried(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	repo := mdb.SongRepository()
	ctx := context.Background()

	id := seedFingerprintSong(t, repo, "/music/audiobook/ep001.m4a")

	// 首轮：待计算列表里应有它
	missing, err := repo.ListLocalWithoutFingerprint(ctx)
	if err != nil {
		t.Fatalf("list missing: %v", err)
	}
	if len(missing) != 1 || missing[0].ID != id {
		t.Fatalf("first round missing: got %+v want 1 item id=%d", missing, id)
	}

	// 模拟计算失败：只标记已尝试，不写指纹
	if err := repo.MarkFingerprintAttempted(ctx, id, time.Now().Unix()); err != nil {
		t.Fatalf("mark attempted: %v", err)
	}

	// 次轮：不应再出现（否则就是死循环重试）
	missing, err = repo.ListLocalWithoutFingerprint(ctx)
	if err != nil {
		t.Fatalf("list missing again: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("second round should skip attempted song, got %+v", missing)
	}

	// 统计里应计入 failed，而不是继续算 missing
	total, computed, failed, err := repo.CountLocalFingerprints(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if total != 1 || computed != 0 || failed != 1 {
		t.Errorf("counts: got total=%d computed=%d failed=%d want 1/0/1", total, computed, failed)
	}
}

// TestClearAllFingerprints_ResetsAttempted 「重新计算全部」必须能重试此前失败的歌曲，
// 否则失败项永远无法恢复。
func TestClearAllFingerprints_ResetsAttempted(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	repo := mdb.SongRepository()
	ctx := context.Background()

	failedID := seedFingerprintSong(t, repo, "/music/broken.mp3")
	okID := seedFingerprintSong(t, repo, "/music/good.mp3")

	if err := repo.MarkFingerprintAttempted(ctx, failedID, time.Now().Unix()); err != nil {
		t.Fatalf("mark attempted: %v", err)
	}
	if err := repo.UpdateFingerprint(ctx, okID, "AQABc...", 210.5, time.Now().Unix()); err != nil {
		t.Fatalf("update fingerprint: %v", err)
	}

	if missing, err := repo.ListLocalWithoutFingerprint(ctx); err != nil || len(missing) != 0 {
		t.Fatalf("before clear: got %+v err=%v want empty", missing, err)
	}

	if err := repo.ClearAllFingerprints(ctx); err != nil {
		t.Fatalf("clear: %v", err)
	}

	missing, err := repo.ListLocalWithoutFingerprint(ctx)
	if err != nil {
		t.Fatalf("list after clear: %v", err)
	}
	if len(missing) != 2 {
		t.Fatalf("after clear both songs should be pending, got %+v", missing)
	}
}

// TestResetFailedFingerprintAttempts_KeepsComputed 「仅重试失败项」必须只重置
// 失败标记，已算好的指纹不能被清空——否则就退化成了「重新计算全部」。
// 场景：ffmpeg 升级新增 mpeg 解复用器后，恢复此前因解码器缺失而失败的 mpg。
func TestResetFailedFingerprintAttempts_KeepsComputed(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	repo := mdb.SongRepository()
	ctx := context.Background()

	failedID := seedFingerprintSong(t, repo, "/music/ktv/song.mpg")
	okID := seedFingerprintSong(t, repo, "/music/good.mp3")

	if err := repo.MarkFingerprintAttempted(ctx, failedID, time.Now().Unix()); err != nil {
		t.Fatalf("mark attempted: %v", err)
	}
	if err := repo.UpdateFingerprint(ctx, okID, "AQABc...", 210.5, time.Now().Unix()); err != nil {
		t.Fatalf("update fingerprint: %v", err)
	}

	if err := repo.ResetFailedFingerprintAttempts(ctx); err != nil {
		t.Fatalf("reset failed attempts: %v", err)
	}

	// 只有失败项回到待计算列表，已算好的不受影响
	missing, err := repo.ListLocalWithoutFingerprint(ctx)
	if err != nil {
		t.Fatalf("list after reset: %v", err)
	}
	if len(missing) != 1 || missing[0].ID != failedID {
		t.Fatalf("after reset only failed song should be pending, got %+v", missing)
	}

	total, computed, failed, err := repo.CountLocalFingerprints(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if total != 2 || computed != 1 || failed != 0 {
		t.Errorf("counts: got total=%d computed=%d failed=%d want 2/1/0", total, computed, failed)
	}
}

// TestListLocalWithoutFingerprint_CarriesCueRange CUE 轨的 FilePath 指向整轨镜像，
// 必须带上区间，否则同一镜像下所有 track 会拿到完全相同的指纹并互判重复。
func TestListLocalWithoutFingerprint_CarriesCueRange(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	repo := mdb.SongRepository()
	ctx := context.Background()

	if err := repo.Create(ctx, &models.Song{
		Type:            models.TypeLocal,
		Title:           "Track 02",
		FilePath:        "/music/album/image.flac",
		CueSourcePath:   "/music/album/image.cue",
		CueAudioPath:    "/music/album/image.flac",
		CueTrackIndex:   2,
		CueStartSeconds: 245.5,
		CueEndSeconds:   480.25,
	}); err != nil {
		t.Fatalf("create cue track: %v", err)
	}

	missing, err := repo.ListLocalWithoutFingerprint(ctx)
	if err != nil {
		t.Fatalf("list missing: %v", err)
	}
	if len(missing) != 1 {
		t.Fatalf("missing count: got %d want 1", len(missing))
	}
	if missing[0].CueStartSeconds != 245.5 || missing[0].CueEndSeconds != 480.25 {
		t.Errorf("cue range: got [%v, %v] want [245.5, 480.25]",
			missing[0].CueStartSeconds, missing[0].CueEndSeconds)
	}
}

// TestFPWorkerCount 并发度必须落在 [1, 4]，且不超过 GOMAXPROCS。
// 硬编码 4 路并发是 #323 里 4 核 NAS 被打满的直接原因。
func TestFPWorkerCount(t *testing.T) {
	n := fpWorkerCount()
	if n < 1 || n > 4 {
		t.Errorf("fpWorkerCount() = %d, want within [1, 4]", n)
	}
	if procs := runtime.GOMAXPROCS(0); n > procs {
		t.Errorf("fpWorkerCount() = %d exceeds GOMAXPROCS %d", n, procs)
	}
}

// TestCancel_DoesNotClobberNewTask 取消等待旧任务收尾期间，若新任务（如扫描尾部的
// runAutoFingerprint）已启动，绝不能把新任务的进度覆写成 cancelled——否则前端轮询
// 看到「已停止」而 ffmpeg 仍在后台跑，正是 songloft-org/songloft#323 要消灭的状态。
func TestCancel_DoesNotClobberNewTask(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	svc := NewFingerprintService(mdb.SongRepository())

	oldDone := make(chan struct{})
	newDone := make(chan struct{})

	svc.mu.Lock()
	svc.running = true
	svc.done = oldDone
	svc.progress = FingerprintProgress{Status: "running", Total: 100, Computed: 10}
	// cancelFn 里模拟「旧任务收尾的同时新任务抢先启动并换掉 done」
	svc.cancelFn = func() {
		go func() {
			svc.mu.Lock()
			svc.done = newDone
			svc.progress = FingerprintProgress{Status: "running", Total: 50}
			svc.mu.Unlock()
			close(oldDone)
		}()
	}
	svc.mu.Unlock()

	// 等待期间新任务顶上来了：Cancel 没有真正停下任何在跑的东西，
	// 必须报 false（否则前端会停掉轮询并显示「已停止」，而 ffmpeg 还在跑）
	if svc.Cancel() {
		t.Error("Cancel() 未能停下新任务时不应返回 true")
	}
	if got := svc.GetProgress().Status; got != "running" {
		t.Errorf("新任务的进度被覆写了: status = %q, want \"running\"", got)
	}
}

// TestFingerprintService_ComputeMissingEmpty 无待计算歌曲时立即完成，不留下 running 状态。
func TestFingerprintService_ComputeMissingEmpty(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	svc := NewFingerprintService(mdb.SongRepository())

	total, err := svc.ComputeMissing()
	if err != nil {
		t.Fatalf("compute missing: %v", err)
	}
	if total != 0 {
		t.Errorf("total: got %d want 0", total)
	}
	if got := svc.GetProgress().Status; got != "done" {
		t.Errorf("status: got %q want \"done\"", got)
	}
	// 没有任务在跑，Cancel 应返回 false 而不是阻塞
	if svc.Cancel() {
		t.Error("Cancel() with no running task: got true want false")
	}
}
