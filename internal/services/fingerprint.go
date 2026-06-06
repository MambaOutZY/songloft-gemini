package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"songloft/internal/database"
)

var (
	fpcalcAvailable     bool
	fpcalcAvailableOnce sync.Once
)

// IsFpcalcAvailable 检测 fpcalc 是否可用（首次调用时检测，结果缓存）。
func IsFpcalcAvailable() bool {
	fpcalcAvailableOnce.Do(func() {
		_, err := exec.LookPath("fpcalc")
		fpcalcAvailable = err == nil
	})
	return fpcalcAvailable
}

type fpcalcOutput struct {
	Fingerprint string  `json:"fingerprint"`
	Duration    float64 `json:"duration"`
}

// ExtractFingerprint 调用 fpcalc -json 提取音频指纹。
func ExtractFingerprint(ctx context.Context, filePath string) (string, float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "fpcalc", "-json", filePath).Output()
	if err != nil {
		return "", 0, fmt.Errorf("fpcalc: %w", err)
	}

	var result fpcalcOutput
	if err := json.Unmarshal(out, &result); err != nil {
		return "", 0, fmt.Errorf("fpcalc parse: %w", err)
	}
	if result.Fingerprint == "" {
		return "", 0, fmt.Errorf("fpcalc returned empty fingerprint")
	}
	return result.Fingerprint, result.Duration, nil
}

// FingerprintProgress 指纹计算进度。
type FingerprintProgress struct {
	Status   string `json:"status"` // idle, running, done
	Computed int64  `json:"computed"`
	Total    int64  `json:"total"`
	Failed   int64  `json:"failed"`
}

// FingerprintService 管理指纹计算的异步任务。
type FingerprintService struct {
	songs SongRepository

	mu       sync.Mutex
	running  bool
	progress FingerprintProgress
}

// NewFingerprintService 创建指纹服务。
func NewFingerprintService(songs SongRepository) *FingerprintService {
	return &FingerprintService{
		songs:    songs,
		progress: FingerprintProgress{Status: "idle"},
	}
}

// GetProgress 返回当前计算进度。
func (s *FingerprintService) GetProgress() FingerprintProgress {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.progress
}

// ComputeMissing 异步为所有缺失指纹的本地歌曲计算指纹。
func (s *FingerprintService) ComputeMissing() (int, error) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return 0, fmt.Errorf("fingerprint computation already in progress")
	}
	s.running = true
	s.mu.Unlock()

	ctx := context.Background()
	missing, err := s.songs.ListLocalWithoutFingerprint(ctx)
	if err != nil {
		s.mu.Lock()
		s.running = false
		s.progress = FingerprintProgress{Status: "idle"}
		s.mu.Unlock()
		return 0, fmt.Errorf("list missing: %w", err)
	}

	total := len(missing)
	s.mu.Lock()
	s.progress = FingerprintProgress{Status: "running", Total: int64(total)}
	s.mu.Unlock()

	if total == 0 {
		s.mu.Lock()
		s.running = false
		s.progress = FingerprintProgress{Status: "done", Total: 0}
		s.mu.Unlock()
		return 0, nil
	}

	go s.doCompute(missing)
	return total, nil
}

const fpWorkers = 4

func (s *FingerprintService) doCompute(items []database.SongIDPath) {
	defer func() {
		s.mu.Lock()
		s.running = false
		s.progress.Status = "done"
		s.mu.Unlock()
	}()

	var computed, failed atomic.Int64
	ch := make(chan database.SongIDPath, fpWorkers*2)

	var wg sync.WaitGroup
	for i := 0; i < fpWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			for item := range ch {
				fp, dur, err := ExtractFingerprint(ctx, item.FilePath)
				if err != nil {
					slog.Info("fingerprint failed", "id", item.ID, "path", item.FilePath, "err", err)
					failed.Add(1)
				} else {
					if err := s.songs.UpdateFingerprint(ctx, item.ID, fp, dur); err != nil {
						slog.Warn("fingerprint save failed", "id", item.ID, "err", err)
						failed.Add(1)
					} else {
						computed.Add(1)
					}
				}
				s.mu.Lock()
				s.progress.Computed = computed.Load()
				s.progress.Failed = failed.Load()
				s.mu.Unlock()
			}
		}()
	}

	for _, item := range items {
		ch <- item
	}
	close(ch)
	wg.Wait()

	slog.Info("fingerprint computation done", "computed", computed.Load(), "failed", failed.Load(), "total", len(items))
}
