package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"songloft/internal/models"
)

// TestAsyncDownloadAndCache_DedupsBySongID 验证流式代理触发的后台全量下载按 song.ID 去重：
// 客户端重试 / 并发 206 分片会多次进入此路径，若不去重会在慢网下发起多个互相抢带宽的
// 下载而全败（songloft-org/songloft#286）。这里预置 inflight 标记模拟「已有下载在跑」，
// 断言第二次同 ID 调用立即短路、不再发起 HTTP 请求。
func TestAsyncDownloadAndCache_DedupsBySongID(t *testing.T) {
	var reqCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount.Add(1)
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("audio-bytes"))
	}))
	defer srv.Close()

	cs := &CacheService{cacheDir: t.TempDir(), downloadClient: srv.Client()}
	song := &models.Song{ID: 4242, Type: "remote", PluginEntryPath: "ytdlp", DedupKey: "vid_p1"}

	// 模拟该歌的后台下载已在进行中：预置 inflight 标记
	cs.asyncCacheInflight.Store(song.ID, struct{}{})

	done := make(chan struct{})
	go func() {
		cs.AsyncDownloadAndCache(context.Background(), song, srv.URL, nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("已有 inflight 时 AsyncDownloadAndCache 未立即短路返回")
	}

	if got := reqCount.Load(); got != 0 {
		t.Fatalf("已有 inflight 时应短路、不发起下载；却发起了 %d 次请求", got)
	}
}
