package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"songloft/internal/models"
)

var ErrUnsupportedTranscodeFormat = errors.New("unsupported transcode format")

// SetFFmpegPath 注入 ffmpeg 可执行文件路径。
func (c *CacheService) SetFFmpegPath(path string) {
	c.ffmpegPath = path
}

// GetOrTranscode 获取转码后的文件路径。
//  1. 原格式==目标格式 → 返回 srcPath
//  2. 转码缓存命中 → 返回缓存路径
//  3. miss → ffmpeg 转码 → 写入缓存 → 返回
//
// srcPath 是原始音频文件路径（本地文件路径或已下载的缓存文件路径）。
// targetFormat 是标准化后的格式名（mp3/ogg/m4a/flac/wav）。
func (c *CacheService) GetOrTranscode(ctx context.Context, srcPath string, song *models.Song, targetFormat string) (string, error) {
	if song == nil {
		return "", errors.New("song is nil")
	}
	if !NeedsTranscode(song.Format, targetFormat) {
		return srcPath, nil
	}

	// 1. 缓存命中
	if p, ok := c.FindTranscodedFile(song, targetFormat); ok {
		return p, nil
	}

	// 2. inflight 去重
	inflightKey := fmt.Sprintf("tc_%d_%s", song.ID, targetFormat)
	state := getSongState()
	state.transcodeInflightMu.Lock()
	if dl, ok := state.transcodeInflight[inflightKey]; ok {
		state.transcodeInflightMu.Unlock()
		<-dl.done
		if dl.err != nil {
			return "", dl.err
		}
		if p, ok := c.FindTranscodedFile(song, targetFormat); ok {
			return p, nil
		}
		return "", fmt.Errorf("transcoded file not found after wait")
	}
	dl := &inflightDownload{done: make(chan struct{})}
	state.transcodeInflight[inflightKey] = dl
	state.transcodeInflightMu.Unlock()
	defer func() {
		state.transcodeInflightMu.Lock()
		delete(state.transcodeInflight, inflightKey)
		state.transcodeInflightMu.Unlock()
		close(dl.done)
	}()

	// 3. 转码
	finalPath, err := c.doTranscode(ctx, srcPath, song, targetFormat)
	if err != nil {
		dl.err = err
		return "", err
	}

	c.touchSongLRU(song.ID)
	go c.EvictLRU()
	return finalPath, nil
}

// FindTranscodedFile 查找已转码的缓存文件。
// 文件名形如 "{id}.{key}.tc.{format}" 或 "{id}.tc.{format}"。
func (c *CacheService) FindTranscodedFile(song *models.Song, targetFormat string) (string, bool) {
	if song == nil {
		return "", false
	}
	name := c.transcodedFileName(song, targetFormat)
	dir, _ := c.getCachePath(song.ID, "")
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		c.touchSongLRU(song.ID)
		return path, true
	}
	return "", false
}

// doTranscode 执行 ffmpeg 转码并写入缓存。
func (c *CacheService) doTranscode(ctx context.Context, srcPath string, song *models.Song, targetFormat string) (string, error) {
	dir, _ := c.getCachePath(song.ID, "")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("mkdir transcode cache dir: %w", err)
	}

	// 临时文件放在目标目录（同设备，rename 无 EXDEV）
	tmp, err := os.CreateTemp(dir, "tc-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()

	if err := c.runFFmpeg(ctx, srcPath, tmpPath, targetFormat); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("ffmpeg transcode: %w", err)
	}

	finalName := c.transcodedFileName(song, targetFormat)
	finalPath := filepath.Join(dir, finalName)
	if _, err := os.Stat(finalPath); err == nil {
		os.Remove(finalPath)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("rename transcoded file: %w", err)
	}

	slog.Info("transcode completed", "songId", song.ID, "format", targetFormat, "path", finalPath)
	return finalPath, nil
}

// runFFmpeg 调用 ffmpeg 执行转码。
func (c *CacheService) runFFmpeg(ctx context.Context, srcPath, dstPath, targetFormat string) error {
	encoder, qualityArgs, err := ffmpegArgs(targetFormat)
	if err != nil {
		return err
	}

	args := []string{"-i", srcPath, "-vn", "-codec:a", encoder}
	args = append(args, qualityArgs...)
	args = append(args, "-y", dstPath)

	ffmpegPath := c.ffmpegPath
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("ffmpeg failed", "args", args, "output", string(output), "error", err)
		return fmt.Errorf("ffmpeg exit: %w", err)
	}
	return nil
}

// transcodedFileName 生成转码缓存文件名。
func (c *CacheService) transcodedFileName(song *models.Song, targetFormat string) string {
	idStr := strconv.FormatInt(song.ID, 10)
	key := cacheKeyOf(song)
	if key != "" {
		return idStr + "." + key + ".tc." + targetFormat
	}
	return idStr + ".tc." + targetFormat
}

// NeedsTranscode 判断是否需要转码。
func NeedsTranscode(srcFormat, targetFormat string) bool {
	if targetFormat == "" {
		return false
	}
	return NormalizeFormat(srcFormat) != NormalizeFormat(targetFormat)
}

// NormalizeFormat 统一格式名称，处理别名。
func NormalizeFormat(f string) string {
	f = strings.ToLower(strings.TrimPrefix(f, "."))
	switch f {
	case "mpeg", "mp3":
		return "mp3"
	case "mp4", "m4a", "aac":
		return "m4a"
	case "ogg", "vorbis":
		return "ogg"
	case "flac":
		return "flac"
	case "wav", "wave":
		return "wav"
	case "wma", "asf":
		return "wma"
	case "ape":
		return "ape"
	}
	return f
}

// ffmpegArgs 根据目标格式返回 ffmpeg 编码器和质量参数。
func ffmpegArgs(targetFormat string) (encoder string, qualityArgs []string, err error) {
	switch NormalizeFormat(targetFormat) {
	case "mp3":
		return "libmp3lame", []string{"-q:a", "0"}, nil
	case "ogg":
		return "libvorbis", []string{"-q:a", "6"}, nil
	case "m4a":
		return "aac", []string{"-b:a", "256k"}, nil
	case "flac":
		return "flac", nil, nil
	case "wav":
		return "pcm_s16le", nil, nil
	default:
		return "", nil, ErrUnsupportedTranscodeFormat
	}
}
