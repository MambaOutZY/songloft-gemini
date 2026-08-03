# WebF 插件页渲染迁移 — 交接文档（songloft-org/songloft#341）

> **这是分支临时交接件，不是产品文档。**
> 刻意**不做**中英双语同步、也刻意从文档站导航里排除（见 `docs/.vitepress/config.mts` 的
> `srcExclude`）。#341 落地后请连同这份文档一起删除。
>
> 面向对象：接手这条分支继续做的下一个 agent / 开发者。
> 最后更新：2026-08-02（Step 5 完成、Step 4 方案重设计、第 7 条上游缺陷、许可合规主体完成）。
>
> **配套的三份临时件**（同样从文档站排除、同样在 #341 落地后一起删）：
> `docs/webf-recon-step456.md`（Step 4/5/6 预研，证伪了两条既定方案）、
> `docs/webf-step4-design.md`（Step 4 四方案对比与选型）、
> `docs/webf-upstream-issues.md`（7 条上游缺陷草稿）。

---

## 0. 一句话现状

WebF 渲染引擎已经**能在真机上跑起来插件页**，选择方式已从「客户端全局运行时开关」改为
**每个插件在 `plugin.json` 里用 `renderEngine` 声明**（默认仍是 `webview`，见 §2.4 —— 这是一次
**用户决策反转**，改之前的文档写的是「不按插件排除」）。
排错闭环已建立（页面 JS 错误与 console 会进客户端日志），
「垫片层（JS 侧）+ 自定义元素（Dart 侧）」两套补缺机制的**框架已就位并各有一个已验证的实例**。
**已完成**：Step 1（JS 垫片框架）、Step 2（Dart 自定义元素框架 + `<songloft-progress-ring>`）、
Step 3（`<songloft-slider>`）、Step 5（安全区 `--sl-safe-*`，见 §2.5）、许可合规（GPL-3.0 全文
随 release 产物 + `NOTICE`）。
**进行中**：Step 4（`<table>` → CSS Grid，方案已重设计）、Step 6（三项经桥下沉）、
上游 7 条缺陷报告起草。

### ⛔ 范围硬边界（用户 2026-08-03 明确划定）

> **只处理 miot / downloader / lyrics 三个插件。其他插件的问题暂时一律不处理。**

这不是「优先级低」，是**不做**。已据此停掉一项在跑的工作（lxmusic 的 WebF 布局崩溃根因分析），
并删掉了它的产物。**不要**因为在下面看到某个插件的缺口就去动它。

**这条边界的一个反直觉后果，务必读懂**：本分支已经建好、已测、已写文档的若干能力，
**现在没有任何在范围内的消费者**：

| 已建好的东西 | 原定消费者 | 现状 |
|---|---|---|
| `input[type=file]` 桥 + 垫片（Step 6） | radio、ytdlp、lxmusic | **在范围内的消费者为零** |
| `<songloft-progress-ring>`（Step 2） | lxmusic 的进度环 | 同上（宿主本就不自动替换，需插件自己改用） |
| `<details>` / `<summary>` 垫片（Step 1） | lxmusic | 同上（自动生效，留着无害） |
| `<audio>` 下沉 | lxmusic | **永久搁置**（原本只是「刻意推迟」） |

**这些已落地的东西不要删** —— 已提交、已测、已进双语文档，留着零成本，
而且一旦将来把某个插件纳入范围就能直接用。但**也不要再去「补完」或扩展它们**
（例如不要为了给 `input[type=file]` 找个用户而去改 radio）。

同理，§3.3 缺口清单里那些只命中范围外插件的行（radio 的 `<table>`、dav / subsonic /
cloudflared / hostc / ytdlp 的 `env()`），**现在都不是待办**。

⚠️ **本文档已发生过多次「旧版结论被证伪 / 被用户推翻」**。三处最容易踩的：
① 引擎选择**不是**全局开关而是逐插件声明（§2.4，用户决策反转）；
② Step 4 **不是**标签改写而是 CSS Grid（§4）；
③ Step 5 **不是**垫片改写 `env()` 而是宿主注入变量（§2.5）。
读到与这三条矛盾的文字时，以被引用的那一节为准。

---

## 1. 分支与推送状态

**六个仓库**参与本次改动。宿主三仓在独立分支上、**都还没合进 main**；三个插件仓**已在 `main` 上
且已发版**（因为插件的 release 工作流从默认分支构建，不发版 `renderEngine` 就不会到用户手里）。

| 仓库 | 分支 | HEAD（截至 2026-08-02 本节更新时） |
|---|---|---|
| 父仓库 `songloft` | `worktree-webf-plugin-render` | `8055506` |
| `songloft-player` | `feat/webf-plugin-render` | `3035b6a` |
| `plugin-toolchain` | `feat/webf-transport-doc` | `8f25a43` |
| `songloft-plugin-miot` | **`main`** | `0e0d945`（**尚未发版**，见下） |
| `songloft-plugin-downloader` | **`main`** | `a99e279`（已发 v2026.8.2） |
| `songloft-plugin-lyrics` | **`main`** | `84d90fe`（已发 v2026.8.2） |

⚠️ **这张表天生会过期，别信它、去跑 `git log`。** 本文档前几版都栽在这里 —— 表里的 HEAD 只反映
到写下它的那一刻，而分支一直在动。留着它只是为了给「哪些仓库参与了」提供索引。

⚠️ **miot 有已提交但未发版的改动**（Step 3 的滑块适配 `e3173a0`、Step 5 的安全区 `0e0d945`）。
release zip 里的 `static/` 还是 v2026.8.2 那份，所以**竖向音量条按默认横向渲染、安全区 3 处仍是
`env()`**。刻意攒到 Step 4/6 做完一并发一次，避免每步各发一版。

工作目录：`/home/ejoydev/work/mimusic/.claude/worktrees/webf-plugin-render`（git worktree，
**不要 `cd` 回主仓库根**）。

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
| player `d03007d` | 运行时开关 pref `plugin_render_engine`（默认 `webview`），设置页「扩展」段的 `SegmentedButton`，`kIsWeb` 时隐藏。**⚠️ 已被后一轮改动整体移除**（开关与 pref 都不再存在），改为逐插件 `renderEngine`，见 §2.4 |
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

### 2.3 补缺机制（Step 1 / Step 2 / Step 3）

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

**Step 3 — 两套机制合用的第一个例子：`<songloft-slider>`**（提交状态以 `git log` 为准）

- Dart 侧 `elements/songloft_slider.dart` —— `<songloft-slider>`，属性
  `value/min/max/step/orientation/disabled/color/track-color`，`CustomPaint` 绘制，
  颜色同样走 CSS `color`（currentColor）。**不用 Material `Slider`** 有两条硬伤理由：它从宿主
  App 的 `Theme` 取色（跟不上插件页的 `--md-*`）；它内部是 `HorizontalDragGestureRecognizer`，
  套 `RotatedBox` 转 90° 后竖向拖动的**全局水平位移是 0**，识别器永不接受 → 拖不动
- JS 侧 `common.js` 的 `rangeSliderShim`（ready 相）—— 扫 `input[type=range]`，在其后插入滑块、
  **隐藏**原 input（`.sl-range-hidden` + inline `display:none`）、遮蔽 `.value` / `.disabled` /
  `matches(':active')`，滑块的 `input`/`change` 转派到原 input（冒泡）。**verified-or-abort**：
  `.value` 访问器装不上（哨兵往返自检失败）就删滑块、还原 input、退回原生表现 —— 刻意不写
  「退化成定时轮询」的第二条路（那条在本环境永远跑不到、也就永远测不到）
- **三道防抖闸**（拖动中不被插件的轮询回写覆盖）：垫片的 `dragging` 标志（带 1500ms 兜底清除）、
  `matches(':active')` 遮蔽、Dart 侧 `_dragging` 时忽略外部 `value`。缺任何一道都会出现
  「手指还没抬起把手就跳回去」
- **手势用与朝向同轴的 drag + `onDown` 定位**，不用 `onTap`（会赢掉 WebF 唯一那个 tap
  recognizer，DOM `click` 就不再派发了）、不用裸 `Listener`（不进竞技场 → 滚动与滑块同时响应）、
  不用 pan（`kPanSlop` = 2×`kTouchSlop`，同轴竞争必输给滚动）
- 插件侧成本：竖向必须写 `data-sl-orientation`（不猜朝向），且要补几行几何 CSS（垫片只拷 inline
  style 不拷 class）。miot 已适配，`data-sl-no-slider` 是退出开关

### 2.4 引擎选择改为逐插件声明（**决策反转，2026-08-02**）

> ⚠️ **这一节记录的是一次用户决策反转。** 本文档更早的版本在 §4「明确不做（用户已定）」里
> 第一条写的是「**按插件排除引擎**（在 `plugin.json` 里标记某插件不用 WebF）—— 用户明确说
> 『不按插件排除』」，配套实现是 player 的**全局运行时开关**（`d03007d`）。
> **用户后来推翻了这个决定**：现在正是「按插件在 `plugin.json` 里声明」，而**全局开关被删掉**。
> 保留这段历史是为了防止后续 agent 照旧文档把设计回滚回去 —— 看到「不按插件排除」字样时，
> 请以本节为准。

**反转的理由**：WebF 是 0.x beta、能力缺口是**逐页面**的（某个插件用了 `<table>` / `input[type=range]`
就在 WebF 下坏，别的插件完全不受影响）。全局开关只能表达「全都用」或「全都不用」，
既让已验证可用的插件被没验证的插件拖住，又要求终端用户去理解「渲染引擎」这个他们不该关心的概念。
逐插件声明把决定权交给**唯一有能力验证的人 —— 插件作者**。

**现在的机制（契约，不要自己改名）**：

- `plugin.json` 字段 **`renderEngine`**，可选，取值 `"webview"` / `"webf"`；**缺失或空串 = `webview`**（宿主默认）
- 插件列表 API 返回 snake_case 的 **`render_engine`**
- 非法取值在后端 **`ValidateManifest`** 阶段报错 → **插件装不上**，不静默回退
- 客户端设置页里原有的全局引擎开关（`plugin_render_engine` pref + `SegmentedButton`）**已删除**
- **Web 端不受该字段影响**：WebF 不支持 Flutter Web（39 处无条件 `import 'dart:ffi'`，是编译失败非降级），
  Web 永远走 iframe 路径
- 三个官方插件（miot 智能音箱、downloader 歌曲下载、lyrics 歌词搜索）本轮标记为 `webf`
- 用户文档：`docs/js-plugin-development-guide.md` §3「renderEngine 渲染引擎声明」+ 英文版同章节
  （双语铁律），§8 的 WebF 章节开头也改成「逐插件选择」的措辞并交叉引用 §3

**风险敞口**：不再有任何全局回退开关。页面在 WebF 下坏掉时，用户的处置只有「禁用该插件」或
「等插件作者发一个把 `renderEngine` 改回 `webview` 的版本」。这是用户明确接受的取舍，见 §6。

**这一条还有一个没预料到的正面副作用**：它把后续每一步的**命中面**都大幅收窄了 —— 只有显式声明
`webf` 的插件才暴露在缺口下。Step 4 因此从「downloader + radio 两个插件」塌缩成「downloader 一个」，
Step 5 从「6 个插件 10 处」塌缩成「miot 3 处」，Step 6 的 `input[type=file]` 直接变成**零暴露**。
评估任何缺口的紧迫性时，**先查 `plugin.json` 的 `renderEngine` 取值**，别照 §3.3 那张表的
「命中的插件」一栏直接下结论。

### 2.5 Step 5 — 安全区 `--sl-safe-*`（**已完成 2026-08-02**）

父 `8055506` + player `3035b6a` + miot `0e0d945`。

**交接文档旧版写的方案（垫片把 CSS 里的 `env()` 改写成 `var()`）已证伪**，两个**独立**死因：

① **CSSOM 没有可用的写入面**：`cssText` 只有 getter，`CSSStyleRule` 既不暴露 `selectorText`
也不暴露 `.style`；唯一写入面是 `insertRule`/`deleteRule`/`replaceSync` 这种「整条规则进出」，
而 `cssText` 是从解析结果**重建**的（简写已展开、WebF 不认识的属性已丢），往返即有损；
`@media` 里的规则 `cssText` 是**空串**且没有 `.cssRules`，delete+insert 会**不可逆地摧毁**它。
另外 `enableBlink: true` 时 `document.styleSheets` 直接返回空。

② **即便能改写也救不了命中面**：miot 的 3 处 `env()` 全都套在 `calc()` / `max()` 里，而
**WebF 没有实现 CSS `max()` / `min()`**（`css/values/calc.dart` 只认 calc 与 clamp），
换成 `var()` 照样是死的。

**现在的机制**：`common.css` 给 `--sl-safe-{top,right,bottom,left}` 备默认值 ——
`:root` 把它们绑到 `env()`（浏览器 / 系统 WebView 拿原生真值，**行为与改动前完全一致**），
`html.webf-engine` 覆盖成确定的 `0px`；宿主再用 `MediaQuery.viewPadding` 的真值经既有的
`_pushToPage` / `window.postMessage` 通道写成 `documentElement` 的**内联**自定义属性覆盖。
插件侧只有一种写法：`var(--sl-safe-bottom)`。

**刻意没照预研建议的「`:root` 直接预置 `0px`」**：那会让浏览器与系统 WebView（**默认引擎**，
绝大多数插件走这条）永久丢掉原生 `env()`，是实打实的回归。

**三条实测事实**（探针第 17 / 17b 组钉住，`out/flutter.log`）：

- **`var(--未定义, env(...))` 求值为 `0`**，连 `env()` 自己的内层兜底都取不到
  （`G(var-fb-env)=0`，而对照的裸 var fallback `F(var-fb)=17` 是通的）
  → **「一份 CSS 带 `env()` 兜底通吃三端」不可用**，必须按引擎给默认值
- **`max()` 是死的**（`C(max)=0`）
- **`clamp()` 的参数里可以塞 `var()`**（`D(clamp+var)=30`、`J(clamp-min)=24`，
  后者证明夹紧真的发生、不是穿透）。⚠️ **这一条与源码判读相反** —— clamp 分支是逐个参数走
  `CSSLength.parseLength`，而它不认 `var(...)`，只有 `calc()` 内部有专门的 `CalcVariableNode`。
  **以实测为准**，`common.css` 的注释里写明了「别照源码把它改回不支持」。
  于是 `clamp(MIN,VAL,MAX) ≡ max(MIN,min(VAL,MAX))` 成了 `max()` 的等价替换，
  miot 的 `.fp-controls` 就是这么改的（**浏览器侧零行为变化**，不需要按引擎分叉）

**未验证**（如实记）：真机刘海（容器里 `MediaQuery.viewPadding` 恒为 0，验的是「注入通道 +
CSS 求值」这一层，注入的是人为非零值）、转屏 / 键盘触发重推、外层 `SafeArea` 不会双重内缩
（源码级结论，`media_query.dart:946-951`）。

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

### 3.2 已确诊的 WebF 上游缺陷（**7 条**，task #12 起草中）

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
6. **`input[type=range]` 那一整行根本不绘制**（Step 3 实测发现，**比本文档旧版记载的「静默变文本框」严重得多**）。
   源码层面「变文本框」没写错：`html/form/input.dart:251-268` 的 `createInput` switch 没有 range 分支，
   落到 `default` → `createInputWidget()` → 一个 Flutter `TextField`。但容器里实测的表现是
   **那一行一个像素都不画** —— 没有文本框，**同一行的兄弟文字与该行自己的 `background` 一起消失**。
   而**盒模型完全正常**（实测 `tagRect=118x40` / `rawRect=120x24`），所以这是**纯绘制层问题**，
   不是布局塌陷。判据已固化在验证探针**第 14b 组**（把那一行染成黄色：整行不出现任何黄色即复现）。
   对插件的杀伤力**不止「滑块没了」，而是「同行内容全没」** —— 作者看到的是「一行莫名空白」，
   既没有报错也没有可疑元素，归因难度比「多了个文本框」高一个量级。
**⬇ 下面 8–11 是 2026-08-03 容器实测新增的，其中 8 与 9 直接命中范围内插件。**

8. **`btoa` 不是二进制安全的**：把 > 0x7F 的码点当字符先做一次 UTF-8 编码，而不是按
   latin1 取字节。实测 `btoa('\x89')` → `"wg=="`（应为 `"iQ=="`）、
   `btoa('\x89PNG')` → `"wolQTg=="`（应为 `"iVBORw=="`），而 `"wolQTg=="` 解码是
   `0xC2 0x89 0x50 0x4E` —— `0xC2 0x89` 正是 U+0089 的 UTF-8 编码。**`atob` 方向是对的**。
   后果：任何含高位字节的二进制（所有图片）经 `btoa` 都会被编坏。
   **而且它不只是「值错」，还会静默丢数据**：拿全部 256 种字节值过一遍，输出**长度是正确的
   344**，但从第 170 个字符起就对不上——它按 UTF-8 展开字节流、却按**字符数**算输出长度，
   于是原始的 `0xC1..0xFF` 共 **63 个字节被直接丢掉**。所以**「长度对得上」同样不能作为
   base64 正确的判据**。
   → 我们的绕法：`common.js` 自带 base64 编码表（`bytesToBase64`，`6f5b3ef`），不用 `btoa`。
   已在真实 WebF 运行时验过：同一个 256 字节 blob 走产品路径产出的 data URL 与预期**逐字符
   相等**（探针 `dataUrl=ok`），顺带说明 `Blob.arrayBuffer()` 的字节是准确的、锅只在 `btoa`
9. **grid `auto` 行高约 7 倍过高**：实测 downloader 页一行占 **281px**（同内容放进等宽 block
   里自然高 **41px**），表头行 72px（应约 39px）。数字与「**在 min-content 宽度下测量子项高度**」
   吻合：表头最高的「艺术家」是 3 个 CJK 字（CJK 每字都是断行点）→ 3 行 ≈ 71；数据行最长的
   艺术家名 12 个 CJK → 13 行 ≈ 280。轨道定义**无关**（裸 `2fr`、`minmax(120px,2fr)`、
   全定宽 px 三种都是 281），`grid-auto-rows: 40px` 则正常 → 是 auto 行高的测量阶段用错了宽度。
   **✅ downloader 已修**（`d629153`）：`.tbl-th, .tbl-td` 加
   `white-space:nowrap; overflow:hidden; text-overflow:ellipsis; overflow-wrap:normal`
   （写在 `style.css` 里随页面加载，**不是** JS 注入，天然满足「行插入前生效」），
   长内容全文放进 `title` 属性（**必须用 `escAttr` 拼**，`esc()` 不转义引号）。
   实测 **行距 281→43 / 表头 72→39 / 可见行数 1→6**（60 行数据区总高 18420→2580）。
   nowrap 下 min-content == max-content，所以那个错误的第一遍也量对了。
   最小合成用例（3×200px 轨道 + 10 个 CJK 字）**不复现**，触发条件未收敛
10. **`Response.headers.get()` 返回 `null`、`Blob.type` 是空串** → 从 fetch 结果推不出 mime
11. **`Infinity or NaN toInt` 不是 lxmusic 特有**：downloader 页 7 次运行里命中 2 次（间歇），
    栈是 `InlineFormattingContext._rectLineIndexCacheKey ← _lineIndexForRect ← layout ←
    RenderFlowLayout._layoutChildren`，**没有一帧是 grid**。§6 里那条「lxmusic 特有」的记法是错的

（第 7 条因为被实测大幅修订，单独放在最后 ↓）

7. **grid 把 `position: sticky` 当脱流处理**（Step 4 重设计时从源码判定，见
   `docs/webf-step4-design.md` §3 Step 1 的完整判据）。`rendering/grid.dart:347-351` 的
   `_isPositionedGridChild()` 把 sticky 与 absolute/fixed **归成同一类**，而这个判据被用在
   **13 处**排除逻辑上（`:385 :471 :872 :913 :2250 :3052 :3077 :3363 :3643 :3925 :4034 :4477`），
   其中 `:2250` 就是**构建 grid item 列表**本身、`:3052`/`:3363` 是**固有宽度计算**。
   于是 sticky 子项**既不占格子、也不参与列轨道定宽**。
   `:1948-1972` 的注释写着「their placeholders can reserve correct space」，但
   **`placeholder` 在整个 `grid.dart` 里只出现在 3 条注释里、没有任何实现**。
   > ⚠️⚠️ **本条曾有一句被实测推翻，已划掉，别再引用它。**
   >
   > ~~对照组证明这是 grid 路径独有的缺陷、不是 WebF 全局不支持 sticky：`rendering/flow.dart`
   > 的在流排除判据（`:425 :1212 :1342`）只判 `isSelfPositioned()`、不含 sticky，
   > 所以块级/流式布局下 sticky 正确留在流内并占据空间。~~
   >
   > **实测结论：`position: sticky` 在 WebF 下压根不生效，且不限于 grid 路径。**
   > 容器里把一个普通 `<div style="position:sticky;top:0">` 放在 `body` 顶部、用**页面级**
   > 滚动（`documentElement.scrollTop=300`，`window.scrollY` 确认为 300），那个 div
   > **整量滚走**（`y = -300`）。downloader 页里内层容器 `scrollTop=400/500` 时表头
   > `deltaY = -400/-500`，也是精确地滚走整个滚动量，而 computed `position` 仍是
   > `"sticky"`、`top` 仍是 `"0px"`（样式没丢）、`scroll` 事件也确实派发了（通知链跑了）。
   >
   > 所以上面那段源码推理只证明了「grid 把 sticky 排除在轨道定宽之外」（这半仍然成立），
   > **不能**据此推出「flow 路径的 sticky 是好的」——`applyStickyChildOffset` 有调用点
   > **不等于**偏移被正确算出并应用。**这是一次典型的「读源码得出乐观结论、实测相反」**，
   > 与 §2.5 里 clamp 那条恰好反向（那次是源码说不行、实测能行）。
   >
   > 残余不确定性（如实记）：合成滚轮（`PointerScrollEvent`）与合成触摸拖动都**无法**驱动
   > WebF 的任何滚动容器（页面级也不动），所以「真实用户滚动下 sticky 是否生效」未验证。
   > 判定为「没实现」而非「时序问题」的依据是：scroll 事件已派发，且页面级最标准的配置
   > 也失败。
   >
   > **✅ downloader 已不再依赖 sticky**（`d629153`）：改成三层结构 —— `.table-wrap` 管横向
   > （表头与数据区都在里面，否则横向滚到最右会错列）／`.tbl` 提供宽度基准与 `min-width`／
   > `.tbl-scroll` 管纵向且**只包数据区**，于是表头压根不需要「贴住」。
   > 滚动条宽度差导致的列错位用**每次 render 后实测**补偿（桌面 Chrome 占位式滚动条实测
   > 4px、WebF 覆盖式实测 **0px**）—— 差值在 CSS 里拿不到，但两条路径用**同一段代码**各自
   > 得到正确值，所以不需要按引擎分叉；量不到就取 0，恰好等于覆盖式滚动条的正确值。
   > 实测 6 种情形（初始／纵向滚 400／滚 1200／横向滚到最右／min-width 900／900+滚到最右）
   > 表头与数据区 `grid-template-columns` 逐字符相同、6 列 x 坐标逐一相同，滚动后表头
   > `delta=0`。**其他插件若要做贴顶表头，照这个结构做，不要用 `position: sticky`。**

### 3.3 缺口清单与真实命中面（已交叉验证）

评估必须基于**构建产物**（builder 用 esbuild 打成 IIFE / es2020，会把
`<script type="module">` 改写成普通 `<script>`），不是 `jsplugins-src/*/static/` 源码。

| 缺口 | 命中的插件 | 现状 |
|---|---|---|
| `<table>` 元素**根本不存在**（退化成嵌套 `display:block`；且 `display:table/table-row/table-cell` 也救不了 —— `css/display.dart` 的 `CSSDisplay` 枚举**没有任何 table 值**，`resolveDisplay` 的 `default` 返回 `CSSDisplay.inline`，比 block **更糟**） | **实际只有** downloader（`webf`）；radio 虽有表格但未声明 = `webview`，**进不到 WebF** | **Step 4 实施中**。原定的「垫片改写成 `<webf-table>`」**已证伪**，改走 CSS Grid，方案见 `docs/webf-step4-design.md` |
| `input[type=range]` **整行不绘制**（源码层面是落到 `TextField`，但实测一个像素都不画，连同行兄弟文字一起消失 —— 见 §3.2 第 6 条） | **仅** miot（2 处） | ✅ Step 3 已提供 `<songloft-slider>` + `common.js` 的 `rangeSliderShim` **自动**替换（隐藏原 input 并双向同步，插件 JS 零改动）。**插件侧仍需两件事**：竖向滑块在原 input 上写 `data-sl-orientation="vertical"`（垫片不猜朝向），以及补几行几何 CSS（新标签匹配不到 `input[type=range]` 选择器，垫片只拷 inline style 不拷 class）。miot 已适配 |
| `env(safe-area-inset-*)` 不求值（`css/keywords.dart` 里那 6 个 `SAFE_AREA_INSET*` / `ENV` 常量是**全库无引用的死常量**，连解析入口都没有；upstream #907 open） | 声明 `webf` 的插件里**只有** miot（3 处）。dav / subsonic / cloudflared / hostc / ytdlp 虽有 `env()` 但都未声明 = `webview`，**不暴露** | ✅ **Step 5 已完成**：宿主注入 `--sl-safe-*` 四个 CSS 变量，插件统一写 `var(--sl-safe-bottom)`。原定的「垫片把 `env()` 改写成 `var()`」**已证伪**（见 §2.5） |
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

### ✅ 插队做掉的一轮：引擎选择改为逐插件声明（2026-08-02，已完成）

去掉客户端全局引擎开关，改成 `plugin.json` 的 `renderEngine` 字段；三个官方插件（miot / downloader /
lyrics）标记为 `webf`；插件开发指南中英双语 + CHANGELOG 已同步。**这是一次决策反转**，
机制与理由见 §2.4，**不要**按本文档旧版的「明确不做 · 按插件排除引擎」把它回滚。
它不改变下面 Step 4–6 的任何结论：缺口清单（§3.3）与命中面完全不变，只是「哪些插件会暴露在
这些缺口下」现在由插件自己声明。

### ✅ Step 3 — `<songloft-slider>` 替换 `input[type=range]`（task #15，**已完成 2026-08-02**）

**这一项不再是「下一个要做的」——下一个是 Step 4（`<table>` 垫片）。** 实现细节见 §2.3 的
Step 3 小节；对插件作者的说明已写进插件开发指南（中英双语）与 CHANGELOG。落地形态与当初的
建议一致：Dart 侧新元素 + `common.js` 的 ready 垫片，**原 `<input>` 保留在 DOM 里只是隐藏**，
`input` / `change` 事件双向打通（`dispatchEvent` 已实测通，值走 `event.data`）。

顺带产出：`input[type=range]` 的真实缺陷比旧版文档记载的严重得多（**整行不绘制**，不是
「变文本框」），已补进 §3.2 第 6 条 —— 上游报 bug（task #12）时请按那一条的措辞写。

### Step 4 — `<table>`（task #16，**实施中**）

> ⚠️ **本文档旧版写的「WebF 自带 `<webf-table>` 系列标签，垫片做的是标签改写而非从零实现」
> 已被证伪。看到那句话时以本节与 `docs/webf-step4-design.md` 为准。**

**证伪理由**：`<webf-table>` 的 `build` 只读**直接 `childNodes`**（`html/table.dart:188-189` 的
`firstWhereOrNull` / `whereType`），`<thead>`/`<tbody>` 不拆就是**一张空表，且不报错不打日志**；
而 downloader 恰好靠 `#tbody` + `innerHTML` 渲染行 —— 保留 `<tbody>` 得空表，拆掉插件 JS 抛
`TypeError`。WebF 又**没有 `MutationObserver`**。更要命的是 `colspan`/`rowspan` 零支持是
**Flutter `Table` widget 的天花板，不是 WebF 没做，上游修也修不了**。

**现在的方案**：CSS Grid（`docs/webf-step4-design.md` 的方案 B'）。WebF 的 Grid 是**已实现**的
（`fr` / `minmax()` / `repeat()` / `auto-fill` / `auto-fit` / `grid-auto-flow` / `span` 全在
`css/grid.dart`），且这是唯一「浏览器 / 系统 WebView / WebF **三条路径共用一套代码**」的方案
—— 自写元素与垫片方案都要靠 `webf-engine` class 分叉两套模板，而 WebF 那份**在本机永远跑不到**
（glibc 2.35 < 2.38），双份实现 + 单份可测最易腐化。

**必须按「双容器 + 同步列宽」写，不要写单容器 + 纯 `fr`** —— grid 子项上的 `position: sticky`
不可用，判据见 §3.2 第 7 条。

**紧迫性**：命中面只有 downloader 一张表（radio 未声明 `renderEngine` = `webview`，进不到 WebF），
但 **downloader 已带 `renderEngine: "webf"` 发版**（v2026.8.2），是目前**唯一已发布、
用户可见的 WebF 回归**。

### ✅ Step 5 — 安全区（task #19，**已完成 2026-08-02**）

实现与三条实测事实见 **§2.5**。原定的「垫片改写 `env()` → `var()`」已证伪，改走
「宿主只注入变量、插件写 `var(--sl-safe-*)`」。

### Step 6 — 三项经桥下沉（task #18，**实施中**）

**优先级**（预研与设计文档已定，按这个顺序）：

1. **`window.open`** —— miot 小米账号二次验证登录，**功能性阻塞**。WebF 的 `window.open` 是
   no-op（`window.cc` 两个重载都 `return this`，**不抛错**）。预研 §3.3 查明产品这边也不是
   no-op，而是**没设 `WebFNavigationDelegate` 落到无条件 cancel**，约 8 行可解（骨架在预研里）
2. **`URL.createObjectURL`** —— miot 带鉴权头拉封面 ×2。WebF 有 `Blob` 但没有产 `blob:` 的入口。
   改 `data:` URL，**注意 `createObjectURL` 是同步的而 blob→base64 是异步的，所以必须改 miot
   的调用点**，不能只做一个假的同步垫片
3. **`input[type=file]`** —— **当前零暴露**（radio 是 `webview`；ytdlp / lxmusic 拿不到源码且
   从未标 `webf`）。桥形状与垫片形状已定形，见 `docs/webf-step4-design.md` §2.3 / §2.4：
   桥返回 `{name, size, text}`、**主载荷是 UTF-8 解码后的字符串**（不返回 path），
   垫片**必须同时拦 `click` 事件与覆写实例 `click` 方法**（radio 是「隐藏 input + 外部按钮代点」）

`file_picker: ^10.3.10` **早就在 `pubspec.yaml` 里了**，所以用它的原生契约哈希代价是**零**。
⚠️ 但**绝不要 bump 它的版本**：版本号进哈希，而它是 caret 约束，**一次 `pub upgrade` 就会静默改变**。

顺带还有一件小事（`docs/webf-step4-design.md` §1.9）：给 `common.js` 加一个**只警告不改写**的
表格垫片，把「插件用了 `<table>` 但它压根不存在」这个**完全静默**的失败
（`element_registry.dart` 的日志默认关）变成一行指路的 `console.warn`。

### ✅ task #3 — 许可合规（**主体已完成 2026-08-02**）

`songloft-player/NOTICE`、双语 README / docs 的许可章节、GPL-3.0 全文随 release 产物、
「完整对应源码」获取方式（`CORRESPONDING-SOURCE.txt`）均已落地。

**三个非显而易见的决策，改这块前先读**：

1. **`LICENSES/GPL-3.0.txt` 刻意不叫 `COPYING` / `LICENSE-*`、也刻意不放仓库根**：
   GitHub 的 licensee **只扫仓库根**，双许可仓库在根上放第二份 license 会让它判成
   `NOASSERTION`，README 的 shields 徽章会从 Apache-2.0 变成 "unknown"。
2. **父仓库与 `songloft-player` 各存一份**（35147 字节，md5 相同）。不是冗余：父仓库的
   `create-release` job checkout 的是 **`ref: main`**、**且不 init 子模块**，拿不到 player 那份。
3. **workflow 里用 `git show ${GITHUB_SHA}:LICENSES/GPL-3.0.txt` 而不是 `cp`**：同上，
   那个 job 的工作区是 main 而不是被构建的那个 commit。用 `cp` 会在 `dev-webf` 阶段
   （main 上还没有这个文件时）**直接搞坏 create-release**。

**剩下的挂账项**（都不阻塞合并）：把 license 文本嵌进每个安装包内部（APK / IPA / DMG / MSIX
都是签名容器，塞进去要动签名流程）、App 内「开源许可」页、首次真实 release 后确认
`CORRESPONDING-SOURCE.txt` 里的版本号没有回落成 `unknown`。

### task #12 — 给 WebF 上游报 §3.2 里的 **7** 条

草稿在 `docs/webf-upstream-issues.md`（同为分支临时件）。
**未经用户确认不要向 `openwebf/webf` 提交** —— 那是对第三方仓库的外发动作。

### 明确不做（用户已定）

- ~~**按插件排除引擎**（在 `plugin.json` 里标记某插件不用 WebF）—— 用户明确说「不按插件排除」~~
  → **这一条已被用户推翻，现在正是按插件在 `plugin.json` 里声明 `renderEngine`**，
  且**全局运行时开关已删除**。机制与反转理由见 §2.4。划掉而不删除，是为了让照着旧版文档
  行动的人能发现结论变了
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
- ~~**`[diag]` 在 Step 2 那轮跑没有输出**，原因未查明~~ → **Step 5 已查明**：
  `WebFController.onLoad` 在 `WebF.fromControllerName` 这条挂载路径下**可能永不触发** ——
  `checkCompleted()` 在 `document.hasPendingRequest` 为真时**直接 return**
  （`launcher/controller.dart:1734`），而探针页第 11 组那两个 `<img src="">` 会被 WebF 请求成
  **文档自身 URL**、解码失败挂住。诊断脚本挂在 `onLoad` 上，所以一行都没跑。
  探针已改成「页面在确定时点经 methodChannel 主动叫 Dart」（与 `slDrag` 同套路），不再依赖 `onLoad`。
  **⚠️ 对产品的含义**：`_pageReady` 就是在 `onLoad` 里置 true，所以安全区、主题、播放状态
  **三条推送共享同一个闸**。真实插件页由后端 `stripEmptySrcAttrs` 剥掉空 `src`、且页面能正常结束
  loading（否则会撞 20s 超时 UI），所以生产上这个闸**应当**会开 —— 但**没有实机验证**。
  新加依赖页面 ready 时序的桥之前，先确认这个闸会开，或学 Step 5 / `slDrag` 让页面主动叫 Dart
- **⚠️ `run.sh` 默认不重建镜像，而 `probe_main.dart` / `entrypoint.sh` 是 `COPY` 进镜像的**
  → 改了它们再跑 `run.sh`，**跑的是旧探针**。表现极具误导性：「我新加的诊断脚本没生效，
  输出的还是内置那份」。与下面 `docker build && run.sh` 那条同源但不同因，为此白跑过一轮。
  改探针后必须 `run.sh --build`
- **⚠️ WebF 的 layout 是异步的**：`el.style.x = ...; void el.offsetHeight;
  el.getBoundingClientRect()` 读到的是**改之前**的布局。所有「改样式 → 量」都必须跨帧
  （`setTimeout`）。曾因此误判出「窄屏列宽不对」与「360 个单元格全是 0×0」两个不存在的缺陷
- **⚠️⚠️ 更进一步：WebF 压根不保证内联样式变更后会重新布局** →
  **「运行时改样式做 A/B 归因」这个手法在 WebF 上不可信**（同一改动两轮分别量到 61 与 43）。
  要对比两个版本，只能**各自新鲜加载**一次。相关地，`removeChild(<style>)`
  在 WebF 下**不撤销样式**（computed style 保持注入后的值），所以「注入→量→移除→再量」
  这种对照法也是无效的
- **首次测量必须留足 settle（≥2s）**：诊断脚本里 `setTimeout(0)` 会读到未完成的布局
  （`cell0H=0`、坐标全 0），**表现与「元素塌陷成 0 高」一模一样**，极易误判成真缺陷
- **绘制层判据不要按坐标取色**：报坐标与抓屏之间若隔了几秒，其间别处的异步读数仍在写 DOM，
  会把下面的行整体推移（实测被推下 96px），于是取到空白处的颜色 → **假阴性**。
  改用**与坐标无关**的判据：「这个颜色在整页出现了多少像素」。
  「data URL 能不能出图」这个问题正是因为用错判据而反复翻转过两次（详见 `common.js` 里
  `blobToDataURL` 的注释），现结论是**能**（4 项全过，各 196px）
- **`document.documentElement.scrollTop = N` 可驱动页面级滚动**（`window.scrollY` 会跟随），
  但 `document.body.scrollTop` **无效**
- **容器里 `100vh = 720`（窗口 1280×717）**，而 downloader 的表格顶在 `y≈735`
  → **整张表默认在折叠线以下**，抓屏前必须先把页面下滚，否则截图里根本没有它
- **容器里没有中日韩字体** → 截图中 CJK 全是豆腐块（拉丁字符正常）。不是 WebF 缺陷，
  但会让人误判「字体挂了」
- **`display:none` 的元素 `getBoundingClientRect()` 未必是 0 尺寸**：实测 downloader 的
  `#empty` 在 `display:none` 下仍返回 728×157。功能上没影响（确实没显示），
  但**拿它的 rect 当「是否可见」的判据会被骗**
- **合成滚动驱动不了 WebF**：`PointerScrollEvent`（即使补了 `PointerAdded` + `PointerHover`、
  落点在可见区内）与合成触摸拖动都无法让任何滚动容器动一下，页面级也不动。
  目前只有程序化 `scrollTop` / `documentElement.scrollTop` 能驱动滚动
- **探针第 16 组本身是 flaky 的**：Step 5 实测 5 次里 2 次 `sldRect=0x0`（滑块被静默漏出拖动目标，
  **表现与「事件没通」一模一样**）。不改一行重跑就好了，probe.html 原有注释已描述过这个失效模式。
  **不要**把它的偶发失败误判成自己的改动坏了
- 断言铁律（与仓库既有的无头浏览器验证一致）：**截图只证明"渲染对了"**，
  交互是否真生效必须落在后端可观测状态上（`curl` 对应 `/settings/<name>`、`play_history` 有无新记录等）。
  数进程用 `pgrep -x`，**不要** `ps -ef | grep | wc -l`

---

## 6. 已知未解问题

- **lxmusic 在 WebF 下有布局崩溃**：`Unsupported operation: Infinity or NaN toInt`、
  `Null check operator used on a null value`。lxmusic 未构建发布、也不在本分支的 `.gitmodules` 里
  （本分支只跟踪 miot / subsonic / cloudflared / dav / hostc / registry / downloader / lyrics / radio，
  **lxmusic / bili / ytdlp 都不是跟踪的子模块**，要验证它们得先自己 clone）。
  → **⛔ 用户已明确划出范围外（2026-08-03），不处理。** 曾派 agent 查根因，中途按该决定停掉。
  留这条只为「以后若把 lxmusic 纳入范围，知道有这么个坑」。
  **⚠️ 「lxmusic 特有」这个记法已被实测推翻**（2026-08-03）：`Infinity or NaN toInt`
  在 **downloader 页也会出现**，7 次运行里命中 2 次（间歇性），栈是
  `InlineFormattingContext._rectLineIndexCacheKey ← _lineIndexForRect ← layout ←
  RenderFlowLayout._layoutChildren` —— **没有一帧是 grid**，所以也不是我们改成 Grid 引入的。
  它是**行内格式化上下文**里的通用缺陷（已登记为 §3.2 第 11 条）。
  也就是说这条已经**命中范围内插件**，不再是「范围外、可以不管」的事；
  但它间歇出现、且目前未观察到可见后果（页面照常渲染），所以按「已知带栈的间歇异常」
  记账，暂不专门开工。**若 downloader / miot / lyrics 出现可见的布局错乱，先怀疑这条。**
- **`<details>` 垫片的一个已知边界**：垫片跑完之后再给 `<details>` 追加直接子节点，
  那个节点会永久留在折叠容器外面（幂等标记会阻止重新包裹）。插件应在插完 HTML 后调
  `SongloftPlugin.applyShims()`，但对"追加单个子节点"这种用法无解
- miot `index.html:1378` 引外部 CDN `marked.min.js`（builder 不打包），离线/内网下 Markdown
  渲染静默失效 —— **与 WebF 无关的既有问题**，顺手记一笔
- **上游风险（缓解手段已变，风险本身没消失）**：WebF 0.x beta，main 分支自 2026-04-19 静默至今，
  30 天下载量 1172，最后 9 个版本几乎全是 flex/inline 布局的正确性修复。
  本文档旧版据此写的结论是「这是**必须保留运行时回退开关**的直接理由」——
  **用户已明确放弃全局回退开关**（见 §2.4），风险敞口改由三件事承担：
  ① 默认 `webview`（不声明就不暴露）；② 逐插件显式声明、由插件作者自己验证；
  ③ 出问题时用户可禁用该插件、或等作者发版把 `renderEngine` 改回 `webview`。
  如实记下来：**上游一旦回归性变坏，没有任何"一键全局切回"的手段**，只能逐插件改 manifest 重新发版；
  受影响插件的用户在新版发出来之前只能禁用插件。这是已知且被接受的取舍，**不要**据此擅自把
  全局开关加回去

---

## 7. 铁律速查（本分支相关）

- `internal/jsplugin/assets/common.js` 服务给**所有**客户端版本与普通浏览器 →
  改动必须**纯增量 + 特性探测**，且 `isWebFHost()` 判定要排在 `isNativeHost()` / `isIframeHost()` **之前**
- `render/elements/` 只能依赖 `flutter` + `webf`（验证探针要跨 package 拷它）
- 改 Dart 后 `cd songloft-player && dart format lib/ test/`；改 Go 后根目录 `gofmt -w .`
- 提交**禁止** `Co-Authored-By`；子仓库引用父仓库 issue 必须写完整路径 `songloft-org/songloft#341`
- 子模块改动流程：子仓库提交 → 回主仓库 `git add <path>` bump 指针 → 主仓库提交
- 本仓库 worktree 的 git stash 栈与主 checkout 共享 → **禁止**裸 `git stash` / `git stash pop`
