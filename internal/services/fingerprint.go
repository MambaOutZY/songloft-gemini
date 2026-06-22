package services

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"songloft/internal/database"
)

var (
	chromaprintAvailable bool
	chromaprintOnce      sync.Once
	resolvedFFmpegPath   string
	durationRe           = regexp.MustCompile(`Duration:\s+(\d+):(\d+):(\d+)\.(\d+)`)
)

// IsChromaprintAvailable 检测 ffmpeg 是否支持 chromaprint muxer（首次调用时检测，结果缓存）。
func IsChromaprintAvailable() bool {
	chromaprintOnce.Do(func() {
		path, err := safeLookPath("ffmpeg")
		if err != nil {
			return
		}
		out, err := exec.Command(path, "-hide_banner", "-muxers").Output()
		if err == nil && strings.Contains(string(out), "chromaprint") {
			chromaprintAvailable = true
			resolvedFFmpegPath = path
		}
	})
	return chromaprintAvailable
}

func parseDurationFromStderr(stderr string) float64 {
	matches := durationRe.FindStringSubmatch(stderr)
	if len(matches) < 5 {
		return 0
	}
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])
	frac, _ := strconv.Atoi(matches[4])
	divisor := 1.0
	for i := 0; i < len(matches[4]); i++ {
		divisor *= 10
	}
	return float64(hours)*3600 + float64(minutes)*60 + float64(seconds) + float64(frac)/divisor
}

// ExtractFingerprint 调用 ffmpeg chromaprint muxer 提取音频指纹。
func ExtractFingerprint(ctx context.Context, filePath string) (string, float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, resolvedFFmpegPath,
		"-i", filePath,
		"-map", "0:a:0",
		"-map_metadata", "-1",
		"-f", "chromaprint", "-fp_format", "base64", "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", 0, fmt.Errorf("ffmpeg chromaprint: %w (%s)", err, stderr.String())
	}

	fingerprint := strings.TrimSpace(stdout.String())
	if nl := strings.IndexByte(fingerprint, '\n'); nl >= 0 {
		fingerprint = fingerprint[:nl]
	}
	if fingerprint == "" {
		return "", 0, fmt.Errorf("ffmpeg chromaprint returned empty fingerprint")
	}

	duration := parseDurationFromStderr(stderr.String())
	return fingerprint, duration, nil
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
	cancelFn context.CancelFunc
	done     chan struct{}
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

// ComputeMissing 为所有缺失指纹的本地歌曲计算指纹。
// 若已有任务在运行，打断旧任务后重新启动。
func (s *FingerprintService) ComputeMissing() (int, error) {
	return s.startCompute(false)
}

// RecomputeAll 清空所有已有指纹后重新计算全部本地歌曲的指纹。
// 若已有任务在运行，打断旧任务后重新启动。
func (s *FingerprintService) RecomputeAll() (int, error) {
	return s.startCompute(true)
}

func (s *FingerprintService) startCompute(clearFirst bool) (int, error) {
	s.mu.Lock()
	if s.running {
		s.cancelFn()
		done := s.done
		s.mu.Unlock()
		<-done
		s.mu.Lock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel
	s.running = true
	s.done = make(chan struct{})
	s.mu.Unlock()

	if clearFirst {
		if err := s.songs.ClearAllFingerprints(ctx); err != nil {
			cancel()
			s.mu.Lock()
			s.running = false
			close(s.done)
			s.progress = FingerprintProgress{Status: "idle"}
			s.mu.Unlock()
			return 0, fmt.Errorf("clear fingerprints: %w", err)
		}
	}

	missing, err := s.songs.ListLocalWithoutFingerprint(ctx)
	if err != nil {
		cancel()
		s.mu.Lock()
		s.running = false
		close(s.done)
		s.progress = FingerprintProgress{Status: "idle"}
		s.mu.Unlock()
		return 0, fmt.Errorf("list missing: %w", err)
	}

	total := len(missing)
	s.mu.Lock()
	s.progress = FingerprintProgress{Status: "running", Total: int64(total)}
	s.mu.Unlock()

	if total == 0 {
		cancel()
		s.mu.Lock()
		s.running = false
		close(s.done)
		s.progress = FingerprintProgress{Status: "done", Total: 0}
		s.mu.Unlock()
		return 0, nil
	}

	go s.doCompute(ctx, missing)
	return total, nil
}

const fpWorkers = 4

func (s *FingerprintService) doCompute(ctx context.Context, items []database.SongIDPath) {
	defer func() {
		s.mu.Lock()
		s.running = false
		s.progress.Status = "done"
		close(s.done)
		s.mu.Unlock()
	}()

	var computed, failed atomic.Int64
	ch := make(chan database.SongIDPath, fpWorkers*2)

	var wg sync.WaitGroup
	for i := 0; i < fpWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				select {
				case <-ctx.Done():
					return
				default:
				}

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

loop:
	for _, item := range items {
		select {
		case <-ctx.Done():
			break loop
		case ch <- item:
		}
	}
	close(ch)
	wg.Wait()

	slog.Info("fingerprint computation done", "computed", computed.Load(), "failed", failed.Load(), "total", len(items))
}
