# 配置与设置 API

<cite>
本文档基于以下源文件编写：

- `internal/handlers/config.go` -- 通用 KV 配置 CRUD 处理器
- `internal/handlers/tab_config_setting.go` -- 底部导航栏 Tab 配置设置端点
- `internal/handlers/log.go` -- 日志等级设置端点
- `internal/handlers/scan.go` -- 音乐路径、自动扫描、扫描标题来源等设置端点
- `internal/handlers/hls.go` -- HLS 代理开关设置端点
- `internal/handlers/jsplugin_registry.go` -- 插件订阅源、HTTP 代理设置端点
- `internal/app/routers.go` -- 配置与设置路由注册
- `internal/models/models.go` -- Config / CreateConfigRequest / UpdateConfigRequest 结构体
</cite>

## 目录

1. [设计概述](#1-设计概述)
2. [通用配置管理 /configs](#2-通用配置管理-configs)
3. [业务设置 /settings](#3-业务设置-settings)
   - [音乐路径](#31-音乐路径)
   - [HLS 代理开关](#32-hls-代理开关)
   - [自动扫描](#33-自动扫描)
   - [扫描标题来源](#34-扫描标题来源)
   - [扫描自动创建歌单](#35-扫描自动创建歌单)
   - [扫描自动创建含子目录](#36-扫描自动创建含子目录)
   - [日志等级](#37-日志等级)
   - [插件订阅源](#38-插件订阅源)
   - [HTTP 代理](#39-http-代理)
   - [底部导航栏 Tab 配置](#310-底部导航栏-tab-配置)

---

## 1. 设计概述

**章节来源**: `AGENTS.md`（配置接口规范）、`internal/app/routers.go`

Songloft 有两类配置接口，用户可见的功能开关一律走业务端点：

| 类型 | 路径风格 | 用途 | 特点 |
|------|----------|------|------|
| **业务设置** | `/settings/<name>` | 用户可见功能开关 | 强类型 JSON、自带默认值、PUT 后触发副作用 |
| **通用 KV** | `/configs/{key}` | admin 编辑器专用 | 纯字符串 KV、PUT 时 key 不存在返回 404、无默认值 |

客户端业务功能一律走 `/settings/*` 端点（`SettingsApi`）；`/configs/{key}` 仅供前端 `config_manager.dart` 通用配置编辑器使用（`ConfigApi`）。两者可读写同一底层 config key（保留双入口），副作用由 `configHandler.SetOnConfigChanged` 回调统一触发。

---

## 2. 通用配置管理 /configs

**章节来源**: `internal/handlers/config.go`

通用 KV 配置端点，供 admin 工具编辑任意配置项。新增业务功能应优先使用 `/settings/*` 端点。

### GET /api/v1/configs

获取配置列表，按 key 升序排列，支持关键词搜索和分页。

- **认证**: Bearer Token
- **查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyword` | string | 否 | 搜索关键词（按 key 匹配） |
| `limit` | int | 否 | 每页数量，默认 20 |
| `offset` | int | 否 | 偏移量，默认 0 |

- **200**: `{"configs": [Config], "total": 5, "limit": 20, "offset": 0}`
- **500**: 服务器错误

### POST /api/v1/configs

创建配置项。

- **认证**: Bearer Token
- **请求体**: `{"key": "music_path", "value": "{\"path\":\"/music\"}"}`（key 和 value 均必填）
- **201**: 返回 `Config` 对象（含 `id`/`key`/`value`/`updated_at`）
- **400**: key 或 value 为空 | **500**: 创建失败

### GET /api/v1/configs/{key}

获取单个配置。

- **认证**: Bearer Token
- **路径参数**: `key`（string）
- **200**: 返回 `Config` 对象
- **404**: 配置不存在

### PUT /api/v1/configs/{key}

更新已有配置。配置必须已存在，否则返回 404。更新后异步触发 `onConfigChanged` 回调（`music_path` 重建 Scanner，`auto_scan` 重启调度）。

- **认证**: Bearer Token
- **路径参数**: `key`（string）
- **请求体**: `{"value": "new_value"}`（value 必填）
- **200**: 返回更新后的 `Config`
- **400**: value 为空 | **404**: key 不存在 | **500**: 更新失败

### DELETE /api/v1/configs/{key}

删除配置项。

- **认证**: Bearer Token
- **路径参数**: `key`（string）
- **200**: `{"message": "配置已删除"}`
- **400**: key 为空 | **500**: 删除失败

---

## 3. 业务设置 /settings

**章节来源**: `internal/handlers/scan.go`、`hls.go`、`log.go`、`jsplugin_registry.go`、`tab_config_setting.go`

所有业务设置端点遵循统一模式：GET 返回当前配置（未配置时返回业务默认值），PUT 写入配置并触发相关副作用。所有端点均需 Bearer Token 认证。

### 3.1 音乐路径

**`GET /api/v1/settings/music-path`** -- 获取音乐路径与扫描排除配置。

**`PUT /api/v1/settings/music-path`** -- 更新配置，`path` 不能为空。

```json
{
  "path": "music",
  "exclude_dirs": ["@eaDir", "tmp"],
  "exclude_paths": []
}
```

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `path` | string | `"music"` | 音乐目录路径 |
| `exclude_dirs` | string[] | `["@eaDir","tmp"]` | 按目录名排除 |
| `exclude_paths` | string[] | `[]` | 按完整路径排除 |

- **PUT 副作用**: 异步触发 `onMusicPathChanged`（重建 Scanner + 清理排除目录歌曲）
- **400**: path 为空 | **500**: 保存失败

### 3.2 HLS 代理开关

**`GET/PUT /api/v1/settings/hls-proxy`** -- `{"enabled": false}`

默认关闭。关闭时电台 `.m3u8` 直接 302 给 player；开启后电台切片字节全部经本机转发，解决源站 Referer/CORS 拦截问题，但所有切片流量走本机带宽。

- **400**: 请求格式错误 | **500**: 保存失败

### 3.3 自动扫描

**`GET/PUT /api/v1/settings/auto-scan`**

```json
{"enabled": false, "interval_seconds": 3600}
```

默认关闭，间隔 3600 秒（1 小时）。PUT 校验 `interval_seconds` 范围 [60, 86400]，更新后立即生效无需重启。

- **PUT 副作用**: 异步触发 `onAutoScanChanged`（重启自动扫描调度器）
- **400**: interval_seconds 超出范围 | **500**: 保存失败

### 3.4 扫描标题来源

**`GET/PUT /api/v1/settings/scan-title-source`** -- `{"title_source": "tag"}`

| 值 | 说明 |
|------|------|
| `tag` | 优先使用音频标签中的标题（默认） |
| `filename` | 始终使用文件名（不含扩展名）作为标题 |

切换后需以"重新导入"模式扫描才能生效。

- **PUT 副作用**: 异步触发 Scanner 重建
- **400**: title_source 不是 `tag` 或 `filename`

### 3.5 扫描自动创建歌单

**`GET/PUT /api/v1/settings/scan-auto-create-playlists`** -- `{"enabled": true}`

默认启用。启用后扫描完成根据音乐目录结构自动创建歌单；关闭则仅入库歌曲不建歌单。

### 3.6 扫描自动创建含子目录

**`GET/PUT /api/v1/settings/scan-auto-create-include-subdirs`** -- `{"enabled": false}`

默认关闭。控制自动创建歌单时是否包含子目录中的歌曲。

### 3.7 日志等级

**`GET/PUT /api/v1/settings/log-level`** -- `{"level": "info"}`

可选值：`debug` / `info` / `warn` / `error`，默认 `info`。PUT 通过共享的 `slog.LevelVar` 即时切换运行时日志等级，同时持久化到 DB，重启后自动恢复。

- **400**: 等级值非法（仅接受上述四个枚举值）
- **500**: 保存失败

### 3.8 插件订阅源

**`GET /api/v1/settings/plugin-registries`** -- 获取订阅源列表（未配置时返回空列表）。

**`PUT /api/v1/settings/plugin-registries`** -- 保存订阅源列表。

```json
{
  "registries": [
    {"url": "https://example.com/registry.json", "name": "官方插件源", "enabled": true}
  ]
}
```

每项包含：`url`（注册表 JSON URL）、`name`（名称）、`enabled`（是否启用）。

- **400**: 请求格式错误 | **500**: 保存失败

### 3.9 HTTP 代理

**`GET/PUT /api/v1/settings/http-proxy`** -- `{"proxy": ""}`

默认空字符串（直连）。设置后所有后端外发 HTTP 请求通过代理转发，包括：插件注册表拉取、插件下载/更新、系统升级检查/下载。支持 HTTP/HTTPS/SOCKS5 协议。loopback 地址自动跳过代理。

- 典型值：`http://192.168.1.1:7890`
- **PUT 副作用**: 调用 `httputil.SetGlobalProxy` 即时更新全局共享 `*http.Transport`
- **400**: 代理地址格式无效 | **500**: 保存失败

### 3.10 底部导航栏 Tab 配置

**`GET /api/v1/settings/tab-config`** -- 获取 Tab 配置（未配置时返回默认值：4 Tab）。

**`PUT /api/v1/settings/tab-config`** -- 保存 Tab 配置。

```json
{
  "show_library": true,
  "show_playlists": true,
  "plugin_tabs": [
    {"plugin_id": 1, "entry_path": "myplugin", "name": "我的插件"}
  ]
}
```

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show_library` | bool | `true` | 是否显示歌曲库 Tab |
| `show_playlists` | bool | `true` | 是否显示歌单 Tab |
| `plugin_tabs` | array | `[]` | 插件 Tab 列表 |

首页和设置固定显示（不在配置中），可选项为歌曲库、歌单和插件 Tab。

**PUT 校验规则**:
- 可选项总数（`show_library` + `show_playlists` 各算 1 + 插件 Tab 数量）不超过 3，加上固定的首页和设置共 5 个
- 每个插件 Tab 的 `entry_path` 和 `name` 不能为空
- `entry_path` 不能重复
- **400**: 校验失败 | **500**: 保存失败
