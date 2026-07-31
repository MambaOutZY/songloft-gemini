-- name: CreateThemePack :exec
INSERT INTO theme_packs (theme_id, name, version, author, description, schema_version, raw_json)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetThemePackByThemeID :one
SELECT id, theme_id, name, version, author, description, schema_version, raw_json, created_at, updated_at
FROM theme_packs WHERE theme_id = ?;

-- name: ListThemePacks :many
SELECT id, theme_id, name, version, author, description, schema_version, raw_json, created_at, updated_at
FROM theme_packs ORDER BY created_at DESC;

-- name: DeleteThemePack :execrows
DELETE FROM theme_packs WHERE theme_id = ?;

-- name: UpdateThemePack :exec
UPDATE theme_packs
SET name = ?, version = ?, author = ?, description = ?, schema_version = ?, raw_json = ?
WHERE theme_id = ?;
