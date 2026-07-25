package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"songloft/internal/models"
	"songloft/internal/services"
)

// VideoHLSHandler 处理视频 HLS 转码播放端点。
// Web 端浏览器不原生支持的视频格式（mpg/flv/wmv/rmvb/avi/mkv 等），
// 通过服务端实时转码为 HLS（H.264+AAC）供 hls.js 播放。
type VideoHLSHandler struct {
	songService  *services.SongService
	cacheService *services.CacheService
}

// NewVideoHLSHandler 构造 VideoHLSHandler。
func NewVideoHLSHandler(songService *services.SongService, cacheService *services.CacheService) *VideoHLSHandler {
	return &VideoHLSHandler{
		songService:  songService,
		cacheService: cacheService,
	}
}

// GetPlaylist 返回视频 HLS 转码的 master.m3u8 播放列表。
// @Summary 获取视频 HLS 播放列表
// @Description 对浏览器不原生支持的视频格式（mpg/flv/wmv/rmvb/avi/mkv 等）实时转码为 HLS（H.264+AAC），返回 master.m3u8 播放列表。多音轨文件会生成 HLS 多音频 rendition（hls.js 原生支持切换）。首次请求会启动转码（转完再播）；后续请求命中缓存。需要 ffmpeg。
// @Tags 歌曲管理
// @Produce application/vnd.apple.mpegurl
// @Param id path int true "歌曲 ID"
// @Success 200 {string} string "HLS 播放列表内容"
// @Failure 400 {object} map[string]string "无效的歌曲 ID"
// @Failure 404 {object} map[string]string "歌曲不存在"
// @Failure 503 {object} map[string]string "ffmpeg 不可用或转码失败"
// @Security BearerAuth
// @Router /songs/{id}/video-hls/playlist.m3u8 [get]
func (h *VideoHLSHandler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	song, srcPath := h.resolveSong(w, r)
	if song == nil {
		return
	}

	playlistPath, err := h.cacheService.VideoHLSTranscode(r.Context(), srcPath, song)
	if err != nil {
		respondError(w, http.StatusServiceUnavailable, "视频转码失败", err)
		return
	}

	h.serveM3U8(w, r, playlistPath)
}

// GetResource 处理 HLS 子资源请求（子 playlist 和 .ts 切片）。
// hls.js 会根据 master.m3u8 中的相对路径请求子播放列表和切片。
// 路由匹配 /songs/{id}/video-hls/* (catch-all)。
// @Summary 获取视频 HLS 子资源
// @Description 返回视频 HLS 转码生成的子播放列表或 .ts 切片文件。由 hls.js 根据 master playlist 自动请求。支持多音轨场景下的子目录结构（stream_0/playlist.m3u8, stream_0/0001.ts 等）。
// @Tags 歌曲管理
// @Param id path int true "歌曲 ID"
// @Param path path string true "子资源路径（如 stream_0/playlist.m3u8 或 stream_0/0001.ts）"
// @Success 200 {file} file "HLS 子资源"
// @Failure 400 {object} map[string]string "无效请求"
// @Failure 404 {object} map[string]string "资源不存在"
// @Security BearerAuth
// @Router /songs/{id}/video-hls/{path} [get]
func (h *VideoHLSHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	songID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || songID <= 0 {
		respondError(w, http.StatusBadRequest, "无效的歌曲 ID", err)
		return
	}

	// 获取通配符路径（chi 的 * 参数）
	subPath := chi.URLParam(r, "*")
	if subPath == "" {
		respondError(w, http.StatusBadRequest, "缺少资源路径", nil)
		return
	}

	// 安全校验：禁止路径穿越
	if strings.Contains(subPath, "..") {
		respondError(w, http.StatusBadRequest, "无效的资源路径", nil)
		return
	}

	dir := h.cacheService.VideoHLSDir(songID)
	fullPath := filepath.Join(dir, filepath.FromSlash(subPath))

	// 确保路径仍在 HLS 目录内（防穿越，追加分隔符避免前缀歧义如 /video_hls/12 vs /video_hls/123）
	absDir, _ := filepath.Abs(dir)
	absPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) && absPath != absDir {
		respondError(w, http.StatusBadRequest, "无效的资源路径", nil)
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// 根据扩展名设置 Content-Type
	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".m3u8":
		h.serveM3U8(w, r, fullPath)
	case ".ts":
		w.Header().Set("Content-Type", "video/mp2t")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		http.ServeFile(w, r, fullPath)
	default:
		http.ServeFile(w, r, fullPath)
	}
}

// serveM3U8 读取 .m3u8 文件并注入 access_token 到所有资源引用中。
func (h *VideoHLSHandler) serveM3U8(w http.ResponseWriter, r *http.Request, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "读取播放列表失败", err)
		return
	}

	// 从请求中获取 access_token，注入到所有资源引用（.ts 和 .m3u8）中
	token := r.URL.Query().Get("access_token")
	if token != "" {
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				// 检查 #EXT-X-MEDIA 中的 URI= 属性，也需要注入 token
				if strings.Contains(trimmed, "URI=\"") {
					lines[i] = injectTokenInURI(trimmed, token)
				}
				continue
			}
			// 非注释、非空行 = 资源引用（.ts 或 .m3u8）
			if strings.HasSuffix(trimmed, ".ts") || strings.HasSuffix(trimmed, ".m3u8") {
				lines[i] = trimmed + "?access_token=" + token
			}
		}
		data = []byte(strings.Join(lines, "\n"))
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(data)
}

// injectTokenInURI 在 HLS 标签的 URI="..." 属性值中追加 access_token。
// 例如：#EXT-X-MEDIA:TYPE=AUDIO,...,URI="stream_1/playlist.m3u8"
// 变为：#EXT-X-MEDIA:TYPE=AUDIO,...,URI="stream_1/playlist.m3u8?access_token=xxx"
func injectTokenInURI(line, token string) string {
	const uriPrefix = "URI=\""
	idx := strings.Index(line, uriPrefix)
	if idx < 0 {
		return line
	}
	start := idx + len(uriPrefix)
	end := strings.Index(line[start:], "\"")
	if end < 0 {
		return line
	}
	end += start
	uri := line[start:end]
	return line[:start] + uri + "?access_token=" + token + line[end:]
}

// resolveSong 解析歌曲 ID 并获取本地文件路径。
func (h *VideoHLSHandler) resolveSong(w http.ResponseWriter, r *http.Request) (*models.Song, string) {
	idStr := chi.URLParam(r, "id")
	songID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || songID <= 0 {
		respondError(w, http.StatusBadRequest, "无效的歌曲 ID", err)
		return nil, ""
	}

	song, err := h.songService.GetByID(r.Context(), songID)
	if err != nil || song == nil {
		respondError(w, http.StatusNotFound, "歌曲不存在", err)
		return nil, ""
	}

	// 仅本地歌曲支持视频 HLS 转码
	if song.Type != models.TypeLocal || song.FilePath == "" {
		respondError(w, http.StatusBadRequest, "仅本地视频歌曲支持 HLS 转码", nil)
		return nil, ""
	}

	return song, song.FilePath
}
