-- name: GetPluginStorage :one
SELECT value FROM plugin_storage WHERE plugin_entry_path = ? AND key = ?;

-- name: SetPluginStorage :exec
INSERT INTO plugin_storage (plugin_entry_path, key, value) VALUES (?, ?, ?)
ON CONFLICT(plugin_entry_path, key) DO UPDATE SET value = excluded.value;

-- name: DeletePluginStorage :execrows
DELETE FROM plugin_storage WHERE plugin_entry_path = ? AND key = ?;

-- name: ListPluginStorageKeys :many
SELECT key FROM plugin_storage WHERE plugin_entry_path = ?;

-- name: DeleteAllPluginStorage :exec
DELETE FROM plugin_storage WHERE plugin_entry_path = ?;

-- name: GetPluginStorageTotalSize :one
SELECT CAST(COALESCE(SUM(LENGTH(value)), 0) AS INTEGER) AS total_size FROM plugin_storage WHERE plugin_entry_path = ?;

-- name: ListPluginStorageEntryPaths :many
SELECT DISTINCT plugin_entry_path FROM plugin_storage;
