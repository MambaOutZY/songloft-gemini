package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"songloft/internal/models"
	"songloft/internal/services"

	"github.com/go-chi/chi/v5"
)

const (
	// maxPlaylistBytes m3u8 响应体上限。HLS 主播放列表 + 媒体播放列表都不会超过这个值；
	// 设上限防止恶意上游返回巨型文件耗尽内存。
	maxPlaylistBytes = 1 * 1024 * 1024

	// hlsContentType m3u8 标准 MIME。
	hlsContentType = "application/vnd.apple.mpegurl"
)

var (
	errNotM3U8      = errors.New("not a valid m3u8: missing #EXTM3U")
	errEmptyContent = errors.New("empty m3u8 content")
)

// hlsAttr 单个 HLS 属性键值对。保留原顺序与引号风格，便于"原样透传 + 仅改 URI"。
type hlsAttr struct {
	Key    string
	Value  string
	Quoted bool
}

// parseAttrLine 解析 attribute list 形式的 HLS 行。
//
// 输入: `#EXT-X-MEDIA:TYPE=AUDIO,URI="a.m3u8",NAME="en"`
// 返回 tag="EXT-X-MEDIA", attrs=[{TYPE,AUDIO,false},{URI,a.m3u8,true},{NAME,en,true}]
//
// 对 `#EXTINF:5.0,title` 这种非 attribute list 形式，attrs 返回 nil。
// 对 `#EXTM3U` 这种无冒号行，attrs 也返回 nil。
func parseAttrLine(line string) (tag string, attrs []hlsAttr) {
	if !strings.HasPrefix(line, "#") {
		return "", nil
	}
	before, rest, ok := strings.Cut(line, ":")
	if !ok {
		return strings.TrimPrefix(line, "#"), nil
	}
	tag = strings.TrimPrefix(before, "#")

	parsed := parseAttrList(rest)
	// 全部 attr 都没有 '='（如 #EXTINF:5.0,title）→ 非 attribute list，返回 nil
	hasEq := false
	for _, a := range parsed {
		if a.Key != "" && (a.Value != "" || a.Quoted) {
			hasEq = true
			break
		}
	}
	if !hasEq {
		return tag, nil
	}
	return tag, parsed
}

// parseAttrList 按"逗号在引号外"切分属性列表。
// 这是手写 HLS 解析最容易出错的点：CODECS="avc1.42c01e,mp4a.40.2" 引号内的逗号不能切。
func parseAttrList(s string) []hlsAttr {
	var out []hlsAttr
	n := len(s)
	i := 0
	for i < n {
		// 跳过前导空白
		for i < n && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= n {
			break
		}
		// 读 key 到 '=' 或 ','
		keyStart := i
		for i < n && s[i] != '=' && s[i] != ',' {
			i++
		}
		key := strings.TrimSpace(s[keyStart:i])
		if i < n && s[i] == ',' {
			if key != "" {
				out = append(out, hlsAttr{Key: key})
			}
			i++
			continue
		}
		if i >= n {
			if key != "" {
				out = append(out, hlsAttr{Key: key})
			}
			break
		}
		i++ // skip '='
		// 读 value
		quoted := false
		var val string
		if i < n && s[i] == '"' {
			quoted = true
			i++
			start := i
			for i < n && s[i] != '"' {
				i++
			}
			val = s[start:i]
			if i < n {
				i++ // skip closing quote
			}
		} else {
			start := i
			for i < n && s[i] != ',' {
				i++
			}
			val = strings.TrimSpace(s[start:i])
		}
		out = append(out, hlsAttr{Key: key, Value: val, Quoted: quoted})
		if i < n && s[i] == ',' {
			i++
		}
	}
	return out
}

// formatAttrLine 把 (tag, attrs) 重新拼回 HLS 行，保留原始属性顺序与引号风格。
func formatAttrLine(tag string, attrs []hlsAttr) string {
	var b strings.Builder
	b.WriteByte('#')
	b.WriteString(tag)
	if len(attrs) == 0 {
		return b.String()
	}
	b.WriteByte(':')
	for i, a := range attrs {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(a.Key)
		if a.Value == "" && !a.Quoted {
			continue
		}
		b.WriteByte('=')
		if a.Quoted {
			b.WriteByte('"')
			b.WriteString(a.Value)
			b.WriteByte('"')
		} else {
			b.WriteString(a.Value)
		}
	}
	return b.String()
}

// isSameOrigin 比较 scheme + hostname(小写) + port。
func isSameOrigin(a, b *url.URL) bool {
	if a == nil || b == nil {
		return false
	}
	return strings.EqualFold(a.Scheme, b.Scheme) &&
		strings.EqualFold(a.Hostname(), b.Hostname()) &&
		a.Port() == b.Port()
}

// looksLikePlaylist 判断 URL path 后缀是否为 .m3u8 / .m3u（忽略 query）。
// 与 music.go:isHLSURL 同义，独立实现避免循环 + 便于 hls_test 单测。
func looksLikePlaylist(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	ext := strings.ToLower(filepath.Ext(u.Path))
	return ext == ".m3u8" || ext == ".m3u"
}

// rewriteRef 解析 ref（可能是相对 URL）为绝对 URL，同源时调 rewrite 回调；
// 非同源 / 解析失败时原样返回 ref。
func rewriteRef(ref string, base *url.URL, rewrite func(absURL string, isPlaylist bool) string, isPlaylist bool) string {
	if ref == "" {
		return ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	abs := base.ResolveReference(refURL)
	if !isSameOrigin(abs, base) {
		return ref
	}
	return rewrite(abs.String(), isPlaylist)
}

// rewriteAttrTagURI 在 attribute list 标签行里替换 URI 类属性，其它属性保持原样。
func rewriteAttrTagURI(line string, base *url.URL, rewrite func(string, bool) string) string {
	tag, attrs := parseAttrLine(line)
	if attrs == nil {
		return line
	}
	changed := false
	for idx, a := range attrs {
		var isPlaylist bool
		var handle bool
		switch tag {
		case "EXT-X-KEY", "EXT-X-SESSION-KEY", "EXT-X-MAP",
			"EXT-X-PRELOAD-HINT", "EXT-X-PART", "EXT-X-SESSION-DATA":
			if a.Key == "URI" {
				handle, isPlaylist = true, false
			}
		case "EXT-X-MEDIA", "EXT-X-RENDITION-REPORT", "EXT-X-I-FRAME-STREAM-INF":
			if a.Key == "URI" {
				handle, isPlaylist = true, true
			}
		case "EXT-X-DATERANGE":
			if a.Key == "X-ASSET-URI" {
				refURL, err := url.Parse(a.Value)
				if err != nil {
					continue
				}
				abs := base.ResolveReference(refURL).String()
				handle, isPlaylist = true, looksLikePlaylist(abs)
			}
			// X-ASSET-LIST 暂不改写：JSON 子代理在 MVP 之外，原样透传 fail-open 到客户端直连
		}
		if !handle {
			continue
		}
		newVal := rewriteRef(a.Value, base, rewrite, isPlaylist)
		if newVal != a.Value {
			attrs[idx].Value = newVal
			changed = true
		}
	}
	if !changed {
		return line
	}
	return formatAttrLine(tag, attrs)
}

// rewriteM3U8 改写 m3u8 内容中所有 URI 指向本机代理。
//
// base: 上游 m3u8 的最终 URL（解析相对 URL + 判定同源用，注意若上游 redirect 需用 resp.Request.URL）
// rewrite: URI 改写回调；absURL 是绝对 URL，isPlaylist=true 表示该 URI 指向 m3u8
//
// 非同源 URL 保持原样（避免开放代理 + 兼容合法跨域 CMAF）。
// 未识别标签原样透传（HLS 草案每 6 个月迭代，未来标签不应导致播放失败）。
func rewriteM3U8(content []byte, base *url.URL, rewrite func(absURL string, isPlaylist bool) string) ([]byte, error) {
	// 跳过 UTF-8 BOM
	content = bytes.TrimPrefix(content, []byte{0xEF, 0xBB, 0xBF})
	if len(content) == 0 {
		return nil, errEmptyContent
	}

	lines := splitLines(content)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "#EXTM3U" {
		return nil, errNotM3U8
	}

	var out bytes.Buffer
	out.Grow(len(content) + 256)

	streamURLPending := false
	for _, raw := range lines {
		trimmed := strings.TrimSpace(raw)

		// STREAM-INF 紧跟的下一行是 playlist URL（HLS spec 强制紧邻）
		if streamURLPending {
			streamURLPending = false
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				out.WriteString(rewriteRef(trimmed, base, rewrite, true))
				out.WriteByte('\n')
				continue
			}
			// 不规范：STREAM-INF 后不是 URL 行。原样输出，下面继续处理
		}

		switch {
		case strings.HasPrefix(trimmed, "#EXT-X-STREAM-INF"):
			streamURLPending = true
			out.WriteString(raw)
			out.WriteByte('\n')

		case strings.HasPrefix(trimmed, "#EXT-X-I-FRAME-STREAM-INF:"),
			strings.HasPrefix(trimmed, "#EXT-X-MEDIA:"),
			strings.HasPrefix(trimmed, "#EXT-X-RENDITION-REPORT:"),
			strings.HasPrefix(trimmed, "#EXT-X-KEY:"),
			strings.HasPrefix(trimmed, "#EXT-X-SESSION-KEY:"),
			strings.HasPrefix(trimmed, "#EXT-X-MAP:"),
			strings.HasPrefix(trimmed, "#EXT-X-PRELOAD-HINT:"),
			strings.HasPrefix(trimmed, "#EXT-X-PART:"),
			strings.HasPrefix(trimmed, "#EXT-X-SESSION-DATA:"),
			strings.HasPrefix(trimmed, "#EXT-X-DATERANGE:"):
			out.WriteString(rewriteAttrTagURI(raw, base, rewrite))
			out.WriteByte('\n')

		case strings.HasPrefix(trimmed, "#") || trimmed == "":
			out.WriteString(raw)
			out.WriteByte('\n')

		default:
			// 裸 URL 行：媒体切片
			out.WriteString(rewriteRef(trimmed, base, rewrite, false))
			out.WriteByte('\n')
		}
	}

	return out.Bytes(), nil
}

// ============================================================
// HLSHandler: 反向代理 m3u8 + 切片
// ============================================================

// HLSHandler 处理 /api/v1/songs/{id}/hls/{playlist,segment} 端点。
// 当 hls_proxy_mode=proxy 时由 serveRadio 直调 ServeProxy；
// player 拉到的 m3u8 内 URL 全部指向本机端点，后续切片/key/init 由 player 自行回访。
type HLSHandler struct {
	songService *services.SongService
	client      *http.Client
	allowHost   func(host string) bool // 默认 services.IsHostnameAllowed；测试可替换
}

// NewHLSHandler 构造 HLSHandler。
// client 无 Timeout：直播切片可能持续数十秒至数分钟，timeout 会被客户端断连+r.Context() 取消接管。
func NewHLSHandler(songService *services.SongService) *HLSHandler {
	return &HLSHandler{
		songService: songService,
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		allowHost: services.IsHostnameAllowed,
	}
}

// ServeProxy 是 serveRadio 在 hls_proxy_mode=proxy 时的入口：直接代理 song.URL 作为顶层 m3u8。
// 后续子 m3u8 / 切片由 player 通过 HandlePlaylist / HandleSegment 端点回访。
func (h *HLSHandler) ServeProxy(w http.ResponseWriter, r *http.Request, song *models.Song) {
	h.servePlaylist(w, r, song, song.URL)
}

// HandlePlaylist 处理 GET/HEAD /api/v1/songs/{id}/hls/playlist?u=<base64url>
func (h *HLSHandler) HandlePlaylist(w http.ResponseWriter, r *http.Request) {
	song, upstreamURL, ok := h.resolveEndpoint(w, r)
	if !ok {
		return
	}
	h.servePlaylist(w, r, song, upstreamURL)
}

// HandleSegment 处理 GET/HEAD /api/v1/songs/{id}/hls/segment?u=<base64url>
func (h *HLSHandler) HandleSegment(w http.ResponseWriter, r *http.Request) {
	song, upstreamURL, ok := h.resolveEndpoint(w, r)
	if !ok {
		return
	}
	h.serveSegment(w, r, song, upstreamURL)
}

// resolveEndpoint 共用：取 song id → 加载 song → base64 解码 u。
func (h *HLSHandler) resolveEndpoint(w http.ResponseWriter, r *http.Request) (*models.Song, string, bool) {
	idStr := chi.URLParam(r, "id")
	songID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || songID <= 0 {
		respondError(w, http.StatusBadRequest, "无效的 song_id", err)
		return nil, "", false
	}

	encoded := r.URL.Query().Get("u")
	if encoded == "" {
		respondError(w, http.StatusBadRequest, "缺少 u 参数", nil)
		return nil, "", false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		respondError(w, http.StatusBadRequest, "u 参数解码失败", err)
		return nil, "", false
	}

	song, err := h.songService.GetByID(r.Context(), songID)
	if err != nil || song == nil {
		http.NotFound(w, r)
		return nil, "", false
	}
	return song, string(decoded), true
}

// servePlaylist 拉上游 m3u8 → 同源校验 → 改写 URI → 回写。
// upstreamURL 来自 ServeProxy(=song.URL) 或 HandlePlaylist 的 ?u=...
func (h *HLSHandler) servePlaylist(w http.ResponseWriter, r *http.Request, song *models.Song, upstreamURL string) {
	songOrigin, upURL, err := h.checkOrigin(song, upstreamURL)
	if err != nil {
		respondError(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	req, err := buildUpstreamRequest(r.Context(), upstreamURL, songOrigin, r.Header.Get("Range"))
	if err != nil {
		respondError(w, http.StatusBadGateway, "构建上游请求失败", err)
		return
	}

	resp, err := h.client.Do(req)
	if err != nil {
		slog.Warn("hls playlist upstream fetch failed", "url", upstreamURL, "error", err)
		http.Error(w, "playlist fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 上游 4xx/5xx：透传 status + body 给 player
	if resp.StatusCode >= 400 {
		slog.Debug("hls playlist upstream error", "url", upstreamURL, "status", resp.StatusCode)
		w.Header().Set("Cache-Control", "no-store")
		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPlaylistBytes+1))
	if err != nil {
		slog.Warn("hls playlist read failed", "url", upstreamURL, "error", err)
		http.Error(w, "playlist read failed", http.StatusBadGateway)
		return
	}
	if len(body) > maxPlaylistBytes {
		slog.Warn("hls playlist too large", "url", upstreamURL, "limit", maxPlaylistBytes)
		http.Error(w, "playlist too large", http.StatusBadGateway)
		return
	}

	// base = resp.Request.URL 处理上游 redirect 后的最终位置；Go 内置 redirect 会更新 Request.URL
	_ = upURL
	base := resp.Request.URL

	rewritten, err := rewriteM3U8(body, base, h.makeRewriter(song.ID))
	if err != nil {
		slog.Warn("hls playlist rewrite failed", "url", upstreamURL, "error", err)
		http.Error(w, "playlist rewrite failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", hlsContentType)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	// HEAD 时 net/http 自动 discard body
	w.Write(rewritten)
}

// serveSegment 透传上游切片字节。无 Timeout client，client 断连由 r.Context() 取消上游。
func (h *HLSHandler) serveSegment(w http.ResponseWriter, r *http.Request, song *models.Song, upstreamURL string) {
	songOrigin, _, err := h.checkOrigin(song, upstreamURL)
	if err != nil {
		respondError(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, nil)
	if err != nil {
		http.Error(w, "segment request build failed", http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "Songloft/1.0")
	req.Header.Set("Referer", songOrigin.Scheme+"://"+songOrigin.Host+"/")
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		slog.Debug("hls segment upstream fetch failed", "url", upstreamURL, "error", err)
		http.Error(w, "segment fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 透传字节级 header；强制 no-store（直播流不能缓存）
	for _, hh := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges"} {
		if v := resp.Header.Get(hh); v != "" {
			w.Header().Set(hh, v)
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) // HEAD 时 stdlib 自动 discard
}

// checkOrigin 同源校验 + IsHostnameAllowed 兜底。
// 返回 song origin 和 upstream 解析结果，便于复用。
func (h *HLSHandler) checkOrigin(song *models.Song, upstreamURL string) (songOrigin, up *url.URL, err error) {
	if song.URL == "" {
		return nil, nil, fmt.Errorf("song.URL 为空")
	}
	songOrigin, err = url.Parse(song.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("song.URL 解析失败")
	}
	if songOrigin.Scheme != "http" && songOrigin.Scheme != "https" {
		return nil, nil, fmt.Errorf("不支持的 song.URL scheme")
	}

	up, err = url.Parse(upstreamURL)
	if err != nil {
		return nil, nil, fmt.Errorf("upstream URL 解析失败")
	}
	if up.Scheme != "http" && up.Scheme != "https" {
		return nil, nil, fmt.Errorf("不支持的 upstream scheme")
	}
	if !isSameOrigin(up, songOrigin) {
		return nil, nil, fmt.Errorf("跨源 URL 不允许代理")
	}
	if h.allowHost != nil && !h.allowHost(up.Hostname()) {
		return nil, nil, fmt.Errorf("目标主机不允许")
	}
	return songOrigin, up, nil
}

// makeRewriter 返回闭包：把绝对上游 URL 改写为本机相对路径。
// 用相对路径（"playlist?u=..." / "segment?u=..."）规避 BASE_PATH 拼接问题——
// player 用当前请求 URL 作为 base 解析相对 URL，正确落到 /api/v1/songs/{id}/hls/{...}。
func (h *HLSHandler) makeRewriter(songID int64) func(absURL string, isPlaylist bool) string {
	_ = songID // 当前实现用相对路径，songID 通过 URL 路径已经携带；保留参数以备日后切绝对路径
	return func(absURL string, isPlaylist bool) string {
		encoded := base64.RawURLEncoding.EncodeToString([]byte(absURL))
		target := "segment"
		if isPlaylist {
			target = "playlist"
		}
		return target + "?u=" + encoded
	}
}

// buildUpstreamRequest 构造带 UA + Referer 的 GET 请求；Range 透传（playlist 通常没用，segment 关键）。
// Referer 用 song origin 的根（scheme + host）—— 多数防盗链只校验 host，不要求具体页面 URL。
func buildUpstreamRequest(ctx context.Context, upstreamURL string, songOrigin *url.URL, rangeHeader string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Songloft/1.0")
	req.Header.Set("Referer", songOrigin.Scheme+"://"+songOrigin.Host+"/")
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	return req, nil
}

// splitLines 按 \r\n / \n / \r 三种行尾切分内容。返回的行不含行尾。
// 末尾若有空行（如以 \n 结尾）会被丢弃，重组时统一用 \n。
func splitLines(content []byte) []string {
	var lines []string
	i := 0
	n := len(content)
	for i < n {
		start := i
		for i < n && content[i] != '\n' && content[i] != '\r' {
			i++
		}
		lines = append(lines, string(content[start:i]))
		if i < n && content[i] == '\r' {
			i++
			if i < n && content[i] == '\n' {
				i++ // \r\n 算一个行尾
			}
		} else if i < n && content[i] == '\n' {
			i++
		}
	}
	return lines
}
