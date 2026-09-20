package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

// respondJSON 返回JSON响应
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("json encode failed", "error", err)
	}
}

// parsePagination 从请求中解析并校验分页参数。
// defaultLimit 为 limit 缺省时的默认值，maxLimit 为允许的上限（<=0 表示不限）。
func parsePagination(r *http.Request, defaultLimit, maxLimit int) (limit, offset int) {
	limit = defaultLimit
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		limit = l
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = o
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// respondError 返回错误响应
func respondError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]string{
		"error": message,
	}
	if err != nil {
		response["detail"] = err.Error()
	}
	respondJSON(w, status, response)
}
