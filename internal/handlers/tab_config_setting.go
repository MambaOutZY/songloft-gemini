package handlers

import (
	"encoding/json"
	"net/http"

	"songloft/internal/models"
)

const tabConfigKey = "tab_config"

// tabConfigSetting 底部导航栏 Tab 配置
type tabConfigSetting struct {
	ShowLibrary   bool             `json:"show_library"`
	ShowPlaylists bool             `json:"show_playlists"`
	PluginTabs    []pluginTabEntry `json:"plugin_tabs"`
}

// pluginTabEntry 插件 Tab 条目
type pluginTabEntry struct {
	PluginID  int    `json:"plugin_id"`
	EntryPath string `json:"entry_path"`
	Name      string `json:"name"`
}

var defaultTabConfig = tabConfigSetting{
	ShowLibrary:   true,
	ShowPlaylists: true,
	PluginTabs:    []pluginTabEntry{},
}

// GetTabConfigSetting 获取底部导航栏 Tab 配置
// @Summary 获取底部导航栏 Tab 配置
// @Description 获取用户自定义的底部导航栏 Tab 配置。首页和设置固定显示，歌曲库和歌单可关闭，最多可添加 2 个插件 Tab，总计不超过 5 个。未配置时返回默认值（4 个 Tab：首页、歌曲库、歌单、设置）。
// @Tags 设置
// @Produce json
// @Success 200 {object} tabConfigSetting "Tab 配置"
// @Security BearerAuth
// @Router /settings/tab-config [get]
func (h *ConfigHandler) GetTabConfigSetting(w http.ResponseWriter, r *http.Request) {
	var cfg tabConfigSetting
	if err := h.configService.GetJSON(tabConfigKey, &cfg); err != nil {
		respondJSON(w, http.StatusOK, defaultTabConfig)
		return
	}
	if cfg.PluginTabs == nil {
		cfg.PluginTabs = []pluginTabEntry{}
	}
	respondJSON(w, http.StatusOK, cfg)
}

// UpdateTabConfigSetting 保存底部导航栏 Tab 配置
// @Summary 保存底部导航栏 Tab 配置
// @Description 保存用户自定义的底部导航栏 Tab 配置。首页和设置固定显示（不在配置中），可选项为歌曲库、歌单和插件 Tab，可选项总数不超过 3 个（总计 5 个 Tab）。每个插件 Tab 的 entry_path 和 name 不能为空，且不能重复。
// @Tags 设置
// @Accept json
// @Produce json
// @Param request body tabConfigSetting true "Tab 配置"
// @Success 200 {object} tabConfigSetting "保存后的 Tab 配置"
// @Failure 400 {object} models.ErrorResponse "请求格式错误或校验失败"
// @Failure 500 {object} models.ErrorResponse "保存配置失败"
// @Security BearerAuth
// @Router /settings/tab-config [put]
func (h *ConfigHandler) UpdateTabConfigSetting(w http.ResponseWriter, r *http.Request) {
	var req tabConfigSetting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if req.PluginTabs == nil {
		req.PluginTabs = []pluginTabEntry{}
	}

	optionalCount := len(req.PluginTabs)
	if req.ShowLibrary {
		optionalCount++
	}
	if req.ShowPlaylists {
		optionalCount++
	}
	if optionalCount > 3 {
		respondError(w, http.StatusBadRequest, "可选标签总数不能超过 3 个（总计最多 5 个标签）", nil)
		return
	}

	seen := make(map[string]bool, len(req.PluginTabs))
	for _, pt := range req.PluginTabs {
		if pt.EntryPath == "" {
			respondError(w, http.StatusBadRequest, "插件 Tab 的 entry_path 不能为空", nil)
			return
		}
		if pt.Name == "" {
			respondError(w, http.StatusBadRequest, "插件 Tab 的 name 不能为空", nil)
			return
		}
		if seen[pt.EntryPath] {
			respondError(w, http.StatusBadRequest, "插件 Tab 的 entry_path 不能重复", nil)
			return
		}
		seen[pt.EntryPath] = true
	}

	if err := h.configService.SetJSON(tabConfigKey, req); err != nil {
		respondError(w, http.StatusInternalServerError, "保存配置失败", err)
		return
	}
	respondJSON(w, http.StatusOK, req)
}

// Ensure models.ErrorResponse is referenced for swagger generation.
var _ = models.ErrorResponse{}
