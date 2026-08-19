package database

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// :memory: 的每个连接都是一张独立的空库（modernc.org/sqlite 无 shared cache），迁移只
// 建在 runMigrations 拿到的那一个连接上。池一旦扩容，新连接就没有任何表。
//
// 这些用例锁住三件事：内存库并发可用、内存库真的被限成单连接、以及文件库没被顺带限死。
// 前者是症状，中间是机制，最后一条防止「修好测试的同时把生产吞吐砍到 1」。

func TestMemoryDBConcurrentQueriesSeeTheSchema(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// 并发度刻意高于原先的 MaxIdleConns(5) 与 MaxOpenConns(10)，这样在修复前必然
	// 逼出新连接。修复前的表现是部分 goroutine 报 "no such table: js_plugins"。
	const n = 24
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var count int
			errs[idx] = db.DB().
				QueryRowContext(context.Background(), "SELECT COUNT(*) FROM js_plugins").
				Scan(&count)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("并发查询 %d 失败（内存库连接池扩容后拿到空库）: %v", i, err)
		}
	}
}

// 上一个用例可能因为调度巧合而没扩容就通过，所以这里直接验机制：内存库上第二个
// 连接必须拿不到——池上限就是 1。
func TestMemoryDBIsLimitedToOneConnection(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	held, err := db.DB().Conn(context.Background())
	if err != nil {
		t.Fatalf("取第一个连接: %v", err)
	}
	defer held.Close()

	// 第一个连接仍被持有时申请第二个：池已满，应当一直等到 ctx 超时。
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	second, err := db.DB().Conn(ctx)
	if err == nil {
		second.Close()
		t.Fatal("内存库同时给出了第二个连接：那个连接看到的是一张没有表的空库")
	}
	if ctx.Err() == nil {
		t.Errorf("期望因池满而等待超时，实际错误: %v", err)
	}
}

// 文件库多个连接指向同一个文件，不存在上面的问题，必须保留并发能力。
// 少了这条，把 MaxOpenConns 无条件设成 1 也能让另外两个用例全绿。
func TestFileDBKeepsConnectionPool(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "pool.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if got := db.DB().Stats().MaxOpenConnections; got != 10 {
		t.Errorf("文件库 MaxOpenConnections = %d, want 10", got)
	}

	ctx := context.Background()
	first, err := db.DB().Conn(ctx)
	if err != nil {
		t.Fatalf("取第一个连接: %v", err)
	}
	defer first.Close()

	// 与内存库相反：这里第二个连接必须立刻拿到。
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	second, err := db.DB().Conn(ctx2)
	if err != nil {
		t.Fatalf("文件库拿不到第二个连接（连接池被误限成单连接）: %v", err)
	}
	defer second.Close()

	var count int
	if err := second.QueryRowContext(ctx, "SELECT COUNT(*) FROM js_plugins").Scan(&count); err != nil {
		t.Errorf("文件库第二个连接查询失败: %v", err)
	}
}

func TestIsMemoryDSN(t *testing.T) {
	tests := []struct {
		dsn  string
		want bool
	}{
		{":memory:", true},
		{"file::memory:", true},
		{"file:test.db?mode=memory&cache=shared", true},
		{"data/songloft.db", false},
		{"/var/lib/songloft/songloft.db", false},
		{"", false},
		// 真实文件名里出现 "memory" 不该被误判成内存库。
		{"data/memory-notes.db", false},
	}
	for _, tt := range tests {
		if got := isMemoryDSN(tt.dsn); got != tt.want {
			t.Errorf("isMemoryDSN(%q) = %v, want %v", tt.dsn, got, tt.want)
		}
	}
}
