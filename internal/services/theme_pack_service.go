package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"songloft/internal/database"
	"songloft/internal/models"
)

const activeThemePackKey = "active_theme_pack"

// ThemePackService 主题包服务
type ThemePackService struct {
	repo          *database.ThemePackRepository
	configService *ConfigService
}

// NewThemePackService 创建主题包服务实例
func NewThemePackService(repo *database.ThemePackRepository, configService *ConfigService) *ThemePackService {
	return &ThemePackService{
		repo:          repo,
		configService: configService,
	}
}

// ThemePackResponse 主题包响应（含完整 JSON 数据）
type ThemePackResponse struct {
	ID            int64                `json:"id"`
	ThemeID       string               `json:"theme_id"`
	Name          string               `json:"name"`
	Version       string               `json:"version"`
	Author        string               `json:"author"`
	Description   string               `json:"description"`
	SchemaVersion int                  `json:"schema_version"`
	Data          *models.ThemePackData `json:"data"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
}

// ThemePackListItem 列表项（不含完整 data，减少传输量）
type ThemePackListItem struct {
	ID            int64  `json:"id"`
	ThemeID       string `json:"theme_id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Author        string `json:"author"`
	Description   string `json:"description"`
	SchemaVersion int    `json:"schema_version"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// Import 导入主题包：解析验证 JSON，存入数据库。已存在相同 ID 时更新。
func (s *ThemePackService) Import(ctx context.Context, rawJSON []byte) (*ThemePackResponse, error) {
	data, err := models.ParseThemePackData(rawJSON)
	if err != nil {
		return nil, err
	}

	// 规范化存储的 JSON（确保格式统一）
	normalized, _ := json.Marshal(data)
	rawStr := string(normalized)

	// 尝试创建，冲突则更新
	err = s.repo.Create(ctx, data.ID, data.Name, data.Version, data.Author, data.Description, data.SchemaVersion, rawStr)
	if errors.Is(err, database.ErrConflict) {
		err = s.repo.Update(ctx, data.ID, data.Name, data.Version, data.Author, data.Description, data.SchemaVersion, rawStr)
		if err != nil {
			return nil, fmt.Errorf("update theme pack: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("import theme pack: %w", err)
	}

	// 重新读取以获取完整时间戳
	return s.Get(ctx, data.ID)
}

// List 列出所有主题包
func (s *ThemePackService) List(ctx context.Context) ([]ThemePackListItem, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list theme packs: %w", err)
	}
	items := make([]ThemePackListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ThemePackListItem{
			ID:            row.ID,
			ThemeID:       row.ThemeID,
			Name:          row.Name,
			Version:       row.Version,
			Author:        row.Author,
			Description:   row.Description,
			SchemaVersion: row.SchemaVersion,
			CreatedAt:     row.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     row.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return items, nil
}

// Get 获取单个主题包详情
func (s *ThemePackService) Get(ctx context.Context, themeID string) (*ThemePackResponse, error) {
	row, err := s.repo.GetByThemeID(ctx, themeID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, models.ErrThemePackNotFound
		}
		return nil, fmt.Errorf("get theme pack: %w", err)
	}

	var data models.ThemePackData
	_ = json.Unmarshal([]byte(row.RawJSON), &data)

	return &ThemePackResponse{
		ID:            row.ID,
		ThemeID:       row.ThemeID,
		Name:          row.Name,
		Version:       row.Version,
		Author:        row.Author,
		Description:   row.Description,
		SchemaVersion: row.SchemaVersion,
		Data:          &data,
		CreatedAt:     row.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     row.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// Delete 删除主题包。若为当前激活主题则同时清除激活状态。
func (s *ThemePackService) Delete(ctx context.Context, themeID string) error {
	// 如果是当前激活主题，先清除
	active := s.GetActiveID()
	if active == themeID {
		s.ClearActive()
	}

	err := s.repo.Delete(ctx, themeID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return models.ErrThemePackNotFound
		}
		return fmt.Errorf("delete theme pack: %w", err)
	}
	return nil
}

// SetActive 设置激活的主题包
func (s *ThemePackService) SetActive(ctx context.Context, themeID string) error {
	// 验证主题包存在
	_, err := s.repo.GetByThemeID(ctx, themeID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return models.ErrThemePackNotFound
		}
		return fmt.Errorf("set active theme pack: %w", err)
	}
	if err := s.configService.Set(activeThemePackKey, themeID); err != nil {
		return fmt.Errorf("save active theme pack: %w", err)
	}
	return nil
}

// GetActiveID 获取当前激活的主题包 ID（空字符串表示默认主题）
func (s *ThemePackService) GetActiveID() string {
	return s.configService.GetString(activeThemePackKey, "")
}

// GetActive 获取当前激活的主题包详情（nil 表示默认主题）
func (s *ThemePackService) GetActive(ctx context.Context) (*ThemePackResponse, error) {
	activeID := s.GetActiveID()
	if activeID == "" {
		return nil, nil
	}
	resp, err := s.Get(ctx, activeID)
	if err != nil {
		// 激活的主题包不存在（被外部删除），自动清除
		if errors.Is(err, models.ErrThemePackNotFound) {
			s.ClearActive()
			return nil, nil
		}
		return nil, err
	}
	return resp, nil
}

// ClearActive 清除激活状态（恢复默认主题）
func (s *ThemePackService) ClearActive() {
	_ = s.configService.Set(activeThemePackKey, "")
}
