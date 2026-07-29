package services

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// countingWriter 记录写入次数，用于断言「降级前绝不向响应写字节」这条不变量。
type countingWriter struct {
	writes int
	buf    bytes.Buffer
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	cw.writes++
	return cw.buf.Write(p)
}

// TestStreamSeekedMP3Unavailable 验证各种「无法开始」的情形都返回 ErrSeekStreamUnavailable
// 且一个字节都没写出——调用方正是靠这个契约无损降级为「从头完整提供文件」。
func TestStreamSeekedMP3Unavailable(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.mp3")
	if err := os.WriteFile(src, []byte("not really mp3"), 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	cases := []struct {
		name string
		cs   *CacheService
		opts SeekStreamOptions
	}{
		{
			name: "ffmpeg 未配置",
			cs:   &CacheService{},
			opts: SeekStreamOptions{SourcePath: src, StartSecond: 30},
		},
		{
			name: "ffmpeg 可执行文件不存在",
			cs:   &CacheService{ffmpegPath: filepath.Join(dir, "nonexistent-ffmpeg")},
			opts: SeekStreamOptions{SourcePath: src, StartSecond: 30},
		},
		{
			name: "ffmpeg 退出但零输出",
			cs:   &CacheService{ffmpegPath: "/bin/true"},
			opts: SeekStreamOptions{SourcePath: src, StartSecond: 30},
		},
		{
			// StartSecond 0 只在 Normalize 时合法（纯均衡流），不均衡时仍是「没活干」。
			name: "起播秒数非正且未开均衡",
			cs:   &CacheService{ffmpegPath: "/bin/echo"},
			opts: SeekStreamOptions{SourcePath: src, StartSecond: 0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cw := &countingWriter{}
			err := tc.cs.StreamSeekedMP3(context.Background(), cw, tc.opts)
			if !errors.Is(err, ErrSeekStreamUnavailable) {
				t.Fatalf("err = %v, want ErrSeekStreamUnavailable", err)
			}
			if cw.writes != 0 {
				t.Errorf("写入次数 = %d，期望 0（降级前不得写响应）", cw.writes)
			}
		})
	}
}

// TestStreamSeekedMP3FFmpegArgs 用 /bin/echo 冒充 ffmpeg，把参数契约钉住：
// 缺 -map 0:a:0 会让双音轨源（.mka，songloft-org/songloft#298）直接失败，
// 缺 -write_xing 0 会让音箱读到错误时长。两者都只会表现为静默降级，极难归因。
func TestStreamSeekedMP3FFmpegArgs(t *testing.T) {
	dir := t.TempDir()

	run := func(t *testing.T, name string) string {
		t.Helper()
		src := filepath.Join(dir, name)
		if err := os.WriteFile(src, []byte("x"), 0644); err != nil {
			t.Fatalf("write src: %v", err)
		}
		cs := &CacheService{ffmpegPath: "/bin/echo"}
		var buf bytes.Buffer
		if err := cs.StreamSeekedMP3(context.Background(), &buf, SeekStreamOptions{
			SourcePath: src, StartSecond: 61.5, RemainingSecond: 100,
		}); err != nil {
			t.Fatalf("StreamSeekedMP3: %v", err)
		}
		return buf.String()
	}

	t.Run("mp3 源走无损 copy", func(t *testing.T) {
		out := run(t, "a.mp3")
		for _, want := range []string{"-ss 61.500", "-map 0:a:0", "-vn", "-codec:a copy", "-write_xing 0", "-f mp3", "pipe:1"} {
			if !bytes.Contains([]byte(out), []byte(want)) {
				t.Errorf("ffmpeg 参数缺 %q，实际: %s", want, out)
			}
		}
		// input seek：-ss 必须在 -i 之前，否则退化为解码丢弃前 N 秒
		if idxSS, idxI := bytes.Index([]byte(out), []byte("-ss")), bytes.Index([]byte(out), []byte("-i")); idxSS > idxI {
			t.Errorf("-ss 出现在 -i 之后（非 input seek）: %s", out)
		}
	})

	t.Run("非 mp3 源按 CBR 重编码", func(t *testing.T) {
		out := run(t, "a.flac")
		if !bytes.Contains([]byte(out), []byte("-codec:a libmp3lame")) {
			t.Errorf("期望 libmp3lame 重编码，实际: %s", out)
		}
		// 必须是 CBR：无 Xing 头的流只能按码率估时长，VBR 会让音箱把时长估短约 7%
		if !bytes.Contains([]byte(out), []byte("-b:a 320k")) {
			t.Errorf("期望 CBR -b:a 320k，实际: %s", out)
		}
	})
}

// TestStreamSeekedMP3NormalizeArgs 钉住「边转边发音量均衡」的参数契约
// （songloft-org/songloft-plugin-miot#61）：
//   - mp3 源也必须重编码，走 copy 的话 loudnorm 滤镜静默失效、听感与落盘产物不一致；
//   - StartSecond 为 0 时不能出现 -ss，纯均衡流应从头开始；
//   - StartSecond > 0 时 seek 与均衡由同一条 ffmpeg 一起做。
func TestStreamSeekedMP3NormalizeArgs(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.mp3")
	if err := os.WriteFile(src, []byte("x"), 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	run := func(t *testing.T, startSecond float64) string {
		t.Helper()
		cs := &CacheService{ffmpegPath: "/bin/echo"}
		var buf bytes.Buffer
		if err := cs.StreamSeekedMP3(context.Background(), &buf, SeekStreamOptions{
			SourcePath: src, StartSecond: startSecond, RemainingSecond: 100, Normalize: true,
		}); err != nil {
			t.Fatalf("StreamSeekedMP3: %v", err)
		}
		return buf.String()
	}

	t.Run("纯均衡不带 seek", func(t *testing.T) {
		out := run(t, 0)
		for _, want := range []string{
			"-af " + loudnormFilter, // 与落盘转码共用同一滤镜串，否则两条路径响度不同
			"-codec:a libmp3lame",   // mp3 源也不能 copy
			"-b:a 320k",             // pipe 无法回写 Xing，只有 CBR 能让音箱估准时长
			"-map 0:a:0", "-vn", "-write_xing 0", "-f mp3", "pipe:1",
		} {
			if !bytes.Contains([]byte(out), []byte(want)) {
				t.Errorf("ffmpeg 参数缺 %q，实际: %s", want, out)
			}
		}
		if bytes.Contains([]byte(out), []byte("-ss")) {
			t.Errorf("StartSecond=0 不该出现 -ss，实际: %s", out)
		}
	})

	t.Run("均衡叠加 seek", func(t *testing.T) {
		out := run(t, 42)
		for _, want := range []string{"-ss 42.000", "-af " + loudnormFilter, "-codec:a libmp3lame"} {
			if !bytes.Contains([]byte(out), []byte(want)) {
				t.Errorf("ffmpeg 参数缺 %q，实际: %s", want, out)
			}
		}
		if idxSS, idxI := bytes.Index([]byte(out), []byte("-ss")), bytes.Index([]byte(out), []byte("-i")); idxSS > idxI {
			t.Errorf("-ss 出现在 -i 之后（非 input seek）: %s", out)
		}
	})
}

// TestSeekStreamUnknownDurationTimeout 钉住「剩余时长未知时的硬超时不能退化成 5 分钟」。
//
// 5 分钟对整首流不是宽限而是硬上限：客户端边播边读时恰好只送出 5 分钟音频就被 CommandContext
// 杀掉，而响应正常闭合、客户端以为播完了。songs.duration == 0 是常态（远程歌曲元数据未刷新、
// 本地文件 tag 与 ffprobe 都拿不到时长），叠上「均衡流每次播放都走这条路」就会真踩到。
func TestSeekStreamUnknownDurationTimeout(t *testing.T) {
	if seekStreamUnknownDurationTimeout <= seekStreamGrace {
		t.Fatalf("未知时长的硬超时 %v 不得小于等于 seekStreamGrace %v——那样长音频会被静默截断",
			seekStreamUnknownDurationTimeout, seekStreamGrace)
	}
	// 至少要覆盖有声书量级；这里取 1 小时作为下界，避免日后有人调小到分钟级
	if seekStreamUnknownDurationTimeout < time.Hour {
		t.Errorf("未知时长的硬超时 %v 太短，覆盖不了长音频", seekStreamUnknownDurationTimeout)
	}
}

// TestStreamSeekedMP3NormalizeRealFFmpeg 用真 ffmpeg 端到端验证纯均衡流：
// 输出是可解的 MP3，且时长与源一致（没被 -ss 或滤镜吃掉开头）。
func TestStreamSeekedMP3NormalizeRealFFmpeg(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "tone.mp3")
	gen := exec.Command(ffmpegPath, "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6", "-codec:a", "libmp3lame", "-y", src)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Skipf("generate sample mp3 failed: %v (%s)", err, out)
	}

	cs := &CacheService{ffmpegPath: ffmpegPath}
	var buf bytes.Buffer
	if err := cs.StreamSeekedMP3(context.Background(), &buf, SeekStreamOptions{
		SourcePath: src, StartSecond: 0, RemainingSecond: 6, Normalize: true,
	}); err != nil {
		t.Fatalf("StreamSeekedMP3: %v", err)
	}

	out := buf.Bytes()
	if len(out) == 0 {
		t.Fatal("输出为空")
	}
	isID3 := bytes.HasPrefix(out, []byte("ID3"))
	isFrameSync := len(out) >= 2 && out[0] == 0xFF && out[1]&0xE0 == 0xE0
	if !isID3 && !isFrameSync {
		t.Errorf("输出不是 MP3 流: % x", out[:min(4, len(out))])
	}
	// 320k CBR 的 6 秒约 240 KB；只要明显大于「被截掉开头」的量级即可，不做精确断言。
	if len(out) < 100*1024 {
		t.Errorf("输出仅 %d 字节，纯均衡流不该丢开头", len(out))
	}
}

// TestStreamSeekedMP3RealFFmpeg 用真 ffmpeg 端到端验证输出确实是从 StartSecond 起的 MP3。
func TestStreamSeekedMP3RealFFmpeg(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "tone.mp3")
	gen := exec.Command(ffmpegPath, "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6", "-codec:a", "libmp3lame", "-y", src)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Skipf("generate sample mp3 failed: %v (%s)", err, out)
	}

	cs := &CacheService{ffmpegPath: ffmpegPath}
	var buf bytes.Buffer
	if err := cs.StreamSeekedMP3(context.Background(), &buf, SeekStreamOptions{
		SourcePath: src, StartSecond: 4, RemainingSecond: 2,
	}); err != nil {
		t.Fatalf("StreamSeekedMP3: %v", err)
	}

	out := buf.Bytes()
	if len(out) == 0 {
		t.Fatal("输出为空")
	}
	// 合法 MP3 流开头：ID3v2 头（ffmpeg mp3 muxer 默认写元数据）或直接是帧同步头
	// （0xFF 后接 0xE0 的高 3 位，即 11 个 1 bit）。
	isID3 := bytes.HasPrefix(out, []byte("ID3"))
	isFrameSync := len(out) >= 2 && out[0] == 0xFF && out[1]&0xE0 == 0xE0
	if !isID3 && !isFrameSync {
		t.Errorf("输出不是 MP3 流: % x", out[:min(4, len(out))])
	}

	full, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read src: %v", err)
	}
	// 6 秒里跳掉 4 秒，剩余字节应显著少于整首（同码率下约 1/3）
	if len(out) >= len(full) {
		t.Errorf("seek 后字节数 %d 不小于整首 %d，input seek 可能未生效", len(out), len(full))
	}
}
