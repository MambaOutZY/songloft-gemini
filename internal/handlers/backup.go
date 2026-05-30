package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"songloft/internal/models"
	"songloft/internal/services"
)

// BackupHandler 歌单备份还原处理器
type BackupHandler struct {
	backupService *services.BackupService
}

// NewBackupHandler 创建备份处理器
func NewBackupHandler(backupService *services.BackupService) *BackupHandler {
	return &BackupHandler{backupService: backupService}
}

// ExportPlaylists 导出所有歌单数据为 JSON 文件
func (h *BackupHandler) ExportPlaylists(w http.ResponseWriter, r *http.Request) {
	data, err := h.backupService.Export(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "导出歌单数据失败", err)
		return
	}

	filename := fmt.Sprintf("songloft-backup-%s.json", time.Now().Format("20060102"))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		respondError(w, http.StatusInternalServerError, "写入导出数据失败", err)
	}
}

// ImportPlaylists 从 JSON 文件导入歌单数据
func (h *BackupHandler) ImportPlaylists(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "请求数据过大或格式错误", err)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "缺少备份文件", err)
		return
	}
	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusBadRequest, "读取备份文件失败", err)
		return
	}

	var data models.BackupData
	if err := json.Unmarshal(body, &data); err != nil {
		respondError(w, http.StatusBadRequest, "无效的备份文件格式", err)
		return
	}

	result, err := h.backupService.Import(r.Context(), &data)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "导入失败", err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}
