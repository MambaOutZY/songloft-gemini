package httputil

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

// slowReader 每次 Read 前 sleep delay,模拟慢但持续推进的连接。
type slowReader struct {
	remaining int
	delay     time.Duration
}

func (s *slowReader) Read(p []byte) (int, error) {
	if s.remaining <= 0 {
		return 0, io.EOF
	}
	time.Sleep(s.delay)
	p[0] = 'x'
	s.remaining--
	return 1, nil
}

// blockingReader 第一次 Read 就阻塞直到 ctx 取消,模拟停滞的死链。
type blockingReader struct {
	ctx context.Context
}

func (b *blockingReader) Read(p []byte) (int, error) {
	<-b.ctx.Done()
	return 0, b.ctx.Err()
}

// 慢但持续推进的下载:每次 Read 间隔 < idle,不应被掐断,能读完。
func TestStallReader_SlowSteadyCompletes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 20 字节 × 5ms 间隔 = 总耗时 ~100ms,远超 idle=30ms,但每次都有进展。
	sr := NewStallReader(&slowReader{remaining: 20, delay: 5 * time.Millisecond}, cancel, 30*time.Millisecond)
	defer sr.Stop()

	n, err := io.Copy(io.Discard, sr)
	if err != nil {
		t.Fatalf("慢但持续的下载不应报错,got: %v", err)
	}
	if n != 20 {
		t.Fatalf("应读完 20 字节,got %d", n)
	}
	if ctx.Err() != nil {
		t.Fatalf("持续推进不应触发 cancel,got ctx.Err()=%v", ctx.Err())
	}
}

// 停滞的死链:idle 内读不到字节,应触发 cancel 使 Read 返回错误。
func TestStallReader_StallTriggersCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sr := NewStallReader(&blockingReader{ctx: ctx}, cancel, 20*time.Millisecond)
	defer sr.Stop()

	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(io.Discard, sr)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("停滞应因 ctx 取消而报错,got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("停滞检测未在预期时间内掐断下载")
	}
}
