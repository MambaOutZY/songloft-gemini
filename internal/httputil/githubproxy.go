package httputil

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

// IsGitHubURL 判断 URL 是否为 GitHub 相关域名。
func IsGitHubURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	return host == "github.com" ||
		host == "raw.githubusercontent.com" ||
		host == "objects.githubusercontent.com" ||
		host == "api.github.com" ||
		strings.HasSuffix(host, ".github.io")
}

// ApplyGithubProxy 将 GitHub 加速代理前缀应用到 URL 上（拼接式，如
// https://ghproxy.com/ + 原始 URL）。代理为空或非 GitHub URL 时原样返回。
func ApplyGithubProxy(rawURL, proxyPrefix string) string {
	if proxyPrefix == "" || !IsGitHubURL(rawURL) {
		return rawURL
	}
	if proxyPrefix[len(proxyPrefix)-1] != '/' {
		proxyPrefix += "/"
	}
	return proxyPrefix + rawURL
}

// GithubGetOptions 是 GetWithGithubProxyFallback 的可选项。
type GithubGetOptions struct {
	// Header 附加到请求上的额外头（如 Authorization），可为 nil。
	Header http.Header
	// AttemptTimeout > 0 时每次尝试（代理/直连各一次）从 ctx 派生独立超时；
	// 0 则依赖 client.Timeout。注意超时覆盖**含 body 读取**的整个生命周期，
	// 只适合小响应体（如 version.json），大文件下载不要设置。
	AttemptTimeout time.Duration
	// ProxyDown 可选的「代理已失效」记忆，供一次业务操作内的多个请求共享，
	// 置位后后续请求直接直连。仅在传输层错误或代理返回 5xx 时置位
	// （4xx 不置位，避免单个资源 404 被误判为代理失效）。可为 nil。
	ProxyDown *atomic.Bool
}

// GetWithGithubProxyFallback 经 GitHub 加速代理 GET rawURL；代理请求失败
// （网络错误/超时/非 2xx）且确实套用了代理时，记 warn 日志后改用原始 URL
// 直连重试一次。直连结果原样返回（含非 2xx，状态码由调用方处理）。
// 调用方主动取消（ctx canceled）不触发降级。
func GetWithGithubProxyFallback(ctx context.Context, client *http.Client, rawURL, proxyPrefix string, opts GithubGetOptions) (*http.Response, error) {
	return getWithFallback(ctx, client, rawURL, ApplyGithubProxy(rawURL, proxyPrefix), proxyPrefix, opts)
}

func getWithFallback(ctx context.Context, client *http.Client, rawURL, proxiedURL, proxyPrefix string, opts GithubGetOptions) (*http.Response, error) {
	if proxiedURL == rawURL {
		return doGet(ctx, client, rawURL, opts)
	}
	if opts.ProxyDown != nil && opts.ProxyDown.Load() {
		return doGet(ctx, client, rawURL, opts)
	}

	resp, err := doGet(ctx, client, proxiedURL, opts)
	if err != nil {
		if ctx.Err() != nil {
			return nil, err
		}
		slog.Warn("github proxy request failed, falling back to direct",
			"failed_url", proxiedURL, "url", rawURL, "proxy", proxyPrefix, "error", err)
		if opts.ProxyDown != nil {
			opts.ProxyDown.Store(true)
		}
		resp, err = doGet(ctx, client, rawURL, opts)
		if err != nil {
			slog.Warn("direct request also failed after proxy fallback",
				"failed_url", rawURL, "error", err)
		}
		return resp, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		drainAndClose(resp.Body)
		slog.Warn("github proxy request failed, falling back to direct",
			"failed_url", proxiedURL, "url", rawURL, "proxy", proxyPrefix, "status", resp.StatusCode)
		// 5xx 多为代理自身故障（停服代理常快速返回网关错误页），置位避免后续
		// 请求继续双倍打代理；4xx 可能只是单个资源不存在，不据此判代理失效
		if resp.StatusCode >= 500 && opts.ProxyDown != nil {
			opts.ProxyDown.Store(true)
		}
		return doGet(ctx, client, rawURL, opts)
	}
	return resp, nil
}

func doGet(ctx context.Context, client *http.Client, url string, opts GithubGetOptions) (*http.Response, error) {
	var cancel context.CancelFunc
	if opts.AttemptTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.AttemptTimeout)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}
	for k, vs := range opts.Header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}
	if cancel != nil {
		// cancel 不能在返回前调用，否则调用方读 body 会被中断；
		// 挂到 body 的 Close 上联动释放。
		resp.Body = &cancelReadCloser{ReadCloser: resp.Body, cancel: cancel}
	}
	return resp, nil
}

type cancelReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelReadCloser) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 4096))
	_ = body.Close()
}
