package services

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
			name: "起播秒数非正",
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
