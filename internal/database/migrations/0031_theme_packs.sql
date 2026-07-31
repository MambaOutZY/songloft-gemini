-- +goose Up

-- 主题包（songloft-org/songloft#337）：声明式 JSON 主题包，用户可自定义配色、圆角、渐变。
-- raw_json 存储完整原始 JSON，前端直接读取渲染；顶层字段做索引/查询用。

-- +goose StatementBegin
CREATE TABLE theme_packs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    theme_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '1.0.0',
    author TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    schema_version INTEGER NOT NULL DEFAULT 1,
    raw_json TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_theme_packs_updated_at
AFTER UPDATE ON theme_packs
FOR EACH ROW
BEGIN
    UPDATE theme_packs SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_theme_packs_updated_at;
DROP TABLE IF EXISTS theme_packs;
