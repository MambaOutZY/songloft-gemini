-- +goose Up

-- +goose StatementBegin
CREATE TABLE plugin_storage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_entry_path TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_entry_path, key)
);
-- +goose StatementEnd

CREATE INDEX idx_plugin_storage_entry_path ON plugin_storage(plugin_entry_path);

-- +goose StatementBegin
CREATE TRIGGER update_plugin_storage_updated_at
AFTER UPDATE ON plugin_storage
FOR EACH ROW
BEGIN
    UPDATE plugin_storage SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS update_plugin_storage_updated_at;
DROP INDEX IF EXISTS idx_plugin_storage_entry_path;
DROP TABLE IF EXISTS plugin_storage;
