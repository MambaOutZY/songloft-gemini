# JS 插件开发指南

本文档详细介绍 Songloft JS 插件系统的架构、API 和开发流程。

---

## 1. 概述

Songloft JS 插件系统允许开发者使用 JavaScript 扩展音乐服务器功能，无需编译 Go 代码。

### 设计理念

系统基于 **Skynet Actor 模型**设计：

- 每个插件是一个独立的 **Actor（JSService）**，拥有自己的 JS 虚拟机
- 插件之间通过 **消息** 通信，互不干扰
- 所有消息由 **ServiceScheduler** 统一调度，保证串行处理
- 双层 SHA256 校验确保插件代码完整性

### 核心特性

| 特性 | 说明 |
|------|------|
| 沙箱隔离 | 每个插件运行在独立的 QuickJS 虚拟机中 |
| 权限控制 | 细粒度权限声明，按需授权 |
| 热更新 | 运行时更新插件，无需重启服务 |
| 插件间通信 | send/call 消息机制 |
| 静态资源 | 内置 Web UI 托管 |
| 健康检查 | 自动检测异常插件并处理 |

### 架构示意

```
Manager（管理器）
  ├── PackageManager（包管理：安装/更新/卸载）
  ├── ServiceScheduler（消息调度器）
  │   ├── JSService[plugin-a]（Actor + QuickJS VM）
  │   ├── JSService[plugin-b]（Actor + QuickJS VM）
  │   └── ...
  ├── HotReloader（热更新监控）
  └── HealthChecker（健康检查）
```

---

## 2. 快速开始

推荐使用官方工具链 [songloft-plugin-toolchain](https://github.com/songloft-org/plugin-toolchain)，5 分钟创建、构建并上传你的第一个 JS 插件。

### Step 1: 用脚手架创建项目

```bash
npx create-songloft-plugin@latest
# 或 pnpm create songloft-plugin
cd <你的插件目录>
npm install   # 或 pnpm install / yarn install
```

脚手架会交互式引导你完成以下配置：

1. **基本信息** — 目录名、插件显示名称、entryPath、简介、作者
2. **权限选择**（多选） — `storage`、`persistent-storage`、`songs.read`、`songs.write`、`playlists.read`、`playlists.write`、`inter-plugin`、`command`、`jsenv`、`fs`、`fs:music`、`fs:external`、`websocket`、`net`
3. **附加功能模板**（多选，可跳过） — 静态页面 (`static/`)、可执行文件管理 (`bin/`)
4. **包管理器** — npm / pnpm / yarn

生成的项目结构（选择全部附加功能时）：

```
my-plugin/
├── plugin.json        # 插件清单（entryHash / zipHash 由 builder 生成）
├── package.json       # npm 依赖（@songloft/plugin-sdk / @songloft/plugin-builder）
├── tsconfig.json
├── src/
│   └── main.ts        # TypeScript 源码入口
├── static/            # [附加功能] 静态资源（HTML + 插件自定义 JS）
│   ├── index.html
│   └── js/
│       └── app.js
└── bin/               # [附加功能] 可执行文件管理（打包/下载/运行外部程序）
```

模板采用叠加层设计：始终包含基础模板，选中的附加功能会额外合并对应文件。

### Step 2: 编写业务逻辑

`src/main.ts` 使用 `@songloft/plugin-sdk` 提供的全局类型与 helper：

```typescript
/// <reference types="@songloft/plugin-sdk" />
import { jsonResponse, createRouter } from '@songloft/plugin-sdk';

const router = createRouter();

router.get('/hello', (req) => jsonResponse({ message: 'Hello!', query: req.query }));

router.get('/songs', async (req) => {
  const songs = await songloft.songs.list({ limit: 10 });
  return jsonResponse({ count: songs.length, songs });
});

function onInit(): void { songloft.log.info('my-plugin initialized'); }
function onDeinit(): void { songloft.log.info('my-plugin deinitialized'); }
function onHTTPRequest(req: HTTPRequest): HTTPResponse { return router.handle(req); }

// @ts-expect-error — QuickJS 全局注入
globalThis.onInit = onInit;
// @ts-expect-error
globalThis.onDeinit = onDeinit;
// @ts-expect-error
globalThis.onHTTPRequest = onHTTPRequest;
```

### Step 3: 启动开发模式（推荐）

```bash
pnpm run dev          # 等价于 songloft-plugin dev
```

首次运行会交互式询问 Songloft 实例地址、用户名与密码，之后：

1. 把账号密码写入项目根目录的 `.songloft-dev.json`（builder 会自动把它追加到 `.gitignore`），后续运行直接静默登录；
2. 立即执行一次构建并上传，首次安装时自动启用插件；
3. 监听 `src/`、`static/`、`plugin.json`，源码变更时自动重建上传，已激活的插件会被后端自动热重载。

> Token 不缓存：每次会话用账号密码即时登录，因此无需关心 token 过期 / 刷新。要换帐号或改密码，编辑（或直接删除）`.songloft-dev.json` 即可。

控制台会打印插件的访问入口（例如 `http://localhost:58091/api/v1/jsplugin/<entryPath>/`），按 `Ctrl+C` 退出。

> 开发模式的详细 CLI 选项、环境变量与配置文件字段见下文 [开发模式详解](#开发模式详解-songloft-plugin-dev)。

### Step 4: 构建生产包

发布前生成可分发的 `.jsplugin.zip`：

```bash
pnpm run build        # 等价于 songloft-plugin build
```

builder 会：

1. 用 esbuild 把 `src/main.ts` 打包为 `build/main.js`（`format: iife`, `target: es2020`，禁止引用 Node 内置模块）；
2. 拷贝 `static/` 到 `build/`，并对 JS/CSS/字体/图片注入内容 hash（可在 `plugin.json` 中设置 `"staticHash": false` 关闭）；
3. 若检测到可用的 `jsc` 工具，将 `main.js` 进一步编译为 `main.jsc` 字节码；
4. 计算 `entryHash = sha256(main 文件)` 与 `zipHash`（规范化算法，排除 `plugin.json` 自身），写回 `build/plugin.json`；
5. 打包为 `dist/<entryPath>.jsplugin.zip`，并生成 `dist/<entryPath>.json` 远程更新元数据。

### Step 5: 安装到目标实例

任选其一：

- **开发模式自动上传** —— `pnpm run dev`（见 Step 3），适合本地迭代；
- **设置页面上传** —— 在 Songloft 客户端的插件管理页选择 `dist/<entryPath>.jsplugin.zip`；
- **目录放置** —— 把 zip 放进服务器的 `data/jsplugins/` 目录，下次启动时自动扫描；
- **API 上传** —— `POST /api/v1/jsplugins/upload`，multipart 字段名 `file`（开发模式底层即此接口）。

安装后，插件的 HTTP API 通过 `/api/v1/jsplugin/<entryPath>/` 访问，静态资源通过 `/api/v1/jsplugin/<entryPath>/static/...` 访问。

### 开发模式详解 (songloft-plugin dev)

`songloft-plugin dev` 把"构建 → 上传 → 热重载"压缩成一个常驻命令，适合本地开发与远程实例联调。

#### 默认行为

| 阶段 | 行为 |
|------|------|
| 启动 | 读取 `.songloft-dev.json`，缺失 `username` / `password` 时交互式询问，登录成功后落地保存 |
| 登录策略 | 不缓存 token；每次启动用账号密码即时登录，会话期间出现 `401` 时自动用同一密码重登 |
| 首次上传 | 调用 `POST /api/v1/jsplugins/upload`，新装后自动调用 `enable` |
| 后续上传 | 同一 `entryPath` 复用 upload 接口，由后端识别为覆盖更新；插件处于活跃状态时自动热重载 |
| 文件监听 | 监听 `src/`、`static/`、`plugin.json`，250ms debounce 触发增量构建 |
| 密码失效 | 若服务器拒绝缓存的密码（如已被修改），自动清除 `.songloft-dev.json` 中的 `password` 字段并提示重新运行 |

#### CLI 选项

```text
songloft-plugin dev [options]

--host <url>        Songloft 实例 URL（默认 http://localhost:58091，
                    亦可读 $MIMUSIC_HOST 或 .songloft-dev.json）
--username <name>   登录用户名（或 $MIMUSIC_USER）
--password <pwd>    登录密码（或 $MIMUSIC_PASSWORD；缺省时静默提示输入）
--token <jwt>       直接使用预签发的 access token（或 $MIMUSIC_TOKEN）
--once              构建+上传一次后退出，跳过 watch
--no-enable         首次安装后不自动启用插件
```

#### 环境变量

| 变量 | 等价选项 |
|------|----------|
| `MIMUSIC_HOST` | `--host` |
| `MIMUSIC_USER` | `--username` |
| `MIMUSIC_PASSWORD` | `--password` |
| `MIMUSIC_TOKEN` | `--token` |

#### `.songloft-dev.json` 字段

dev 命令自动在项目根目录维护下面的配置文件（同时把它追加到 `.gitignore`）：

```json
{
  "host": "http://localhost:58091",
  "username": "admin",
  "password": "your-password",
  "pluginId": 12,
  "entryPath": "my-plugin"
}
```

| 字段 | 写入时机 | 说明 |
|------|----------|------|
| `host` | 首次启动 | Songloft 实例 URL |
| `username` / `password` | 首次启动交互输入后写入，亦可手填 | 用于每次会话登录；明文存储，**切勿提交** |
| `pluginId` / `entryPath` | 首次上传后写入 | 仅供参考，dev 命令实际通过 `entryPath` 与后端对账 |

> 不存在 `accessToken` / `refreshToken` 字段：dev 命令不缓存 token。
>
> 不想让密码明文落地？改用 `--token <jwt>` 或 `$MIMUSIC_TOKEN` 提供预签发的 access token；token 模式下不会读写 `.songloft-dev.json` 中的凭据字段。
>
> 删除整个文件等同于重置登录状态。

---

## 3. 插件结构

### ZIP 打包格式

插件以 `.jsplugin.zip` 格式分发，文件名规则：`{entryPath}.jsplugin.zip`

ZIP 内部结构（所有文件在根级别，不含父目录）：

```
plugin.json          # 插件清单（必须）
main.js              # 入口文件（必须，或 main.jsc 字节码）
static/              # 静态资源目录（可选）
  ├── index.html
  └── js/
      └── app.js
```

> 公共资源（CSS 变量/reset/MD3 组件样式、字体、API 工具库）由主程序自动注入，插件无需打包。详见 [§8. 静态资源](#8-静态资源)。

### plugin.json 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 插件名称（2-50 字符） |
| `version` | string | 是 | 语义化版本号（如 `1.0.0`） |
| `description` | string | 否 | 插件描述 |
| `author` | string | 否 | 作者 |
| `homepage` | string | 否 | 主页 URL |
| `license` | string | 否 | 许可证 |
| `entryPath` | string | 是 | 路由前缀（小写字母+数字+连字符，如 `my-plugin`） |
| `main` | string | 是 | 入口文件路径（必须以 `.js` 结尾） |
| `minHostVersion` | string | 否 | 最低宿主版本要求 |
| `permissions` | string[] | 是 | 权限列表（可为空数组 `[]`） |
| `renderEngine` | string | 否 | 客户端渲染插件页用的引擎，`webview` / `webf`，缺失或空串等同 `webview`。详见 [renderEngine 渲染引擎声明](#renderengine-渲染引擎声明) |
| `updateUrl` | string | 否 | 远程更新检查 URL |
| `download_url` | string | 否 | 插件下载 URL |
| `entryHash` | string | 是 | `sha256(main.js)` 64 位小写 hex，由 `@songloft/plugin-builder` 自动生成，请勿手动编辑 |
| `zipHash` | string | 是 | zip 内除 `plugin.json` 外所有文件的规范化 sha256 64 位小写 hex，由 `@songloft/plugin-builder` 自动生成，请勿手动编辑 |

> `entryHash` / `zipHash` 为强制校验字段，缺失或与实际内容不匹配时，安装与加载均会被后端拒绝。`zipHash` 计算范围**不含** `plugin.json` 自身，避免 hash 写回 `plugin.json` 引起的循环依赖。

### entryPath 命名规则

- 仅允许小写字母、数字和连字符
- 必须以小写字母开头
- 正则：`^[a-z][a-z0-9-]*$`
- 示例：`example-basic`、`music-sync`、`metadata-helper`

### renderEngine 渲染引擎声明

原生客户端渲染插件页有两条路径：系统 WebView，和 [WebF](https://openwebf.com/)（纯 Flutter 渲染的 W3C 运行时）。用哪条**由插件自己在 `plugin.json` 里声明**——宿主侧**没有**全局引擎开关，插件之间互不影响。

```json
{
  "entryPath": "my-plugin",
  "renderEngine": "webf"
}
```

| 取值 | 含义 |
|------|------|
| 字段缺失 / `""` | 等同 `webview`，即宿主默认 |
| `"webview"` | 系统 WebView 渲染（默认） |
| `"webf"` | WebF 渲染 |

- **其它取值一律非法**：后端 `ValidateManifest` 阶段直接报错，插件**装不上**（不会静默回退到 `webview`）
- 插件列表 API 以 snake_case 的 `render_engine` 字段返回该值
- 该字段可随版本改：不想再用 WebF 就发一个把它改回 `webview`（或删掉）的新版本

#### 什么时候该声明 `webf`

**只有插件页在 WebF 下已经实测可用时才声明。** WebF 不是浏览器，有一批 HTML/CSS 能力缺失（内建元素、`env()`、`window.open`、`URL.createObjectURL` 等），能力边界、已垫掉的缺口、以及可用的原生元素见 [§8 · WebF 渲染引擎与原生元素](#webf-渲染引擎与原生元素)——那一节是判断「我的页面能不能上 WebF」的依据，**不要**只按浏览器里的表现下结论。

声明之后作者要承担的事：

- **自己验证**每个页面在 WebF 下的渲染与交互，包括表格、滑块、文件选择、外链跳转这类容易静默降级的控件
- 需要按引擎分叉时用 `html.webf-engine` class（主程序自动加），见 [§8 · WebF 渲染引擎与原生元素](#webf-渲染引擎与原生元素)
- WebF 目前是 **0.x beta**。宿主**不保留任何全局回退开关**：页面在 WebF 下出问题时，用户能做的只是**禁用这个插件**，或等你发一个改回 `webview` 的版本。把这条当成声明 `webf` 的成本

#### 平台限制（声明前必看）

- **Web 端（浏览器里的 Songloft Web）完全不受该字段影响**：WebF 不支持 Flutter Web，Web 端**永远**走 iframe 路径。声明 `webf` 不会改变 Web 端的任何行为
- **Linux 端覆盖面很窄**：WebF Linux 仅支持 x86-64 且 glibc ≥ 2.38，**没有 arm64**——NAS、Debian 12、树莓派等常见环境都在覆盖之外，拿不到 WebF 渲染面
- 因此**插件页必须在系统 WebView / 普通浏览器里同样可用**：`webf` 是「在支持的平台上换一个更好的渲染面」，不是「只为 WebF 写页面」的许可

---

## 4. 生命周期

插件有三个核心生命周期回调函数：

### onInit()

插件加载完成后调用。用于初始化资源、设置定时器等。

```javascript
async function onInit() {
    songloft.log.info("Plugin initialized");
    await songloft.storage.set("start_time", new Date().toISOString());
}
```

**注意**：`onInit()` 失败不会阻止插件运行，插件仍可响应 HTTP 请求。

### onDeinit()

插件卸载前调用。用于清理资源、保存状态。

```javascript
function onDeinit() {
    songloft.log.info("Plugin shutting down, saving state...");
}
```

### onHTTPRequest(req)

收到 HTTP 请求时调用。这是插件对外提供服务的主要入口。

**参数 `req` 结构：**

```javascript
{
    method: "GET",           // HTTP 方法
    path: "/songs",          // 请求路径（相对于插件的 entryPath）
    headers: {},             // 请求头 map
    body: "",                // 请求体（POST/PUT 时）
    query: "limit=10&offset=0"  // URL 查询字符串
}
```

**返回值结构：**

```javascript
{
    statusCode: 200,          // HTTP 状态码
    headers: {                // 响应头
        "Content-Type": "application/json"
    },
    body: "..."               // 响应体（字符串）
}
```

**示例：路由分发**

```javascript
function onHTTPRequest(req) {
    switch (req.path) {
        case "/":
        case "":
            return { statusCode: 200, body: "Hello!", headers: {} };
        case "/api/data":
            if (req.method === "POST") {
                return handlePost(req);
            }
            return handleGet(req);
        default:
            return { statusCode: 404, body: "Not Found", headers: {} };
    }
}
```

### onWebSocket(req, socket)

客户端连接 `/api/v1/jsplugin/{entryPath}/...` 并发起 WebSocket upgrade 时调用。插件必须声明 `websocket` 权限。`onWebSocket` 应注册消息/关闭/错误回调后返回，连接生命周期由宿主托管。

**参数 `req` 结构：**

```javascript
{
    method: "GET",
    path: "/api/inbound",
    headers: {},
    query: "access_token=...",
    remoteAddr: "127.0.0.1:12345"
}
```

**`socket` 常用方法：**

- `socket.send(string | Uint8Array | ArrayBuffer)`：发送文本或二进制消息
- `socket.close(code?, reason?)`：关闭连接
- `socket.onMessage(fn)` / `socket.onClose(fn)` / `socket.onError(fn)`：注册事件回调
- `socket.onmessage = fn` / `socket.addEventListener(...)`：兼容浏览器 WebSocket 风格

**示例：Echo 服务**

```javascript
globalThis.onWebSocket = async function(req, socket) {
    socket.onMessage(async function(event) {
        await socket.send(event.data);
    });
};
```

---

## 5. API 参考

所有 API 通过全局 `songloft` 对象访问。

> **重要：所有 `songloft.*` 方法均为异步、返回 Promise，必须在 `async` 函数中 `await`。** 这与 `fetch` 等 Web 标准 API 行为一致。下文示例均置于 `async` 函数上下文中。**例外：** `songloft.log.*`（同步本地日志）和 `songloft.comm.onMessage(...)`（同步注册回调）无需 `await`。

### HTTP 请求（全局 fetch）

使用标准全局 `fetch` 函数发起 HTTP 请求（由运行时 polyfill 提供，返回 Promise）。**无需声明权限**。

```javascript
// GET
const resp = await fetch("https://example.com/api");
const data = await resp.json();

// POST
const postResp = await fetch("https://example.com/api", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ hello: "world" })
});
const text = await postResp.text();
```

请求头里可以使用两个运行时内部控制头：`X-Fetch-No-Redirect` 禁止自动跟随重定向，`X-Fetch-Timeout-Ms` 设置单次请求超时（100-30000ms）。这两个头只影响运行时行为，不会转发给目标服务器。

**`Response` 对象字段：**
- `ok` — `status >= 200 && status < 300`
- `status` — HTTP 状态码
- `statusText` — 状态文本
- `headers` — 响应头对象
- `json()` — 返回 `Promise<unknown>`，解析 JSON
- `text()` — 返回 `Promise<string>`，原始文本

`onHTTPRequest`、`onWebSocket` 和事件回调都可以是 `async function`，框架会等待 Promise settle。

### Crypto（全局 crypto）

运行时提供轻量 `crypto` 工具对象。**无需声明权限**。

```javascript
const md5 = crypto.md5("data");
const sha256 = crypto.sha256Bytes(Buffer.from("data", "utf8")).toString("hex");

const key = Buffer.from("1234567890abcdef", "utf8");
const iv = Buffer.from("abcdef1234567890", "utf8");
const encrypted = crypto.aesEncrypt("hello", "cbc", key, iv).toString("base64");
const decrypted = crypto.aesDecrypt(encrypted, "cbc", key, iv).toString("utf8");
```

常用方法：`md5(str)`、`sha1(str)`、`sha256Bytes(buffer)`、`rc4(key, data)`、`aesEncrypt(buffer, "cbc" | "ecb", key, iv?)`、`aesDecrypt(buffer, "cbc" | "ecb", key, iv?)`、`rsaEncrypt(buffer, publicKeyPEM)`、`randomBytes(size)`。AES 使用 PKCS7 padding；`aesDecrypt` 的字符串密文默认按 base64 解析，传入 `Buffer` 时按原始字节解析。

### 定时器（全局 setTimeout / setInterval）

使用标准全局定时器 API（由运行时 polyfill 提供）。**无需声明权限**，插件卸载时运行时会自动清理未清除的定时器。

```javascript
// 一次性延迟
const t = setTimeout(() => songloft.log.info("tick"), 1000);
clearTimeout(t);

// 周期执行
const i = setInterval(() => songloft.log.info("heartbeat"), 60000);
clearInterval(i);
```

**注意：** 定时器回调在独立的后台 goroutine 中执行（每 500ms 检查一次到期定时器），使用 TryLock 机制确保**不阻塞 HTTP 请求处理**。当 HTTP 请求正在处理时，定时器自动让步等待下一轮。`setInterval` 的最小间隔被限制为 10ms。

### songloft.storage — 持久化存储

需要权限：`storage`

```javascript
async function storageExample() {
    // 读取值（异步返回原类型值或 null）
    var value = await songloft.storage.get("key");

    // 写入值（值经 JSON 自动序列化，可直接存对象/数组）
    await songloft.storage.set("config", { volume: 80, list: [1, 2, 3] });

    // 删除键
    await songloft.storage.delete("key");

    // 获取所有键名
    var keys = await songloft.storage.keys();  // ["key1", "key2", ...]
}
```

**存储限制：**
- 键名为字符串
- 值经 JSON 自动序列化，可直接存对象/数组/数字；`get` 异步返回原类型值或 null
- 每个插件有独立的存储空间

### songloft.songs — 歌曲操作

需要权限：`songs.read`

```javascript
async function songsExample() {
    // 获取歌曲列表
    var songs = await songloft.songs.list({ limit: 20, offset: 0 });

    // 根据 ID 获取歌曲
    var song = await songloft.songs.getById(123);

    // 搜索歌曲
    var results = await songloft.songs.search("关键词");
}
```

**Song 对象结构：**
```javascript
{
    id: 1,
    type: "local",        // "local" | "remote" | "radio"
    title: "歌曲名",
    artist: "艺术家",
    album: "专辑名",
    duration: 240.5,      // 秒
    file_path: "/path/to/file.mp3",
    url: "",
    cover_url: "",        // 封面 URL（CoverPath 内部字段不会序列化输出）
    is_video: false       // 是否为视频容器
}
```

### songloft.playlists — 歌单操作

需要权限：`playlists.read`（读取）或 `playlists.write`（修改）；或者通配符糖 `playlists.*`。

```javascript
async function playlistsExample() {
    // 需要 playlists.read
    var playlists = await songloft.playlists.list();
    var playlist = await songloft.playlists.getById(1);
    var songs = await songloft.playlists.getSongs(1, { limit: 50, offset: 0 });
}
```

### songloft.comm — 插件间通信

需要权限：`inter-plugin`

```javascript
async function commExample() {
    // 发送消息（fire-and-forget）
    await songloft.comm.send("target-plugin", "action-name", { data: "hello" });

    // 请求-响应调用（等待响应，超时默认 10s）
    var resp = await songloft.comm.call("target-plugin", "action-name", { data: "hello" }, 5000);
    // resp = { success: true, data: { ... } }
}

// 注册消息处理器（同步注册，无需 await）
songloft.comm.onMessage("action-name", function(payload, from) {
    // payload: 发送方传递的数据
    // from: 发送方的 entryPath
    return { result: "processed" };  // 返回值作为 call 的响应
});
```

### songloft.log — 日志

无需权限。

```javascript
songloft.log.info("informational message");
songloft.log.warn("warning message");
songloft.log.error("error message");
```

日志输出到服务器标准日志，带 `[plugin]` 前缀。

### songloft.plugin — 插件信息

无需权限。

```javascript
async function pluginInfoExample() {
    // 获取插件的 JWT Token（用于访问宿主 API，如音乐文件、封面等需认证的资源）
    var token = await songloft.plugin.getToken();

    // 获取宿主服务的基础 URL（如 http://192.168.1.100:58091）
    var hostUrl = await songloft.plugin.getHostUrl();
}
```

**典型用法：构建带认证的资源 URL**

```javascript
async function getMusicUrl(songId) {
    var host = await songloft.plugin.getHostUrl();
    var token = await songloft.plugin.getToken();
    return host + "/music/" + encodedPath + "?access_token=" + token;
}
```

**方法说明：**
- `getToken()` — 返回当前有效的 JWT access_token 字符串，可用于访问宿主的受保护 API
- `getHostUrl()` — 返回宿主服务的基础 URL，用于构建完整的 API 或资源地址

### songloft.lyrics — 歌词提供者

无需权限。

插件可以注册为歌词提供者，在歌曲没有歌词时由宿主自动调用。

#### 注册与取消

```javascript
// 注册为歌词提供者
songloft.lyrics.registerProvider();

// 取消注册
songloft.lyrics.unregisterProvider();
```

#### 实现搜索端点

注册后，宿主会通过 `InvokeHTTP` 调用插件的 `/lyric-search` 端点搜索歌词。插件需自行实现该路由。

**请求参数（Query String）：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `title` | string | 歌曲标题 |
| `artist` | string | 艺术家 |
| `album` | string | 专辑名 |
| `duration` | number | 时长（秒） |
| `fingerprint` | string | 音频指纹（Chromaprint，可选，有值时才传） |
| `isrc` | string | ISRC 国际标准录音编码（可选，有值时才传） |

**响应格式（HTTP 200 + JSON）：**

```json
{
  "lyric": "[00:01.00]歌词第一行\n[00:05.00]歌词第二行",
  "tlyric": "[00:01.00]翻译第一行",
  "rlyric": "[00:01.00]罗马音第一行",
  "lxlyric": "[00:01.00]逐字歌词"
}
```

- `lyric`（必填）：主歌词，LRC 格式
- `tlyric`（可选）：翻译歌词
- `rlyric`（可选）：罗马音歌词
- `lxlyric`（可选）：逐字歌词

无结果时返回 HTTP 404 或空 body。

#### 完整示例

```typescript
/// <reference types="@songloft/plugin-sdk" />
import { createRouter, jsonResponse, parseQuery } from '@songloft/plugin-sdk';

const router = createRouter();
let registered = false;

router.get('/lyric-search', async (req: HTTPRequest) => {
  const q = parseQuery(req.query);
  const result = await searchFromMySource(
    q.title, q.artist, q.album,
    parseFloat(q.duration) || 0,
    q.fingerprint,  // 可选，用于精确匹配
    q.isrc           // 可选，用于精确匹配
  );
  if (!result) return jsonResponse(null, 404);
  return jsonResponse(result);  // { lyric, tlyric?, rlyric?, lxlyric? }
});

globalThis.onInit = async () => {
  songloft.lyrics.registerProvider();
  registered = true;
};

globalThis.onDeinit = async () => {
  if (registered) songloft.lyrics.unregisterProvider();
};

globalThis.onHTTPRequest = (req: HTTPRequest) => router.handle(req);
```

#### 工作流程

1. 用户播放无歌词的歌曲，客户端请求 `GET /api/v1/songs/{id}/lyric`
2. 宿主发现歌词为空，遍历所有已注册的歌词提供者插件
3. 对每个插件调用 `GET /lyric-search?title=...&artist=...`（15 秒超时）
4. 第一个返回 HTTP 200 + 非空歌词的插件胜出，停止遍历
5. 搜到的歌词异步写入数据库缓存（`lyric_source=scraped`），后续请求直接返回缓存
6. 本地歌曲还会将歌词嵌入音频文件标签

### songloft.covers — 封面提供者

无需权限。

插件可以注册为封面提供者，在歌曲没有封面时由宿主自动调用。

#### 注册与取消

```javascript
// 注册为封面提供者
songloft.covers.registerProvider();

// 取消注册
songloft.covers.unregisterProvider();
```

#### 实现搜索端点

注册后，宿主会通过 `InvokeHTTP` 调用插件的 `/cover-search` 端点搜索封面。插件需自行实现该路由。

**请求参数（Query String）：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `title` | string | 歌曲标题 |
| `artist` | string | 艺术家 |
| `album` | string | 专辑名 |
| `fingerprint` | string | 音频指纹（Chromaprint，可选，有值时才传） |
| `isrc` | string | ISRC 国际标准录音编码（可选，有值时才传） |

**响应格式（HTTP 200 + JSON）：**

```json
{
  "cover_url": "https://example.com/covers/album.jpg"
}
```

- `cover_url`（必填）：封面图片的完整 URL

无结果时返回 HTTP 404 或空 body。

#### 完整示例

```typescript
/// <reference types="@songloft/plugin-sdk" />
import { createRouter, jsonResponse, parseQuery } from '@songloft/plugin-sdk';

const router = createRouter();
let registered = false;

router.get('/cover-search', async (req: HTTPRequest) => {
  const q = parseQuery(req.query);
  const coverUrl = await searchCoverFromMySource(
    q.title, q.artist, q.album,
    q.fingerprint,  // 可选，用于精确匹配
    q.isrc           // 可选，用于精确匹配
  );
  if (!coverUrl) return jsonResponse(null, 404);
  return jsonResponse({ cover_url: coverUrl });
});

globalThis.onInit = async () => {
  songloft.covers.registerProvider();
  registered = true;
};

globalThis.onDeinit = async () => {
  if (registered) songloft.covers.unregisterProvider();
};

globalThis.onHTTPRequest = (req: HTTPRequest) => router.handle(req);
```

#### 工作流程

1. 用户播放无封面的歌曲，客户端请求 `GET /api/v1/songs/{id}/cover`
2. 宿主发现封面为空，遍历所有已注册的封面提供者插件
3. 对每个插件调用 `GET /cover-search?title=...&artist=...`（15 秒超时）
4. 第一个返回 HTTP 200 + 非空 `cover_url` 的插件胜出，停止遍历
5. 搜到的封面异步持久化：
   - **本地歌曲**：下载封面图片 → 保存到本地 `cover_path` → 嵌入音频文件标签
   - **远程歌曲**：存储 `cover_url` 到数据库
6. 后续请求直接返回缓存，不再调用插件

### 提供者机制通用说明

歌词和封面提供者共享相同的架构：

- **多插件支持**：多个插件可同时注册为同一类型的提供者，宿主按 first-match-wins 策略依次尝试
- **空闲驱逐安全**：插件被空闲驱逐（内存回收）后，提供者注册不会丢失；下次搜索时宿主会自动重新加载插件
- **惰性清理**：已禁用或已删除的插件会在搜索遍历时被自动从提供者集合中移除
- **指纹与 ISRC**：建议插件优先使用 `fingerprint` 和 `isrc` 进行精确匹配（如果有值），再 fallback 到 title/artist 模糊搜索

---

## 6. 权限系统

插件必须在 `plugin.json` 的 `permissions` 字段中声明所需权限。运行时调用 API 时会校验权限，未声明的权限将被拒绝。

### 可用权限列表

与后端 `internal/jsplugin/permissions.go` 的 `AllPermissions` 保持一致：

| 权限 | 说明 |
|------|------|
| `storage` | 读写插件私有持久化存储 |
| `songs.read` | 读取歌曲元数据 |
| `songs.write` | 修改/写入歌曲元数据 |
| `songs.*` | 歌曲读写通配符（一把梭糖） |
| `playlists.read` | 读取歌单及歌单中的歌曲 |
| `playlists.write` | 创建/修改/删除歌单及其歌曲 |
| `playlists.*` | 歌单读写通配符（一把梭糖） |
| `inter-plugin` | 插件间通信 |
| `command` | 执行外部命令/管理可执行文件 |
| `jsenv` | 创建/执行子 JS 沙箱环境 |
| `fs` | 读写插件数据目录内文件 |
| `fs:music` | 访问 music_path 音乐目录 |
| `fs:external` | 访问管理员配置的外部目录 |
| `websocket` | 使用 `new WebSocket(...)` 主动连接外部服务，或处理入站 `onWebSocket` upgrade |
| `persistent-storage` | 读写卸载插件后仍保留的持久化存储 |
| `net` | 使用原始网络 socket（UDP / 出站 TCP） |

> 注意：网络请求 (`fetch`)、定时器 (`setTimeout/setInterval`)、日志等能力**无需权限声明**，是默认宿主能力。

### 通配符糖

以 `.*` 结尾的权限在声明层作为一把梭糖，runner 在检查时用前缀匹配。例如声明 `playlists.*`
既包括 `playlists.read` 也包括 `playlists.write`；而单声明 `playlists.read` 时无法调用写接口。

### 最小权限原则

只声明实际需要的权限，减少安全风险：

```json
{
  "permissions": ["storage", "songs.read"]
}
```

---

## 7. 插件间通信

插件可以通过消息机制相互协作。

### 异步发送（Send）

发送方不等待响应，适合通知类场景：

```javascript
// Plugin A: 通知 Plugin B
async function notifyB() {
    await songloft.comm.send("plugin-b", "data-updated", { source: "plugin-a" });
}
```

### 同步调用（Call）

发送方等待接收方处理并返回结果：

```javascript
// Plugin A: 调用 Plugin B 的服务
async function fetchFromB() {
    var response = await songloft.comm.call("plugin-b", "get-data", { id: 123 }, 5000);
    if (response.success) {
        var data = response.data;
    }
}
```

### 注册处理器（onMessage）

接收方注册处理特定 action 的函数：

```javascript
// Plugin B: 注册 action handler
songloft.comm.onMessage("get-data", function(payload, from) {
    songloft.log.info("Request from: " + from);
    // payload = { id: 123 }
    return { name: "example", value: 42 };
});

songloft.comm.onMessage("data-updated", function(payload, from) {
    songloft.log.info("Got notification from: " + from);
    // 无需返回值（send 场景）
});
```

### 通信权限

通信双方都需要 `inter-plugin` 权限。

---

## 8. 静态资源

插件可以通过 `static/` 目录提供 Web UI。

### 目录结构

```
my-plugin/
├── plugin.json
├── main.js
└── static/
    ├── index.html
    └── js/
        └── app.js       # 插件自定义逻辑
```

> 公共资源（设计令牌/reset/MD3 组件样式、字体文件、API 工具库 `common.js`）由主程序自动注入，**无需**在插件中打包。

### 主程序自动注入

后端在返回插件 HTML 页面时，会在 `<head>` 顶部自动注入以下内容（按顺序）：

1. **`<base>`** — 设置相对路径基准，HTML 中可直接用相对路径引用 `static/...` 和插件 API
2. **Auth bridge 脚本** — 从 URL `?access_token=` 写入 localStorage、fetch 503 自动重试
3. **`theme.css`** — MD3 颜色令牌（含亮/暗双主题）、字体声明、CSS reset、`body` 基样、安全区默认值
4. **`components.css`** — 与客户端 Flutter 组件对齐的共享组件库（`.card`/`.btn*`/`.switch`/`.tab-bar` 等）
5. **`webf-shims.css`** — WebF 引擎垫片样式（仅 WebF 渲染面命中）
6. **`common.js`** — embed 检测、主题桥接、`window.SongloftPlugin` 全局 API
7. **`webf-shims.js`** — WebF 引擎能力垫片（空 `img src`、`<details>`、滑块、文件选择、安全区等；仅 WebF 生效）

> `common.css` 已退役并拆成 `theme.css` + `components.css`；WebF 兼容层从 `common.js` 抽离到 `webf-shims.js`。插件无需改动——公共 API（`window.SongloftPlugin`）表面不变。

因此插件 HTML **不需要**：
- `<link>` 引用 fonts.css 或 style.css（主程序提供）
- embed 检测脚本（主程序提供）
- 打包字体文件（主程序通过 `/api/v1/jsplugin-assets/fonts/` 提供）

### window.SongloftPlugin — 浏览器端全局 API

主程序注入的 `common.js` 暴露 `window.SongloftPlugin` 全局对象，提供以下方法：

```javascript
// API 请求
SongloftPlugin.getAuthToken()        // 从 localStorage 读取 JWT token
SongloftPlugin.apiGet(path)          // GET 请求，返回 Promise<JSON>
SongloftPlugin.apiPost(path, body)   // POST 请求
SongloftPlugin.apiPut(path, body)    // PUT 请求
SongloftPlugin.apiDelete(path)       // DELETE 请求

// 主题
SongloftPlugin.getTheme()            // 返回 'light' | 'dark'
SongloftPlugin.onThemeChange(cb)     // 监听主题变化，cb(theme: 'light' | 'dark')
SongloftPlugin.getColorScheme()      // 宿主真实色板 {primary:'#415F91', ...}；未下推时 null
SongloftPlugin.forceStyleRecalc()    // 自行改过根节点 CSS 变量后调用（WebF 专用，别处为空操作）

// 页面内导航
SongloftPlugin.onHostBack(fn)        // 注册返回键钩子，fn 返回 true 表示已消费（仅 WebF）
```

插件 JS 可直接使用：

```javascript
const { apiGet, getTheme, onThemeChange } = SongloftPlugin;

const data = await apiGet('/api/hello');
console.log('当前主题:', getTheme());
onThemeChange(theme => console.log('主题切换到:', theme));
```

如果插件有多个 JS 文件，每个文件顶部直接从全局解构即可：

```javascript
const { apiGet, apiPost } = SongloftPlugin;
```

### 客户端 SDK —— 调用宿主播放器（webview 页面专用）

在 Songloft 客户端中打开的插件页面，可通过 `window.SongloftPlugin.host` / `.player` 调用宿主客户端能力——最常见的是改写宿主的「正在播放队列」。

> - 生效范围：**native 客户端**（Android/iOS/macOS/Windows/Linux）的 webview 插件页；**Web 端插件页**（Tab 内嵌页与首页/全屏页均在宿主 iframe 内打开，走 postMessage 桥接）。
> - 不生效：仅当用户通过「在浏览器中打开」把插件页在独立新浏览器标签打开时（无宿主父窗口）——此时 `host.isAvailable()` 返回 `false`，调用会抛错，务必先 feature-detect。
> - 能力由宿主客户端注入，跟随客户端版本。请在 `plugin.json` 设置合适的 `minHostVersion`，并用 `host.getInfo().capabilities` 做能力协商。

```javascript
const { host, player } = SongloftPlugin;

if (host && host.isAvailable()) {
  // 能力协商
  const info = await host.getInfo();   // { version, platform, capabilities: ['player'] }

  // 用歌曲 id 替换正在播放队列并从第 0 首开始播（id 通常来自你自己的搜索结果，
  // 先经服务端 songs.create 持久化拿到 id）
  await player.setQueue([101, 102, 103], { startIndex: 0 });

  // 追加到队列末尾（不打断当前播放）
  await player.addToQueue([104]);

  // 读取状态 / 订阅状态变更
  const state = await player.getState();     // { queue, current_index, is_playing, ... }
  const off = player.onStateChange(s => console.log('当前第', s.current_index, '首'));
}
```

`player` 命名空间方法：`getState` / `setQueue` / `addToQueue` / `insertToQueue` / `removeFromQueue` / `reorderQueue` / `clearQueue` / `play(id?)` / `pause` / `togglePlay` / `next` / `prev` / `seek(seconds)` / `setVolume(0-100)` / `setPlayMode('order'|'loop'|'single'|'random'|'singlePlay')` / `playPlaylistById(id)` / `onStateChange(cb)`。

用 TypeScript / 构建工具（如 Vue 模板）开发时，安装 [`@songloft/client-sdk`](https://github.com/songloft-org/plugin-toolchain/tree/main/packages/client-sdk) 获得完整类型与便捷封装：

```ts
import { player, host, isClient } from '@songloft/client-sdk';

if (isClient()) {
  await player.setQueue([101, 102], { startIndex: 0 });
}
```

免构建的 vanilla 静态页面无需安装，直接用注入的 `window.SongloftPlugin.player` 即可（仅少了类型提示）。

### Cookie 读取桥 —— 获取第三方站点登录态（native 专用）

插件若需要第三方站点的会话 Cookie（如 FN Connect 网关的 `os-access-code`、自建 NAS 的登录态等），可使用 `getCookies` 桥接——由宿主原生层从 WebView Cookie Store 读取，不受浏览器同源策略和 HttpOnly 限制。

```javascript
// 前提：用户已在应用内 WebView 中打开目标站点并完成登录
const cookies = await SongloftPlugin.getCookies('https://pcyear.5ddd.com');
// cookies: { 'os-access-code': 'xxx', 'music-token': 'yyy', ... }
```

**参数说明：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `origin` | `string` | 目标站点 origin，必须含协议+主机（+端口），如 `https://example.com`。路径忽略 |

**返回值：** `Promise<Record<string, string>>` — Cookie 名→值映射。该 origin 无 Cookie 时返回空对象 `{}`。

**平台限制：**

| 平台 | 支持 | 说明 |
|------|------|------|
| Android / iOS / macOS / Windows / Linux | ✅ | 原生 `CookieManager` 读取 WebView Cookie Store |
| Web | ❌ | 浏览器同源策略硬限制，调用会 reject |

> ⚠️ 使用前建议检测平台：
> ```javascript
> const info = await SongloftPlugin.host.getInfo();
> if (info.platform !== 'web') {
>   const cookies = await SongloftPlugin.getCookies(origin);
> }
> ```

**典型使用流程（以 FN Connect 为例）：**

1. 用户添加飞牛音乐音源，插件拼出 origin（如 `https://pcyear.5ddd.com`）
2. 插件引导用户在应用内 WebView 中打开目标站点并登录
3. 登录完成后，插件调用 `getCookies(origin)` 获取会话 Cookie
4. 将 Cookie 存入插件后端配置（通过 `apiPost` 等），后续请求携带该会话

**TypeScript 用法：**

```ts
import { getCookies, host } from '@songloft/client-sdk';

const info = await host.getInfo();
if (info.platform !== 'web') {
  const cookies = await getCookies('https://pcyear.5ddd.com');
  // 传给插件后端保存
  await SongloftPlugin.apiPost('/config/cookies', cookies);
}
```

### 主题适配

主程序的 `theme.css` 在 `:root` 下定义了 `--md-*` CSS 变量（亮色），并在 `html[data-theme="dark"]` 下覆盖为暗色值。插件页面使用这些变量即可自动适配主题：

```css
/* 插件自定义样式 — 引用 --md-* 变量自动跟随主题 */
.my-card {
    background: var(--md-surface-container);
    color: var(--md-on-surface);
    border: 1px solid var(--md-outline-variant);
}
```

主题变化时（用户在主程序设置中切换），`common.js` 会：
1. 更新 `<html>` 的 `data-theme` 属性和 `theme-light`/`theme-dark` CSS class
2. 用宿主推来的真实色板改写 `--md-*`（见下）
3. 派发 `songloft-theme-change` CustomEvent
4. 写入 `localStorage['songloft-theme']`

插件 JS 可通过 `SongloftPlugin.onThemeChange(callback)` 监听主题变化做额外处理。

#### 变量清单

`--md-*` 与 Flutter 的 `ColorScheme` 字段**逐一对应**（camelCase → kebab-case），方便两侧对照：

| 分组 | 变量 |
|---|---|
| 主色 | `--md-primary` `--md-on-primary` `--md-primary-container` `--md-on-primary-container` |
| 次色 | `--md-secondary` `--md-on-secondary` `--md-secondary-container` `--md-on-secondary-container` |
| 第三色 | `--md-tertiary` `--md-on-tertiary` `--md-tertiary-container` `--md-on-tertiary-container` |
| 错误 | `--md-error` `--md-on-error` `--md-error-container` `--md-on-error-container` |
| 表面 | `--md-surface` `--md-on-surface` `--md-on-surface-variant` `--md-surface-dim` `--md-surface-bright` |
| 表面阶梯 | `--md-surface-container-lowest` `--md-surface-container-low` `--md-surface-container` `--md-surface-container-high` `--md-surface-container-highest` |
| 描边 / 反色 | `--md-outline` `--md-outline-variant` `--md-inverse-surface` `--md-on-inverse-surface` `--md-inverse-primary` |
| 本项目自有（M3 无此角色，**不参与下推**） | `--md-success` `--md-success-container` `--md-warning` `--md-warning-container` |
| 派生别名（组件在用：switch 轨道 / progress 底） | `--md-surface-variant`→`surfaceContainerHighest` |
| 圆角刻度（对齐 Flutter `AppRadius`） | `--md-radius-sm` 8 · `-md` 12 · `-lg` 16 · `-xl` 24 · `-xxl` 28 · `-full` 50px |
| 阴影 | `--md-shadow-1` `--md-shadow-2` `--md-shadow-3` |

> 旧的 `--md-surface-1` / `--md-surface-2` 别名已移除，请直接写 `--md-surface-container` / `--md-surface-container-high`。

**`surface` 与 `surface-container` 的关系别搞反**：`--md-surface` 是**页面底色**，卡片 / 输入框 / hover 用 `container` 阶梯依次加深。公共 `.card` 已是 `SectionCard` 形制（描边无阴影 + 16 圆角），组标题用卡外的 `.section-title`（大写小字）——直接套用即与客户端一致，通常无需自定义：

```css
/* 仅当要自绘卡片时才需要，公共 .card 已经是这个形制 */
.my-section-card {
    background: var(--md-surface-container);
    border: 1px solid var(--md-outline-variant);
    border-radius: var(--md-radius-lg);
}
```

#### 宿主实时下推真实色板

`theme.css` 里的静态值只是**首帧兜底**（由默认 seed `#415F91` 导出）。页面就绪后宿主会把**真实的** `ColorScheme` 随 `songloft-theme` 消息推来——含用户自定义 ThemePack——写成 `documentElement` 的**内联**自定义属性。内联优先级最高，连插件自己在 `:root` 里重定义的同名变量也会被压住。

于是纯 CSS 的插件**什么都不用改**就与主程序同色。要在 JS 里读色则必须用 `getColorScheme()`：

```javascript
// 宿主还没推到时返回 null（此时页面用的是静态兜底色）
const cs = SongloftPlugin.getColorScheme();
const primary = (cs && cs.primary) || '#415F91';

// 到达 / 变更时收通知。注意事件派发在 document 上且**不冒泡**，监听 window 收不到。
document.addEventListener('songloft-color-scheme-change', e => {
    console.log('新主色:', e.detail.colors.primary);
});
```

> ⚠️ **WebF 下这是唯一能读到色值的途径。** `getComputedStyle(document.documentElement).getPropertyValue('--md-primary')` 在 WebF 里一律返回空串；而 `<flutter-cupertino-*>` 的属性值不展开 `var()`、只吃字面 hex（如 `activeColor`）。两条常见套路都走不通，只能靠这个 API 拿字面值再喂进属性。
>
> 色板保证在 `songloft-theme-change` 派发**之前**就已落地，所以在 `onThemeChange` 回调里直接调 `getColorScheme()` 拿到的一定是新值。切换主题包时亮暗可能没变、只有色值变——那种情况只有 `songloft-color-scheme-change` 会派发，两个事件都监听才完整。

### 页面内导航与返回键

插件页是**单页**的，如果内部有多级视图（如「主页 → 设置页」），需要两件事：

1. **自绘返回按钮**——全平台都要有，这是主路径。
2. **让宿主的返回键也跟着走**——Android 硬件返回键、全屏页 AppBar 的返回箭头。

```javascript
// 返回 true = 本次返回已被页面消费，宿主不退出路由 / 不退出应用
SongloftPlugin.onHostBack(() => {
    if (currentPage !== 'main') { currentPage = 'main'; return true; }
    return false;
});
```

> ⚠️ **`onHostBack` 只对 WebF 渲染的插件页生效**（宿主的注册点有 `isWebFHost()` 闸门）。系统 WebView / iframe / 浏览器三条链路的宿主走各自 `controller.canGoBack()`，读的是真实浏览历史，JS 侧拦不到。这个不对称是设计使然：那些环境下自绘返回按钮已经够用。
>
> ⚠️ **不要用 `history.pushState` 表达页面内层级。** WebF 不实现 SPA history 路由（官方方案是 `@openwebf/*-router` 的原生屏栈）。`pushState` 之后 `history.length > 1` 会让宿主的兜底判断误报「已消费」，而 WebF 又不 fire `popstate`——页面毫无变化，**返回键直接变成死键**。

### WebF 渲染引擎与原生元素

新版客户端在部分平台上可以用 [WebF](https://openwebf.com/)（自研 W3C 运行时，纯 Flutter 渲染）替代系统 WebView 渲染插件页。**这是逐插件选择的**：只有在 `plugin.json` 里声明 `"renderEngine": "webf"` 的插件才走 WebF 渲染面，默认仍是系统 WebView，Web 端永远走 iframe——字段语义与平台限制见 [§3 · renderEngine 渲染引擎声明](#renderengine-渲染引擎声明)。WebF **不是浏览器**，有一批 HTML/CSS 能力缺失，主程序的 `webf-shims.js`（配套样式 `webf-shims.css`）会统一垫掉常见缺口（空 `img src`、`<details>` 折叠等），并给 `<html>` 加上 `webf-engine` class 供插件按引擎分叉：

```css
html.webf-engine .only-in-webf { display: block; }
```

#### WebF 下内联 SVG 的真实限制

WebF 的 `<svg>` 实现是：把整棵 svg 子树**重新序列化成字符串**，交给 `flutter_svg` 渲染。因此：

- svg 的子节点只作数据存在，**没有真实 box** —— 拿不到布局（`getBoundingClientRect()` 无意义）
- 单个 `<path>` / `<circle>` **无法单独做 CSS 动画、也无法命中测试**（点不到）
- **任何子节点的属性 / 样式 / 子树变更都会触发整棵 SVG 重建**（重新拼字符串 + 重新解析 + 重新光栅化）

结论：**高频更新的内联 SVG 在 WebF 下性能最差**。典型反例就是「每秒改 `stroke-dashoffset` 的 SVG 进度环」—— 每一次进度变化都在重建整棵 SVG。

#### `<songloft-progress-ring>` —— 原生环形进度条

为此主程序提供了一个原生元素：进度变化只走一次 Flutter `CustomPaint` 重绘，没有字符串序列化、没有 SVG 重解析。

```html
<songloft-progress-ring value="30" max="100" stroke-width="5"
                        style="width:48px;height:48px"></songloft-progress-ring>
```

```js
// 进度更新就是改属性，没有其它 API
document.querySelector('songloft-progress-ring').setAttribute('value', '65');
```

属性一览：

| 属性 | 默认值 | 说明 |
|------|--------|------|
| `value` | `0` | 当前进度值 |
| `min` | `0` | 区间下界 |
| `max` | `100` | 区间上界。`max <= min` 是退化区间，按 0 处理（只画轨道） |
| `stroke-width` | `4` | 线宽（px），夹到 `(0, 短边/2]` |
| `color` | CSS `color` 的值 | 进度弧颜色，**只接受具体色值**（见下） |
| `track-color` | 进度色的 24% 不透明度 | 轨道颜色 |
| `line-cap` | `butt` | 设为 `round` 得圆头端点 |

- **尺寸**走 CSS `width` / `height`（未指定时为 36×36），`display` 默认 `inline-block`
- 弧从 12 点方向起顺时针增长
- **非法值一律夹紧或忽略，不会抛异常**：`value="oops"` 当 0，越界 value 夹到 `[min, max]`，负 `stroke-width` 夹到最小值
- **暂不支持不确定态（indeterminate）动画**，需要转圈请自己用 CSS 动画包一层容器

#### 颜色如何跟随主题

Flutter 侧拿不到插件页的 `--md-*` CSS 变量，所以颜色由**页面**决定，两条路：

```css
/* ① 推荐：CSS color 属性（currentColor 语义）。
   这是唯一能跟着 --md-* 变量走的路 —— 用户在主程序里切换亮/暗主题时，
   环的颜色会跟着重绘。 */
songloft-progress-ring {
    color: var(--md-primary);
}
```

```html
<!-- ② 需要「进度色与文字色不同」或颜色由 JS 算出来时，用属性覆盖 -->
<songloft-progress-ring value="30" color="#4caf50" track-color="#e0e0e0"
                        style="width:48px;height:48px"></songloft-progress-ring>
```

`color` 是可继承属性，所以**什么都不配也能跟主题**：它会取继承来的文字色，而 `theme.css` 已把文字色绑到 `--md-on-surface`。

两个已实测的坑（不要踩）：

- **属性里写 `var(--md-primary)` 不生效**：WebF 的属性值不经过 CSS 变量展开，元素会按「无效颜色」忽略它并退回 ①。要跟变量就用 CSS `color`。
- **`getComputedStyle(el).getPropertyValue('--md-primary')` 在 WebF 下一律返回空串**：WebF 的 getComputedStyle 不暴露自定义属性，所以「用 JS 读变量再写进属性」这条常见套路走不通。**替代方案是 `SongloftPlugin.getColorScheme()`**（见 [主题适配 · 宿主实时下推真实色板](#宿主实时下推真实色板)）——它给的是字面 hex，正好能喂进属性。
- **运行时改 `<html>` 上的 CSS 变量，后代不会重新求值**：WebF 的变量变更只通知**同一个元素**自己的依赖，不向后代遍历。`--md-*` 与 `--sl-safe-*` 的切换由 `common.js` 内部补一次「带后代」的样式重算来兜住，**插件无需改动**；但插件如果自己在运行时改根节点变量（自定义配色、密度开关等），改完必须调 `SongloftPlugin.forceStyleRecalc()`（非 WebF 下是空操作，可以无条件调）。

#### 兼容性与降级

- 该元素**只在 WebF 渲染面下存在**。在普通浏览器、系统 WebView（旧版客户端）里它是未知标签，会渲染成一个空盒子 —— 主程序**不会**把插件里的 SVG 自动替换成它（SVG 是任意图形，机械判定「哪个 svg 是进度环」必然误伤正常 SVG），替换与降级都由插件自己控制
- 需要两端都好看时，两套实现共存 + 用 `html.webf-engine` 二选一：

```css
.ring-native { display: none; }                        /* 默认藏起原生元素 */
html.webf-engine .ring-native { display: inline-block; }
html.webf-engine .ring-svg { display: none; }          /* WebF 下藏起 SVG 版 */
```

#### `<songloft-slider>` —— 原生滑块（`input[type=range]` 的替身）

**大多数插件什么都不用做。** WebF 没有实现 `input[type=range]`——实测那一整行在 WebF 下**一个像素都不画**：既没有滑块也没有文本框，同一行的兄弟文字与该行自己的 `background` 会一起消失。所以主程序的 `webf-shims.js` 垫片在 WebF 下会自动：

1. 扫描页面里所有 `input[type="range"]`，在每个 input **后面**插入一个 `<songloft-slider>`；
2. 把原 `<input>` **隐藏**（加 `.sl-range-hidden` class + inline `display:none`）而**不是移除**；
3. 双向同步两者。

因此插件既有的 JS **一行都不用改**：

- `el.value` 读写照常（垫片在实例上装了访问器；JS 写入会同步给滑块，拖动期间除外）
- `el.disabled = true / false` 照常（滑块会跟着变灰并停止响应手势）
- `el.addEventListener('input' / 'change', ...)` 照常（滑块的交互会在原 input 上派发**冒泡的** `input` / `change`）
- `el.matches(':active')` 照常（拖动中返回 `true`）。这条是插件「用户正在拖，别用轮询结果覆盖」的标准写法；隐藏后的 input 在 WebF 里永远进不了真正的 `:active`，所以垫片遮蔽了 `matches`

垫片幂等（`data-sl-range-shim` 标记），动态插入 HTML 后调 `SongloftPlugin.applyShims()` 即可给新出现的 range 补上滑块。若 `.value` 的访问器装不上（哨兵往返自检失败），垫片会**整体放弃**：删掉滑块、还原原 input、打一条 `console.warn` —— 宁可退回 WebF 的原生表现，也不要「input 被隐藏了、值又同步不上」。

属性一览（走垫片时由垫片从原 input 转写；手写该元素时自己给）。垫片还会一并转写 `aria-label` 与原 input 的 inline `style`，并给滑块加上 `.sl-range-slider` class 和 `data-sl-for="<原 input 的 id>"`：

| 属性 | 默认值 | 说明 |
|------|--------|------|
| `value` | `min` | 当前值 |
| `min` | `0` | 区间下界 |
| `max` | `100` | 区间上界。`max <= min` 是退化区间，停在起点且不响应拖动 |
| `step` | `1` | 步长。`any` 或 `<= 0` 视为连续 |
| `orientation` | `horizontal` | 设为 `vertical` 得竖向滑块（**min 在下、max 在上**） |
| `disabled` | 不存在 | 存在即禁用（`false` / `0` 例外，视为未禁用）。禁用时整体 38% 不透明度且不响应手势 |
| `color` | CSS `color` 的值 | 已填充轨道 + 把手的颜色，**只接受具体色值** |
| `track-color` | 填充色的 24% 不透明度 | 未填充轨道的颜色 |

- **尺寸**走 CSS `width` / `height`；未指定时按朝向兜底为**横向 160×28 / 竖向 28×160**，`display` 默认 `inline-block`
- 事件：拖动与**点击轨道**（点击会让把手跳到点击处，与浏览器一致）都派发 `input`，抬手派发 `change`；新值放在 `event.data` 里（字符串，整数不带 `.0`）
- **交互期元素不回写自己的 `value` 属性**——真值由页面侧持有。所以直接使用该元素时请从 `event.data` 取值，**不要**读 `getAttribute('value')`（那只是你上一次推给它的值）
- `min` / `max` / `step` 必须写成**属性**：WebF 没实现这三个的属性反射，`el.min` 读出来是空串
- 非法值一律忽略并在客户端日志里留一条提示，不抛异常

##### 竖向滑块：必须显式声明 `data-sl-orientation`

垫片**不猜**朝向，要竖向就在原 `<input>` 上写出来：

```html
<input type="range" id="volumeSlider" min="0" max="100" value="50"
       aria-label="音量" data-sl-orientation="vertical">
```

为什么不能自动推断：浏览器里的竖向 range 通常是 `transform: rotate(-90deg)` 转出来的，而 WebF 的 `getComputedStyle` 支持面不可靠（连自定义属性都不暴露），读 transform 反推**猜错了是静默的错朝向**——比要求一行声明糟得多。

不想要滑块（想保留 WebF 的原生表现，或插件自己已经处理了这个 range）就写 `data-sl-no-slider`，垫片会跳过它：

```html
<input type="range" data-sl-no-slider>
```

##### 插件通常需要补几行 CSS

`<songloft-slider>` 是**新标签**，匹配不到插件原有的 `input[type="range"]` 选择器，因此拿不到原有几何。垫片只把原 input 的 **inline `style`** 拷过去（`style="width:100%"` 这种因此自动生效），**刻意不拷 class**——class 上挂的往往是「让原生 range 长得像滑块」的规则（`-webkit-appearance`、`::-webkit-slider-thumb`、`accent-color`），拷过来只会带进无意义甚至有害的声明。

不补 CSS 也能用，只是拿到元素默认尺寸（横向 160×28）、不合版面。可用的三个选择器：`songloft-slider`、垫片加的 class `.sl-range-slider`、以及 `[data-sl-for="<原 input 的 id>"]`（原 id 留在 input 上，不会挪到滑块上）。

第一方插件 miot 的实际写法（竖向音量条，原本是 `width: 110px` + `rotate(-90deg)`）：

```css
songloft-slider {
    color: var(--md-primary);
}

/* 竖向元素自己就是竖着画的，不需要 transform；
   原来 rotate 前的 width 现在对应 height */
.volume-panel .volume-slider-wrap songloft-slider {
    width: 28px;
    height: 110px;
}
```

这些规则在浏览器 / 系统 WebView 下永不匹配（那儿没有 `songloft-slider` 元素），属纯增量，不必包在 `html.webf-engine` 里。

颜色跟随主题的结论与 `<songloft-progress-ring>` 完全一致（CSS `color` / currentColor 走得通，**属性里写 `var()` 不生效**），见上面「[颜色如何跟随主题](#颜色如何跟随主题)」，不再重复。

##### 生效范围与一个已知残留风险

- 垫片**只在 WebF 渲染面下跑**：普通浏览器与系统 WebView 里原生 `input[type=range]` 照常工作，页面不会有任何变化。手写 `<songloft-slider>` 时它在非 WebF 环境是未知标签（空盒子），需要两端都好看就像进度环那样两套实现 + `html.webf-engine` 二选一
- **竖向滑块放进竖向滚动容器里可能抢不到手势**：滑块用与朝向同轴的 drag 手势与滚动**竞争**（这是正确行为，否则会出现「页面在滚 + 滑块同时在动」），胜负依赖手势竞技场的「命中更深者先接受」。miot 的音量面板是弹出层所以不受影响；真要放进长列表里请实测

#### 安全区：用 `--sl-safe-*`，不要写 `env(safe-area-inset-*)`

**WebF 完全没有实现 CSS 的 `env()`** —— 不是求值不准，是连解析入口都不存在。所以刘海屏 / 圆角屏 / 手势条设备上，写 `env(safe-area-inset-bottom)` 的插件页会**顶到状态栏或被下巴切掉**。

主程序改为把真实安全区（Flutter 的 `MediaQuery.viewPadding`）注入成四个 CSS 变量，插件侧统一写 `var()`：

| 变量 | 语义 |
|------|------|
| `--sl-safe-top` | 上安全区（状态栏 / 刘海） |
| `--sl-safe-right` | 右安全区（横屏刘海 / 圆角） |
| `--sl-safe-bottom` | 下安全区（Home 手势条） |
| `--sl-safe-left` | 左安全区 |

```css
/* 推荐写法：一份 CSS 通吃三种运行环境，不需要分叉 */
.player-bar {
    padding: 6px 16px calc(4px + var(--sl-safe-bottom));
}
```

`theme.css` 已经在 `:root` 上给这四个变量备好了默认值（WebF 下再由 `webf-shims.css` 压成 0px 后被宿主注入的真值覆盖），所以**三种环境下都有确定值，插件只写一种形式**：

| 环境 | `var(--sl-safe-bottom)` 的值 |
|------|------|
| 普通浏览器 / 系统 WebView（默认渲染引擎） | `env(safe-area-inset-bottom, 0px)`，即原生真值（桌面浏览器上是 `0px`） |
| WebF + 新版客户端 | 宿主注入的真实 `MediaQuery.viewPadding`（转屏 / 进退全屏 / 页面重挂都会重推） |
| WebF + 旧版客户端（不推安全区） | `0px`，等价于「无安全区」，与不做这件事时的表现一致 |

因此：**只写 `var(--sl-safe-bottom)` 就够了，不要再画蛇添足加 `env()` 兜底。**

三条已实测的硬约束（都在 WebF 下踩过）：

- **`var(--x, env(...))` 这种带 `env()` 兜底的写法在 WebF 下求值为 `0`** —— fallback 链在 `env()` 处断掉，连 `env()` 自己的内层兜底（`env(safe-area-inset-bottom, 19px)` 里的 `19px`）也取不到。所以它不是「更安全的写法」，只是把变量的默认值白白覆盖掉
- **WebF 没有实现 CSS `max()` / `min()`**，整条声明会失效（不只是安全区那一项）。想表达「至少留 24px，安全区更大时按安全区」，把 `max(24px, ...)` 换成 `clamp()`：

  ```css
  /* clamp(MIN, VAL, MAX) 的定义就是 max(MIN, min(VAL, MAX))，
     对任何 ≤ 96px 的安全区与 max(24px, …) 完全等价（真机最大约 34px）。
     浏览器侧零行为变化，WebF 侧实测可用。 */
  .fp-controls {
      padding-bottom: clamp(24px, var(--sl-safe-bottom), 96px);
  }
  ```

  只想「安全区之上再加固定间距」时用 `calc()` 更直白：`calc(24px + var(--sl-safe-bottom))`
- **`clamp()` 可用、参数里也能塞 `var()`**（已实测），但 `calc()` 之外的其它 CSS 数学函数（`max` / `min` / `round` / `mod` 等）都不要用

注意宿主注入的值是**剩给页面自己处理**的那部分：客户端外层已有 `SafeArea` 消化掉一部分安全区（插件 Tab 页消化了上 / 左 / 右，把下方留给页面），所以不会出现「上层让开了、页面又内缩一次」的双重留白。

#### 文件选择：`input[type=file]` 自动接管，但结果**不在 `input.files` 里**

**页面 HTML 一行都不用改，读结果的 JS 必须改。**

WebF 没有实现 `input[type=file]`：它的 `<input>` build 分支只认 radio / checkbox / button / submit / date / time，`type=file` 落到 default → 渲染成一个 Flutter 文本框，**点了什么都不会发生**（也不报错）。所以 `webf-shims.js` 的垫片在 WebF 下会自动：

1. 扫描页面里所有 `input[type="file"]`，把它**隐藏**（加 `.sl-file-hidden` class + inline `display:none`）；
2. 既拦它的 `click` 事件、也覆写它的实例 `click()` 方法 —— 所以「隐藏 input + 外部按钮调 `fileInput.click()`」这种常见写法照样能弹出选择器；
3. 弹出**宿主的原生文件选择器**，把用户选中的文件经桥送回页面；
4. 把结果写进 `SongloftPlugin.lastPickedFiles`，然后在原 input 上派发一个冒泡的 `change`。

为什么垫片必须自己隐藏原 input：实测 **WebF 不认 HTML `hidden` 属性**（带与不带 `hidden` 的 file input 盒子都是 170×24），插件刻意隐藏的那个 input 在 WebF 下会实打实占掉一行，而且是个点不动的空文本框。

##### 读结果：主通道是 `SongloftPlugin.lastPickedFiles`

```js
fileInput.addEventListener('change', function () {
    // ✅ 主通道：一个普通 JS 数组，一定可读
    var files = (window.SongloftPlugin && SongloftPlugin.lastPickedFiles) || [];
    if (!files.length) return;
    importPlaylist(files[0].text);   // as=text（默认）时是解码后的字符串
});
```

- **`input.files` / `FileReader` / `FileList` 在 WebF 下都不能用**：后两者**压根不存在**（实测 `typeof` 均为 `undefined`），所以宿主刻意**不去伪造** `input.files` —— 假 `File` 配不上真 `FileReader`，而真 `FileReader` 根本没有。用 `new FileReader()` 读文件的代码在 WebF 下会直接抛异常。
- `change` 事件上**也会尝试**挂一份 `event.data = {files: [...]}`，但那只是锦上添花：WebF 的 `Event` 是 binding object，能不能挂自定义属性**没有契约保证**。**不要**把它当主通道。
- **用户取消时不派发 `change`**（与浏览器一致），所以不必担心「空 change 走进读取失败分支、弹一个用户没做错任何事的报错」。
- `lastPickedFiles` 是**全局单值**（未选过时为 `null`），页面里有多个 file input 时它存的是最近一次的结果 —— 在 `change` 回调里立刻取走即可。宿主调用失败时只打一条 `console.warn` 且不派发 `change`。

##### 载荷形态：`data-sl-file-as`

```html
<!-- 只要元信息，不读内容 -->
<input type="file" id="pick" accept=".m3u,.m3u8,.json" data-sl-file-as="none">
```

| 取值 | 载荷 | 什么时候用 |
|------|------|-----------|
| `text`（默认） | `text` 字符串 + `encoding`（+ 可能的 `textLossy`） | 导入 m3u / json / lrc 这类文本 |
| `bytes` | `bytesBase64`（base64 字符串） | 二进制文件，或需要自己按 GBK 等编码解码 |
| `none` | 只有 `name` / `size` | 只要文件名与大小 |

**默认是 `text` 而不是 `bytes`**：真实用例（导入 m3u / json）只要文本，而 base64 会把一个 20 MB 的文件变成约 27 MB 的字符串，还要跨两次序列化桥（Dart → C++ → QuickJS），默认不该付这个钱。

每个文件对象的字段：

| 字段 | 出现条件 | 说明 |
|------|---------|------|
| `name` | 总是 | 文件名（不含路径） |
| `size` | 总是 | 字节数 |
| `text` | `as=text` 且读取成功 | 解码后的字符串（BOM 已剥离） |
| `encoding` | 同上 | `utf-8` / `utf-8-lossy` / `utf-16le` / `utf-16be` |
| `textLossy` | 解码有损时为 `true` | 宿主只按 BOM + 严格 UTF-8 判定，**不猜 GBK**；GBK 文件会走「容错解码」并打上这个标记 —— 要精确处理请改用 `as="bytes"` 自己解码 |
| `bytesBase64` | `as=bytes` 且读取成功 | base64 |
| `error` | 读取失败时 | `too_large`（超过单文件 32 MB 上限，同时带 `limit`）/ `read_failed`。**刻意不静默截断**：半截的 m3u 解析出来是「导入成功但少了一半」，比报错难查得多 |

- **拿不到文件路径**，这是有意的：桌面端是真实路径、Android SAF 是 content URI，跨平台语义不一致，对页面 JS 也毫无用处，还是不必要的信息泄露。
- `accept` 原样透传给宿主，但**只有 `.ext` 扩展名形式会变成真的过滤器**；写成 MIME（`text/plain`、`image/*`）或与扩展名混写时，宿主**整体放弃过滤**（宁可多几个可选项——插件自己还会校验，也不要把用户本该能选的文件挡住）。
- `multiple` 存在时可多选，否则只取第一项。
- 一次只允许一个选择器在飞：用户连点不会挂起两次宿主调用（否则回来的两个 `change` 里后到的那个未必是用户最后选的文件）。

##### 退出开关、幂等与两端兼容

想保留 WebF 的原生表现（或插件自己已经处理了这个 input）就写 `data-sl-no-file-picker`，垫片会跳过它：

```html
<input type="file" data-sl-no-file-picker>
```

垫片幂等（`data-sl-file-shim` 标记），动态插入 HTML 后调 `SongloftPlugin.applyShims()` 即可给新出现的 file input 补上接管。垫片**只在 WebF 渲染面下跑**：普通浏览器与系统 WebView 里原生 file input 与 `FileReader` 照常工作。所以推荐**两条路都留**，一份代码通吃三种运行环境：

```js
function readPickedFile(input, cb) {
    var picked = window.SongloftPlugin && SongloftPlugin.lastPickedFiles;
    if (picked && picked.length) return cb(picked[0].text);   // WebF
    var f = input.files && input.files[0];                    // 浏览器 / 系统 WebView
    if (!f) return;
    var r = new FileReader();
    r.onload = function () { cb(r.result); };
    r.readAsText(f);
}
```

#### `URL.createObjectURL` 不存在：改用 `SongloftPlugin.blobToDataURL()`（**异步**）

**WebF 里 `URL.createObjectURL` 压根不存在**（实测 `typeof URL.createObjectURL === 'undefined'`）。`Blob` 本身是有的，但没有任何入口能产出 `blob:` URL，而且**也不可能垫一个**：`blob:` 要资源加载器配合，而 WebF 的加载器只认 `http` / `https` / `assets` / `file` / `data:`，其余 scheme 直接抛错 —— 就算 JS 侧造出一个 `blob:xxx` 字符串，加载那一步必然失败。

典型受害写法是「带鉴权头 fetch 一张图 → 显示」：`fetch` 拿到的是 `Blob`，而 `<img src>` 不能直接吃 Blob。宿主给出的替代是 `data:` URL（WebF 原生支持）：

```js
var url = await SongloftPlugin.blobToDataURL(blob);   // 'data:image/jpeg;base64,...'
```

签名 `blobToDataURL(blob, mimeType?) → Promise<string>`。`mimeType` 用来覆盖 `blob.type`（`blob.type` 为空时用得上），两者都没有时按 `application/octet-stream`。

**三条渲染路径共用同一份实现**：它用的 `Blob.prototype.arrayBuffer` + `btoa` 在普通浏览器与系统 WebView 下同样存在，所以**不必按引擎分叉**，一份异步写法通吃。

##### ⚠️ 它是异步的，所以你必须改调用点

`createObjectURL` 是**同步**的，而 `blobToDataURL` 返回 **Promise**。这不是实现偷懒，而是无法弥合的形状差异：`Blob → base64` 只能经 `arrayBuffer()`（`FileReader` 在 WebF 下不存在），而那本身就是异步的。所以**不要指望宿主给一个同步替身** —— 改调用点是唯一的路：

```js
// ❌ 改写前：WebF 下 URL.createObjectURL 是 undefined，这行直接抛 TypeError
function showCover(blob) {
    var url = URL.createObjectURL(blob);
    img.src = url;
    bg.style.backgroundImage = 'url(' + url + ')';
    // 用完还得记着 URL.revokeObjectURL(url)
}

// ✅ 改写后：函数变异步，拿到字符串后的用法完全不变
async function showCover(blob) {
    var url = await SongloftPlugin.blobToDataURL(blob);
    img.src = url;
    bg.style.backgroundImage = 'url(' + url + ')';
    // 不需要 revoke
}
```

注意「改调用点」会**往上传染**：`showCover()` 变成 async 之后，它的调用者要么跟着 `await` / `.then`，要么接受「图片晚一拍出现」。这是移植时最容易漏的一环——漏掉不会报错，只是图不出来。

##### 两个已实测可用的消费点

- `<img src="data:…">`
- CSS `background-image: url(data:…)` —— 这条尤其值得写明：它走的是与 `<img>` **完全不同**的代码路径，而 data URL 里含逗号与分号，CSS `url()` 的词法本来可能切错。实测能出图，所以「同一张图既当封面 `<img>` 又当模糊背景」可以沿用**同一个** data URL，不必另找出路。

##### 生命周期语义变了

data URL **不需要**（也没有）`revokeObjectURL`：它不是句柄，就是一个字符串。代价是它**常驻内存**（base64 比原始字节大约 4/3），只要还有元素的 `src` / `style` 引用它，或者你自己把它存进了变量 / 数组 / DOM 属性，那份字符串就不会被回收。因此：

- 大图、长列表缩略图这类场景，不要无脑把 data URL 攒进数组当缓存 —— 该丢的时候把引用清掉。
- 更省的做法是**能直接用 URL 就别绕 Blob**：只要那张图能通过一个可直接访问的 URL 拿到（不需要自定义请求头），把 `<img src>` 指过去即可，完全绕开这一整套。

#### `window.open`：外链改由**系统浏览器**打开（插件无需改动）

**插件侧一行都不用改**，但要知道它现在的语义。

WebF 的 `window.open` 曾经是**彻底静默**的：不抛错、也什么都不发生（归因是没装导航代理时，WebF 的默认导航策略把外链无条件 cancel 掉了）。所以「点『去网页登录』什么反应都没有」这种 bug 在 WebF 下既没有报错也没有日志。

新版客户端在 WebF 渲染面上装了导航代理，行为变成三档：

| 目标 | 行为 |
|------|------|
| `#` 开头的页内锚点 | 照常跳转（`pushState` + `hashchange` 正常工作） |
| **外部** `http(s)` / `mailto:` / `tel:` | 用**系统浏览器 / 系统默认应用**打开 |
| **同源**（插件页自己）的 `http(s)` 跳转 | **被拦下**，并在客户端日志里留一条 warn |

```js
// 两种调用形态都已实测可用（单参、带 target 的双参都会真的转发到宿主）
window.open('https://account.xiaomi.com/oauth2/authorize?...');
window.open('https://example.com/help', '_blank');
```

三条要点：

- **它打开的是外部浏览器，而不是页内新窗口**。所以「弹窗里完成操作后由弹窗回填数据到父页」（`window.opener`、给返回的 window 对象赋值、跨窗口 `postMessage`）这类流程在这里**走不通** —— 请改成「用户回到插件页后由插件轮询 / 或提供一个『我已完成』按钮触发回调」的形态。
- **同源整页跳转被刻意拦掉**：WebF 里那条路是「把整个插件页 `load()` 成新地址」，会把宿主注入的上下文、loading 状态、返回键行为全部弄错。**WebF 下不要做多页跳转**，单页 + 页内切换视图，返回键用 [`onHostBack`](#页面内导航与返回键) 对接（**别用 `history.pushState`**，理由见那一节）。
- 其余 scheme（相对路径、`javascript:`、自定义 scheme）一律不放行。若插件依赖自定义 scheme 唤起第三方 App，请当它在 WebF 下不可用并另做降级。

#### `<table>` 在 WebF 下**不存在**：改用 CSS Grid

> **先读下一节。** 如果这张「表」本质上是**一个可滚动的列表**（每行结构相同、行数可能很多），
> 优先用 [`<webf-list-view>` + flex 行](#webf-ui-原生组件flutter-cupertino--webf-list-view)，
> 而不是本节的 CSS Grid —— 那条路能一并绕开下面「grid `auto` 行高」与「sticky 表头」两个最难缠的坑。
> 本节适用于**真正的二维表格**（列宽必须跨行严格对齐、行数有限、不需要虚拟滚动）。

WebF 的元素注册表里 `table` / `thead` / `tbody` / `tr` / `th` / `td` **一个都没有注册**，全部落到未知元素（`display:block`）。后果不是「样式差一点」，而是**信息结构丢失** —— 一张 6 列的表会竖排成 6 行，几十条数据变成几百行无标题文本。而且**完全静默**：不报错、不打日志。

宿主只能帮你**发现**它，不能帮你修：WebF 渲染面下 `webf-shims.js` 会给页面上每个 `<table>` 打上 `data-sl-table-unsupported` 属性并在 console 打一条 warn（你就是照那条 warn 找到这一节的）。**宿主刻意不做自动改写**，原因见下。

**改法：CSS Grid。** 一行 = 连续 N 个单元格 div，靠 grid 的自动放置换行：

```css
.tbl-head, .tbl-body {
    display: grid;
    /* 两个容器共用同一份轨道定义，这是「列宽跨行对齐」的全部秘密 */
    grid-template-columns: 36px minmax(0, 3fr) minmax(0, 2fr) minmax(0, 2fr) 90px 60px;
}
```

```html
<div class="table-wrap">            <!-- 横向滚动：overflow-x 在这里，表头与数据区都在里面 -->
  <div class="tbl">                 <!-- 普通 block：宽度基准 + min-width 下限 -->
    <div class="tbl-head">…6 个表头格…</div>     <!-- 留在纵向滚动容器外面，不用 sticky -->
    <div class="tbl-scroll">       <!-- 纵向滚动：max-height + overflow-y，只包数据区 -->
      <div class="tbl-body">…6×N 个数据格…</div>
    </div>
  </div>
</div>
```

**六条硬约束（每一条都是踩出来的，别自己重新发现）**：

- **不要写 `display: table` / `table-row` / `table-cell`**。WebF 的 `CSSDisplay` 枚举里**没有任何 table 取值**，`resolveDisplay` 落到 `default` 返回 **`inline`** —— 比默认的 `block` **更糟**，是负收益。`display: contents`（浏览器里透明化行元素的标准招数）同样不支持，所以**不能保留 `<tr>` 包裹元素**。
- **单元格必须 `white-space: nowrap` + `overflow: hidden` + `text-overflow: ellipsis`，不能让内容换行**。WebF 的 grid `auto` 行高是**在 min-content 宽度下**测量子项高度的（已确诊的上游缺陷）：可换行时 CJK 每个字都是断行点，「艺术家」3 个字被测成 3 行、12 个字的名字被测成 13 行 —— 实测一行占 **281px**（同内容自然高 41px）、表头 72px，可见区只装得下 1 行，用户看到的是**一张几乎空的表**。`nowrap` 下 min-content == max-content，那个错误的测量也就量对了（实测行距 41 / 表头 39）。这条属性**必须写在随页面加载的 CSS 里**，不能 JS 事后注入 —— 要在行插入之前生效才能从第一次布局起就正确。长内容用 `title` 属性给桌面端悬停看全（拼属性一律用转义引号的函数，`textContent → innerHTML` 那种 `esc()` 不转义引号，含双引号的内容会截断属性）。
- **表头不要用 `position: sticky` 贴顶，让它根本不需要 sticky**。实测 WebF 下 `position: sticky` **压根不生效**，而且**不限于 grid 路径** —— 把一个普通 div 放在 `body` 顶部、用页面级滚动（`documentElement.scrollTop = 300`）也整量滚走（`y = -300`），而 computed `position` 仍是 `"sticky"`、`top` 仍是 `"0px"`、`scroll` 事件也确实派发了（**样式没丢、通知链也跑了，只是偏移没被应用**）。所以结构上避开它：**只让数据区滚动**，表头是它的兄弟节点、留在纵向滚动容器**外面**。这个结构在浏览器 / 系统 WebView 下同样正确，仍是一套代码通吃三条路径。
  - 顺带一条源码事实（对「为什么表头必须是独立 grid 容器」仍然成立）：WebF 的 grid 布局把 `position: sticky` 子项与 absolute/fixed **归成同一类脱流元素**，于是 sticky 表头单元格**既不占格子、也不参与列轨道定宽**。所以无论如何都得是**两个 grid 容器**（表头 + 数据区）共用同一份 `grid-template-columns`。
- **纵向滚动条会让表头与数据区错开一个滚动条宽度，必须补偿**。数据区在自己的滚动容器里，占位式滚动条（桌面浏览器）只吃它的内容宽度，而表头在外面吃不到（实测差十几像素；WebF 与移动端是覆盖式滚动条、差 0）。CSS 里拿不到这个宽度，只能实测：`scrollEl.getBoundingClientRect().width - bodyEl.getBoundingClientRect().width` 写进一个自定义属性、表头 `padding-right` 抵掉它。两个要点：① **必须跨帧量**（WebF 的 layout 是异步的，刚写完 `innerHTML` 立刻量到的是改之前的布局，包一层 `setTimeout` 即可）；② 量不到就当 0，那恰好是覆盖式滚动条的正确值。给滚动容器加 `scrollbar-gutter: stable` 可以让「有没有滚动条」不再改变内容宽度，少一次跨阈值跳动。
- **轨道定义里不要用 `auto` / `min-content` / `max-content`**。只用定宽 `px` 与 `minmax(0, Nfr)`，这样每列宽度是「可用宽度」的**纯函数**，与两个容器各自装了什么内容无关 —— 这是两个独立容器能对齐的前提。
- **窄屏不要用 `@media` + `display:none` 隐藏某几列**。WebF 里 `display:none` 的元素**仍会挂一个 0 尺寸的 box、照样占掉一个 grid 格子**，后面所有单元格会整体错位一格。改为给 `.tbl` 设 `min-width`、低于该宽度整表横向滚动（横向滚动容器必须同时包住表头与数据区，否则滚到右边两者就错列了）。

**为什么宿主不自动改写成 WebF 自带的 `<webf-table>` 家族**：它是 Flutter `Table` widget 的薄封装，能力上限由上游锁死 —— `colspan`/`rowspan` 零支持、CSS `width` 完全无效（只认表头单元格的 `column-width` 属性）、CSS `position:sticky` 无效（要换成 `sticky` 属性）、行必须是**直接子节点**（`<thead>`/`<tbody>` 不拆就渲染出一张**空表且不报错**）。更关键的是那些标签在普通浏览器与系统 WebView 下**根本不存在**，用它就必须长期维护两套模板。CSS Grid 是标准 CSS，**三条渲染路径共用同一套 HTML/CSS/JS、同一套外观**。

**两处不可避免的降级**：`tr:hover` 整行高亮变成单个单元格高亮（展平后 DOM 里没有行元素，纯 CSS 无法表达）；表格无障碍语义丢失（缓解：给每行的交互控件补 `aria-label`，容器补 `role="group"`）。**这两处降级正是「列表类内容改用 `<webf-list-view>` 更好」的理由** —— 那条路上行是真实的行元素，两者都不丢。

#### webf-ui 原生组件（`<flutter-cupertino-*>` / `<webf-list-view>`）

前面几节都是「某个 Web 能力在 WebF 下缺失，怎么绕」。**webf-ui 是相反的方向**：直接用映射到
Flutter widget 的原生元素，绕开整个 CSS 布局层。列表与表单控件用它通常比自己拿 HTML/CSS
拼更省事，也更不容易撞上 WebF 的布局缺陷。

主程序客户端已内置 `webf_cupertino_ui`，所以下面这些标签在 WebF 渲染面下**开箱可用**，
插件侧无需安装任何运行时（npm 上的 `@openwebf/vue-cupertino-ui` / `react-cupertino-ui`
**只是类型声明包**，装它只为编辑器补全）：

| 类别 | 标签 |
|---|---|
| 按钮 | `flutter-cupertino-button` |
| 输入 | `flutter-cupertino-input`、`flutter-cupertino-search-text-field` |
| 开关与选择 | `flutter-cupertino-switch`、`flutter-cupertino-checkbox`、`flutter-cupertino-radio`、`flutter-cupertino-slider`、`flutter-cupertino-sliding-segmented-control` |
| 列表与表单 | `flutter-cupertino-list-section`、`flutter-cupertino-list-tile`（含 `-leading` / `-subtitle` / `-trailing` / `-additional-info` 子标签）、`flutter-cupertino-form-section`、`flutter-cupertino-form-row`、`flutter-cupertino-text-form-field-row` |
| 弹层 | `flutter-cupertino-alert`、`flutter-cupertino-action-sheet`、`flutter-cupertino-modal-popup`、`flutter-cupertino-context-menu` |
| 导航 | `flutter-cupertino-tab-scaffold`、`flutter-cupertino-tab-bar`、`flutter-cupertino-tab-view` |
| 其它 | `flutter-cupertino-icon`（1300+ 图标名）、`flutter-cupertino-date-picker` |

另外 **`<webf-list-view>` 来自 `webf` 包本身**（不属于 cupertino 那一批），映射到 Flutter 的
ListView，自带 view 回收 —— 长列表的首选。

**Cupertino 里没有任意选项的下拉选择器**（`flutter-cupertino-picker` 在
`installWebFCupertinoUI()` 里是**注释掉的**，只有 `date-picker` 注册了）。

> ⛔ **别用原生 `<select>` 兜这一块 —— 它在 WebF 下选中值传不回 JS。**
> 它能画出来、能弹出菜单，所以很容易被当成「可用」；但 WebF 的 `HTMLSelectElement`
> 只暴露 `value` / `selectedIndex` / `disabled` / `multiple` / `required`，**没有 `options`**。
> Vue 的 `v-model` 在 `<select>` 上走 `vModelSelect` 指令，而它整个实现建立在 `el.options`
> 上（change 监听器是 `Array.prototype.filter.call(el.options, o => o.selected)`），
> 于是 `filter.call(undefined, …)` 抛 TypeError —— **任何框架**的 `<select>` 双向绑定都会踩。
> 绕开 v-model 改成显式 `@change` 读 `el.value` 也**实测不通**（剩下的断点在 Dart 侧、
> 从 JS 观测不到）。不报错、不打日志。
>
> **判据陷阱：「下拉显示更新了」不能当成「数据通了」。** WebF 的 select 是 `WidgetElement`，
> 它先改自己的 `selectedIndex` 再派发 `change`，显示的标签由 Flutter 侧维护，与 JS 收不收到
> 值完全无关。

**推荐做法：触发按钮 + 常规流里的内联面板，选项行用普通 `<div>`。**

```vue
<script setup>
const open = ref(false);
// 面板行 = 占位项 + 全部选项；点哪一行就 emit 那一行的 value
const rows = computed(() => [{ value: '', label: '全部' }, ...options.value]);
</script>
<template>
  <SlButton :label="currentLabel" trailing-icon="chevron" @click="open = !open" />
  <!-- v-if 而不是 display:none：WebF 里 display:none 的元素仍会挂 0 尺寸 render box -->
  <div v-if="open" class="panel">
    <div v-for="r in rows" :key="r.value" class="opt"
         @click="open = false; pick(r.value)">{{ r.label }}</div>
  </div>
</template>
```

这套只用三个**核心**原语：cupertino button 的 `click`、常规流块盒、普通元素的 `click`
（DOM click 由 WebF 唯一那个全局 tap recognizer 派发）。值只在你自己的 JS 里流动，
不碰任何 WebF 元素的属性读写。代价是展开时把下方内容顶下去。

刻意**不用**浮层（`position: absolute/fixed` 要赌 WebF 的层叠与命中测试，面板往往得盖在
某个 Flutter widget 上）、**不嵌** `<webf-list-view>`（要赌 tap 穿过 Flutter ListView 的
手势竞技场）、**不靠 `overflow` 滚动**（选项多时让页面自身滚动）。

> ℹ️ **官方的 `<flutter-cupertino-action-sheet>` 也能做下拉，但要知道它的代价。**
> 契约（读 `action_sheet.dart`，不是那份 React 口径的 `.md`）：`el.show(config)`，config
> 可以是对象**或 JSON 字符串**（传字符串更稳）；选中派发
> `CustomEvent('select', detail: {text, event, isDefault, isDestructive, index})`，
> `index` 是 `actions` 下标、**`cancelButton` 不带 `index`**、点遮罩关闭**不派发**。
> 风险在于：宿主元素 build 出来的是 `SizedBox.shrink()`，而 `show()` 的实现是
> `state?._showActionSheetImpl(args)` —— state 还没建立时是**静默 no-op**，不抛异常、
> 不打日志，于是「点了什么都不发生」与「正常工作」在代码里**无法区分**。所以宿主元素
> 必须常驻、不能用 `display:none` 藏；而且一旦不工作，你没有可打的日志、只能靠人肉试。
> downloader 就是因此改用了上面的自绘方案。

反过来说，`v-model` 用在**组件**上是安全的（编译成 `:modelValue` + `@update:modelValue`，
纯 Vue 逻辑，不碰原生指令）；`<input type=checkbox>` 的 `vModelCheckbox` 也安全（只依赖
`el.checked` / `el.value`，WebF 两个都有）。要提防的只有原生 `<select>`。

> 🔍 **图标画成 `?` 方框时，别去改图标名 —— 那是客户端缺字体，不是你的锅。**
> 判据一句话：**看得见问号 = 名字对、字体缺；什么都看不见 = 名字错。**
> 因为 `icon.dart` 对查不到的 `type` 返回的是 `SizedBox.shrink()`（写错名字是**看不见**）；
> 而问号方框是**字体缺失**的表现 —— 图标码点落在私用区，缺字体时由系统兜底字体渲染成
> 「未知字符」占位符（macOS 上正是圆角框里一个 `?`）。根因是 `webf_cupertino_ui` 没有依赖
> `cupertino_icons`，客户端必须自己补这条依赖（Songloft 客户端已补，见父仓
> `docs/webf/handoff.md` 第 15 条）。老客户端上遇到只能等客户端更新，插件侧无法绕开。

##### 必须做特性探测，不能只判 `window.webf`

webf-ui 的标签在浏览器 / 系统 WebView / Web 端 iframe 里都是**未知元素**，所以用了它就必须
准备一条 HTML 回落路径（见下一小节）。而选择走哪条路径时，**判据不能是「我是不是跑在 WebF 里」**：

客户端与插件是**各自独立发版**的，`minHostVersion` 只约束**服务端**版本，没有任何字段能约束
客户端。于是「新插件 + 老客户端」这个组合必然出现，那时 `<flutter-cupertino-*>` 会落到
`_UnknownHTMLElement`（一个空的 `display:block` 盒子），用户看到的是**所有控件凭空消失，
且不报任何错**。

所以探的是「元素到底注册上了没有」—— 注册上时 bindings 会在实例上定义对应的 JS 属性：

```js
function hasCupertinoUI() {
    if (!window.webf) return false;
    try {
        return document.createElement('flutter-cupertino-switch').checked !== undefined;
    } catch (e) {
        return false;
    }
}

// <webf-list-view> 来自 webf 包、与 cupertino 是两件事，单独探
function hasListView() {
    if (!window.webf) return false;
    try {
        return typeof document.createElement('webf-list-view').finishLoad === 'function';
    } catch (e) {
        return false;
    }
}
```

探测结果在页面生命周期内不会变，算一次存成常量即可。

##### 引擎分叉收敛到叶子组件，业务代码只写一套

**不要**在业务代码里到处 `if (isWebF)`，也不要维护两套页面模板（那条路已经反复证明会腐化 ——
其中一份在开发机上永远跑不到，改坏了没人发现）。把分叉压到一层薄包装组件里：

```vue
<!-- ui/SlSwitch.vue —— 唯一的分叉点 -->
<template>
  <flutter-cupertino-switch v-if="useNativeUI" ref="el" @change="onNativeChange" />
  <label v-else class="my-switch">
    <input type="checkbox" :checked="modelValue" @change="onHtmlChange" />
    <span class="my-switch-track" />
  </label>
</template>
```

```vue
<!-- 业务代码只见包装组件 -->
<SlSwitch :model-value="settings.embedMetadata" @update:model-value="save" />
```

##### 属性契约的三个坑（都会静默失效，不报错）

**① HTML 属性是 kebab-case，JS 属性才是 camelCase。**

`webf_cupertino_ui` 的 `*_bindings_generated.dart` 里同时注册了两套名字：

```dart
attributes['active-color'] = ElementAttributeProperty(...);   // HTML 属性
'activeColor': StaticDefinedBindingProperty(...)              // JS 属性
```

模板里写 `activeColor` 会被当成 `activecolor` 属性（HTML 属性名大小写不敏感），**不匹配任何
注册项、静默无效**。一律写 kebab-case。

**② 输入框的值属性叫 `val`，不是 `value`；而且它是受控的。**

`<flutter-cupertino-input>` 内部每次 build 都做
`if (_controller.text != val) { 整段替换文本并把光标收到末尾 }`。所以如果你在回写路径上做了
类型转换（比如数字输入框存成 `Number`），用户敲到一半的中间态就会被改写、光标跳走。
**数字字段请在状态里存字符串，提交时才 `parseInt`。**

> ⚠️⚠️ **「受控」还意味着一条会让整页白屏的崩溃链（debug 构建）**：对**已挂载**的输入框
> 回写 `val` 会走 `_Editable.updateRenderObject` → `RenderEditable.text=` →
> `TextPainter.markNeedsLayout`；若同一帧鼠标正停在插件页上，`MouseTracker.updateAllDevices`
> 的 hit test 会打到这个 layout 未就绪的 `RenderEditable`，`getClosestGlyphForOffset` 撞
> `Text layout not available` 断言，异常把 `_debugDuringDeviceUpdate` 永久置位，之后
> **每帧**刷 `mouse_tracker.dart` 的 `!_debugDuringDeviceUpdate` 断言、帧循环烂掉
> （2026-08-05 在 downloader 实测，栈已确证；触发要求鼠标停在页面上，截图脚本永远撞不到）。
>
> **结论：WebF 原生输入按非受控用 —— `val` 只在挂载时给一次初值，之后只读 `input`
> 事件，永不回写。** 外部确实要改显示值时（规范化、清空），改 `:key` 让元素**重挂载**，
> 新值走 mount 而不是 update。downloader 的 `ui/SlInput.vue` 就是这个模式的完整实现。
> 换 WebF 自己的 HTML `<input>` 也绕不开：它底下同样是 Flutter `TextField`
> （`webf/lib/src/html/form/base_input.dart`），同一个 `RenderEditable`。

**③ 布尔属性的两个入口语义不同，而框架走哪个是启发式的。**

同一个 `checked`：

```dart
// HTML 属性 setter —— 认字符串
setter: (value) => checked = value == 'true' || value == ''
// JS 属性 setter —— 认真布尔
set checked(value) { final bool next = value == true; ... }
```

字符串 `'true'` 走 JS 属性入口会变成 **false**（Dart 里 `'true' == true` 为假）。而 Vue / React
对自定义元素是**启发式**决定走 prop 还是 attr 的，插件侧无法确定它选了哪条 —— 选错就是
「开关点了没反应」这种最难查的无声故障。

**结论：布尔属性绕开模板绑定，拿到元素引用后命令式赋 JS 属性，传真布尔。**

```js
// 依赖变化时重新写一遍；flush:'post' 保证元素已挂载
watchEffect(() => {
    if (el.value) el.value.checked = !!props.modelValue;
}, { flush: 'post' });
```

字符串类属性（`val` / `placeholder` / `type` / `variant` / `active-color`）两条入口都会
`toString()`，模板绑定是安全的。

##### 子节点契约：只认第一个子节点

`<flutter-cupertino-button>` 的实现就是
`childNodes.isEmpty ? SizedBox() : childNodes.first.toWidget()`，官方 `button.md` 亦明写
"The first child is used as the primary content"。所以「图标 + 文字」两个并列子节点时，
文字会被整段丢弃 —— **不报错、不打日志**，只是画出个缺内容的按钮。

**结论：内容包成恰好一个子元素，文字再包一层元素。**

```html
<!-- ✗ 文字丢失（只取第一个子节点） -->
<flutter-cupertino-button><flutter-cupertino-icon type="arrow_clockwise" />刷新</flutter-cupertino-button>
<!-- ✓ -->
<flutter-cupertino-button><span class="btn-inner"
  ><flutter-cupertino-icon type="arrow_clockwise" /><span>刷新</span
></span></flutter-cupertino-button>
```

包装组件里**别开放插槽**，把文字与图标做成 prop（`<SlButton icon="refresh" label="刷新" />`），
让调用方没机会违反。另外文字那层要 `white-space: nowrap` —— 按钮宽度按内容的固有宽度
算，可换行的 CJK 标签会被量成「一字一行」的又窄又高的按钮。

> **裸文本子节点（`<flutter-cupertino-button>刷新</...>`）可以用。** upstream 的 `button.md`
> 快速上手示例就是这种写法。本文这里一度写过「裸文本画不出来」，后来查明那次观察取自一个
> **进程内缓存的旧 bundle**（重装插件后没重启客户端），已撤回。仍然建议按上面的写法包一层
> 元素 —— 反正「图标 + 文字」时本来就必须包，统一写法少一个分叉。

##### 装饰交给 widget 画，不要在 CSS 上再写一份

`<flutter-cupertino-input>` 内部的 `CupertinoTextField` 直接把 `renderStyle.decoration`
（也就是你写的 CSS `border` / `border-radius` / `background`）当成自己的 `decoration`。而 WebF
自己的 render box **也**会画同一份 CSS 装饰，于是屏幕上出现**两个框**：外框在 border box 上，
widget 那份落在 content box 上、被 `padding` 内缩一圈。

所以这类元素在 CSS 里只写尺寸（`width` / `height`），装饰留给 widget —— 背景 / 边框 / 圆角 /
阴影一个都不写时 `renderStyle.decoration` 返回 `null`，widget 回落到自带的 `systemGrey6` +
8px 圆角，并且跟着 CupertinoTheme 亮/暗自适应。同理 `font-size` / `color` 对它是无效的
（widget 不读 renderStyle 的文本样式），写了只会误导后来人。

##### 事件都在 `event.detail` 上

| 元素 | 事件 | `detail` |
|---|---|---|
| `switch` / `checkbox` | `change` | boolean |
| `input` / `search-text-field` | `input`、`submit` | string |
| `input` | `focus` / `blur` / `clear` | 无 |
| `button` | `click` | 无（是普通 `Event`） |

注意 **`<flutter-cupertino-input>` 没有 `change` 事件**，只有 `input` / `submit` / `blur`。
要实现「改完即存」而不是每敲一个字符发一次请求，用 `blur`（HTML 回落分支用 `change`）。

> ⚠️ **但 `blur` 没有去重，「blur == 用户改完了一个值」这个前提不成立。** 上游实现是
> `_focusNode.addListener(() { hasFocus ? dispatch('focus') : dispatch('blur') })`
> （`input.dart` 的 `initState`）—— **没有记住上一次的焦点态**，FocusNode 只要在未聚焦状态下
> 发出任何一次通知，就会再派发一个 `blur`。实测表现是十几条内容**完全相同**的保存请求，
> 界面同时发卡。
>
> 所以「改完即存」必须在业务侧加一道「值真的变了才提交」的守卫，而不是去猜 FocusNode 什么时候
> 会通知：
>
> ```js
> let savedFingerprint = null;          // 从服务端读回来之后要登记一次
> function save() {
>   const fp = fingerprint();           // ← 必须和提交体用**同一个**规范化函数
>   if (fp === savedFingerprint) return Promise.resolve(null);
>   savedFingerprint = fp;
>   return api.save(state);
> }
> ```
>
> 指纹与提交体共用规范化函数这点别省：若指纹按用户输入的 `-4` 算、提交体规范化成 `0`，
> 两者永远不等，守卫就退化成「每次都提交」。

##### `<webf-list-view>` 的两条硬约束

```html
<webf-list-view shrink-wrap="false" scroll-direction="vertical">
  <div class="row">…</div>   <!-- 每一项必须是**直接子节点** -->
  <div class="row">…</div>
</webf-list-view>
```

- **列表项必须是直接子节点** —— Flutter ListView 靠此做 view 回收。中间套一层
  `<div>` 会让 Flutter 只看到一个子节点，回收失效。
- **`shrink-wrap` 默认是 `true`，通常必须显式设为 `false`。** true 时列表高度等于内容总高、
  不在内部滚动，几百行会一路撑下去。设为 false 后必须给它**确定的高度**（例如
  `height: clamp(240px, calc(100vh - 420px), 720px)`）—— 不能留无界约束，WebF 在无界约束下
  解析 flex 会触发 `Infinity or NaN toInt`。

属性只有这两个；事件是 `refresh` / `loadmore`，方法是 `finishRefresh()` / `finishLoad()` /
`resetHeader()` / `resetFooter()`。**如果接了 `loadmore` 就必须调 `finishLoad('success' |
'noMore' | 'fail')`**，否则加载指示器会永久转圈。不做分页就别接这两个事件，少一个失败面。

##### 行内布局用 flex，不要用 grid

`<webf-list-view>` 里的行用 flex 排列单元格、列宽用 CSS 变量在表头与行之间共享。**不要用
grid** —— 会撞上「grid `auto` 行高在 min-content 宽度下测量」那个缺陷（见上一节）。单元格
仍然建议 `white-space: nowrap` + 省略号 + `title` 属性放全文。

##### 参考实现

官方插件 [songloft-plugin-downloader](https://github.com/songloft-org/songloft-plugin-downloader)
（有 `frontend/` 目录的那一版起）是这套做法的完整样例：Vue 3 + Vite，`frontend/src/ui/` 是包装层、
`frontend/src/engine.js` 是特性探测、`frontend/src/ui/native-props.js` 是布尔属性的命令式绑定。

### 访问路径

安装后，静态文件通过以下路径访问（注意：运行时路由是单数 `jsplugin`，与管理 API `/api/v1/jsplugins`（复数）不同）：

```
GET /api/v1/jsplugin/{entryPath}/                 → static/index.html（自动注入）
GET /api/v1/jsplugin/{entryPath}/static           → static/index.html
GET /api/v1/jsplugin/{entryPath}/static/<file>    → 任意静态资源
GET /api/v1/jsplugin-assets/*                     → 主程序公共资源（CSS/JS/字体）
```

### 注意事项

- 静态文件在安装时从 ZIP 解压到 `data/jsplugins_data/{entryPath}/static/`
- 更新插件时会重新解压静态文件
- 建议使用相对路径引用插件 API
- 公共资源由主程序提供，插件不需要也不应该打包自己的 CSS 变量/字体/API 工具库

---

## 9. 安全机制

### 双层 Hash 校验

插件系统使用两层 SHA256 校验保护代码完整性：

1. **Layer 1 — ZIP Hash**：整个 ZIP 文件的 SHA256
2. **Layer 2 — Entry Hash**：入口文件（main.js）内容的 SHA256

#### 校验流程

```
加载插件时：
1. 计算 ZIP 文件 SHA256 → 与数据库中的 zip_hash 比对
2. 若不匹配：
   - 检查文件 mtime 是否变化
   - mtime 未变 = 文件被篡改 → 拒绝加载
   - mtime 已变 = 合法更新 → 允许并更新 hash
3. 从 ZIP 内存中读取 main.js（不落盘）
4. 计算 main.js SHA256 → 与 entry_hash 比对
5. 若不匹配且 ZIP hash 未变 → 拒绝（内部篡改）
```

### main.js 不落盘

入口文件从 ZIP 直接读入内存，不写入磁盘文件系统，减少被篡改风险。

### 权限隔离

- 每个插件声明权限，运行时严格校验
- 未声明权限的 API 调用会被拒绝
- QuickJS 虚拟机提供运行时隔离

---

## 10. 打包发布

### 打包步骤

```bash
# 1. 确保目录结构正确
my-plugin/
├── plugin.json
├── main.js
└── static/
    └── index.html

# 2. 进入插件目录
cd my-plugin/

# 3. 打包为 ZIP（文件在根级别，不含父目录）
zip -r ../my-plugin.jsplugin.zip plugin.json main.js static/

# 4. 验证 ZIP 结构
unzip -l ../my-plugin.jsplugin.zip
# 应该看到:
#   plugin.json
#   main.js
#   static/index.html
```

### 文件命名

ZIP 文件名格式：`{entryPath}.jsplugin.zip`

系统会从文件名提取 entryPath：`my-plugin.jsplugin.zip` → `my-plugin`

### 安装方式

1. **开发模式（推荐）**：`songloft-plugin dev` 在本地迭代，参见 [§2.6](#26-开发模式详解-songloft-plugin-dev)
2. **UI 上传**：通过 Songloft 客户端的设置页面 → 插件管理上传 ZIP
3. **目录放置**：将 ZIP 放入服务器的 `data/jsplugins/` 目录，服务启动时自动发现
4. **API 上传**：`POST /api/v1/jsplugins/upload`，multipart 字段名 `file`（开发模式底层即此接口）

### 更新已有插件

- 重新上传同 `entryPath` 的新版本 ZIP 即可（`/upload` 端点同时处理新装与覆盖更新，由后端用响应状态码 `201` / `200` 区分）
- 也可显式调用 `PUT /api/v1/jsplugins/{id}` 上传新 ZIP
- 或直接替换 `data/jsplugins/` 目录中的 ZIP 文件

无论哪种方式，原插件若处于 `active` 状态，更新成功后后端会自动触发热重载。

---

## 11. 热更新

插件支持运行时更新，无需重启 Songloft 服务。

### 热更新流程

```
1. 检测到 ZIP 文件变化（mtime 改变）
2. 冻结当前服务（停止接收新消息）
3. 调用 onDeinit() 回调
4. 销毁旧的 QuickJS 虚拟机
5. 从新 ZIP 重新加载代码
6. 创建新的 QuickJS 虚拟机
7. 调用 onInit() 回调
8. 解冻服务，恢复消息处理
```

### 自动检测

系统每 30 秒轮询 `data/jsplugins/` 目录，检测 ZIP 文件 mtime 变化。若检测到变化，自动触发热更新。

### 手动触发

目前未提供独立的 `reload` 端点。重新触发热更新的常用做法：

- **开发期**：保持 `songloft-plugin dev` 运行，保存源码即可；
- **运维**：重新上传同 `entryPath` 的 ZIP（`POST /api/v1/jsplugins/upload`）或调用 `PUT /api/v1/jsplugins/{id}`，后端在更新成功后会自动对处于 `active` 状态的插件触发热重载；
- **远程更新**：调用 `POST /api/v1/jsplugins/{id}/update` 拉取 `updateUrl` 中的新版本，同样会自动热重载。

### 错误回滚

如果新版本加载失败，系统会尝试回滚到旧版本。若回滚也失败，则将插件标记为 `error` 状态。

### 注意事项

- 热更新期间，正在处理的请求会完成后再切换
- 定时器和存储状态在热更新后需要重新初始化
- 建议在 `onInit()` 中恢复必要状态

---

## 12. 最佳实践

### 性能建议

1. **避免长时间阻塞** — `onHTTPRequest` 应快速返回
2. **合理使用定时器** — 定时器回调在独立线程中执行，不阻塞 HTTP 请求。但回调中的 `fetch` 等网络操作仍会占用 VM 锁，建议避免在单次回调中执行多个串行网络请求
3. **缓存计算结果** — 使用 `songloft.storage` 缓存频繁访问的数据
4. **控制响应体大小** — 避免返回过大的 JSON 响应
5. **定时器间隔** — 建议 `setInterval` 间隔不低于 1 秒；系统每 500ms 检查一次到期定时器

### 错误处理

```javascript
function onHTTPRequest(req) {
    try {
        // 业务逻辑
        var data = processRequest(req);
        return {
            statusCode: 200,
            body: JSON.stringify(data),
            headers: { "Content-Type": "application/json" }
        };
    } catch (e) {
        songloft.log.error("Request failed: " + e.message);
        return {
            statusCode: 500,
            body: JSON.stringify({ error: e.message }),
            headers: { "Content-Type": "application/json" }
        };
    }
}
```

### 版本管理

- 遵循语义化版本（SemVer）
- 在 `plugin.json` 中设置 `updateUrl` 支持远程更新检查
- 重大变更时更新主版本号

### 开发调试

1. 查看服务器日志中 `[plugin]` 前缀的输出
2. 使用 `songloft.log.info/warn/error` 输出调试信息
3. 健康检查失败会在日志中记录

### 存储使用模式

```javascript
// 存储复杂对象（storage 自动 JSON 序列化，直接存对象即可）
async function saveConfig(config) {
    await songloft.storage.set("config", config);
}

async function loadConfig() {
    var config = await songloft.storage.get("config");
    return config || { defaultKey: "defaultValue" };
}
```

### 插件间协作模式

```javascript
// 服务提供者模式
songloft.comm.onMessage("get-service", function(payload, from) {
    switch (payload.method) {
        case "translate":
            return { text: translate(payload.text) };
        case "summarize":
            return { summary: summarize(payload.text) };
        default:
            return { error: "unknown method" };
    }
});

// 服务消费者模式
async function useTranslation(text) {
    var resp = await songloft.comm.call("translator-plugin", "get-service", {
        method: "translate",
        text: text
    }, 5000);
    if (resp.success && resp.data) {
        return resp.data.text;
    }
    return text; // fallback
}
```

---

## 附录：完整示例

参见 [plugin-toolchain/examples/basic](https://github.com/songloft-org/plugin-toolchain/tree/main/examples/basic) 目录，包含基于官方工具链的完整示例插件代码。
