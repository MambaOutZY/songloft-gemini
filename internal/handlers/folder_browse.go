package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"songloft/internal/database"
	"songloft/internal/models"
)

// ListFolders 文件夹浏览：返回指定路径下的子文件夹（含递归歌曲计数）和直属歌曲。
// @Summary 文件夹浏览
// @Description 按目录层级浏览本地歌曲。path 为相对于 music_path 的路径（空 = 根目录），返回下一级子文件夹及其递归歌曲数和当前目录的直属歌曲。
// @Tags 歌曲管理
// @Produce json
// @Param path query string false "相对路径，空 = 音乐根目录"
// @Param keyword query string false "搜索过滤（文件夹名 / 歌曲标题）"
// @Success 200 {object} map[string]any "成功 {path, parent_path, music_path, folders:[{name,path,song_count}], total_folders, songs:[...], total_songs}"
// @Failure 400 {object} map[string]string "未设置音乐目录 / 非法路径"
// @Failure 500 {object} map[string]string "服务器错误"
// @Security BearerAuth
// @Router /songs/folders [get]
func (h *SongHandler) ListFolders(w http.ResponseWriter, r *http.Request) {
	musicPath := ""
	if h.getMusicPath != nil {
		musicPath = h.getMusicPath()
	}
	if musicPath == "" {
		respondError(w, http.StatusBadRequest, "未设置音乐目录", nil)
		return
	}

	q := r.URL.Query()
	relPath := q.Get("path")
	keyword := q.Get("keyword")

	// Sanitize path to prevent directory traversal
	if relPath != "" {
		relPath = filepath.ToSlash(filepath.Clean(relPath))
		if strings.Contains(relPath, "..") {
			respondError(w, http.StatusBadRequest, "非法路径", nil)
			return
		}
		relPath = strings.TrimPrefix(relPath, "/")
	}

	// Build absolute prefix with trailing slash
	var absPrefix string
	if relPath == "" {
		absPrefix = musicPath + "/"
	} else {
		absPrefix = musicPath + "/" + relPath + "/"
	}

	ctx := r.Context()

	// Get subfolders
	folderList, err := h.songService.ListFolders(ctx, absPrefix, keyword)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "查询文件夹失败", err)
		return
	}
	if folderList == nil {
		folderList = []database.FolderInfo{}
	}

	// Build relative paths for each folder
	for i := range folderList {
		if relPath == "" {
			folderList[i].Path = folderList[i].Name
		} else {
			folderList[i].Path = relPath + "/" + folderList[i].Name
		}
	}

	// Get direct songs in this folder
	songList, err := h.songService.ListDirectSongs(ctx, absPrefix, keyword)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "查询歌曲失败", err)
		return
	}
	if songList == nil {
		songList = []*models.Song{}
	}

	// Compute parent path
	parentPath := ""
	if relPath != "" {
		parentPath = filepath.ToSlash(filepath.Dir(relPath))
		if parentPath == "." {
			parentPath = ""
		}
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"path":          relPath,
		"parent_path":   parentPath,
		"music_path":    musicPath,
		"folders":       folderList,
		"total_folders": len(folderList),
		"songs":         songList,
		"total_songs":   len(songList),
	})
}
