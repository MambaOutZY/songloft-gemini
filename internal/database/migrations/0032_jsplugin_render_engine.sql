-- +goose Up
-- 插件页渲染引擎声明（songloft-org/songloft#341）：由插件在自己的 plugin.json 里用
-- renderEngine 字段声明用哪个引擎渲染（"webview" / "webf"），取代客户端的全局开关。
-- 空串是合法值，语义 = 「跟随宿主默认」（当前默认 webview）。刻意不把已有行回填成
-- 'webview'，这样将来改宿主默认值不必再迁移一次。
ALTER TABLE js_plugins ADD COLUMN render_engine TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE js_plugins DROP COLUMN render_engine;
