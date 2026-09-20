package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// respondJSON 返回JSON响应
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("json encode failed", "error", err)
	}
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
