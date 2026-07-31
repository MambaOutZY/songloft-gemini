package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"songloft/internal/models"
	"songloft/internal/services"
)

// ThemePackHandler 主题包 HTTP 处理器
type ThemePackHandler struct {
	service       *services.ThemePackService
	configService *services.ConfigService
}

// NewThemePackHandler 创建主题包处理器
func NewThemePackHandler(service *services.ThemePackService, configService *services.ConfigService) *ThemePackHandler {
	return &ThemePackHandler{service: service, configService: configService}
}

// ListThemePacks 列出所有已安装主题包
// @Summary 列出主题包
// @Description 返回所有已安装的主题包列表
// @Tags 主题包
// @Produce json
// @Success 200 {array} services.ThemePackListItem "主题包列表"
// @Security BearerAuth
// @Router /theme-packs [get]
func (h *ThemePackHandler) ListThemePacks(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取主题包列表失败", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

// GetThemePack 获取单个主题包详情
// @Summary 获取主题包详情
// @Description 返回主题包的完整信息，含主题配置数据
// @Tags 主题包
// @Produce json
// @Param themeID path string true "主题包 ID"
// @Success 200 {object} services.ThemePackResponse "主题包详情"
// @Failure 404 {object} models.ErrorResponse "主题包不存在"
// @Security BearerAuth
// @Router /theme-packs/{themeID} [get]
func (h *ThemePackHandler) GetThemePack(w http.ResponseWriter, r *http.Request) {
	themeID := chi.URLParam(r, "themeID")
	if themeID == "" {
		respondError(w, http.StatusBadRequest, "缺少主题包 ID", nil)
		return
	}

	resp, err := h.service.Get(r.Context(), themeID)
	if err != nil {
		if errors.Is(err, models.ErrThemePackNotFound) {
			respondError(w, http.StatusNotFound, "主题包不存在", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "获取主题包失败", err)
		return
	}
	respondJSON(w, http.StatusOK, resp)
}

// ImportThemePack 导入主题包
// @Summary 导入主题包
// @Description 接收 .songloft-theme JSON 内容，验证后存入。相同 ID 的主题包会被更新。
// @Tags 主题包
// @Accept json
// @Produce json
// @Param request body object true "主题包 JSON（完整 .songloft-theme 文件内容）"
// @Success 200 {object} services.ThemePackResponse "导入成功"
// @Failure 400 {object} models.ErrorResponse "格式错误"
// @Security BearerAuth
// @Router /theme-packs [post]
func (h *ThemePackHandler) ImportThemePack(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB 限制
	if err != nil {
		respondError(w, http.StatusBadRequest, "读取请求体失败", err)
		return
	}
	if len(body) == 0 {
		respondError(w, http.StatusBadRequest, "请求体为空", nil)
		return
	}

	// 验证是合法 JSON
	if !json.Valid(body) {
		respondError(w, http.StatusBadRequest, "无效的 JSON 格式", nil)
		return
	}

	resp, err := h.service.Import(r.Context(), body)
	if err != nil {
		if errors.Is(err, models.ErrThemePackInvalidSchema) ||
			errors.Is(err, models.ErrThemePackMissingID) ||
			errors.Is(err, models.ErrThemePackMissingName) ||
			errors.Is(err, models.ErrThemePackInvalidColor) ||
			errors.Is(err, models.ErrThemePackInvalidRadius) ||
			errors.Is(err, models.ErrThemePackInvalidJSON) {
			respondError(w, http.StatusBadRequest, "主题包格式错误", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "导入主题包失败", err)
		return
	}
	respondJSON(w, http.StatusOK, resp)
}

// DeleteThemePack 删除主题包
// @Summary 删除主题包
// @Description 删除指定主题包，若为当前激活主题则同时恢复默认
// @Tags 主题包
// @Produce json
// @Param themeID path string true "主题包 ID"
// @Success 200 {object} models.SuccessResponse "删除成功"
// @Failure 404 {object} models.ErrorResponse "主题包不存在"
// @Security BearerAuth
// @Router /theme-packs/{themeID} [delete]
func (h *ThemePackHandler) DeleteThemePack(w http.ResponseWriter, r *http.Request) {
	themeID := chi.URLParam(r, "themeID")
	if themeID == "" {
		respondError(w, http.StatusBadRequest, "缺少主题包 ID", nil)
		return
	}

	err := h.service.Delete(r.Context(), themeID)
	if err != nil {
		if errors.Is(err, models.ErrThemePackNotFound) {
			respondError(w, http.StatusNotFound, "主题包不存在", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "删除主题包失败", err)
		return
	}
	respondJSON(w, http.StatusOK, models.SuccessResponse{Message: "删除成功"})
}

// SetActiveThemePack 设置激活主题包
// @Summary 设置激活主题包
// @Description 将指定主题包设为当前使用的主题
// @Tags 主题包
// @Accept json
// @Produce json
// @Param request body object true "激活请求" example({"theme_id":"example.neon-night"})
// @Success 200 {object} models.SuccessResponse "设置成功"
// @Failure 400 {object} models.ErrorResponse "请求格式错误"
// @Failure 404 {object} models.ErrorResponse "主题包不存在"
// @Security BearerAuth
// @Router /theme-packs/active [put]
func (h *ThemePackHandler) SetActiveThemePack(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ThemeID string `json:"theme_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if req.ThemeID == "" {
		respondError(w, http.StatusBadRequest, "缺少 theme_id", nil)
		return
	}

	err := h.service.SetActive(r.Context(), req.ThemeID)
	if err != nil {
		if errors.Is(err, models.ErrThemePackNotFound) {
			respondError(w, http.StatusNotFound, "主题包不存在", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "设置激活主题包失败", err)
		return
	}
	respondJSON(w, http.StatusOK, models.SuccessResponse{Message: "设置成功"})
}

// GetActiveThemePack 获取当前激活的主题包
// @Summary 获取激活主题包
// @Description 返回当前激活的主题包详情，无激活主题时返回 null
// @Tags 主题包
// @Produce json
// @Success 200 {object} services.ThemePackResponse "激活主题包（可为 null）"
// @Security BearerAuth
// @Router /theme-packs/active [get]
func (h *ThemePackHandler) GetActiveThemePack(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetActive(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取激活主题包失败", err)
		return
	}
	respondJSON(w, http.StatusOK, resp)
}

// ClearActiveThemePack 清除激活主题包（恢复默认）
// @Summary 恢复默认主题
// @Description 清除当前激活的主题包，恢复为默认主题
// @Tags 主题包
// @Produce json
// @Success 200 {object} models.SuccessResponse "已恢复默认主题"
// @Security BearerAuth
// @Router /theme-packs/active [delete]
func (h *ThemePackHandler) ClearActiveThemePack(w http.ResponseWriter, r *http.Request) {
	h.service.ClearActive()
	respondJSON(w, http.StatusOK, models.SuccessResponse{Message: "已恢复默认主题"})
}
