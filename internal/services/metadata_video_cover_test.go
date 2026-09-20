package services

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestExtractCoverFromVideoFile_FFmpegNotConfigured：FFMpegPath 为空时应直接返回错误，不 spawn 任何进程。
func TestExtractCoverFromVideoFile_FFmpegNotConfigured(t *testing.T) {
	ext := NewMetadataExtractor(&MetadataConfig{FFMpegPath: ""})
	_, err := ext.ExtractCoverFromVideoFile(context.Background(), "/does/not/matter.mp4", 10)
	if err == nil {
		t.Fatal("expected error when ffmpeg not configured, got nil")
	}
}

// TestExtractCoverFromVideoFile_RealFFmpeg：用真 ffmpeg 生成 3s testsrc mp4，抽帧应产出 JPEG。
// 校验点：1) 返回的 coverPath 非空；2) 文件存在且以 JPEG magic bytes (0xFF 0xD8) 开头；3) 大小 <= maxCoverSize。
func TestExtractCoverFromVideoFile_RealFFmpeg(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not available")
	}

	tmp := t.TempDir()
	videoPath := filepath.Join(tmp, "sample.mp4")
	// lavfi testsrc：3 秒、24fps、128x72 彩条；-pix_fmt yuv420p 确保 H.264 兼容。
	genCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	genArgs := []string{
		"-y",
		"-f", "lavfi",
		"-i", "testsrc=duration=3:size=128x72:rate=24",
		"-pix_fmt", "yuv420p",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		videoPath,
	}
	if out, err := exec.CommandContext(genCtx, ffmpegPath, genArgs...).CombinedOutput(); err != nil {
		t.Skipf("cannot synthesize test video (libx264 unavailable?): %v\n%s", err, string(out))
	}

	coverStorage := t.TempDir()
	ext := NewMetadataExtractor(&MetadataConfig{
		FFMpegPath:       ffmpegPath,
		CoverStoragePath: coverStorage,
	})

	coverPath, err := ext.ExtractCoverFromVideoFile(context.Background(), videoPath, 3.0)
	if err != nil {
		t.Fatalf("ExtractCoverFromVideoFile failed: %v", err)
	}
	if coverPath == "" {
		t.Fatal("expected non-empty coverPath")
	}

	data, err := os.ReadFile(coverPath)
	if err != nil {
		t.Fatalf("read cover: %v", err)
	}
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		t.Fatalf("expected JPEG magic bytes, got % x ...", data[:min(4, len(data))])
	}
	if len(data) > maxCoverSize {
		t.Fatalf("cover size %d exceeds maxCoverSize %d", len(data), maxCoverSize)
	}
}

// TestExtractCoverFromVideoFile_ZeroDuration：duration<=0 时走 -ss 0 分支，仍应能抽出首帧。
func TestExtractCoverFromVideoFile_ZeroDuration(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not available")
	}

	tmp := t.TempDir()
	videoPath := filepath.Join(tmp, "sample.mp4")
	genCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(genCtx, ffmpegPath,
		"-y", "-f", "lavfi", "-i", "testsrc=duration=1:size=64x64:rate=10",
		"-pix_fmt", "yuv420p", "-c:v", "libx264", "-preset", "ultrafast", videoPath,
	).CombinedOutput(); err != nil {
		t.Skipf("cannot synthesize test video: %v\n%s", err, string(out))
	}

	coverStorage := t.TempDir()
	ext := NewMetadataExtractor(&MetadataConfig{FFMpegPath: ffmpegPath, CoverStoragePath: coverStorage})
	coverPath, err := ext.ExtractCoverFromVideoFile(context.Background(), videoPath, 0)
	if err != nil {
		t.Fatalf("ExtractCoverFromVideoFile duration=0 failed: %v", err)
	}
	if coverPath == "" {
		t.Fatal("expected non-empty coverPath")
	}
}
