package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"songloft/internal/database/sqlc"
)

// ThemePackRow 主题包数据行
type ThemePackRow struct {
	ID            int64
	ThemeID       string
	Name          string
	Version       string
	Author        string
	Description   string
	SchemaVersion int
	RawJSON       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ThemePackRepository 主题包仓储
type ThemePackRepository struct {
	db      sqlc.DBTX
	queries *sqlc.Queries
}

// NewThemePackRepository 创建仓储实例
func NewThemePackRepository(db sqlc.DBTX) *ThemePackRepository {
	return &ThemePackRepository{db: db, queries: sqlc.New(db)}
}

// Create 创建主题包，theme_id 冲突时返回 ErrConflict
func (r *ThemePackRepository) Create(ctx context.Context, themeID, name, version, author, description string, schemaVersion int, rawJSON string) error {
	err := r.queries.CreateThemePack(ctx, sqlc.CreateThemePackParams{
		ThemeID:       themeID,
		Name:          name,
		Version:       version,
		Author:        author,
		Description:   description,
		SchemaVersion: int64(schemaVersion),
		RawJson:       rawJSON,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrConflict
		}
		return fmt.Errorf("create theme pack: %w", err)
	}
	return nil
}

// GetByThemeID 按 theme_id 获取
func (r *ThemePackRepository) GetByThemeID(ctx context.Context, themeID string) (*ThemePackRow, error) {
	row, err := r.queries.GetThemePackByThemeID(ctx, themeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get theme pack: %w", err)
	}
	return r.toRow(row), nil
}

// List 列出所有主题包
func (r *ThemePackRepository) List(ctx context.Context) ([]ThemePackRow, error) {
	rows, err := r.queries.ListThemePacks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list theme packs: %w", err)
	}
	out := make([]ThemePackRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, *r.toRow(row))
	}
	return out, nil
}

// Delete 删除主题包，未命中时返回 ErrNotFound
func (r *ThemePackRepository) Delete(ctx context.Context, themeID string) error {
	rows, err := r.queries.DeleteThemePack(ctx, themeID)
	if err != nil {
		return fmt.Errorf("delete theme pack: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Update 更新主题包
func (r *ThemePackRepository) Update(ctx context.Context, themeID, name, version, author, description string, schemaVersion int, rawJSON string) error {
	err := r.queries.UpdateThemePack(ctx, sqlc.UpdateThemePackParams{
		ThemeID:       themeID,
		Name:          name,
		Version:       version,
		Author:        author,
		Description:   description,
		SchemaVersion: int64(schemaVersion),
		RawJson:       rawJSON,
	})
	if err != nil {
		return fmt.Errorf("update theme pack: %w", err)
	}
	return nil
}

func (r *ThemePackRepository) toRow(row sqlc.ThemePack) *ThemePackRow {
	return &ThemePackRow{
		ID:            row.ID,
		ThemeID:       row.ThemeID,
		Name:          row.Name,
		Version:       row.Version,
		Author:        row.Author,
		Description:   row.Description,
		SchemaVersion: int(row.SchemaVersion),
		RawJSON:       row.RawJson,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}
