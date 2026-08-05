# WebF 插件页渲染迁移 —— 文档索引（songloft-org/songloft#341）

> **本目录整体是分支临时件，不是产品文档。**
> 刻意**不做**中英双语同步（没有 `docs/en/webf/` 对应件），也整体从文档站排除
> （`docs/.vitepress/config.mts` 的 `srcExclude: ['webf/**']` + `ignoreList: ['webf']`）。
> **#341 落地后连同整个目录一起删除**，那两条配置也一并删掉。
>
> 产品面向用户的 WebF 内容**不在这里**，在 `docs/js-plugin-development-guide.md` §8
> 「WebF 渲染引擎与原生元素」及其英文版（双语铁律适用），以及 `CHANGELOG.md`。
>
> 最后更新：2026-08-05（**flex 套 flex → 子树静默不绘制**，定案并修复，见 `handoff.md`
> 第 29 条与 `upstream-issues.md` 第 8 条新增的小节；同日：白屏根因定案为
> 「渲染面重新挂载 + controller 命中缓存」，
> `handoff.md` 第 26 条被推翻、新增第 27 条 + `upstream-issues.md` 第 10 条；
> 同日此前：downloader 的 flex `wrap` base size 缺陷修复 + `console` 取证通道被推翻，
> 见 `handoff.md` 第 21 / 22 条与 `upstream-issues.md` 第 8 条）。

---

## 这是什么

把 JS 插件页的渲染从系统 WebView 迁到 [WebF](https://github.com/openwebf/webf)
（自建 W3C 运行时，用 Flutter 绘制 HTML/CSS）。选择方式是**每个插件在 `plugin.json` 里用
`renderEngine` 声明**，缺省仍是 `webview`。

---

## 先读哪份

| 文档 | 什么时候读 | 行数量级 |
|---|---|---|
| **[handoff.md](handoff.md)** | **接手就读这份，它是入口。** 现状、硬约束、11 条已确诊上游缺陷、剩余工作、验证环境配方、铁律 | ~1100 |
| [step4-design.md](step4-design.md) | 要动表格 / `input[type=file]` 时读。四方案对比与选型依据 | ~1150 |
| [recon-step456.md](recon-step456.md) | 想知道「为什么不那样做」时读。Step 4/5/6 预研，**证伪了两条既定方案** | ~880 |
| [upstream-issues.md](upstream-issues.md) | 要给 WebF 上游报 bug 时读。10 条草稿，**英文正文** | ~1350 |

**只想干活、不想读完**：`handoff.md` 的 §0（现状 + 范围边界）→ §3.1（API 事实）→ §3.2（缺陷台账）
→ §5（验证环境）。§3.2 是全套文档里复用率最高的一节。

---

## ⚠️ 2026-08-04 之后的路线变更（先读这条）

三个插件的 `renderEngine` 曾被主仓 `397f4bd` 全部回退成 `webview`。之后 **downloader 换了思路
重做**：不再用浏览器语义的 HTML/CSS 硬凑，改用 **webf-ui 原生组件**（`webf_cupertino_ui` 的
31 个 `<flutter-cupertino-*>` + `webf` 内建的 `<webf-list-view>`），前端换成 Vue 3 + Vite，
并已重新声明 `webf`（v2026.8.4）。

**后果：下表里 Step 4（`<table>` → CSS Grid）已降级为备选方案** —— 列表类内容一律优先
`<webf-list-view>`，那条路直接绕开「grid `auto` 行高」与「sticky 表头」两个最难缠的缺陷。
完整说明与本轮新核实的 10 条事实见 `handoff.md` 的「🆕 2026-08-04 之后：webf-ui 路线」一节
（**它晚于本目录其余内容，冲突时以它为准**）。

---

## 当前状态速览

**范围硬边界**：**只处理 miot / downloader / lyrics 三个插件**，其他插件的问题一律不处理
（用户 2026-08-03 划定，详见 `handoff.md` §0）。

| 步骤 | 状态 |
|---|---|
| Step 1 JS 垫片框架 | ✅ |
| Step 2 Dart 自定义元素框架 + `<songloft-progress-ring>` | ✅ |
| Step 3 `<songloft-slider>`（替 `input[type=range]`） | ✅ 已实测 |
| Step 4 `<table>` → CSS Grid | ✅ 已实测（downloader 已发版 v2026.8.3） |
| Step 5 安全区 `--sl-safe-*` | ✅ 已实测 |
| Step 6 `window.open` / `blobToDataURL` / `input[type=file]` | ✅ 已实测 |
| 许可合规（GPL-3.0） | ✅ 含 App 内许可页 + 许可全文随包分发 |
| 上游报 bug | 📝 10 条草稿已写，**未提交**（需用户批准 + 人工排重） |

**未验证的三件事**（只能靠真机手测）：真机刘海安全区（容器里 `MediaQuery.viewPadding` 恒为 0）；
真实用户手指滚动下的表现（合成滚轮驱动不了 WebF 的滚动容器）；系统 WebView 路径
（三条渲染路径里唯一没单独验过的）。

---

## 这套文档最值钱的三样东西

如果时间只够看三处，看这三处 —— 它们是反复返工换来的：

1. **`handoff.md` §3.2 的缺陷台账（11 条）** —— 每条都带 `file:line` 与复现判据。
   报上游、判断「这是不是我改坏的」都靠它。
2. **`handoff.md` §5 的验证环境坑列表** —— 本机跑不了 WebF（glibc 2.35 < 2.38），
   容器是唯一途径，而那套环境本身有十来个能让你误判的陷阱（改了探针不 `--build` 就是在跑旧的、
   layout 异步导致「改样式→同帧量」读到旧值、按坐标取色的假阴性……）。
3. **被证伪 / 被推翻的结论清单**（见下）。

---

## ⚠️ 已被推翻的结论（照旧文档行动会走错）

这套文档**反复发生过「先写下结论、后被推翻」**。每一处都保留原文并划掉、加交叉引用，
但如果你只搜关键词、不读上下文，仍可能捡到废弃结论。清单：

| 曾经的结论 | 现在 | 出处 |
|---|---|---|
| 引擎选择用**客户端全局开关** | 改为逐插件 `plugin.json` 声明，全局开关**已删除**（用户决策反转） | `handoff.md` §2.4 |
| Step 4 用垫片把 `<table>` **改写成 `<webf-table>`** | 证伪，改走 **CSS Grid** | `handoff.md` §4、`step4-design.md` |
| Step 5 用垫片把 CSS 里的 `env()` **改写成 `var()`** | 证伪（CSSOM 无写入面 + `max()` 也没实现），改为**宿主注入变量、插件写 `var()`** | `handoff.md` §2.5 |
| `clamp()` 的参数**不接受** `var()`（源码判读） | **实测能行**（`D=30`/`J=24`） | `handoff.md` §2.5 |
| sticky 只在 **grid 路径**坏，flow 路径是好的（源码判读） | **实测：全局都不生效**，页面级最标准配置也整量滚走 | `handoff.md` §3.2 第 7 条 |
| data URL **出不了图**（据一次目视观察） | **实测能出图**（4 项绘制判据全过，各 196px） | `common.js` 的 `blobToDataURL` 注释 |
| `Infinity or NaN toInt` 是 **lxmusic 特有** | downloader 页也会出现（栈在 `InlineFormattingContext`，无 grid 帧） | `handoff.md` §3.2 第 11 条、§6 |
| 「必须保留运行时全局回退开关」 | 用户明确放弃，风险改由「默认 `webview` + 逐插件声明」承担 | `handoff.md` §6 |
| 插件页 `console.log` 是**真机上可靠的取证通道**（多处文档与代码注释这么写） | **在最常见的那条路径上整体失效**：`onJSLog` 在 `createController` 里赋值，而 controller 命中预加载/进程内缓存时那条路径不跑。2026-08-05 实测：整份日志 `[plugin][console]` 零命中，而页面确实画出来了。布局问题改用**截图量像素** | `handoff.md` 第 22 条（及第 16 条的后果 ④） |
| webf-ui 路线让「布局相关的整批缺陷不再命中」 | **收窄**：`flex-wrap: wrap` 下 WidgetElement 与嵌套 flex 的 base size 仍被测成容器宽度（每项独占一行铺满）。原生元素绕开了 grid/table，但没绕开 flex 的 base size 测量 | `handoff.md` 第 21 条 |
| 下拉面板**不能用浮层**，只能用常规流块盒（怕 WebF 的层叠与命中测试） | **层叠部分证伪**：`position:absolute` + `z-index` 能正常盖住 `<webf-list-view>` 这个 Flutter widget。已改为浮层（命中测试仍待验证）。连带发现：祖先链上的 `overflow:hidden` 会把浮层整段切掉 | `handoff.md` 第 24 条 |
| 「用截图脚本验证就够了」 | **不够**：受控文本框被外部改写 + 鼠标停在页面上 = 整页白屏（debug），而截图脚本运行时鼠标从不在窗口里，这条**永远撞不到**。真机手动操作是不可替代的一环 | `handoff.md` 第 25 条 |
| 白屏 = 第 16 条的加载超时 / hot reload / 文本框断言 / **同一 controller 被两个渲染面抢占** | **四条全错**（文本框断言真实存在但不是这次）。真根因：**渲染面重新挂载 + controller 命中缓存** → `createController`/`onLoad`/`onJSLog` 全不跑 + `_adoptPreloadedController()` 无条件报成功 → 任何失败都静默白屏。同一个白屏连错三次归因 | `handoff.md` 第 27 条（第 26 条为被推翻的第三次） |
| Tab 页靠 shell 层 **Offstage 保活、永不释放**（据此推断两个渲染面必然共存） | **桌面端根本不保活**：Offstage 保活只对 Web / 移动端生效，Windows/macOS/Linux 是「切走即销毁」（规避 #246 的 WebView2 灰块）。所以 macOS 上两个渲染面从不共存，而是先后挂载 —— 为验证共存而加的探针**一次都没触发**，我却在它给出否定答案前就把结论写进了文档 | `handoff.md` 第 26 条 |
| flex `wrap` 的 base size 缺陷「已修复并真机验证」（第 21 条，给 `.dl-filter-item` 补 `width`） | **只修对了一半**：主轴排版确实正常了、像素也量过，但同一个「试排一遍」的测量层还会把**嵌套 flex 子树留在 `needsLayout`**，而 Flutter 会**静默跳过绘制**它 → 整片内容消失。「排版对了」≠「这条缺陷绕开了」 | `handoff.md` 第 29 条 |
| cupertino button 的 `variant` 可以直接用来表达按钮层级 | **不能**：`CupertinoButton.filled` 的底色固定是 `CupertinoTheme.primaryColor`、构造器不接受 color，圆角默认还是直角。统一走 `plain` + CSS 给外观 | `handoff.md` 第 23 条 |

**方法论教训（两次方向相反的翻车，各记一次）**：
源码推理与实测冲突时**以实测为准**（sticky、data URL 都是源码/目视给了乐观或悲观的错误结论）；
反过来，源码说「不支持」也可能是错的（`clamp()` + `var()`）。**能进容器验的就别只读源码。**

---

## 判据本身出过错（比结论出错更危险）

三条已知的坏判据，别再用：

- **「盒子有尺寸」≠「图被画出来了」** —— `getBoundingClientRect` 通过不代表绘制成功。
- **「按坐标取色」有假阴性** —— 报坐标与抓屏之间别处的异步读数仍在写 DOM，会把行整体推移
  （实测被推下 96px），取到空白处。改用与坐标无关的「该颜色在整页出现多少像素」。
- **「base64 长度对得上」≠ 内容对** —— WebF 的 `btoa` 输出长度正确（344）但第 170 字符起就错，
  静默丢掉 63 个字节。二进制编解码的回归用例**必须含 > 0x7F 的字节**。

---

## 相关但不在本目录的东西

- **产品文档**：`docs/js-plugin-development-guide.md` §8 + 英文版（**改一边必须同步另一边**）
- **宿主垫片层**：`internal/jsplugin/assets/common.js`（`isWebFEngine()` 分支内）、`common.css`
- **Dart 渲染层**：`songloft-player/lib/features/home/presentation/render/`
  （`elements/` 子目录**只能** import `flutter` 与 `webf`，验证探针要跨 package 拷它）
- **验证探针**：`songloft-player/scripts/webf-verify/`（用法见 `handoff.md` §5）
- **合并前必须撤掉的临时改动**：两个 release 工作流发布到独立 tag `dev-webf`
  （见 `handoff.md` §1，**里面夹着两处真 bug 修复，撤销时要保留**）
