package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"
)

// 主题包相关错误
var (
	ErrThemePackInvalidSchema = errors.New("unsupported schema version")
	ErrThemePackMissingID     = errors.New("missing theme pack id")
	ErrThemePackMissingName   = errors.New("missing theme pack name")
	ErrThemePackInvalidColor  = errors.New("invalid color format, expected #RRGGBB")
	ErrThemePackInvalidRadius = errors.New("radius value out of range (0-100)")
	ErrThemePackAlreadyExists = errors.New("theme pack with this id already exists")
	ErrThemePackNotFound      = errors.New("theme pack not found")
	ErrThemePackInvalidJSON   = errors.New("invalid theme pack JSON")
)

// ThemePack 数据库行映射
type ThemePack struct {
	ID            int64     `json:"id"`
	ThemeID       string    `json:"theme_id"`
	Name          string    `json:"name"`
	Version       string    `json:"version"`
	Author        string    `json:"author"`
	Description   string    `json:"description"`
	SchemaVersion int       `json:"schema_version"`
	RawJSON       string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ThemePackData 从 RawJSON 解析出的完整主题配置
type ThemePackData struct {
	SchemaVersion    int              `json:"schemaVersion"`
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Version          string           `json:"version"`
	Author           string           `json:"author"`
	Description      string           `json:"description"`
	Light            *ThemePackColors `json:"light,omitempty"`
	Dark             *ThemePackColors `json:"dark,omitempty"`
	PlayerGradient   []string         `json:"playerGradient,omitempty"`
	CardRadius       *float64         `json:"cardRadius,omitempty"`
	ControlRadius    *float64         `json:"controlRadius,omitempty"`
	NavigationRadius *float64         `json:"navigationRadius,omitempty"`
}

// ThemePackColors 亮色或暗色配色方案
type ThemePackColors struct {
	SeedColor       string `json:"seedColor"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	SurfaceColor    string `json:"surfaceColor,omitempty"`
}

var colorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// ParseThemePackData 解析并验证主题包 JSON
func ParseThemePackData(rawJSON []byte) (*ThemePackData, error) {
	var data ThemePackData
	if err := json.Unmarshal(rawJSON, &data); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrThemePackInvalidJSON, err)
	}
	if err := data.Validate(); err != nil {
		return nil, err
	}
	return &data, nil
}

// Validate 验证主题包数据
func (d *ThemePackData) Validate() error {
	if d.SchemaVersion < 1 {
		return ErrThemePackInvalidSchema
	}
	if d.ID == "" {
		return ErrThemePackMissingID
	}
	if d.Name == "" {
		return ErrThemePackMissingName
	}

	if d.Light != nil {
		if err := d.Light.validate(); err != nil {
			return fmt.Errorf("light: %w", err)
		}
	}
	if d.Dark != nil {
		if err := d.Dark.validate(); err != nil {
			return fmt.Errorf("dark: %w", err)
		}
	}

	for _, c := range d.PlayerGradient {
		if !colorRegex.MatchString(c) {
			return fmt.Errorf("%w: %s", ErrThemePackInvalidColor, c)
		}
	}

	if d.CardRadius != nil && (*d.CardRadius < 0 || *d.CardRadius > 100) {
		return ErrThemePackInvalidRadius
	}
	if d.ControlRadius != nil && (*d.ControlRadius < 0 || *d.ControlRadius > 100) {
		return ErrThemePackInvalidRadius
	}
	if d.NavigationRadius != nil && (*d.NavigationRadius < 0 || *d.NavigationRadius > 100) {
		return ErrThemePackInvalidRadius
	}

	return nil
}

func (c *ThemePackColors) validate() error {
	if c.SeedColor == "" {
		return fmt.Errorf("%w: seedColor is required", ErrThemePackInvalidColor)
	}
	if !colorRegex.MatchString(c.SeedColor) {
		return fmt.Errorf("%w: %s", ErrThemePackInvalidColor, c.SeedColor)
	}
	if c.BackgroundColor != "" && !colorRegex.MatchString(c.BackgroundColor) {
		return fmt.Errorf("%w: %s", ErrThemePackInvalidColor, c.BackgroundColor)
	}
	if c.SurfaceColor != "" && !colorRegex.MatchString(c.SurfaceColor) {
		return fmt.Errorf("%w: %s", ErrThemePackInvalidColor, c.SurfaceColor)
	}
	return nil
}

// ThemeCatalogIndex 远程主题目录索引
type ThemeCatalogIndex struct {
	Version int                 `json:"version"`
	Themes  []ThemeCatalogEntry `json:"themes"`
}

// ThemeCatalogEntry 目录中的单个主题条目
type ThemeCatalogEntry struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Author       string `json:"author"`
	Description  string `json:"description"`
	URL          string `json:"url"`
	SHA256       string `json:"sha256"`
	InstallState string `json:"install_state,omitempty"`
}
