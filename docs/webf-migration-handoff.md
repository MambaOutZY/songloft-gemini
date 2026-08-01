# WebF 插件页渲染迁移 — 交接文档（songloft-org/songloft#341）

> **这是分支临时交接件，不是产品文档。**
> 刻意**不做**中英双语同步、也刻意从文档站导航里排除（见 `docs/.vitepress/config.mts` 的
> `srcExclude`）。#341 落地后请连同这份文档一起删除。
>
> 面向对象：接手这条分支继续做的下一个 agent / 开发者。
> 最后更新：2026-08-02。

---

## 0. 一句话现状

WebF 渲染引擎已经**能在真机上跑起来插件页**（运行时开关，默认仍是系统 WebView），
排错闭环已建立（页面 JS 错误与 console 会进客户端日志），
「垫片层（JS 侧）+ 自定义元素（Dart 侧）」两套补缺机制的**框架已就位并各有一个已验证的实例**。
剩下的是**按缺口逐项补组件**（Step 3–6）+ 上游报 bug + 许可文档。

---

## 1. 分支与推送状态

三个仓库都在独立分支上，**都还没合进 main**：

| 仓库 | 分支 | 本地 HEAD | 远端 HEAD | 状态 |
|---|---|---|---|---|
| 父仓库 `songloft` | `worktree-webf-plugin-render` | `629d39a` | `03b7606` | **本地领先 2 个提交**，且 `songloft-player` 指针待 bump（`git status` 显示 ` M songloft-player`） |
| `songloft-player` | `feat/webf-plugin-render` | `9c56833` | `4fe5d16` | **本地领先 2 个提交** |
| `plugin-toolchain` | `feat/webf-transport-doc` | `a6fa726` | 已推 | 干净 |

工作目录：`/home/ejoydev/work/mimusic/.claude/worktrees/webf-plugin-render`（git worktree，
**不要 `cd` 回主仓库根**）。

**下一步要做的第一件事**：把 Step 1 / Step 2 的 4 个提交推上去（player 先推，再在父仓库
`git add songloft-player` bump 指针并提交，然后推父仓库）。

### ⚠️ 合并前必须撤掉的临时改动

`.github/workflows/release.yml`（父仓库）与 `songloft-player/.github/workflows/build-and-release.yml`
被改成发布到独立 tag **`dev-webf`**，为的是不覆盖共享的 `dev`（那是所有人在用的滚动 dev 版）。
相关提交：父 `2f8c804`、player `c775382`，提交信息里已写明"合并前必须撤掉"。

实现细节（避免误改）：新增的 `release_tag` output 与既有的 `version` / `version_tag`
**是三个不同的东西** —— 后两者驱动 `-tags dev` 与 `FRONTEND_VERSION`，不能一起改；
另外顺手修了两处硬编码的 `dev`（补丁 manifest 的 `TAG`、`version.json` 的
`download_url_prefix`），**那两处是真 bug，撤销临时改动时要保留修复**。

已核实：共享的 `dev` tag 未被本分支污染（其 release 时间仍是 2026-05-29）。

---

## 2. 已完成（按提交顺序，含"为什么"）

### 2.1 基础设施

| 提交 | 内容 |
|---|---|
| player `878ac08` | **WebF 验证容器**（`scripts/webf-verify/`）。本机跑不了 WebF，这是唯一验证途径，用法见 §5 |
| player `41c67f1` | 加 `webf: ^0.24.27` 依赖 + `NOTICE` 声明 GPL-3.0。提交带 `!`，因为改了 `GeneratedPluginRegistrant` → **原生契约哈希变化 → 本次必须整包发版**（见 player `AGENTS.md` 的 Kotlin 冻结规则第 5 条） |
| player `1656d91` | **抽出引擎无关的渲染层** `lib/features/home/presentation/render/`。顺带消除了 `plugin_tab_page_native.dart` 与 `plugin_webview_page_native.dart` 各持一份的 token 注入脚本 / 20s 超时 / 错误 UI / `_reloadSeq` 重复代码。这一步**行为零变化** |
| player `f559130` | WebF 渲染面 + 图标字体预注册 |
| player `d03007d` | 运行时开关 pref `plugin_render_engine`（默认 `webview`），设置页「扩展」段的 `SegmentedButton`，`kIsWeb` 时隐藏 |
| 父 `1fe44f8` | `common.js` 加**第三条**宿主传输（WebF methodChannel）+ `requestBack` 回报 |
| toolchain `a6fa726` | client-sdk 传输列表文档同步 |

### 2.2 排错修复（每一条都是踩过的坑）

| 提交 | 修的什么 |
|---|---|
| 父 `ffdf69b` | **`common.js` 不再用 `dataset` 写主题**。WebF 里 `<head>` 阻塞脚本执行期 `document.documentElement.dataset` 是 `null` → `TypeError` → **整个 IIFE 中断** → `window.SongloftPlugin` 压根不存在。改成 `setAttribute('data-theme', th)` / `getAttribute`，调用点再包 try/catch。这是"主题不生效"的**真根因**（我先后追了 4 个假设全部 PASS，最后靠页面内省才抓到） |
| player `76ba993` | **插件页 URL 补尾斜杠**（`ensurePluginPathTrailingSlash`）。WebF **不采纳 `<base href>`**，导致 `app.bundle.<hash>.js` 被解析成 `/api/v1/jsplugin/static/js/...`（少一层）→ 401 → 白屏 |
| 父 `2b14a17` | **服务端剥掉空 `img src`**（`stripEmptySrcAttrs`，`internal/jsplugin/routes.go`）。`<img src="">` 在 WebF 里会解析成文档自身 URL 去请求，拿到 HTML 后报 `Failed to decode image (mime=text/html)`。**必须服务端做**：静态 HTML 的图片加载发生在解析期，任何 JS 垫片都来不及 |
| player `4fe5d16` | 关掉 WebF 的 HTTP 缓存（`enableHttpCache: false`）+ **转发页面 JS 错误与 console 到 `debugPrint`**。后半段才是关键：`Bytecode are not valid to execute.` 是**次级症状**（前面有未捕获 JS 异常污染了 QuickJS 上下文），它既不带 URL 也不带原始异常，没有 `onJSError` / `onJSLog` 转发根本无法归因 |
| player `17c736d` | 探针加跨表 CSS 变量用例 + 页面内省能力（`DIAGNOSE=1`） |

### 2.3 补缺机制（Step 1 / Step 2）

**Step 1 — JS 垫片层**（父 `28fe3f0` + player `d92b915` 回归用例）

`internal/jsplugin/assets/common.js` 里建了完整的垫片框架：

```
isWebFEngine()              引擎探测，所有垫片都关在这个分支里
runShims(shims, phase)      逐个 try/catch，一个失败只影响它自己（console.warn）
collectByTag(tag)           querySelectorAll → 退回 getElementsByTagName，快照成数组
earlyShims = [...]          installEarly()  立即执行（<head> 阻塞期，<body> 还没解析）
readyShims  = [...]         applyOnReady()  DOMContentLoaded 后执行
SongloftPlugin.applyShims   导出，供插件动态插入 HTML 后手动重跑（幂等）
```

**边界判据（写在源码注释里，后续加垫片时照它判）**：
「这个缺口会不会在**解析期**就产生不可撤销的副作用（发请求 / 起解码）」
→ 会 → 只能服务端改 HTML（如空 `img src`）；不会 → 放 `applyOnReady`。

同时 `installEarly()` 给 `<html>` 加 class **`webf-engine`**，供插件按引擎分叉。

已实现的垫片：`emptyImgSrcAccessorShim`（early，属性访问器兜底）、`emptyImgSrcSweepShim`（ready）、
`detailsShim`（ready，`<details>`/`<summary>` 折叠）。

`detailsShim` 的关键决策：**保留原标签**（`<details>`/`<summary>` 不换名），
因为插件是按标签名 `querySelector` 的；把非 `<summary>` 子节点包进 `.sl-details-content`；
用 Material Symbols 连字（`arrow_drop_down` / `arrow_right`）画箭头；
在实例上定义 `open` 访问器；派发 `toggle` 事件；`data-sl-details-shim` 标记保证幂等。

**Step 2 — Dart 自定义元素**（player `9c56833` + 父 `629d39a` 文档）

`lib/features/home/presentation/render/elements/`
- `songloft_custom_elements.dart` — 幂等注册入口 `ensureRegistered()`，逐元素 try/catch
- `songloft_progress_ring.dart` — `<songloft-progress-ring>`，属性
  `value/min/max/stroke-width/color/track-color/line-cap`，`CustomPaint` 绘制，
  颜色跟 CSS `color`（currentColor）从而跟随主题

**这个目录只允许依赖 `flutter` 与 `webf`** —— 因为验证探针（另一个 package）会把它整目录拷进去编译，
拷不动产品的其他依赖。约束写在两边的头注释里，`analysis_options.yaml` 也为此排除了 `probe_main.dart`。

---

## 3. WebF 硬约束与已确诊缺陷（**接手前必读，能省掉大量返工**）

### 3.1 API 事实（都被我踩过一次）

- `WebF(...)` **不能直接 new** —— `WebF._` 是 `@protected`。唯一公开挂载入口是静态方法
  `WebF.fromControllerName(controllerName:, bundle:, createController:, loadingWidget:, errorBuilder:)`
- `controller.onJSLog` 是**字段不是构造参数**，只能构造后赋值
- `controller.view.evaluateJavaScripts(code)` 返回 `Future<void>`，**拿不到返回值**
- `WebFControllerManager` 两级限额语义不同：超 `maxAttachedInstances` 只 **detach**（状态保留），
  超 `maxAliveInstances` 才 **dispose**（重挂会自动重建但 JS 状态归零 + 闪 loading）。
  产品取 `maxAliveInstances: 8, maxAttachedInstances: 3`
- `WebF.defineCustomElement(tag, creator)` **要求标签名带连字符**（首字符 a-z + 至少一个 `-`），
  且**重复注册抛异常**，且必须早于任何 controller 创建。
  → **无法用它覆盖 `<svg>`/`<audio>`/`<table>`/`<input>` 这类内建标签**，所有原生组件只能是新标签，
  也就是「插件必须显式改用我们的标签」
- `WidgetElement` 子类要实现 **`WebFWidgetElementState createState()`**（不是 `build(BuildContext)`）；
  基类会在 `attributeDidUpdate` **之前**调 `requestUpdateState()`；可用
  `WebFRenderWidgetAdaptor` 承载 HTML 子节点
- WebF 无 `canGoBack`、controller 上也无 `goBack()` → 返回键靠反向 methodChannel 问页面
  （`common.js` 的 `requestBack` handler 判 `history.length`）
- `window.postMessage` 在 WebF 是**同窗口自发自收** → `common.js` 的接收侧一行都不用改

### 3.2 已确诊的 WebF 上游缺陷（task #12 待报）

1. **`css/font_face.dart:396`** URL 加载分支用 `bundle.data!.buffer.asByteData()`（取**整个底层
   buffer**、无视视图 offset/length），而 `data:` 分支 `:375` 用的是正确的
   `ByteData.sublistView(content)`。表现：TTF 拿到 HTTP 200 但仍是豆腐块。
   另外 `supportedFonts = ['ttc','ttf','otf','data']`（**woff2 不支持**），
   且 format 是从 **URL 扩展名**推断的、完全无视 CSS 里的 `format()` 声明，挑不到源就静默 `return`
   （不发请求、不打日志）。
   → 我们的绕法：Flutter 侧 `FontLoader` 预注册（`plugin_render_fonts.dart`），
   只注册 `Material Symbols Outlined`；**刻意不注册 Roboto**（那会用 37 KB 的拉丁子集
   全局覆盖 Flutter Material 的默认字体）
2. **`<base href>` 不被采纳**
3. **`script.dart:66` 的 isBytecode 分支没有回退** —— 同为字节码执行的 `to_native.dart:397`
   有「失败即删缓存、退回原始 JS」的自愈，前者没有，于是脚本静默不执行
4. **`Bytecode are not valid to execute.` 不带任何归因信息**（无 URL、无原始异常）
5. **`documentElement.dataset` 在 `<head>` 阻塞脚本期是 `null`**

### 3.3 缺口清单与真实命中面（已交叉验证）

评估必须基于**构建产物**（builder 用 esbuild 打成 IIFE / es2020，会把
`<script type="module">` 改写成普通 `<script>`），不是 `jsplugins-src/*/static/` 源码。

| 缺口 | 命中的插件 | 现状 |
|---|---|---|
| `<table>` 元素**根本不存在**（退化成嵌套 `display:block`） | downloader、radio | **Step 4 待做** |
| `input[type=range]` 静默变文本框 | **仅** miot（2 处） | **Step 3 待做** |
| `env(safe-area-inset-*)` 不求值（`style_declaration.dart:736`，upstream #907 open） | miot 3、dav 3、subsonic 1、cloudflared 1、hostc 1、ytdlp 1 | **Step 5 待做** |
| `window.open` 是 no-op（`window.cc:157-168` 两个重载都 `return this`，不抛错） | **仅** miot `js/auth.js:95`（小米账号二次验证） | **Step 6 待做** |
| `input[type=file]` 静默变文本框 | ytdlp、radio、lxmusic 各 1 | **Step 6 待做** |
| `URL.createObjectURL` 不存在（`Blob` 有，无入口产 `blob:`） | **仅** miot `js/playback.js:422`、`js/fullscreen-player.js:201` | **Step 6 待做** |
| `<details>`/`<summary>` | lxmusic | ✅ Step 1 已垫 |
| 内联 `<svg>` 无真实 box（整棵子树重新序列化交给 `flutter_svg`，高频更新性能最差） | lxmusic 进度环 | ✅ Step 2 提供了 `<songloft-progress-ring>` 替代（插件需自己改用，宿主不自动替换） |
| `<audio>` | lxmusic | **刻意推迟** —— 要下沉到宿主播放器，是 UX 变更，且 lxmusic 未构建发布 |
| `backdrop-filter` / `mask-image` / `color-mix()` 不渲染 | dav、lxmusic、miot（歌词渐隐用 `mask-image`）、subsonic | **接受降级**（已定，不做） |
| `getComputedStyle` 不暴露自定义属性；元素**属性值里的 `var()` 不展开** | — | 已写进插件开发文档 |
| `HTMLElement.click()` 是异步的；内联 `style.xxx = ''` 不可靠 | — | 已记录 |

**已排除的伪阻塞**（省下大量工作，别再去查）：
CSS Grid **已实现**（experimental，193 KB 实现，issue 原文写"不支持"是错的）；
`matchMedia` 3 处全是一次性读 `.matches`、**没有** `addEventListener('change')`，所以 WebF
只有旧式 `addListener` 的缺陷不影响本项目；token 注入不需要 pre-inject API（已走 `?access_token=`
+ 后端内联 `authBridgeScriptTpl`）；`gap` / `-webkit-line-clamp` / `aspect-ratio` /
`navigator.clipboard.writeText` / `NodeList.forEach` / `WebSocket` / `localStorage` /
`select` / `textarea` / `input[type=time]` 全部可用。

---

## 4. 未完成（按建议顺序）

### Step 3 — `<songloft-slider>` 替换 `input[type=range]`（task #15，**下一个要做的**）

- **命中面极小：只有 miot 2 处。** 别为它做通用化过度设计
- 做法：Dart 侧 MD3 风格 `<songloft-slider>` 自定义元素 + `common.js` 里的 ready 垫片扫
  `input[type=range]`
- **强烈建议：不要把原 `<input>` 从 DOM 里移除。** 把它**隐藏**并与 `<songloft-slider>`
  双向同步 —— 这是 Step 1 的教训：插件按标签名 `querySelector` 并直接读 `.value`，
  换掉标签会静默打断插件自己的 JS
- 必须派发 **`input` 与 `change` DOM 事件**（这是本项目第一次真用 `dispatchEvent`，
  **需要实测证明 JS 侧确实收到**，不能假定）
- 验收：容器里渲染 miot 真实页面（起本机后端，见 §5），拖动滑块后**在后端可观测状态上断言**
  （音量/进度真的变了），不能只看截图

### Step 4 — `<table>` 垫片（task #16）

- 命中 downloader、radio
- WebF **自带** `<webf-table>` / `<webf-table-header>` / `<webf-table-row>` / `<webf-table-cell>`，
  垫片做的是**标签改写**而非从零实现
- 注意 radio 还叠了 sticky 表头

### Step 5 — `env(safe-area-inset-*)` 垫片（task #19）

- Dart 侧读 `MediaQuery.padding`，经桥注入 CSS 变量（如 `--sl-safe-top`），
  垫片把 `env(safe-area-inset-top)` 改写成 `var(--sl-safe-top)`
- 命中 6 个插件共 10 处，是**收益/成本比最高**的一项

### Step 6 — 三项经桥下沉（task #18）

- `window.open` → Dart `url_launcher`（miot 小米二次验证登录）
- `input[type=file]` → Dart `file_picker`（ytdlp、radio）
- `URL.createObjectURL` → `data:` URL 或落盘（miot 带鉴权头拉封面 ×2）

### task #3 — 许可合规文档

`songloft-player/NOTICE` **已完成**（新增「DISTRIBUTION LICENSE — PLEASE READ」段 + WebF 列为第 1 项，
其余重新编号到 8 含 Material Symbols）。**剩下**：README / docs 的许可章节，
且必须遵守**文档双语同步铁律**（`README.md` ↔ `README.en.md`）。
另外 GPLv3 强制要求 release 产物随附 GPL-3.0 全文与「完整对应源码」获取方式，**这条还没做**。

### task #12 — 给 WebF 上游报 §3.2 里的 5 条

### 明确不做（用户已定）

- **按插件排除引擎**（在 `plugin.json` 里标记某插件不用 WebF）—— 用户明确说「不按插件排除」
- **Web 端迁移** —— WebF 不支持 Flutter Web（39 处无条件 `import 'dart:ffi'`，是编译失败非降级），
  iframe 路径**永久保留**，渲染路径 2 → 3 条
- **Linux 插件页缺口** —— WebF Linux 仅 x86-64 + glibc ≥ 2.38、无 arm64，覆盖不到 NAS /
  Debian 12 / 树莓派。用户定「只记录，不在本次处理」
- App Store / Google Play（GPL 版不能上）与「保持 Apache-2.0 分发二进制」—— 用户已明确放弃两条

### Phase 3（翻默认之后才做，现在别动）

默认切 `webf` 并观察一个版本周期后，才可以删 native 侧的 platform-view hack。
删每一个之前**先确认原 issue 根因确实点名 platform view**：
`core/utils/webview_environment.dart` 整文件（#271）、`window_visibility.dart` 的 HWND 卸载链路（#293）、
`useHybridComposition: false`（#273）、`shell_layout.dart` 的 `isNativeDesktop` 切走即销毁分支（#246）。

**⚠️ 不要动 Web 侧的**：`plugin_iframe_diagnostics.dart`（#278）、
`core/a11y/web_semantics_controller.dart` + `semantics_pointer_override_web.dart`（#295）、
两个 `_stub.dart`。这些服务于永久保留的 iframe 路径，删了就是回归。

---

## 5. 验证环境（**本机跑不了 WebF，这是唯一途径**）

宿主 glibc **2.35** < WebF 要求的 2.38，且缺 `clang++`/`ninja`、无 Android SDK/Xcode。
唯一可用目标 Chrome(web) 恰是 WebF 不支持的平台。

```bash
cd songloft-player
./scripts/webf-verify/run.sh                    # 跑探针，产出 out/probe.png
./scripts/webf-verify/run.sh --build            # 强制重建镜像（首次 10-20 分钟）

# 环境变量
FONT_FIX=1|2|3   字体修复方案对照（1=woff2+ttf 双 src, 2=base64 data:, 3=Flutter FontLoader）
DIAGNOSE=1       页面加载完注入内省脚本，输出经 console 回传到 out/flutter.log
HOST_NETWORK=1   容器用 host 网络 —— 渲染**真实插件页**（PROBE_URL 指向宿主上的 Go 后端）时必须置
PROBE_URL=...    换渲染目标
SETTLE=<秒>      抓屏前等待
```

产出在 `songloft-player/scripts/webf-verify/out/`：`probe.png`（截图）、`build.log`、
`flutter.log`、`codepoints.txt`、`env.txt`、`elements.sha1`。

**探针页 `probe.html`** 现在是**两列布局**、13 个检查组 + `13b` 文本断言，每组的通过/失败判据
都写在它自己的注释里。`entrypoint.sh` 会把产品的 `render/elements/` 拷进探针 `lib/elements/`，
**源目录缺失直接 `exit 1`**，并把 sha1 写到 `out/elements.sha1`（保证测的是产品那一份）。

### 验证真实插件页的配方

```bash
# 起本机后端（lite 版够用，省掉前端构建）
go build -tags "dev lite" -o /tmp/songloft-webf .
/tmp/songloft-webf -port 58191 -db /tmp/webf/test.db -music <musicdir>

# 渲染插件页，注意 URL **必须带尾斜杠**（WebF 不采纳 base href）
HOST_NETWORK=1 PROBE_URL='http://127.0.0.1:58191/api/v1/jsplugin/miot/?embed=&theme=dark&access_token=<t>' \
  ./scripts/webf-verify/run.sh
```

### 验证环境自身的坑

- **`docker build ... && run.sh; echo done`** —— build 失败时 `&&` 短路但 `echo` 照样打印
  「done」，然后你读到的是**上一轮的旧截图**。我为此误判了两次。**分开跑，检查 build 退出码**
- **`DIAGNOSE`** 最终落到 Dart 的 `bool.fromEnvironment`，它**只认字面 `"true"`**，
  传 `1` 会被静默当 false（`run.sh` 已做归一化）
- **两列布局的纵向预算有限**，加检查组前先确认新行不会把已有行挤出截图
- **`[diag]` 在 Step 2 那轮跑没有输出**，原因未查明（既有探针问题，不影响 Step 1/2 的结论）
- 断言铁律（与仓库既有的无头浏览器验证一致）：**截图只证明"渲染对了"**，
  交互是否真生效必须落在后端可观测状态上（`curl` 对应 `/settings/<name>`、`play_history` 有无新记录等）。
  数进程用 `pgrep -x`，**不要** `ps -ef | grep | wc -l`

---

## 6. 已知未解问题

- **lxmusic 在 WebF 下有布局崩溃**：`Unsupported operation: Infinity or NaN toInt`、
  `Null check operator used on a null value`。lxmusic 未构建发布、也不在本分支的 `.gitmodules` 里
  （本分支只跟踪 miot / subsonic / cloudflared / dav / hostc / registry / downloader / lyrics / radio，
  **lxmusic / bili / ytdlp 都不是跟踪的子模块**，要验证它们得先自己 clone）
- **`<details>` 垫片的一个已知边界**：垫片跑完之后再给 `<details>` 追加直接子节点，
  那个节点会永久留在折叠容器外面（幂等标记会阻止重新包裹）。插件应在插完 HTML 后调
  `SongloftPlugin.applyShims()`，但对"追加单个子节点"这种用法无解
- miot `index.html:1378` 引外部 CDN `marked.min.js`（builder 不打包），离线/内网下 Markdown
  渲染静默失效 —— **与 WebF 无关的既有问题**，顺手记一笔
- **上游风险**：WebF 0.x beta，main 分支自 2026-04-19 静默至今，30 天下载量 1172，
  最后 9 个版本几乎全是 flex/inline 布局的正确性修复 → 这是**必须保留运行时回退开关**的直接理由

---

## 7. 铁律速查（本分支相关）

- `internal/jsplugin/assets/common.js` 服务给**所有**客户端版本与普通浏览器 →
  改动必须**纯增量 + 特性探测**，且 `isWebFHost()` 判定要排在 `isNativeHost()` / `isIframeHost()` **之前**
- `render/elements/` 只能依赖 `flutter` + `webf`（验证探针要跨 package 拷它）
- 改 Dart 后 `cd songloft-player && dart format lib/ test/`；改 Go 后根目录 `gofmt -w .`
- 提交**禁止** `Co-Authored-By`；子仓库引用父仓库 issue 必须写完整路径 `songloft-org/songloft#341`
- 子模块改动流程：子仓库提交 → 回主仓库 `git add <path>` bump 指针 → 主仓库提交
- 本仓库 worktree 的 git stash 栈与主 checkout 共享 → **禁止**裸 `git stash` / `git stash pop`
