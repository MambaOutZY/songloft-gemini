package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"songloft/internal/database/sqlc"
)

type PluginStorageRepository struct {
	db      sqlc.DBTX
	queries *sqlc.Queries
}

func NewPluginStorageRepository(db sqlc.DBTX) *PluginStorageRepository {
	return &PluginStorageRepository{db: db, queries: sqlc.New(db)}
}

func (r *PluginStorageRepository) Get(ctx context.Context, entryPath, key string) (string, error) {
	value, err := r.queries.GetPluginStorage(ctx, sqlc.GetPluginStorageParams{
		PluginEntryPath: entryPath,
		Key:             key,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("get plugin storage: %w", err)
	}
	return value, nil
}

func (r *PluginStorageRepository) Set(ctx context.Context, entryPath, key, value string) error {
	if err := r.queries.SetPluginStorage(ctx, sqlc.SetPluginStorageParams{
		PluginEntryPath: entryPath,
		Key:             key,
		Value:           value,
	}); err != nil {
		return fmt.Errorf("set plugin storage: %w", err)
	}
	return nil
}

func (r *PluginStorageRepository) Delete(ctx context.Context, entryPath, key string) error {
	rows, err := r.queries.DeletePluginStorage(ctx, sqlc.DeletePluginStorageParams{
		PluginEntryPath: entryPath,
		Key:             key,
	})
	if err != nil {
		return fmt.Errorf("delete plugin storage: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PluginStorageRepository) Keys(ctx context.Context, entryPath string) ([]string, error) {
	keys, err := r.queries.ListPluginStorageKeys(ctx, entryPath)
	if err != nil {
		return nil, fmt.Errorf("list plugin storage keys: %w", err)
	}
	return keys, nil
}

func (r *PluginStorageRepository) DeleteAll(ctx context.Context, entryPath string) error {
	if err := r.queries.DeleteAllPluginStorage(ctx, entryPath); err != nil {
		return fmt.Errorf("delete all plugin storage: %w", err)
	}
	return nil
}

func (r *PluginStorageRepository) TotalSize(ctx context.Context, entryPath string) (int64, error) {
	size, err := r.queries.GetPluginStorageTotalSize(ctx, entryPath)
	if err != nil {
		return 0, fmt.Errorf("get plugin storage total size: %w", err)
	}
	return size, nil
}

func (r *PluginStorageRepository) ListEntryPaths(ctx context.Context) ([]string, error) {
	paths, err := r.queries.ListPluginStorageEntryPaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("list plugin storage entry paths: %w", err)
	}
	return paths, nil
}
