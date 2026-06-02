package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"songloft/internal/services"
)

// logLevelConfigKey 是日志等级在 configs 表中的 key。
// 业务封装（GetLevel / SetLevel / GetLevelSetting / UpdateLevelSetting）是唯一访问入口，
// 通用 /api/v1/configs/{key} 不预置此 key，避免双入口造成不一致。
const logLevelConfigKey = "log_level"

// ParseLogLevel 将字符串映射为 slog.Level。
// 导出供 app 包启动时复用同一份映射表。
func ParseLogLevel(s string) (slog.Level, bool) {
	switch s {
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "warn":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	}
	return 0, false
}

// LogHandler 暴露 /settings/log-level 端点：读写日志等级配置，PUT 时同步修改
// 全局 slog.LevelVar 让新等级即时生效（无需重启）。
type LogHandler struct {
	configService *services.ConfigService
	levelVar      *slog.LevelVar // 注入自 App.Init 创建的同一实例
}

// NewLogHandler 构造 LogHandler。
// configService 可为 nil（测试场景下不走读写时），此时 GetLevel 返回默认 "info"，SetLevel 报错。
// levelVar 可为 nil（仅持久化、不切运行时等级的场景）；为 nil 时 SetLevel 仍写库，但不影响活跃 logger。
func NewLogHandler(configService *services.ConfigService, levelVar *slog.LevelVar) *LogHandler {
	return &LogHandler{configService: configService, levelVar: levelVar}
}

// GetLevel 返回当前持久化的日志等级。缺失或非法值时退化为默认 "info"。
func (h *LogHandler) GetLevel() string {
	if h.configService == nil {
		return "info"
	}
	level := h.configService.GetString(logLevelConfigKey, "info")
	if _, ok := ParseLogLevel(level); !ok {
		return "info"
	}
	return level
}

// SetLevel 持久化日志等级并立即切换运行时 LevelVar。
// 非法等级返回错误，不写库也不改 LevelVar。
func (h *LogHandler) SetLevel(level string) error {
	lvl, ok := ParseLogLevel(level)
	if !ok {
		return fmt.Errorf("无效的日志等级: %s（仅支持 debug/info/warn/error）", level)
	}
	if h.configService == nil {
		return fmt.Errorf("configService 未注入，无法持久化日志等级")
	}
	if err := h.configService.Set(logLevelConfigKey, level); err != nil {
		return err
	}
	if h.levelVar != nil {
		h.levelVar.Set(lvl)
	}
	return nil
}

// logLevelSettingRequest /settings/log-level PUT 请求体。
type logLevelSettingRequest struct {
	Level string `json:"level"`
}

// GetLevelSetting 处理 GET /api/v1/settings/log-level
// @Summary 获取日志等级
// @Description 返回当前 slog 全局日志等级
// @Tags 设置
// @Produce json
// @Success 200 {object} map[string]string "返回 level 字段：debug/info/warn/error"
// @Security BearerAuth
// @Router /settings/log-level [get]
func (h *LogHandler) GetLevelSetting(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"level": h.GetLevel()})
}

// UpdateLevelSetting 处理 PUT /api/v1/settings/log-level
// @Summary 更新日志等级
// @Description 切换 slog 全局日志等级并持久化。新等级即时生效，重启后从 DB 恢复。
// @Tags 设置
// @Accept json
// @Produce json
// @Param request body logLevelSettingRequest true "等级请求（debug/info/warn/error）"
// @Success 200 {object} map[string]string "返回更新后的 level"
// @Failure 400 {object} map[string]string "请求格式错误或等级非法"
// @Failure 500 {object} map[string]string "保存配置失败"
// @Security BearerAuth
// @Router /settings/log-level [put]
func (h *LogHandler) UpdateLevelSetting(w http.ResponseWriter, r *http.Request) {
	var req logLevelSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	if _, ok := ParseLogLevel(req.Level); !ok {
		respondError(w, http.StatusBadRequest, "无效的日志等级",
			fmt.Errorf("level 必须是 debug/info/warn/error 之一，收到 %q", req.Level))
		return
	}
	if err := h.SetLevel(req.Level); err != nil {
		respondError(w, http.StatusInternalServerError, "保存配置失败", err)
		return
	}
	slog.Info("日志等级已切换", "level", req.Level)
	respondJSON(w, http.StatusOK, map[string]string{"level": req.Level})
}
