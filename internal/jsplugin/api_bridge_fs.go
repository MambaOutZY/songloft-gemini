package jsplugin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxFSFileSize = 10 << 20 // 10MB

func (h *BridgeHandler) handleFS(action, data string) (string, error) {
	switch action {
	case "fs.readFile":
		return h.fsReadFile(data)
	case "fs.writeFile":
		return h.fsWriteFile(data)
	case "fs.appendFile":
		return h.fsAppendFile(data)
	case "fs.readdir":
		return h.fsReaddir(data)
	case "fs.unlink":
		return h.fsUnlink(data)
	case "fs.exists":
		return h.fsExists(data)
	case "fs.mkdir":
		return h.fsMkdir(data)
	case "fs.stat":
		return h.fsStat(data)
	case "fs.rename":
		return h.fsRename(data)
	default:
		return "", fmt.Errorf("unknown fs action: %s", action)
	}
}

func (h *BridgeHandler) resolveFSPath(relPath string) (string, error) {
	if relPath == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	if strings.Contains(relPath, "..") {
		return "", fmt.Errorf("path cannot contain '..'")
	}
	base := h.pluginDataDir()
	joined := filepath.Join(base, relPath)
	absClean, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	sep := string(filepath.Separator)
	if absClean != base && !strings.HasPrefix(absClean, base+sep) {
		return "", fmt.Errorf("path escapes plugin data directory")
	}
	return absClean, nil
}

func (h *BridgeHandler) fsReadFile(data string) (string, error) {
	var req struct {
		Path     string `json:"path"`
		Encoding string `json:"encoding"`
	}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return "", fmt.Errorf("fs.readFile: %w", err)
	}

	absPath, err := h.resolveFSPath(req.Path)
	if err != nil {
		return "", fmt.Errorf("fs.readFile: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("fs.readFile: %w", err)
	}
	if info.Size() > maxFSFileSize {
		return "", fmt.Errorf("fs.readFile: file exceeds %dMB limit", maxFSFileSize>>20)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("fs.readFile: %w", err)
	}

	if req.Encoding == "base64" {
		return base64.StdEncoding.EncodeToString(content), nil
	}
	return string(content), nil
}

func (h *BridgeHandler) fsWriteFile(data string) (string, error) {
	var req struct {
		Path     string `json:"path"`
		Data     string `json:"data"`
		Encoding string `json:"encoding"`
	}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return "", fmt.Errorf("fs.writeFile: %w", err)
	}

	absPath, err := h.resolveFSPath(req.Path)
	if err != nil {
		return "", fmt.Errorf("fs.writeFile: %w", err)
	}

	var content []byte
	if req.Encoding == "base64" {
		content, err = base64.StdEncoding.DecodeString(req.Data)
		if err != nil {
			return "", fmt.Errorf("fs.writeFile: invalid base64: %w", err)
		}
	} else {
		content = []byte(req.Data)
	}

	if len(content) > maxFSFileSize {
		return "", fmt.Errorf("fs.writeFile: data exceeds %dMB limit", maxFSFileSize>>20)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("fs.writeFile: mkdir: %w", err)
	}

	if err := os.WriteFile(absPath, content, 0o644); err != nil {
		return "", fmt.Errorf("fs.writeFile: %w", err)
	}
	return "", nil
}

func (h *BridgeHandler) fsAppendFile(data string) (string, error) {
	var req struct {
		Path     string `json:"path"`
		Data     string `json:"data"`
		Encoding string `json:"encoding"`
	}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return "", fmt.Errorf("fs.appendFile: %w", err)
	}

	absPath, err := h.resolveFSPath(req.Path)
	if err != nil {
		return "", fmt.Errorf("fs.appendFile: %w", err)
	}

	var content []byte
	if req.Encoding == "base64" {
		content, err = base64.StdEncoding.DecodeString(req.Data)
		if err != nil {
			return "", fmt.Errorf("fs.appendFile: invalid base64: %w", err)
		}
	} else {
		content = []byte(req.Data)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("fs.appendFile: mkdir: %w", err)
	}

	f, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("fs.appendFile: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return "", fmt.Errorf("fs.appendFile: %w", err)
	}
	return "", nil
}

func (h *BridgeHandler) fsReaddir(data string) (string, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return "", fmt.Errorf("fs.readdir: %w", err)
	}

	absPath, err := h.resolveFSPath(req.Path)
	if err != nil {
		return "", fmt.Errorf("fs.readdir: %w", err)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("fs.readdir: %w", err)
	}

	type dirEntry struct {
		Name  string `json:"name"`
		IsDir bool   `json:"isDir"`
	}
	result := make([]dirEntry, 0, len(entries))
	for _, e := range entries {
		result = append(result, dirEntry{Name: e.Name(), IsDir: e.IsDir()})
	}

	out, _ := json.Marshal(result)
	return string(out), nil
}

func (h *BridgeHandler) fsUnlink(data string) (string, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return "", fmt.Errorf("fs.unlink: %w", err)
	}

	absPath, err := h.resolveFSPath(req.Path)
	if err != nil {
		return "", fmt.Errorf("fs.unlink: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("fs.unlink: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("fs.unlink: cannot unlink directory, use rmdir or remove via readdir")
	}

	if err := os.Remove(absPath); err != nil {
		return "", fmt.Errorf("fs.unlink: %w", err)
	}
	return "", nil
}

func (h *BridgeHandler) fsExists(data string) (string, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return "", fmt.Errorf("fs.exists: %w", err)
	}

	absPath, err := h.resolveFSPath(req.Path)
	if err != nil {
		return "", fmt.Errorf("fs.exists: %w", err)
	}

	_, err = os.Stat(absPath)
	if err != nil {
		return "false", nil
	}
	return "true", nil
}

func (h *BridgeHandler) fsMkdir(data string) (string, error) {
	var req struct {
		Path      string `json:"path"`
		Recursive bool   `json:"recursive"`
	}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return "", fmt.Errorf("fs.mkdir: %w", err)
	}

	absPath, err := h.resolveFSPath(req.Path)
	if err != nil {
		return "", fmt.Errorf("fs.mkdir: %w", err)
	}

	if req.Recursive {
		err = os.MkdirAll(absPath, 0o755)
	} else {
		err = os.Mkdir(absPath, 0o755)
	}
	if err != nil {
		return "", fmt.Errorf("fs.mkdir: %w", err)
	}
	return "", nil
}

func (h *BridgeHandler) fsStat(data string) (string, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return "", fmt.Errorf("fs.stat: %w", err)
	}

	absPath, err := h.resolveFSPath(req.Path)
	if err != nil {
		return "", fmt.Errorf("fs.stat: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("fs.stat: %w", err)
	}

	result, _ := json.Marshal(struct {
		Size    int64 `json:"size"`
		ModTime int64 `json:"modTime"`
		IsDir   bool  `json:"isDir"`
	}{
		Size:    info.Size(),
		ModTime: info.ModTime().UnixMilli(),
		IsDir:   info.IsDir(),
	})
	return string(result), nil
}

func (h *BridgeHandler) fsRename(data string) (string, error) {
	var req struct {
		OldPath string `json:"oldPath"`
		NewPath string `json:"newPath"`
	}
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return "", fmt.Errorf("fs.rename: %w", err)
	}

	absOld, err := h.resolveFSPath(req.OldPath)
	if err != nil {
		return "", fmt.Errorf("fs.rename: oldPath: %w", err)
	}
	absNew, err := h.resolveFSPath(req.NewPath)
	if err != nil {
		return "", fmt.Errorf("fs.rename: newPath: %w", err)
	}

	newDir := filepath.Dir(absNew)
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return "", fmt.Errorf("fs.rename: mkdir: %w", err)
	}

	if err := os.Rename(absOld, absNew); err != nil {
		return "", fmt.Errorf("fs.rename: %w", err)
	}
	return "", nil
}
