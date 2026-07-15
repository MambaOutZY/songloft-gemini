package source

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// rtFunc 把函数适配成 http.RoundTripper。
type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// newFakeClient 返回一个直接吐预设 body 的 client,用于在不经真实网络栈的前提下
// 构造"body 干净 EOF 却短于 Content-Length"这种真实网络栈会自行拦成 ErrUnexpectedEOF、
// 从而无法覆盖到 downloadToTemp 显式截断检查的场景。
func newFakeClient(body []byte, contentLength int64) *http.Client {
	return &http.Client{Transport: rtFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Body:          io.NopCloser(bytes.NewReader(body)),
			ContentLength: contentLength,
			Header:        make(http.Header),
		}, nil
	})}
}

func newFetcherWithClient(c *http.Client) *SourceFetcher {
	return NewSourceFetcher(FetcherOpts{HTTPClient: c})
}

// Content-Length 声明 1000 但只吐 500 字节 → 判为截断,报错且不留临时文件。
func TestDownloadToTemp_Truncated(t *testing.T) {
	body := bytes.Repeat([]byte("a"), 500)
	f := newFetcherWithClient(newFakeClient(body, 1000))

	path, n, err := f.downloadToTemp(context.Background(), "http://x/audio", nil)
	if err == nil {
		t.Fatalf("expected truncation error, got path=%q n=%d", path, n)
	}
	if !strings.Contains(err.Error(), "truncated") {
		t.Fatalf("expected truncated error, got %v", err)
	}
	if path != "" {
		if _, statErr := os.Stat(path); statErr == nil {
			t.Fatalf("temp file should be removed on truncation: %s", path)
		}
	}
}

// Content-Length 未知(-1,chunked / gzip 透明解压)→ 不误判,下载成功。
func TestDownloadToTemp_UnknownLengthNoFalsePositive(t *testing.T) {
	body := bytes.Repeat([]byte("a"), 500)
	f := newFetcherWithClient(newFakeClient(body, -1))

	path, n, err := f.downloadToTemp(context.Background(), "http://x/audio", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(path)
	if n != 500 {
		t.Fatalf("expected 500 bytes, got %d", n)
	}
}

// Content-Length 与实际一致 → 成功。
func TestDownloadToTemp_ExactLength(t *testing.T) {
	body := bytes.Repeat([]byte("a"), 500)
	f := newFetcherWithClient(newFakeClient(body, 500))

	path, n, err := f.downloadToTemp(context.Background(), "http://x/audio", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(path)
	if n != 500 {
		t.Fatalf("expected 500 bytes, got %d", n)
	}
}
