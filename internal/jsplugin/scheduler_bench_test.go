package jsplugin

import (
	"context"
	"testing"
)

// BenchmarkSchedulerCall 量单插件串行队列的同步 Call 往返吞吐（不含 JS 执行）。
func BenchmarkSchedulerCall(b *testing.B) {
	sched := NewServiceScheduler(1)
	defer sched.Close()

	resp := &Message{}
	handler := &mockHandler{response: resp}
	if err := sched.RegisterService("bench", handler, defaultQueueSize); err != nil {
		b.Fatalf("RegisterService: %v", err)
	}

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := sched.Call(ctx, "bench", "", MsgHostCall, nil, 0); err != nil {
			b.Fatalf("Call: %v", err)
		}
	}
}

// BenchmarkSchedulerSend 量异步 Send 的投递吞吐。
func BenchmarkSchedulerSend(b *testing.B) {
	sched := NewServiceScheduler(1)
	defer sched.Close()

	handler := &mockHandler{}
	if err := sched.RegisterService("bench", handler, b.N+16); err != nil {
		b.Fatalf("RegisterService: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sched.Send(&Message{Type: MsgHostCall, Target: "bench"})
	}
}
