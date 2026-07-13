package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"songloft/internal/httputil"
	"songloft/internal/jsplugin"
)

const pluginRegistriesConfigKey = "plugin_registries"

// pluginAutoUpdateConfigKey 自动更新开关配置键。
// 与 jsplugin 包内后台 ticker 读取的键保持一致。
const pluginAutoUpdateConfigKey = "plugin_auto_update"

var defaultPluginRegistries = pluginRegistriesSetting{
	Registries: []jsplugin.RegistryConfig{
		{
			URL:     "https://raw.githubusercontent.com/songloft-org/songloft-plugin-registry/main/registry.json",
			Name:    "Songloft 官方插件",
			Enabled: true,
		},
	},
}

// --- Settings: GET/PUT /api/v1/settings/plugin-registries ---

// pluginRegistriesSetting 订阅源列表配置。
type pluginRegistriesSetting struct {
	Registries []jsplugin.RegistryConfig `json:"registries"`
}

// GetRegistriesSetting 获取插件订阅源列表
// @Summary 获取插件订阅源列表
// @Description 获取用户保存的所有插件注册表订阅源 URL。未配置时返回空列表。
// @Tags JS插件管理
// @Produce json
// @Success 200 {object} pluginRegistriesSetting "订阅源列表"
// @Security BearerAuth
// @Router /settings/plugin-registries [get]
func (h *JSPluginHandler) GetRegistriesSetting(w http.ResponseWriter, r *http.Request) {
	var cfg pluginRegistriesSetting
	if err := h.configService.GetJSON(pluginRegistriesConfigKey, &cfg); err != nil {
		respondJSON(w, http.StatusOK, defaultPluginRegistries)
		return
	}
	if cfg.Registries == nil {
		cfg.Registries = []jsplugin.RegistryConfig{}
	}
	respondJSON(w, http.StatusOK, cfg)
}

// UpdateRegistriesSetting 保存插件订阅源列表
// @Summary 保存插件订阅源列表
// @Description 保存用户配置的插件注册表订阅源 URL 列表。每个源包含 URL、名称和是否启用。
// @Tags JS插件管理
// @Accept json
// @Produce json
// @Param request body pluginRegistriesSetting true "订阅源列表"
// @Success 200 {object} pluginRegistriesSetting "保存后的订阅源列表"
// @Failure 400 {object} models.ErrorResponse "请求格式错误"
// @Failure 500 {object} models.ErrorResponse "保存配置失败"
// @Security BearerAuth
// @Router /settings/plugin-registries [put]
func (h *JSPluginHandler) UpdateRegistriesSetting(w http.ResponseWriter, r *http.Request) {
	var req pluginRegistriesSetting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if req.Registries == nil {
		req.Registries = []jsplugin.RegistryConfig{}
	}
	if err := h.configService.SetJSON(pluginRegistriesConfigKey, req); err != nil {
		respondError(w, http.StatusInternalServerError, "保存配置失败", err)
		return
	}
	respondJSON(w, http.StatusOK, req)
}

// --- Settings: GET/PUT /api/v1/settings/plugin-auto-update ---

// pluginAutoUpdateSetting 插件自动更新开关配置。
type pluginAutoUpdateSetting struct {
	Enabled bool `json:"enabled"`
}

// GetPluginAutoUpdateSetting 获取插件自动更新开关
// @Summary 获取插件自动更新开关
// @Description 获取“后台自动更新已安装插件”开关的当前状态。开启后，服务会在启动后延迟数分钟检查一次、之后每 6 小时定时检查所有具有远程更新源的插件并自动更新。默认关闭。
// @Tags 设置
// @Produce json
// @Success 200 {object} pluginAutoUpdateSetting "返回 enabled 字段表示开关状态"
// @Security BearerAuth
// @Router /settings/plugin-auto-update [get]
func (h *JSPluginHandler) GetPluginAutoUpdateSetting(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, pluginAutoUpdateSetting{
		Enabled: h.configService.GetBool(pluginAutoUpdateConfigKey, false),
	})
}

// UpdatePluginAutoUpdateSetting 保存插件自动更新开关
// @Summary 保存插件自动更新开关
// @Description 开启/关闭插件后台自动更新。开启后后台 ticker 会定时对有更新源的插件执行“检查更新 + 下载安装 + 热重载”。开关即时生效，无需重启。
// @Tags 设置
// @Accept json
// @Produce json
// @Param request body pluginAutoUpdateSetting true "开关请求"
// @Success 200 {object} pluginAutoUpdateSetting "返回 enabled 字段表示更新后的开关状态"
// @Failure 400 {object} models.ErrorResponse "请求格式错误"
// @Failure 500 {object} models.ErrorResponse "保存配置失败"
// @Security BearerAuth
// @Router /settings/plugin-auto-update [put]
func (h *JSPluginHandler) UpdatePluginAutoUpdateSetting(w http.ResponseWriter, r *http.Request) {
	var req pluginAutoUpdateSetting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	value := "false"
	if req.Enabled {
		value = "true"
	}
	if err := h.configService.Set(pluginAutoUpdateConfigKey, value); err != nil {
		respondError(w, http.StatusInternalServerError, "保存配置失败", err)
		return
	}
	respondJSON(w, http.StatusOK, req)
}

// --- Registry: POST /api/v1/jsplugins/registry/refresh ---

// registryRefreshRequest 刷新注册表请求。
type registryRefreshRequest struct {
	RegistryURL string `json:"registry_url"`
	AllSources  bool   `json:"all_sources"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
	Search      string `json:"search"`
	GithubProxy string `json:"github_proxy"`
	Token       string `json:"token,omitempty"`
}

// registryPluginEntry 注册表插件条目（含安装状态）。
type registryPluginEntry struct {
	Name             string `json:"name"`
	EntryPath        string `json:"entry_path"`
	Version          string `json:"version"`
	Description      string `json:"description,omitempty"`
	Author           string `json:"author,omitempty"`
	Homepage         string `json:"homepage,omitempty"`
	Icon             string `json:"icon,omitempty"`
	DownloadURL      string `json:"download_url"`
	Installed        bool   `json:"installed"`
	InstalledVersion string `json:"installed_version,omitempty"`
	HasUpdate        bool   `json:"has_update"`
	// SourceURL 该插件所属订阅源 URL（仅「全部」聚合模式返回），
	// 安装时回传给后端以按源解析私有源 token。
	SourceURL string `json:"source_url,omitempty"`
}

// registryRefreshResponse 刷新注册表响应。
type registryRefreshResponse struct {
	Plugins  []registryPluginEntry `json:"plugins"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Warnings []string              `json:"warnings,omitempty"`
}

// handleRegistryRefresh 拉取订阅源的插件列表
// @Summary 刷新插件注册表
// @Description 拉取订阅源（含递归 includes），去重合并后返回分页的可用插件列表。每个插件标注是否已安装及是否有更新。默认拉取单个 registry_url，可选传入 token 字段访问需要认证的私有源（如 GitHub 私有仓库 PAT）。当 all_sources=true 时忽略 registry_url/token，改为聚合已保存的所有启用订阅源（各源用自身存储的 token），跨源按 entry_path 去重、高版本优先。
// @Tags JS插件管理
// @Accept json
// @Produce json
// @Param request body registryRefreshRequest true "刷新请求"
// @Success 200 {object} registryRefreshResponse "插件列表"
// @Failure 400 {object} models.ErrorResponse "请求格式错误"
// @Failure 500 {object} models.ErrorResponse "拉取注册表失败"
// @Security BearerAuth
// @Router /jsplugins/registry/refresh [post]
func (h *JSPluginHandler) handleRegistryRefresh(w http.ResponseWriter, r *http.Request) {
	var req registryRefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if !req.AllSources && req.RegistryURL == "" {
		respondError(w, http.StatusBadRequest, "registry_url 不能为空", nil)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	svc := jsplugin.NewRegistryService()
	var (
		entries  []jsplugin.RegistryEntry
		warnings []string
	)
	if req.AllSources {
		// 聚合所有启用源：读配置 → 过滤 enabled → 逐源拉取合并去重
		var cfg pluginRegistriesSetting
		if err := h.configService.GetJSON(pluginRegistriesConfigKey, &cfg); err != nil {
			cfg = defaultPluginRegistries
		}
		enabled := make([]jsplugin.RegistryConfig, 0, len(cfg.Registries))
		for _, src := range cfg.Registries {
			if src.Enabled {
				enabled = append(enabled, src)
			}
		}
		entries, warnings = svc.FetchAndMergeMulti(r.Context(), enabled, req.GithubProxy)
	} else {
		var err error
		entries, warnings, err = svc.FetchAndMerge(r.Context(), req.RegistryURL, req.GithubProxy, req.Token)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "拉取注册表失败", err)
			return
		}
	}

	// 获取已安装插件，构建 entryPath -> version 映射
	installedMap := h.buildInstalledMap(r.Context())

	// 搜索过滤
	search := strings.ToLower(strings.TrimSpace(req.Search))
	var filtered []registryPluginEntry
	for _, entry := range entries {
		if search != "" {
			if !strings.Contains(strings.ToLower(entry.Name), search) &&
				!strings.Contains(strings.ToLower(entry.Description), search) &&
				!strings.Contains(strings.ToLower(entry.Author), search) &&
				!strings.Contains(strings.ToLower(entry.EntryPath), search) {
				continue
			}
		}
		p := registryPluginEntry{
			Name:        entry.Name,
			EntryPath:   entry.EntryPath,
			Version:     entry.Version,
			Description: entry.Description,
			Author:      entry.Author,
			Homepage:    entry.Homepage,
			Icon:        entry.Icon,
			DownloadURL: entry.DownloadURL,
			SourceURL:   entry.SourceURL,
		}
		if installedVer, ok := installedMap[entry.EntryPath]; ok {
			p.Installed = true
			p.InstalledVersion = installedVer
			p.HasUpdate = entry.Version != installedVer
		}
		filtered = append(filtered, p)
	}

	total := len(filtered)
	start := (req.Page - 1) * req.PageSize
	if start >= total {
		filtered = nil
	} else {
		end := min(start+req.PageSize, total)
		filtered = filtered[start:end]
	}
	if filtered == nil {
		filtered = []registryPluginEntry{}
	}

	respondJSON(w, http.StatusOK, registryRefreshResponse{
		Plugins:  filtered,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Warnings: warnings,
	})
}

func (h *JSPluginHandler) buildInstalledMap(ctx context.Context) map[string]string {
	installed := make(map[string]string)
	plugins, err := h.repo.GetAll(ctx)
	if err != nil {
		slog.Warn("failed to load installed plugins for registry comparison", "error", err)
		return installed
	}
	for _, p := range plugins {
		installed[p.EntryPath] = p.Version
	}
	return installed
}

// --- Registry: POST /api/v1/jsplugins/registry/install ---

// registryInstallRequest 从注册表安装插件请求。
type registryInstallRequest struct {
	DownloadURL string `json:"download_url"`
	GithubProxy string `json:"github_proxy"`
	Token       string `json:"token,omitempty"`
	// SourceURL 插件所属订阅源 URL。「全部」聚合模式安装时回传：
	// 当未显式提供 token 时，后端据此从 plugin_registries 配置解析该源的 token。
	SourceURL string `json:"source_url,omitempty"`
}

// resolveSourceToken 根据订阅源 URL 从配置中查出其存储的 token。
// 找不到匹配源时返回空字符串。
func (h *JSPluginHandler) resolveSourceToken(sourceURL string) string {
	if sourceURL == "" {
		return ""
	}
	var cfg pluginRegistriesSetting
	if err := h.configService.GetJSON(pluginRegistriesConfigKey, &cfg); err != nil {
		cfg = defaultPluginRegistries
	}
	for _, src := range cfg.Registries {
		if src.URL == sourceURL {
			return src.Token
		}
	}
	return ""
}

// handleRegistryInstall 从注册表 download_url 安装插件
// @Summary 从注册表安装插件
// @Description 从注册表中的 download_url 下载 ZIP 并安装插件。如果 entry_path 已存在则自动走更新路径。支持 GitHub 代理。可选传入 token 字段用于从需要认证的私有源下载；若未提供 token 但提供了 source_url（「全部」聚合模式），后端会自动从 plugin_registries 配置解析该源存储的 token。
// @Tags JS插件管理
// @Accept json
// @Produce json
// @Param request body registryInstallRequest true "安装请求"
// @Success 200 {object} jsPluginUploadResponse "安装结果（更新已有插件）"
// @Success 201 {object} jsPluginUploadResponse "安装结果（新插件）"
// @Failure 400 {object} models.ErrorResponse "请求格式错误"
// @Failure 500 {object} models.ErrorResponse "下载或安装失败"
// @Security BearerAuth
// @Router /jsplugins/registry/install [post]
func (h *JSPluginHandler) handleRegistryInstall(w http.ResponseWriter, r *http.Request) {
	var req registryInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if req.DownloadURL == "" {
		respondError(w, http.StatusBadRequest, "download_url 不能为空", nil)
		return
	}

	// 「全部」聚合模式：未显式给 token 时，按来源源 URL 从配置解析 token
	if req.Token == "" && req.SourceURL != "" {
		req.Token = h.resolveSourceToken(req.SourceURL)
	}

	var zipData []byte

	// GitHub browser-style release URLs don't accept Bearer tokens for private
	// repos. When a token is provided and the URL matches, use the GitHub API.
	if req.Token != "" {
		if owner, repo, tag, filename, ok := parseGitHubReleaseURL(req.DownloadURL); ok {
			data, err := downloadGitHubReleaseAsset(r.Context(), owner, repo, tag, filename, req.Token)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "下载插件失败", err)
				return
			}
			zipData = data
		}
	}

	if zipData == nil {
		downloadURL := req.DownloadURL
		if req.GithubProxy != "" {
			proxyPrefix := req.GithubProxy
			if proxyPrefix[len(proxyPrefix)-1] != '/' {
				proxyPrefix += "/"
			}
			downloadURL = proxyPrefix + downloadURL
		}
		data, err := downloadZIP(r.Context(), downloadURL, req.Token)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "下载插件失败", err)
			return
		}
		zipData = data
	}

	plugin, wasUpdate, err := h.packageMgr.InstallFromUpload(zipData)
	if err != nil {
		respondJSON(w, http.StatusOK, jsPluginUploadResponse{
			Total:   1,
			Success: 0,
			Failed:  1,
			Results: []jsPluginUploadResult{{
				FileName: req.DownloadURL,
				Error:    err.Error(),
				Success:  false,
			}},
			Message: "安装插件失败",
		})
		return
	}

	if h.manager != nil {
		if wasUpdate && plugin.Status == jsplugin.JSPluginStatusActive {
			if reloadErr := h.manager.ReloadPlugin(r.Context(), plugin.EntryPath); reloadErr != nil {
				slog.Warn("reload plugin after registry install failed", "entryPath", plugin.EntryPath, "error", reloadErr)
			}
		} else if !wasUpdate {
			if enableErr := h.manager.EnablePlugin(r.Context(), plugin.ID); enableErr != nil {
				slog.Warn("auto-enable plugin after registry install failed", "entryPath", plugin.EntryPath, "error", enableErr)
			} else {
				plugin.Status = jsplugin.JSPluginStatusActive
			}
		}
	}

	var (
		message string
		status  int
	)
	if wasUpdate {
		message = fmt.Sprintf("插件已更新到 v%s", plugin.Version)
		status = http.StatusOK
	} else {
		message = fmt.Sprintf("插件 %s 安装成功", plugin.EntryPath)
		status = http.StatusCreated
	}

	respondJSON(w, status, jsPluginUploadResponse{
		Total:   1,
		Success: 1,
		Failed:  0,
		Results: []jsPluginUploadResult{{
			FileName: req.DownloadURL,
			Plugin:   plugin,
			Success:  true,
		}},
		Message: message,
	})
}

func downloadZIP(ctx context.Context, url string, token string) ([]byte, error) {
	client := httputil.NewClient(60 * time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d from %s", resp.StatusCode, url)
	}

	const maxZIPSize = 50 << 20 // 50 MB
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxZIPSize+1))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if len(data) > maxZIPSize {
		return nil, fmt.Errorf("zip file exceeds %d bytes", maxZIPSize)
	}
	return data, nil
}

// parseGitHubReleaseURL extracts owner, repo, tag, and filename from a GitHub
// browser-style release download URL. Returns ok=false for non-matching URLs.
func parseGitHubReleaseURL(rawURL string) (owner, repo, tag, filename string, ok bool) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host != "github.com" {
		return
	}
	// /owner/repo/releases/download/tag/filename
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) == 6 && parts[2] == "releases" && parts[3] == "download" {
		return parts[0], parts[1], parts[4], parts[5], true
	}
	return
}

// downloadGitHubReleaseAsset downloads a release asset from a private GitHub
// repo via the GitHub API. Browser-style release URLs (github.com/.../releases/
// download/...) don't accept Bearer tokens for private repos—only the API does.
func downloadGitHubReleaseAsset(ctx context.Context, owner, repo, tag, filename, token string) ([]byte, error) {
	client := httputil.NewClient(60 * time.Second)

	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create release request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get release by tag: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get release %s/%s@%s: http status %d", owner, repo, tag, resp.StatusCode)
	}

	var release struct {
		Assets []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parse release: %w", err)
	}

	var assetID int64
	for _, a := range release.Assets {
		if a.Name == filename {
			assetID = a.ID
			break
		}
	}
	if assetID == 0 {
		return nil, fmt.Errorf("asset %q not found in release %s/%s@%s", filename, owner, repo, tag)
	}

	assetURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/assets/%d", owner, repo, assetID)
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create asset request: %w", err)
	}
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Accept", "application/octet-stream")

	resp2, err := client.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("download asset: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download asset %s/%s@%s/%s: http status %d", owner, repo, tag, filename, resp2.StatusCode)
	}

	const maxZIPSize = 50 << 20
	data, err := io.ReadAll(io.LimitReader(resp2.Body, maxZIPSize+1))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if len(data) > maxZIPSize {
		return nil, fmt.Errorf("zip file exceeds %d bytes", maxZIPSize)
	}
	return data, nil
}

// --- Settings: GET/PUT /api/v1/settings/http-proxy ---

const httpProxyConfigKey = "http_proxy"

// httpProxySetting HTTP 代理配置。
type httpProxySetting struct {
	Proxy string `json:"proxy"`
}

// GetHttpProxySetting 获取 HTTP 代理配置
// @Summary 获取 HTTP 代理配置
// @Description 获取全局 HTTP 代理地址。所有后端外发请求（插件下载、注册表拉取、升级检查等）会通过此代理转发。未配置时返回空字符串（直连）。
// @Tags 设置
// @Produce json
// @Success 200 {object} httpProxySetting "代理配置"
// @Security BearerAuth
// @Router /settings/http-proxy [get]
func (h *JSPluginHandler) GetHttpProxySetting(w http.ResponseWriter, r *http.Request) {
	var cfg httpProxySetting
	if err := h.configService.GetJSON(httpProxyConfigKey, &cfg); err != nil {
		respondJSON(w, http.StatusOK, httpProxySetting{Proxy: ""})
		return
	}
	respondJSON(w, http.StatusOK, cfg)
}

// UpdateHttpProxySetting 保存 HTTP 代理配置
// @Summary 保存 HTTP 代理配置
// @Description 设置全局 HTTP 代理地址（如 http://192.168.1.1:7890）。设为空字符串则关闭代理。保存后即时生效，无需重启。
// @Tags 设置
// @Accept json
// @Produce json
// @Param request body httpProxySetting true "代理配置"
// @Success 200 {object} httpProxySetting "保存后的代理配置"
// @Failure 400 {object} models.ErrorResponse "请求格式错误或代理地址无效"
// @Failure 500 {object} models.ErrorResponse "保存配置失败"
// @Security BearerAuth
// @Router /settings/http-proxy [put]
func (h *JSPluginHandler) UpdateHttpProxySetting(w http.ResponseWriter, r *http.Request) {
	var req httpProxySetting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if err := httputil.SetGlobalProxy(req.Proxy); err != nil {
		respondError(w, http.StatusBadRequest, "代理地址无效", err)
		return
	}
	if err := h.configService.SetJSON(httpProxyConfigKey, req); err != nil {
		respondError(w, http.StatusInternalServerError, "保存配置失败", err)
		return
	}
	slog.Info("HTTP 代理已更新", "proxy", req.Proxy)
	respondJSON(w, http.StatusOK, req)
}

// --- Settings: GET/PUT /api/v1/settings/plugin-keep-alive ---

const pluginKeepAliveConfigKey = "plugin_keep_alive"

// pluginKeepAliveSetting 插件常驻白名单配置。
type pluginKeepAliveSetting struct {
	Plugins []string `json:"plugins"`
}

// GetPluginKeepAliveSetting 获取插件常驻白名单
// @Summary 获取插件常驻白名单
// @Description 获取不会被自动休眠的插件 entryPath 列表。白名单中的插件即使空闲超过 10 分钟也不会被卸载。未配置时返回空列表。
// @Tags 设置
// @Produce json
// @Success 200 {object} pluginKeepAliveSetting "常驻白名单"
// @Security BearerAuth
// @Router /settings/plugin-keep-alive [get]
func (h *JSPluginHandler) GetPluginKeepAliveSetting(w http.ResponseWriter, r *http.Request) {
	var cfg pluginKeepAliveSetting
	if err := h.configService.GetJSON(pluginKeepAliveConfigKey, &cfg); err != nil {
		respondJSON(w, http.StatusOK, pluginKeepAliveSetting{Plugins: []string{}})
		return
	}
	if cfg.Plugins == nil {
		cfg.Plugins = []string{}
	}
	respondJSON(w, http.StatusOK, cfg)
}

// UpdatePluginKeepAliveSetting 保存插件常驻白名单
// @Summary 保存插件常驻白名单
// @Description 设置不会被自动休眠的插件 entryPath 列表。保存后即时生效，白名单中的插件将跳过空闲检查，始终保持运行。
// @Tags 设置
// @Accept json
// @Produce json
// @Param request body pluginKeepAliveSetting true "常驻白名单"
// @Success 200 {object} pluginKeepAliveSetting "保存后的白名单"
// @Failure 400 {object} models.ErrorResponse "请求格式错误"
// @Failure 500 {object} models.ErrorResponse "保存配置失败"
// @Security BearerAuth
// @Router /settings/plugin-keep-alive [put]
func (h *JSPluginHandler) UpdatePluginKeepAliveSetting(w http.ResponseWriter, r *http.Request) {
	var req pluginKeepAliveSetting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if req.Plugins == nil {
		req.Plugins = []string{}
	}
	if err := h.configService.SetJSON(pluginKeepAliveConfigKey, req); err != nil {
		respondError(w, http.StatusInternalServerError, "保存配置失败", err)
		return
	}
	respondJSON(w, http.StatusOK, req)
}
