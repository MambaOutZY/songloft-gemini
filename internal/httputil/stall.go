package httputil

import (
	"context"
	"io"
	"time"
)

// StallReader 包裹下载 body 做「停滞检测」：每次成功读到字节就重置计时器；
// 连续 idle 时长内读不到任何数据（真死链）才调用 cancel 掐断底层请求，
// 使阻塞中的 Read 立即返回错误。
//
// 与 http.Client.Timeout 不同，它从不限制传输的**总时长**——慢但持续推进的
// 下载（如通过慢速代理/梯子拉大文件）能一直跑完，不会在固定秒数处被判死。
// (issue #265: "write temp: context deadline exceeded")
//
// 用法：搭配无整请求超时的 client（NewStreamingClient）与可取消的请求 ctx，
//
//	dlCtx, cancel := context.WithCancel(ctx)
//	req, _ := http.NewRequestWithContext(dlCtx, ...)
//	resp, _ := client.Do(req)
//	sr := NewStallReader(resp.Body, cancel, idle)
//	defer sr.Stop()
//	_, err := io.Copy(dst, sr)
type StallReader struct {
	r     io.Reader
	timer *time.Timer
	idle  time.Duration
}

// NewStallReader 创建停滞检测 Reader。idle 内无任何字节到达即触发 cancel。
// cancel 应为承载该 HTTP 请求的 context 的取消函数；重复调用无害（幂等）。
func NewStallReader(r io.Reader, cancel context.CancelFunc, idle time.Duration) *StallReader {
	return &StallReader{
		r:     r,
		timer: time.AfterFunc(idle, cancel),
		idle:  idle,
	}
}

func (s *StallReader) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	if n > 0 {
		// 有进展就续命；若计时器已触发过 cancel，此处 Reset 无实际影响，
		// 因为底层 ctx 已取消，后续 Read 会自然报错退出。
		s.timer.Reset(s.idle)
	}
	return n, err
}

// Stop 停止计时器，应在下载结束后 defer 调用，避免残留 goroutine 误触发 cancel。
func (s *StallReader) Stop() {
	s.timer.Stop()
}
