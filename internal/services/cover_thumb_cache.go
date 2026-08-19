package services

import (
	"container/heap"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	coverThumbCacheMaxSize    = 200 << 20 // 200MB
	coverThumbCacheEvictRatio = 0.9       // 淘汰到 90%
)

// CoverThumbCache 缩略图磁盘缓存。按 ETag hash 存储已生成的 JPEG 缩略图，
// 避免重复解码+缩放 CPU 开销。
type CoverThumbCache struct {
	dir       string
	maxSize   int64
	mu        sync.Mutex
	totalSize int64
	ready     chan struct{} // 异步扫描完成后关闭
}

// NewCoverThumbCache 创建缩略图缓存，dataDir 为数据根目录。
func NewCoverThumbCache(dataDir string) *CoverThumbCache {
	dir := filepath.Join(dataDir, "cover_thumbs")
	os.MkdirAll(dir, 0755)
	c := &CoverThumbCache{
		dir:     dir,
		maxSize: coverThumbCacheMaxSize,
		ready:   make(chan struct{}),
	}
	go c.scan()
	return c
}

// Get 查找缓存。hash 为不含引号的十六进制 SHA1。
// 命中返回文件绝对路径和 true；未命中返回 "", false。
func (c *CoverThumbCache) Get(hash string) (string, bool) {
	p := c.pathFor(hash)
	info, err := os.Stat(p)
	if err != nil {
		return "", false
	}
	if info.Size() == 0 {
		return "", false
	}
	// 更新 atime 用于 LRU（仅 touch mtime，便携且无需 syscall）
	now := time.Now()
	os.Chtimes(p, now, now)
	return p, true
}

// Put 写入缩略图缓存。data 为 JPEG 字节。返回写入的文件路径。
func (c *CoverThumbCache) Put(hash string, data []byte) (string, error) {
	p := c.pathFor(hash)
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建缩略图缓存目录失败：%w", err)
	}

	// 原子写入：同目录临时文件 + rename，不存在跨设备问题
	tmp, err := os.CreateTemp(dir, ".thumb-*.tmp")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败：%w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return "", fmt.Errorf("写入临时文件失败：%w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("关闭临时文件失败：%w", err)
	}
	if err := os.Rename(tmpName, p); err != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("重命名缓存文件失败：%w", err)
	}

	c.mu.Lock()
	c.totalSize += int64(len(data))
	needEvict := c.totalSize > c.maxSize
	c.mu.Unlock()

	if needEvict {
		go c.evict()
	}
	return p, nil
}

// Clear 清空全部缩略图缓存。
func (c *CoverThumbCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := os.RemoveAll(c.dir); err != nil {
		return err
	}
	os.MkdirAll(c.dir, 0755)
	c.totalSize = 0
	return nil
}

// Stats 返回缓存统计。
func (c *CoverThumbCache) Stats() (totalSize int64, fileCount int) {
	c.mu.Lock()
	size := c.totalSize
	c.mu.Unlock()

	count := 0
	filepath.WalkDir(c.dir, func(_ string, d os.DirEntry, _ error) error {
		if d != nil && !d.IsDir() {
			count++
		}
		return nil
	})
	return size, count
}

func (c *CoverThumbCache) pathFor(hash string) string {
	if len(hash) < 4 {
		return filepath.Join(c.dir, hash+".jpg")
	}
	return filepath.Join(c.dir, hash[0:2], hash[2:4], hash+".jpg")
}

// scan 启动时异步统计已有缓存大小。
func (c *CoverThumbCache) scan() {
	defer close(c.ready)
	var total int64
	filepath.WalkDir(c.dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if info, e := d.Info(); e == nil {
			total += info.Size()
		}
		return nil
	})
	c.mu.Lock()
	c.totalSize += total
	c.mu.Unlock()
	slog.Debug("缩略图缓存扫描完成", "dir", c.dir, "size_mb", total>>20)
}

// evict 淘汰最旧的缓存文件，直到 totalSize <= maxSize * evictRatio。
func (c *CoverThumbCache) evict() {
	c.mu.Lock()
	if c.totalSize <= c.maxSize {
		c.mu.Unlock()
		return
	}
	target := int64(float64(c.maxSize) * coverThumbCacheEvictRatio)
	c.mu.Unlock()

	// 收集所有缓存文件
	var entries thumbEntryHeap
	filepath.WalkDir(c.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		info, e := d.Info()
		if e != nil {
			return nil
		}
		entries = append(entries, thumbEntry{path: path, size: info.Size(), mtime: info.ModTime()})
		return nil
	})

	// 按 mtime 升序（最旧在前）
	heap.Init(&entries)

	var freed int64
	c.mu.Lock()
	currentSize := c.totalSize
	c.mu.Unlock()

	for entries.Len() > 0 && currentSize-freed > target {
		e := heap.Pop(&entries).(thumbEntry)
		if err := os.Remove(e.path); err == nil {
			freed += e.size
			// 尝试清理空父目录（两级）
			parent := filepath.Dir(e.path)
			os.Remove(parent) // 非空时自动失败
			grandparent := filepath.Dir(parent)
			if grandparent != c.dir {
				os.Remove(grandparent)
			}
		}
	}

	c.mu.Lock()
	c.totalSize -= freed
	if c.totalSize < 0 {
		c.totalSize = 0
	}
	c.mu.Unlock()

	if freed > 0 {
		slog.Debug("缩略图缓存淘汰完成", "freed_mb", freed>>20, "remaining_mb", (currentSize-freed)>>20)
	}
}

// thumbEntry 用于 LRU 堆排序
type thumbEntry struct {
	path  string
	size  int64
	mtime time.Time
}

type thumbEntryHeap []thumbEntry

func (h thumbEntryHeap) Len() int            { return len(h) }
func (h thumbEntryHeap) Less(i, j int) bool  { return h[i].mtime.Before(h[j].mtime) }
func (h thumbEntryHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *thumbEntryHeap) Push(x interface{}) { *h = append(*h, x.(thumbEntry)) }
func (h *thumbEntryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	e := old[n-1]
	*h = old[:n-1]
	return e
}
