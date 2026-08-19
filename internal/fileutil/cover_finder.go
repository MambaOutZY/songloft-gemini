package fileutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// coverFileNames 外部封面文件名约定，按优先级排列。
var coverFileNames = []string{
	"cover",
	"Cover",
	"album",
	"folder",
	".folder",
	"AlbumArtSmall",
}

// coverImageExts 支持的封面图片扩展名，按优先级排列。
var coverImageExts = []string{".png", ".jpg", ".jpeg", ".webp", ".bmp", ".gif"}

// FindExternalCover 在指定目录中按约定查找外部封面图片文件。
//
// 查找顺序：
//  1. 按约定文件名 (cover, Cover, album, folder, .folder, AlbumArtSmall) 查找
//  2. 兜底：目录内唯一的一张图片文件
//  3. 未找到返回空字符串
func FindExternalCover(dirPath string) (string, error) {
	// 1. 按约定文件名优先级查找
	for _, name := range coverFileNames {
		for _, ext := range coverImageExts {
			candidate := filepath.Join(dirPath, name+ext)
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				return candidate, nil
			}
			upperExt := strings.ToUpper(ext)
			if upperExt != ext {
				candidate := filepath.Join(dirPath, name+upperExt)
				if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
					return candidate, nil
				}
			}
		}
	}

	// 2. 兜底：目录内唯一图片
	if path := findSingleImage(dirPath); path != "" {
		return path, nil
	}

	return "", nil
}

// SaveExternalCover 读取外部封面文件并保存到 coverStoragePath 下的分层目录（按内容哈希去重）。
// 返回保存后的封面路径，extPath 不存在或读取失败时返回空字符串。
func SaveExternalCover(extPath, coverStoragePath string) (string, error) {
	data, err := os.ReadFile(extPath)
	if err != nil {
		return "", fmt.Errorf("read external cover: %w", err)
	}
	if len(data) == 0 {
		return "", nil
	}

	ext := strings.TrimPrefix(filepath.Ext(extPath), ".")
	if ext == "" {
		ext = "jpg"
	}

	coverPath := GenerateCoverPath(data, ext, coverStoragePath)

	if err := os.MkdirAll(filepath.Dir(coverPath), 0755); err != nil {
		return "", fmt.Errorf("create cover dir: %w", err)
	}
	if err := os.WriteFile(coverPath, data, 0644); err != nil {
		return "", fmt.Errorf("write cover: %w", err)
	}

	return coverPath, nil
}

// GenerateCoverPath 根据图片内容哈希生成分层目录的封面文件路径。
func GenerateCoverPath(coverData []byte, ext, coverStoragePath string) string {
	hash := sha256.Sum256(coverData)
	hashStr := hex.EncodeToString(hash[:])
	dir1 := hashStr[0:2]
	dir2 := hashStr[2:4]
	fileExt := "." + strings.ToLower(ext)
	filename := hashStr + fileExt
	return filepath.Join(coverStoragePath, dir1, dir2, filename)
}

func findSingleImage(dirPath string) string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return ""
	}

	var imagePath string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if isImageExt(ext) {
			if imagePath != "" {
				return ""
			}
			imagePath = filepath.Join(dirPath, entry.Name())
		}
	}
	return imagePath
}

func isImageExt(ext string) bool {
	for _, e := range coverImageExts {
		if ext == e {
			return true
		}
	}
	return false
}
