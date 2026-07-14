package httputil

import (
	"bufio"
	"io"
)

// ICYDeinterleaveReader 把交织了 ICY 元数据的 Shoutcast/Icecast 音频流「去交织」，
// 只吐出纯音频字节，丢弃元数据块。
//
// 背景：部分 Shoutcast v1 服务器**无条件**在音频流里按 icy-metaint 间隔插入元数据，
// 即使客户端没在请求里带 Icy-MetaData 头。浏览器 <audio>/HTML5 不解析这种交织块，
// 会把元数据字节当音频解码，一个 metaint 间隔后即崩断（约 2-3 秒）。给这类下游代理时
// 用本 Reader 剥掉元数据，浏览器就能拿到连续纯音频。(issue #275)
//
// 交织流字节协议（循环）：
//
//	[ metaint 字节纯音频 ][ 1 字节长度 L ][ L*16 字节元数据 ][ metaint 字节音频 ][ L2 ]...
//
// L 为 uint8，元数据块实际长度 = L*16 字节（L=0 表示本轮无元数据，很常见）。
//
// 注意：只用于下游客户端**没有**请求 ICY 元数据的场景。原生播放器（ExoPlayer/libmpv）
// 会自带 Icy-MetaData 头并自行解析交织块，那种路径应原样透传，不要套本 Reader。
type ICYDeinterleaveReader struct {
	src       *bufio.Reader // 包裹上游 body
	metaint   int           // 每段音频固定字节数
	remaining int           // 本段音频还剩多少字节没吐给下游；初值 = metaint
}

// NewICYDeinterleaveReader 创建去交织 Reader。metaint 为上游 icy-metaint 头的值（>0）。
func NewICYDeinterleaveReader(r io.Reader, metaint int) *ICYDeinterleaveReader {
	return &ICYDeinterleaveReader{
		src:       bufio.NewReader(r),
		metaint:   metaint,
		remaining: metaint,
	}
}

func (d *ICYDeinterleaveReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// 到达元数据边界：读长度字节并跳过其后的元数据块。
	if d.remaining == 0 {
		lb, err := d.src.ReadByte()
		if err != nil {
			// 边界处 EOF 即流干净结束；其它错误一并透传，io.Copy 会停止。
			return 0, err
		}
		if metaLen := int(lb) * 16; metaLen > 0 {
			if _, err := d.src.Discard(metaLen); err != nil {
				// 元数据块被截断：视为流结束。
				return 0, err
			}
		}
		d.remaining = d.metaint
	}

	// 吐音频：读取上限 cap 到 remaining，绝不越界读入元数据区。
	want := min(len(p), d.remaining)
	n, err := d.src.Read(p[:want])
	d.remaining -= n
	return n, err
}
