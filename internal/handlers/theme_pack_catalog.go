package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"songloft/internal/httputil"
	"songloft/internal/models"
)

const (
	themeCatalogURLKey     = "theme_catalog_url"
	defaultThemeCatalogURL = "https://raw.githubusercontent.com/songloft-org/songloft-themes/main/index.json"
	catalogCacheTTL        = 5 * time.Minute
)

// catalogCache 简单的内存缓存
type catalogCache struct {
	mu        sync.RWMutex
	data      *models.ThemeCatalogIndex
	fetchedAt time.Time
	url       string
}

func (c *catalogCache) get(url string) *models.ThemeCatalogIndex {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.url == url && c.data != nil && time.Since(c.fetchedAt) < catalogCacheTTL {
		return c.data
	}
	return nil
}

func (c *catalogCache) set(url string, data *models.ThemeCatalogIndex) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.url = url
	c.data = data
	c.fetchedAt = time.Now()
}

var themeCache catalogCache

// RefreshCatalog 获取远程主题目录
// @Summary 获取在线主题目录
// @Description 从远程 URL 获取主题目录索引，并标注各主题的安装状态
// @Tags 主题包
// @Accept json
// @Produce json
// @Param request body object true "目录请求"
// @Success 200 {object} object "目录响应"
// @Security BearerAuth
// @Router /theme-packs/catalog/refresh [post]
func (h *ThemePackHandler) RefreshCatalog(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CatalogURL  string `json:"catalog_url"`
		GithubProxy string `json:"github_proxy"`
		Force       bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}

	catalogURL := req.CatalogURL
	if catalogURL == "" {
		catalogURL = h.getCatalogURL()
	}

	// 检查缓存
	if !req.Force {
		if cached := themeCache.get(catalogURL); cached != nil {
			entries := h.resolveInstallStates(r.Context(), cached, catalogURL)
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"themes": entries,
				"total":  len(entries),
			})
			return
		}
	}

	// 获取远程目录
	catalog, err := fetchCatalog(r.Context(), catalogURL, req.GithubProxy)
	if err != nil {
		respondError(w, http.StatusBadGateway, "获取主题目录失败", err)
		return
	}

	themeCache.set(catalogURL, catalog)

	entries := h.resolveInstallStates(r.Context(), catalog, catalogURL)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"themes": entries,
		"total":  len(entries),
	})
}

// InstallFromCatalog 从在线目录安装主题包
// @Summary 从在线目录安装主题包
// @Description 下载远程主题包 JSON，验证 SHA-256 后导入
// @Tags 主题包
// @Accept json
// @Produce json
// @Param request body object true "安装请求"
// @Success 200 {object} services.ThemePackResponse "安装成功"
// @Failure 400 {object} models.ErrorResponse "校验失败"
// @Security BearerAuth
// @Router /theme-packs/catalog/install [post]
func (h *ThemePackHandler) InstallFromCatalog(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL         string `json:"url"`
		GithubProxy string `json:"github_proxy"`
		SHA256      string `json:"sha256"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if req.URL == "" {
		respondError(w, http.StatusBadRequest, "缺少主题包 URL", nil)
		return
	}

	// 下载主题包
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := httputil.GetWithGithubProxyFallback(r.Context(), client, req.URL, req.GithubProxy, httputil.GithubGetOptions{})
	if err != nil {
		respondError(w, http.StatusBadGateway, "下载主题包失败", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respondError(w, http.StatusBadGateway, fmt.Sprintf("下载主题包失败: HTTP %d", resp.StatusCode), nil)
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		respondError(w, http.StatusBadGateway, "读取主题包失败", err)
		return
	}

	// SHA-256 校验
	if req.SHA256 != "" {
		hash := sha256.Sum256(body)
		actual := hex.EncodeToString(hash[:])
		if actual != req.SHA256 {
			respondError(w, http.StatusBadRequest,
				fmt.Sprintf("SHA-256 校验失败: 期望 %s, 实际 %s", req.SHA256, actual), nil)
			return
		}
	}

	// 导入
	result, err := h.service.Import(r.Context(), body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "导入主题包失败", err)
		return
	}

	// 清除缓存以刷新安装状态
	themeCache.set("", nil)

	respondJSON(w, http.StatusOK, result)
}

// GetCatalogURLSetting 获取主题目录 URL 设置
func (h *ThemePackHandler) GetCatalogURLSetting(w http.ResponseWriter, r *http.Request) {
	url := h.getCatalogURL()
	respondJSON(w, http.StatusOK, map[string]string{"url": url})
}

// UpdateCatalogURLSetting 更新主题目录 URL 设置
func (h *ThemePackHandler) UpdateCatalogURLSetting(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if req.URL == "" {
		respondError(w, http.StatusBadRequest, "URL 不能为空", nil)
		return
	}
	if err := h.configService.Set(themeCatalogURLKey, req.URL); err != nil {
		respondError(w, http.StatusInternalServerError, "保存设置失败", err)
		return
	}
	// 清除缓存
	themeCache.set("", nil)
	respondJSON(w, http.StatusOK, map[string]string{"url": req.URL})
}

func (h *ThemePackHandler) getCatalogURL() string {
	return h.configService.GetString(themeCatalogURLKey, defaultThemeCatalogURL)
}

// resolveInstallStates 标注每个目录条目的安装状态
func (h *ThemePackHandler) resolveInstallStates(ctx context.Context, catalog *models.ThemeCatalogIndex, baseURL string) []models.ThemeCatalogEntry {
	installed, _ := h.service.List(ctx)
	installedMap := make(map[string]string) // themeID -> version
	for _, item := range installed {
		installedMap[item.ThemeID] = item.Version
	}

	// 计算 base URL（去掉最后的文件名）
	base := baseURL
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[:idx+1]
	}

	entries := make([]models.ThemeCatalogEntry, len(catalog.Themes))
	for i, entry := range catalog.Themes {
		entries[i] = entry

		// 拼接完整 URL
		if !strings.HasPrefix(entry.URL, "http://") && !strings.HasPrefix(entry.URL, "https://") {
			entries[i].URL = base + entry.URL
		}

		// 标注安装状态
		if ver, ok := installedMap[entry.ID]; ok {
			if ver == entry.Version {
				entries[i].InstallState = "installed"
			} else {
				entries[i].InstallState = "has_update"
			}
		} else {
			entries[i].InstallState = "not_installed"
		}
	}
	return entries
}

// fetchCatalog 从远程获取并解析目录索引
func fetchCatalog(ctx context.Context, catalogURL, githubProxy string) (*models.ThemeCatalogIndex, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := httputil.GetWithGithubProxyFallback(ctx, client, catalogURL, githubProxy, httputil.GithubGetOptions{})
	if err != nil {
		return nil, fmt.Errorf("fetch catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch catalog: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read catalog: %w", err)
	}

	var catalog models.ThemeCatalogIndex
	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}

	return &catalog, nil
}
