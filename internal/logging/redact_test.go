package logging

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRedact(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string // 期望脱敏后包含
		notWant string // 期望脱敏后不再包含（敏感原文）
	}{
		{
			name:    "slog 文本键值密钥",
			in:      `time=... level=INFO msg=start jwt_secret=abcDEF123 foo=bar`,
			notWant: "abcDEF123",
			want:    "jwt_secret=***",
		},
		{
			name:    "JSON 形式 access_token",
			in:      `{"access_token":"eyJhbGciOi","expires":3600}`,
			notWant: "eyJhbGciOi",
			want:    `"access_token":"***`,
		},
		{
			name:    "默认密码明文",
			in:      `默认管理员账号: admin，默认密码: s3cr3t`,
			notWant: "s3cr3t",
			want:    "***",
		},
		{
			name:    "Authorization 头",
			in:      `headers=map[Authorization:tokenvalue123 X-Foo:bar]`,
			notWant: "tokenvalue123",
			want:    "Authorization:***",
		},
		{
			name:    "Bearer token",
			in:      `Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.abc`,
			notWant: "eyJhbGciOiJIUzI1NiJ9.abc",
			want:    "Bearer ***",
		},
		{
			name:    "URL 内嵌凭证",
			in:      `proxy=http://user:passwd@192.168.1.1:7890`,
			notWant: "user:passwd@",
			want:    "http://***:***@",
		},
		{
			name:    "客户端 IP 脱敏保留网段",
			in:      `remote=203.0.113.45:51234`,
			notWant: "203.0.113.45",
			want:    "203.0.*.*",
		},
		{
			name: "回环地址不脱敏",
			in:   `listening on 127.0.0.1:58091`,
			want: "127.0.0.1",
		},
		{
			name:    "Unix 用户目录",
			in:      `music path=/home/alice/Music`,
			notWant: "/home/alice",
			want:    "/home/<user>",
		},
		{
			name:    "Windows 用户目录",
			in:      `path=C:\Users\Bob\AppData`,
			notWant: `\Users\Bob`,
			want:    `Users\<user>`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Redact(c.in)
			if c.notWant != "" && strings.Contains(got, c.notWant) {
				t.Errorf("脱敏后仍包含敏感内容 %q\n输入: %s\n输出: %s", c.notWant, c.in, got)
			}
			if c.want != "" && !strings.Contains(got, c.want) {
				t.Errorf("脱敏后未包含期望内容 %q\n输入: %s\n输出: %s", c.want, c.in, got)
			}
		})
	}
}

func TestRedactStream(t *testing.T) {
	in := "line1 password=hunter2\nline2 ok\n"
	var sb strings.Builder
	if err := RedactStream(&sb, strings.NewReader(in)); err != nil {
		t.Fatalf("RedactStream 出错: %v", err)
	}
	out := sb.String()
	if strings.Contains(out, "hunter2") {
		t.Errorf("流式脱敏后仍含密码: %s", out)
	}
	if !strings.Contains(out, "line2 ok") {
		t.Errorf("流式脱敏丢失了正常行: %s", out)
	}
	if strings.Count(out, "\n") != 2 {
		t.Errorf("流式脱敏行数不符，期望 2 个换行: %q", out)
	}
}

// buildMixedLog 造一份跨多个并行块的日志：混入每一条规则各自的命中样本、
// 完全不含敏感内容的普通行，以及多字节字符，用来对比并行与串行输出。
func buildMixedLog(lines int) string {
	samples := []string{
		`time=... level=INFO msg=start jwt_secret=abcDEF123 foo=bar`,
		`{"access_token":"eyJhbGciOi","expires":3600}`,
		`默认管理员账号: admin，默认密码: s3cr3t`,
		`headers=map[Authorization:tokenvalue123 X-Foo:bar]`,
		`Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.abc`,
		`proxy=http://user:passwd@192.168.1.1:7890`,
		`remote=203.0.113.45:51234`,
		`listening on 127.0.0.1:58091`,
		`music path=/home/alice/Music`,
		`path=C:\Users\Bob\AppData`,
		`扫描完成，新增 12 首歌曲，耗时 340ms`,
		`plain line with no sensitive content at all`,
		`ſecret=unicodefold`,
		`api_key: k-123456`,
	}
	var sb strings.Builder
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&sb, "%06d %s\n", i, samples[i%len(samples)])
	}
	return sb.String()
}

// 并行脱敏的输出必须与串行逐字节一致。这是 RedactStream 并行化的正确性根据：
// 不依赖对日志内容的任何假设，破坏块内顺序或块间顺序都会让本用例变红。
func TestRedactStreamParallelMatchesSerial(t *testing.T) {
	// 行数刻意远大于 redactChunkMaxLines，保证跨多个块并且最后一块不满。
	in := buildMixedLog(redactChunkMaxLines*7 + 13)

	var serial strings.Builder
	if err := redactStreamSerial(&serial, strings.NewReader(in)); err != nil {
		t.Fatalf("串行脱敏出错: %v", err)
	}

	for _, workers := range []int{2, 3, 8} {
		var parallel strings.Builder
		if err := redactStreamParallel(&parallel, strings.NewReader(in), workers); err != nil {
			t.Fatalf("workers=%d 并行脱敏出错: %v", workers, err)
		}
		if parallel.String() != serial.String() {
			t.Errorf("workers=%d 并行输出与串行不一致", workers)
		}
	}

	// 导出端点走的是 RedactStream 这个入口，一并锁住。
	var viaEntry strings.Builder
	if err := RedactStream(&viaEntry, strings.NewReader(in)); err != nil {
		t.Fatalf("RedactStream 出错: %v", err)
	}
	if viaEntry.String() != serial.String() {
		t.Errorf("RedactStream 输出与串行不一致（GOMAXPROCS=%d）", runtime.GOMAXPROCS(0))
	}
}

// 行序必须严格保留。并行实现里块间定序一旦丢失，脱敏内容看起来仍然是对的，
// 只有逐行序号能发现——所以单独用一条不含任何敏感内容的输入来钉。
func TestRedactStreamPreservesLineOrder(t *testing.T) {
	const lines = redactChunkMaxLines*5 + 7
	var sb strings.Builder
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&sb, "line-%06d\n", i)
	}

	var out strings.Builder
	if err := redactStreamParallel(&out, strings.NewReader(sb.String()), 4); err != nil {
		t.Fatalf("并行脱敏出错: %v", err)
	}

	got := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(got) != lines {
		t.Fatalf("行数不符：期望 %d，实得 %d", lines, len(got))
	}
	for i, line := range got {
		if want := fmt.Sprintf("line-%06d", i); line != want {
			t.Fatalf("第 %d 行乱序：期望 %q，实得 %q", i, want, line)
		}
	}
}

// 字面量闸门不能挡掉真正该脱敏的行。每条加了闸门的规则各来一例。
func TestRedactGatesDoNotSkipRealMatches(t *testing.T) {
	cases := []struct {
		rule    string
		in      string
		notWant string
	}{
		{"reSecretCN", `默认密码: s3cr3t`, "s3cr3t"},
		{"reSecretCN 令牌", `令牌：tok-abc123`, "tok-abc123"},
		{"reURLCred", `proxy=http://user:passwd@example.com`, "user:passwd@"},
		// 非 http scheme 单独一例：闸门判据必须是 `://` 而不是某个具体 scheme，
		// 否则 redis/ftp/socks5 这类 URL 里的凭证会漏出去。
		{"reURLCred 非 http scheme", `cache=redis://admin:s3cr3t@10.0.0.5:6379`, "admin:s3cr3t@"},
		{"reURLCred socks5", `proxy=socks5://u:p@example.com:1080`, "u:p@"},
		{"reIPv4", `remote=203.0.113.45`, "203.0.113.45"},
		{"reUnixHome /home/", `path=/home/alice/Music`, "/home/alice"},
		{"reUnixHome /Users/", `path=/Users/alice/Music`, "/Users/alice"},
		{"reWinHome", `path=C:\Users\Bob\AppData`, `\Users\Bob`},
		{"reWinHome 小写盘符", `path=d:\users\bob\AppData`, `\users\bob`},
	}
	for _, c := range cases {
		t.Run(c.rule, func(t *testing.T) {
			if got := Redact(c.in); strings.Contains(got, c.notWant) {
				t.Errorf("闸门挡掉了该脱敏的内容 %q\n输入: %s\n输出: %s", c.notWant, c.in, got)
			}
		})
	}
}

// reSecretKV / reAuthHeader / reBearer 三条带 `(?i)`，而 Go 的 `(?i)` 是 Unicode
// 简单折叠：`(?i)secret` 也匹配 `ſecret`（U+017F）、`(?i)token` 也匹配带 Kelvin
// 记号（U+212A）的 `toKen`。给它们加 ASCII 折叠的字面量闸门在真实日志上测不出
// 差异，却会让这两行漏脱敏——把密钥贴进公开 issue。本用例就是那道闸门的反向守卫：
// 谁给这三条规则加了 ASCII 闸门，这里立刻变红。
func TestRedactUnicodeFoldedKeywordsStillRedacted(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		notWant string
	}{
		{"长 s 的 secret", "\u017Fecret=abcDEF123", "abcDEF123"},
		{"Kelvin 记号的 token", "to\u212Aen=abcDEF123", "abcDEF123"},
		{"Kelvin 记号的 api_key", "api_\u212Aey=abcDEF123", "abcDEF123"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Redact(c.in); strings.Contains(got, c.notWant) {
				t.Errorf("Unicode 折叠关键字未脱敏 %q\n输入: %q\n输出: %q", c.notWant, c.in, got)
			}
		})
	}
}

// failAfterWriter 写满 limit 字节后开始报错，模拟客户端中途断开。
type failAfterWriter struct {
	limit   int
	written int
}

var errWriteFailed = errors.New("下游写失败")

func (w *failAfterWriter) Write(p []byte) (int, error) {
	if w.written >= w.limit {
		return 0, errWriteFailed
	}
	w.written += len(p)
	return len(p), nil
}

// countingReader 记录一共被读了多少字节。
type countingReader struct {
	inner io.Reader
	read  int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.inner.Read(p)
	atomic.AddInt64(&r.read, int64(n))
	return n, err
}

// 下游写失败必须原样返回、不能吊死，**并且生产者要经 abort 提前收工**——
// 客户端断开后把剩下 10 MiB 读完再脱敏一遍，是白烧一整个请求的 CPU。
//
// 只断言「错误被返回」是不够的：去掉 close(abort) 之后错误照样返回，
// 差别只在有没有白干。所以这里直接量源被读了多少字节。
// 上界是硬的：在飞块数被通道容量卡在 work(workers) + order(2×workers) 以内，
// 再加 Scanner 最多 64KiB 的预读，因此远小于整份输入。
func TestRedactStreamWriteErrorPropagates(t *testing.T) {
	in := buildMixedLog(redactChunkMaxLines * 60)
	src := &countingReader{inner: strings.NewReader(in)}

	done := make(chan error, 1)
	go func() {
		done <- redactStreamParallel(&failAfterWriter{limit: 1}, src, 4)
	}()

	select {
	case err := <-done:
		if !errors.Is(err, errWriteFailed) {
			t.Errorf("期望返回下游写错误，实得 %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("写失败后并行脱敏未返回（疑似死锁）")
	}

	read := atomic.LoadInt64(&src.read)
	if limit := int64(len(in)) / 2; read >= limit {
		t.Errorf("写失败后仍读了 %d/%d 字节（≥%d），生产者没有提前收工", read, len(in), limit)
	}
}

// 末行不带换行的输入，输出仍补一个换行——与串行实现原有行为一致。
func TestRedactStreamTrailingNewline(t *testing.T) {
	for _, in := range []string{"a\nb", "a\nb\n"} {
		var serial, parallel strings.Builder
		if err := redactStreamSerial(&serial, strings.NewReader(in)); err != nil {
			t.Fatalf("串行脱敏出错: %v", err)
		}
		if err := redactStreamParallel(&parallel, strings.NewReader(in), 4); err != nil {
			t.Fatalf("并行脱敏出错: %v", err)
		}
		if serial.String() != "a\nb\n" {
			t.Errorf("输入 %q 串行输出为 %q，期望 %q", in, serial.String(), "a\nb\n")
		}
		if parallel.String() != serial.String() {
			t.Errorf("输入 %q 并行输出 %q 与串行 %q 不一致", in, parallel.String(), serial.String())
		}
	}
}

// 超过 bufio.Scanner 默认 64KiB 上限的长行必须完整通过（放宽到 1MiB 的那条约束）。
func TestRedactStreamLongLine(t *testing.T) {
	long := strings.Repeat("x", 200*1024)
	in := "short\n" + long + " password=hunter2\nafter\n"

	var out strings.Builder
	if err := redactStreamParallel(&out, strings.NewReader(in), 4); err != nil {
		t.Fatalf("并行脱敏出错: %v", err)
	}
	got := out.String()
	if strings.Contains(got, "hunter2") {
		t.Error("长行里的密码未脱敏")
	}
	if !strings.Contains(got, long) {
		t.Error("长行被截断")
	}
	if !strings.Contains(got, "after") {
		t.Error("长行之后的行丢失")
	}
}

// 分块的双上限。字节这条只影响内存不影响输出，比对输出的用例抓不到它被删掉，
// 所以单独钉：单行就超过字节上限时必须立刻派块，否则 512 行 × 1 MiB 的最坏情况
// 会让一个块攒到几百 MB，把 RedactStream「不整份读进内存」的约束丢掉。
func TestRedactBlockFullHonoursBothLimits(t *testing.T) {
	cases := []struct {
		name  string
		lines int
		bytes int
		want  bool
	}{
		{"都没到上限", redactChunkMaxLines - 1, redactChunkBytes - 1, false},
		{"行数到上限", redactChunkMaxLines, 0, true},
		{"单行就超过字节上限", 1, redactChunkBytes, true},
		{"少量长行超过字节上限", 3, redactChunkBytes + 1, true},
		{"空块", 0, 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := redactBlockFull(c.lines, c.bytes); got != c.want {
				t.Errorf("redactBlockFull(%d, %d) = %v，期望 %v", c.lines, c.bytes, got, c.want)
			}
		})
	}
}

// 一整片长行（每行都超过分块字节上限）也要逐字节与串行一致。
// 这条走的是「单行即成块」那条分支，与普通短行的攒块路径不同。
func TestRedactStreamLongLineStream(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&sb, "%03d %s password=hunter%d\n", i, strings.Repeat("x", redactChunkBytes), i)
	}
	in := sb.String()

	var serial, parallel strings.Builder
	if err := redactStreamSerial(&serial, strings.NewReader(in)); err != nil {
		t.Fatalf("串行脱敏出错: %v", err)
	}
	if err := redactStreamParallel(&parallel, strings.NewReader(in), 4); err != nil {
		t.Fatalf("并行脱敏出错: %v", err)
	}
	if parallel.String() != serial.String() {
		t.Error("长行流的并行输出与串行不一致")
	}
	if strings.Contains(parallel.String(), "hunter7") {
		t.Error("长行流里的密码未脱敏")
	}
}
