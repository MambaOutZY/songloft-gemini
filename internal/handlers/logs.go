package handlers

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"songloft/internal/logging"
)

const defaultLogExportMaxBytes int64 = 10 << 20 // 10 MiB

type logFileSegment struct {
	path   string
	offset int64
	length int64
}

// LogExportHandler 暴露 GET /api/v1/logs/export：把后端落盘的日志文件按时间顺序
// 拼接、逐行脱敏后作为附件返回，供用户下载并附到 issue。
//
// 与 LogHandler（日志等级配置）分开：本 handler 只依赖日志落盘目录，不涉及配置读写。
type LogExportHandler struct {
	logDir        string // 日志落盘目录（<data_dir>/logs/）；与 app 侧 RotateWriter 一致
	maxExportSize int64
}

// NewLogExportHandler 构造 LogExportHandler。logDir 为空时导出会返回空内容（附提示行）。
func NewLogExportHandler(logDir string) *LogExportHandler {
	return &LogExportHandler{
		logDir:        logDir,
		maxExportSize: defaultLogExportMaxBytes,
	}
}

// ExportLogs 处理 GET /api/v1/logs/export
// @Summary 导出后端日志
// @Description 导出后端最近 10 MiB 原始日志，按时间从旧到新拼接并逐行脱敏后作为纯文本附件返回；日志总量超过上限时会在首行注明截断。脱敏会抹除密钥/token/密码、Authorization/Cookie 头、URL 内嵌凭证、客户端 IP 主机位、用户主目录名等敏感信息，便于用户安全地附到 issue。远程服务器、桌面 Bundle、移动 Bundle 三种模式下均可用（均由同一份后端提供该端点）。无日志文件时返回仅含提示行的文本。
// @Tags 设置
// @Produce plain
// @Success 200 {file} binary "脱敏后的后端日志（text/plain）"
// @Failure 500 {object} map[string]string "读取日志目录失败"
// @Security BearerAuth
// @Router /logs/export [get]
func (h *LogExportHandler) ExportLogs(w http.ResponseWriter, r *http.Request) {
	filename := fmt.Sprintf("songloft-backend-logs-%s.log", time.Now().Format("20060102"))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	if h.logDir == "" {
		fmt.Fprintln(w, "# 无可用的后端日志（日志落盘未启用）")
		return
	}

	files, err := logging.ListLogFiles(h.logDir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "读取日志目录失败", err)
		return
	}
	if len(files) == 0 {
		fmt.Fprintln(w, "# 暂无后端日志文件")
		return
	}

	segments, totalSize := selectRecentLogSegments(files, h.maxExportSize)
	if totalSize > h.maxExportSize {
		fmt.Fprintf(
			w,
			"# 后端日志已截断：原始日志共 %d 字节，仅导出最近 %d MiB。\n",
			totalSize,
			h.maxExportSize>>20,
		)
	}

	// 逐文件、逐行流式脱敏写出，避免把整份日志读进内存。
	// 头部已发出，途中出错无法改状态码，只能记录并中断。
	for _, segment := range segments {
		f, err := os.Open(segment.path)
		if err != nil {
			slog.Warn("导出日志：打开文件失败，跳过", "path", segment.path, "error", err)
			continue
		}
		reader := io.NewSectionReader(f, segment.offset, segment.length)
		buffered := bufio.NewReader(reader)
		if segment.offset > 0 && !startsAtLineBoundary(f, segment.offset) {
			if err := discardPartialLine(buffered); err != nil {
				slog.Warn("导出日志：跳过截断行失败", "path", segment.path, "error", err)
				f.Close()
				continue
			}
		}
		if err := logging.RedactStream(w, buffered); err != nil {
			slog.Warn("导出日志：脱敏写出中断", "path", segment.path, "error", err)
			f.Close()
			return
		}
		f.Close()
	}
}

func selectRecentLogSegments(paths []string, maxBytes int64) ([]logFileSegment, int64) {
	if maxBytes <= 0 {
		maxBytes = defaultLogExportMaxBytes
	}
	type sizedLogFile struct {
		path string
		size int64
	}
	files := make([]sizedLogFile, 0, len(paths))
	var totalSize int64
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			if err != nil {
				slog.Warn("导出日志：读取文件信息失败，跳过", "path", path, "error", err)
			}
			continue
		}
		files = append(files, sizedLogFile{path: path, size: info.Size()})
		totalSize += info.Size()
	}

	remaining := maxBytes
	reversed := make([]logFileSegment, 0, len(files))
	for i := len(files) - 1; i >= 0 && remaining > 0; i-- {
		length := min(files[i].size, remaining)
		reversed = append(reversed, logFileSegment{
			path:   files[i].path,
			offset: files[i].size - length,
			length: length,
		})
		remaining -= length
	}
	segments := make([]logFileSegment, len(reversed))
	for i := range reversed {
		segments[len(reversed)-1-i] = reversed[i]
	}
	return segments, totalSize
}

func startsAtLineBoundary(f *os.File, offset int64) bool {
	if offset <= 0 {
		return true
	}
	previous := []byte{0}
	_, err := f.ReadAt(previous, offset-1)
	return err == nil && previous[0] == '\n'
}

func discardPartialLine(reader *bufio.Reader) error {
	for {
		_, err := reader.ReadSlice('\n')
		switch err {
		case nil, io.EOF:
			return nil
		case bufio.ErrBufferFull:
			continue
		default:
			return err
		}
	}
}
