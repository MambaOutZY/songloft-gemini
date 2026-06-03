-- +goose Up
INSERT INTO configs (key, value)
VALUES ('plugin_registries', '{"registries":[{"url":"https://raw.githubusercontent.com/songloft-org/songloft-plugin-registry/main/registry.json","name":"Songloft 官方插件","enabled":true}]}')
ON CONFLICT(key) DO NOTHING;

-- +goose Down
DELETE FROM configs WHERE key = 'plugin_registries';
