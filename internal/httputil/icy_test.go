package httputil

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// buildICYStream 把音频段与元数据块交织成一条 Shoutcast 流。
// audioSegs 是各段音频(每段应恰为 metaint 字节)；metaBlocks 是各段之后插入的元数据原文
// (会自动 pad 到 16 的倍数，空串表示该轮无元数据即长度字节 0)。段数应与元数据块数一致。
func buildICYStream(audioSegs []string, metaBlocks []string) []byte {
	var buf bytes.Buffer
	for i, seg := range audioSegs {
		buf.WriteString(seg)
		meta := ""
		if i < len(metaBlocks) {
			meta = metaBlocks[i]
		}
		if meta == "" {
			buf.WriteByte(0) // 长度字节 0：本轮无元数据
			continue
		}
		// pad 到 16 的倍数
		padded := meta
		if r := len(padded) % 16; r != 0 {
			padded += strings.Repeat("\x00", 16-r)
		}
		buf.WriteByte(byte(len(padded) / 16))
		buf.WriteString(padded)
	}
	return buf.Bytes()
}

func TestICYDeinterleaveReader(t *testing.T) {
	const metaint = 16
	audio := []string{
		strings.Repeat("A", metaint),
		strings.Repeat("B", metaint),
		strings.Repeat("C", metaint),
	}
	// 第一段后有元数据，第二段后无(L=0)，第三段后有。
	metas := []string{"StreamTitle='x';", "", "StreamTitle='yy';"}
	want := "AAAAAAAAAAAAAAAABBBBBBBBBBBBBBBBCCCCCCCCCCCCCCCC"

	// 用不同 Read buffer 大小驱动，覆盖部分读与缓冲边界。
	for _, bufSize := range []int{1, 3, 7, 16, 64, 4096} {
		t.Run("bufSize="+itoa(bufSize), func(t *testing.T) {
			raw := buildICYStream(audio, metas)
			d := NewICYDeinterleaveReader(bytes.NewReader(raw), metaint)

			var out bytes.Buffer
			p := make([]byte, bufSize)
			for {
				n, err := d.Read(p)
				out.Write(p[:n])
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("read: %v", err)
				}
			}
			if out.String() != want {
				t.Errorf("去交织输出=%q\n期望=%q", out.String(), want)
			}
		})
	}
}

// TestICYDeinterleaveReaderTruncated 验证流在各处被截断时不 panic，且已解出的纯音频保留。
func TestICYDeinterleaveReaderTruncated(t *testing.T) {
	const metaint = 16
	full := buildICYStream([]string{strings.Repeat("A", metaint), strings.Repeat("B", metaint)}, []string{"StreamTitle='x';", ""})

	// 在多个截断点各跑一次：音频段中途、长度字节处、元数据块中途。
	for _, cut := range []int{0, 5, metaint, metaint + 1, metaint + 8, len(full) - 1} {
		if cut > len(full) {
			continue
		}
		d := NewICYDeinterleaveReader(bytes.NewReader(full[:cut]), metaint)
		out, err := io.ReadAll(d)
		if err != nil {
			t.Fatalf("cut=%d ReadAll: %v", cut, err)
		}
		// 输出只能是纯音频字节(A/B)，绝不含元数据字符。
		for _, b := range out {
			if b != 'A' && b != 'B' {
				t.Fatalf("cut=%d 输出含非音频字节 %q，元数据泄漏", cut, b)
			}
		}
	}
}

// itoa 小工具，避免为测试引入 strconv。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
