package jsplugin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

// InvokeHTTP 不经 chi 路由,直接走 onHTTPRequest 调用插件的 HTTP handler。
//
// 用于服务端业务逻辑(如 SourceFetcher、SourceResolver)调用插件接口,避免:
//   - HTTP 回环开销
//   - JWT 鉴权链(本地内部调用本来就有信任)
//   - chi middleware 链(日志/CORS 等)
//
// path 必须以 "/" 开头(若不带前导斜杠会自动补);query 用 url.Values 序列化为 query string;
// body 直接透传(支持二进制,含非 UTF-8 字节时内部 base64 编码)。
//
// 签名故意做成元组而非 *InvokeResult,以匹配 source.PluginInvoker 接口,避免跨包指针类型。
//
// 错误语义:
//   - 插件未加载/EnsureLoaded 失败 → 返回非 nil err
//   - 调度器超时/插件抛错 → 返回非 nil err(原始错误)
//   - 插件返回非 200 状态码 → err 为 nil,statusCode 反映;调用方自行判断
func (m *Manager) InvokeHTTP(
	ctx context.Context,
	entryPath, method, path string,
	query interface{}, // 实际为 url.Values 或 nil;用 interface{} 让 source.PluginInvoker 不必导 net/url
	body []byte,
) (statusCode int, respHeaders map[string]string, respBody []byte, err error) {
	if m == nil {
		return 0, nil, nil, fmt.Errorf("jsplugin.Manager is nil")
	}
	if entryPath == "" {
		return 0, nil, nil, fmt.Errorf("entryPath is empty")
	}
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}

	if _, err := m.EnsureLoaded(ctx, entryPath); err != nil {
		return 0, nil, nil, fmt.Errorf("plugin %s not available: %w", entryPath, err)
	}

	reqData := &HTTPRequestData{
		Method:  strings.ToUpper(method),
		Path:    path,
		Headers: map[string]string{"Content-Type": "application/json"},
	}
	if v, ok := query.(url.Values); ok && v != nil {
		reqData.Query = v.Encode()
	}
	if len(body) > 0 {
		if utf8.Valid(body) {
			reqData.Body = string(body)
		} else {
			reqData.Body = base64.StdEncoding.EncodeToString(body)
			reqData.BodyEncoding = "base64"
		}
	}

	// 调度器 Call 超时：默认 0（→ scheduler 内部 30s）。但当调用方在 ctx 上显式设了更长的
	// deadline 时（如下载路径的 acquireAudio 用 5min 预算等待 music/url 解析排队，issue #265），
	// 把剩余时间作为 Call 超时传入，让调用方意图生效，而不是被硬编码的 30s 判死。
	// 无 deadline 的调用（普通播放解析、探测等）仍走 30s 默认，行为不变。
	callTimeout := time.Duration(0)
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > defaultCallTimeout {
			callTimeout = remaining
		}
	}
	resp, err := m.scheduler.Call(ctx, entryPath, "", MsgHTTPRequest, reqData, callTimeout)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("plugin %s call failed: %w", entryPath, err)
	}
	if resp == nil || resp.Data == nil {
		return 0, nil, nil, fmt.Errorf("plugin %s returned empty response", entryPath)
	}
	respData, ok := resp.Data.(*HTTPResponseData)
	if !ok {
		return 0, nil, nil, fmt.Errorf("plugin %s returned invalid response type", entryPath)
	}

	// StatusCode == 0 在 2026 修复后只可能出现在调用方没设 statusCode 字段(JS 侧 jsonResponse
	// 等 helper 都会显式带上),视作插件协议违例返 502,而不是悄悄升级成 200 + 空 body
	// 让上游 source.fetcher 在 json.Unmarshal 时报 "unexpected end of JSON input"。
	status := respData.StatusCode
	bodyStr := respData.Body
	if status == 0 {
		slog.Warn("jsplugin-http: plugin returned StatusCode=0, treating as 502",
			"entryPath", entryPath, "method", method, "path", path, "bodyLen", len(bodyStr))
		status = http.StatusBadGateway
		if bodyStr == "" {
			errBody, _ := json.Marshal(map[string]string{
				"error":  "plugin protocol error",
				"detail": "plugin returned StatusCode=0 (likely onHTTPRequest resolved to undefined)",
			})
			bodyStr = string(errBody)
		}
	}
	return status, respData.Headers, []byte(bodyStr), nil
}

// ListActive 返回当前 active 状态的所有插件元数据,供 SourceResolver fan-out 时枚举音乐源。
//
// TODO(capabilities): 当前所有插件都视为"音乐源插件",未来若需要区分(如有"工具型"插件),
// 在 PluginManifest 加 Capabilities 字段后,此处按 capability 过滤。
func (m *Manager) ListActive() []*JSPlugin {
	services := m.ListServices()
	plugins := make([]*JSPlugin, 0, len(services))
	for _, s := range services {
		if s == nil {
			continue
		}
		// 只返回成功加载/运行中的;停止或错误状态的不参与 fan-out
		status := s.Status()
		if status == ServiceStatusStopped {
			continue
		}
		plugins = append(plugins, s.plugin)
	}
	return plugins
}
