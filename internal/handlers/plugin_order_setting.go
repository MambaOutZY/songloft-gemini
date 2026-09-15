package handlers

import (
	"encoding/json"
	"net/http"
)

const pluginOrderKey = "plugin_order"

// pluginOrderSetting 主页插件网格的自定义排序。
// order 元素为插件 entry_path；顺序即用户在主页拖拽后的显示顺序。
// 不覆盖已启用插件的显隐（那由插件自身 status 决定），也与
// tab_config 完全独立：一个插件可以在主页排第 3、在底部 Tab 排第 1。
type pluginOrderSetting struct {
	Order []string `json:"order"`
}

var defaultPluginOrder = pluginOrderSetting{Order: []string{}}

// GetPluginOrderSetting 获取主页插件网格排序
// @Summary 获取主页插件网格排序
// @Description 获取用户在主页插件网格中拖拽形成的排序（entry_path 列表）。未在此列表中的已安装插件由客户端按 GET /jsplugins 返回顺序追加到末尾；卸载插件在保存时或插件卸载时被静默清理。未配置时返回空列表（等价于沿用后端默认顺序）。
// @Tags 设置
// @Produce json
// @Success 200 {object} pluginOrderSetting "插件排序配置"
// @Security BearerAuth
// @Router /settings/plugin-order [get]
func (h *ConfigHandler) GetPluginOrderSetting(w http.ResponseWriter, r *http.Request) {
	var cfg pluginOrderSetting
	if err := h.configService.GetJSON(pluginOrderKey, &cfg); err != nil {
		respondJSON(w, http.StatusOK, defaultPluginOrder)
		return
	}
	if cfg.Order == nil {
		cfg.Order = []string{}
	}
	respondJSON(w, http.StatusOK, cfg)
}

// UpdatePluginOrderSetting 保存主页插件网格排序
// @Summary 保存主页插件网格排序
// @Description 保存用户在主页插件网格中拖拽形成的排序。order 元素为插件 entry_path，不能为空、不能重复；插件已不存在的条目会在保存时被静默清理（保存即自愈，客户端应以响应为准）。禁用插件的条目刻意保留：重新启用后位置自动恢复，与 tab-config 语义一致。
// @Tags 设置
// @Accept json
// @Produce json
// @Param request body pluginOrderSetting true "插件排序配置"
// @Success 200 {object} pluginOrderSetting "保存后的插件排序（孤儿条目已清理）"
// @Failure 400 {object} models.ErrorResponse "请求格式错误或校验失败"
// @Failure 500 {object} models.ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /settings/plugin-order [put]
func (h *ConfigHandler) UpdatePluginOrderSetting(w http.ResponseWriter, r *http.Request) {
	var req pluginOrderSetting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if req.Order == nil {
		req.Order = []string{}
	}

	seen := make(map[string]bool, len(req.Order))
	for _, ep := range req.Order {
		if ep == "" {
			respondError(w, http.StatusBadRequest, "entry_path 不能为空", nil)
			return
		}
		if seen[ep] {
			respondError(w, http.StatusBadRequest, "entry_path 不能重复: "+ep, nil)
			return
		}
		seen[ep] = true
	}

	plugins, err := h.pluginRepo.GetAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取插件列表失败", err)
		return
	}
	installed := make(map[string]bool, len(plugins))
	for _, p := range plugins {
		installed[p.EntryPath] = true
	}

	// 清理孤儿条目（插件已卸载）：与 tab_config PUT 的自愈语义一致。
	// 禁用插件的条目保留：重启用后位置自动恢复。
	cleaned := make([]string, 0, len(req.Order))
	for _, ep := range req.Order {
		if installed[ep] {
			cleaned = append(cleaned, ep)
		}
	}
	req.Order = cleaned

	if err := h.configService.SetJSON(pluginOrderKey, req); err != nil {
		respondError(w, http.StatusInternalServerError, "保存配置失败", err)
		return
	}
	respondJSON(w, http.StatusOK, req)
}
