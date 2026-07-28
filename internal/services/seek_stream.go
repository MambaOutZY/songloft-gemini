package services

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ErrSeekStreamUnavailable 表示 seek 流无法开始（ffmpeg 未配置 / 启动失败 / 并发已满，
// 或进程在吐出任何音频字节之前就退出）。调用方拿到此错误时 **尚未** 向 w 写入任何字节，
// 可安全降级为「从头完整提供该文件」。
var ErrSeekStreamUnavailable = errors.New("seek stream unavailable")

// seekStreamMaxConcurrent 同时存活的 seek 流上限。
//
// seek 流不占用 transcodeSem（见 StreamSeekedMP3 注释），所以必须自带闸门：它由「用户每按一次
// 暂停/续播」触发，频率比电台转码高一个数量级，且分组播放会给组内每台音箱各开一个流。满了直接
// 降级为从头播而 **不排队**——排队意味着音箱那端干等，比从头播更糟。
const seekStreamMaxConcurrent = 4

// seekStreamGrace 是 ffmpeg 硬超时相对于「剩余音频时长」的宽限。
// 音箱只挂着连接却不再读数据时 r.Context() 不会取消，靠这个上限回收孤儿进程。
const seekStreamGrace = 5 * time.Minute

var seekStreamSem = make(chan struct{}, seekStreamMaxConcurrent)

// SeekStreamOptions seek 流参数。
type SeekStreamOptions struct {
	// SourcePath 本机已存在的音频文件：原始文件、转码产物或 CUE 提取产物。
	// 必须是「曲内偏移 0 对应该文件开头」的文件——CUE 轨要先提取再 seek，
	// 否则两个 -ss 叠加会跳到整轨镜像的绝对位置。
	SourcePath string
	// StartSecond 起播秒数，必须 > 0。
	StartSecond float64
	// RemainingSecond 预计剩余音频时长（秒），用于设置 ffmpeg 硬超时；<= 0 表示未知（用固定兜底）。
	RemainingSecond float64
}

// StreamSeekedMP3 阻塞运行 ffmpeg，把 SourcePath 从 StartSecond 起的音频以 MP3 流写入 w，
// 直到 ffmpeg 结束或 ctx 取消（客户端断开）。
//
// 用于不支持 HTTP Range seek 的推流客户端：小爱音箱经 player_play_url 收到 URL 后只会从头拉，
// 想「从第 N 秒续播」只能让服务端产出一条以第 N 秒为开头的流（songloft-plugin-miot#60）。
//
// 与 runFFmpeg 不同，本函数 **不占用** c.transcodeSem：进程会存活整首歌的剩余时长，
// 持串行信号量会长时间饿死其他有限文件的转码。并发由 seekStreamSem 单独限。
//
// 返回值语义与 StreamTranscodedRadio 一致：
//   - nil：正常结束（含客户端断开导致的 ctx 取消）。
//   - ErrSeekStreamUnavailable：在写出任何字节前失败，调用方应降级为从头提供文件（此时 w 未被写入）。
//   - 其他 error：已写出部分字节后中途失败，无法再降级。
func (c *CacheService) StreamSeekedMP3(ctx context.Context, w io.Writer, opts SeekStreamOptions) error {
	if opts.StartSecond <= 0 {
		return fmt.Errorf("%w: non-positive start second", ErrSeekStreamUnavailable)
	}
	ffmpegPath := c.ffmpegPath
	if ffmpegPath == "" {
		return fmt.Errorf("%w: ffmpeg not configured", ErrSeekStreamUnavailable)
	}

	select {
	case seekStreamSem <- struct{}{}:
		defer func() { <-seekStreamSem }()
	default:
		return fmt.Errorf("%w: %d concurrent seek streams already running", ErrSeekStreamUnavailable, seekStreamMaxConcurrent)
	}

	// 源已是 MP3 时无损 copy：input seek + copy 几乎零 CPU（实测 320k 整首 0.14s）。
	//
	// 需要重编码时用 CBR 而非项目别处的 VBR（-q:a 0）：pipe 不可 seek → 无法回写 Xing 帧
	// （见下面 -write_xing 0），拿到流的客户端只能按「首帧码率 × 字节数」估算时长。VBR 下这个
	// 估算会偏差 7%+（实测 105.4s 的流被估成 97.6s），CBR 下几乎精确。音箱正是这类只能靠估算的
	// 客户端，估短了可能提前判定播完。320k CBR 对无损源（flac/ape）也不构成可闻损失。
	//
	// 已知残留：源本身是 VBR MP3（含 force_mp3/normalize 转码产物，那条路径用 -q:a 0）时走 copy，
	// 这个时长估算同样会偏几个百分点。只影响客户端自己显示的总时长——插件的进度与自动切歌都由
	// 它本地计时驱动（见 PlaylistManager.playCurrent）——故不为此放弃零开销的 copy 快路径。
	encoder := "libmp3lame"
	var qualityArgs []string
	if NormalizeFormat(filepath.Ext(opts.SourcePath)) == "mp3" {
		encoder = "copy"
	} else {
		qualityArgs = []string{"-b:a", "320k"}
	}

	args := []string{"-hide_banner", "-loglevel", "error"}
	// -ss 放在 -i 之前用 input seek（O(1)），放后面会解码丢弃前 N 秒。
	args = append(args, "-ss", fmt.Sprintf("%.3f", opts.StartSecond), "-i", opts.SourcePath)
	// -map 0:a:0 必需：mp3 muxer 只接受单条音频流，双音轨源（.mka，songloft-org/songloft#298）
	// 不显式选轨会让 ffmpeg 直接报错退出，静默降级回「从头播」。
	args = append(args, "-map", "0:a:0", "-vn", "-codec:a", encoder)
	args = append(args, qualityArgs...)
	// -write_xing 0 必需：pipe 不可 seek，mp3 muxer 无法回写 Xing 帧的真实时长，
	// 留下的占位值会让音箱显示错误时长甚至提前判定播完。
	args = append(args, "-write_xing", "0", "-f", "mp3", "pipe:1")

	// 音箱只挂着连接不再读数据时 r.Context() 不取消，用硬超时回收孤儿 ffmpeg。
	timeout := seekStreamGrace
	if opts.RemainingSecond > 0 {
		timeout += time.Duration(opts.RemainingSecond) * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, ffmpegPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("%w: stdout pipe: %v", ErrSeekStreamUnavailable, err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &capWriter{w: &stderr, remaining: 4096}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: start: %v", ErrSeekStreamUnavailable, err)
	}

	reader := bufio.NewReaderSize(stdout, 64*1024)
	// 预读第一字节：只有确认 ffmpeg 真的产出了音频才提交响应。若在任何字节之前就 EOF/失败
	// （seek 越过文件尾、源不可解码、选轨失败等），返回 ErrSeekStreamUnavailable 让调用方
	// 降级为从头提供文件——此时尚未写出任何字节。
	first, _ := reader.Peek(1)
	if len(first) == 0 {
		_ = cmd.Wait()
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = "no output"
		}
		return fmt.Errorf("%w: %s", ErrSeekStreamUnavailable, msg)
	}

	// 已确认有音频输出：从此刻起本函数拥有响应，流式转发直到结束或断开。
	_, copyErr := io.Copy(&flushingWriter{w: w}, reader)
	waitErr := cmd.Wait()

	if ctx.Err() != nil {
		return nil
	}
	if copyErr != nil {
		return copyErr
	}
	if waitErr != nil {
		slog.Warn("seek stream ffmpeg exited with error",
			"src", opts.SourcePath, "seek", opts.StartSecond,
			"stderr", strings.TrimSpace(stderr.String()), "error", waitErr)
	}
	return nil
}
