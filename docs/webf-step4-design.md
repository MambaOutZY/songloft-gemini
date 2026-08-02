# WebF 迁移 — Step 4 方案重设计 + Step 6 `input[type=file]` 定形（songloft-org/songloft#341）

> **这是 songloft-org/songloft#341 的分支临时件。** 落地后连同
> `docs/webf-migration-handoff.md`、`docs/webf-recon-step456.md` 一起删除。
> **刻意不做中英双语同步**，也从文档站导航排除（`docs/.vitepress/config.mts` 的 `srcExclude`）。
>
> 面向对象：接手 Step 4 / Step 6 实施的下一个 agent。
> 前置必读：`docs/webf-recon-step456.md` §2（Step 4 预研，**已证伪既定方案**）与 §3（Step 6 预研）、
> `docs/webf-migration-handoff.md` §3.3 / §4。
> 最后更新：2026-08-02。

**本文档的标注约定**（沿用预研文档）：
`✅ 已核实` = 附 `文件:行号`，读过源码；
`⚠️ 需实测` = 从源码推不出来 / 源码不随 pub 包发布 / 只有运行时能确认。
**混淆这两类是本文档唯一不可接受的错误。**

---

## 0. 源码获取结果（第一件事）

| 插件 | 是否拿到源码 | 方式 |
|---|---|---|
| **radio**（电台导入） | ✅ 拿到 | 本仓库跟踪的子模块，`git submodule update --init jsplugins-src/songloft-plugin-radio` 拉下，HEAD `fbc7d6fcb08f2b7e2b680f549bec25521cf0d9fd` |
| **ytdlp** | ❌ **拿不到** | `git clone https://github.com/songloft-org/songloft-plugin-ytdlp.git` → `ERROR: Repository not found.`（同名 `git ls-remote` 亦然）。仓库名不是 `songloft-org/songloft-plugin-ytdlp`，或为私有 / 尚未公开 |
| **lxmusic** | ❌ **拿不到** | 同上，`songloft-org/songloft-plugin-lxmusic` 也是 `Repository not found` |
| **bili** | 未尝试 | 本文档不需要它（表格与 `input[type=file]` 的命中面里都没有 bili） |

> **ytdlp / lxmusic 需另行获取**（正确的仓库地址 / 访问权限）。
> 本文档**不猜**它们的代码长什么样；凡涉及这两个插件的判断一律标 `无法核实`。
> 好消息：§2 会说明它们当前对 Step 6 **零紧迫性**，所以拿不到源码不阻塞任何事。

**⚠️ 注意**：这两次 clone 都落在 `/tmp`（`/tmp/sl-ytdlp` / `/tmp/sl-lxmusic`，均已因失败而不存在），
**没有**写进 `jsplugins-src/`。后续 agent 若拿到正确地址，也请 clone 到 `/tmp`
——放进 `jsplugins-src/` 会污染工作树、让别人误提交一个未登记的子模块目录。

### 顺带核实：「构建产物 vs `static/` 源码」这一条对表格**不构成差异**

交接文档 §3.3 强调评估必须基于构建产物。已核实 builder 对 HTML **只做两件事**，
**都不触碰 `<table>` 结构**：

- `plugin-toolchain/packages/plugin-builder/src/build.ts:99-102`：`cpSync(static/, build/static/)`
  —— 整目录**原样拷贝**；
- `:112-151`：若存在 `static/js/app.js`，用 esbuild 打成 `format: 'iife'` / `target: 'es2020'`
  的 `app.bundle.js`（`:116-123`），删掉其余 `.js`，然后**只**用正则
  `/<script\b[^>]*\bsrc="(?:static\/)?js\/app\.js"[^>]*><\/script>/` 换掉那一个 script 标签（`:138-141`）；
- `:154-164` → `static-assets.ts`：给资源文件名注内容 hash，改写 HTML 的 `<script src>` / `<link href>`
  与 CSS 的 `url()`。

→ **`<table>/<thead>/<tbody>/<tr>/<th>/<td>` 在构建产物里与 `static/index.html` 逐字节相同。**
另外 downloader 与 radio 的页面脚本本来就是普通 `<script src="static/js/app.js">`
（`downloader/static/index.html:131`、`radio/static/index.html:121`），
**不是** `<script type="module">`，所以「builder 会改写 module 标签」这条在这两个插件上不适用。
本文档后续对表格的所有结论因此可以直接引 `static/` 的行号。

---

## 1. Step 4 方案重设计（核心产出）

> **一句话结论**：**真实命中面只剩 downloader 一个插件的一张 6 列静态表**
> （radio 那张表没声明 `renderEngine`，压根跑不到 WebF 上）。
> 在这个规模下，「宿主写一个 `<songloft-table>` 元素」和「宿主写一层 table 展平垫片」
> **两者的成本都远超收益**。
> **推荐方案 B'**：downloader 自己把那张表改成 **CSS Grid**（宿主零改动、浏览器/WebView 路径同一套
> 代码同一套外观、sticky 表头用已实现的 `position: sticky`），
> 宿主侧只做一件**廉价且高价值**的事：在 `common.js` 里加一个**只警告不改写**的探测垫片
> （检测到 `webf-engine` 下存在原生 `<table>` 就 `console.warn` 指向文档），
> 把「静默一张空表 / 一坨嵌套 block」变成「日志里一行明确指路」。
> 完整对比见 §1.8。

### 1.1 先定紧迫性：`renderEngine` 的实际取值（**这一条改变了 Step 4 的全部范围**）

✅ **已核实**（直接读 `jsplugins-src/*/plugin.json`）：

| 插件 | `plugin.json` 的 `renderEngine` | 有没有 `<table>` | 真的暴露在这个缺口下？ |
|---|---|---|---|
| **downloader**（歌曲下载） | `"webf"`（`plugin.json:13`） | ✅ 1 张，6 列 | ✅ **是。唯一一个** |
| **radio**（电台导入） | **字段不存在** → 按 §2.4 契约 = `webview` | ✅ 1 张，4 列 | ❌ **否**（走 iframe/WebView，原生 `<table>` 正常渲染） |
| **lyrics**（歌词搜索） | `"webf"`（`plugin.json:12`） | ❌ 零表格标签 | ❌ 否 |
| **miot**（智能音箱） | `"webf"`（`plugin.json:13`） | ❌ 零表格标签 | ❌ 否 |
| subsonic / dav / hostc / cloudflared / registry | 字段不存在 = `webview` | ❌ 零表格标签 | ❌ 否 |

统计方式：对每个跟踪子模块的 `static/` + `src/` 做
`grep -rniEo '<(table|thead|tbody|tfoot|tr|th|td|caption|colgroup)\b'`，
只有 downloader 与 radio 有命中，其余 7 个插件**一个表格标签都没有**。

**由此得到三条会直接改写实施计划的结论：**

1. **真实命中面从「downloader + radio」收缩到「downloader」。**
   交接文档 §3.3 的「命中的插件：downloader、radio」是**按标签统计**的，
   在 §2.4 引入 `renderEngine` 之后**已经过时** —— radio 声明的是 webview，
   它那张表连 WebF 都进不去，`<webf-table>` 的 sticky 对齐问题、
   「radio 还叠了 sticky 表头」这条注意事项**当前一律不适用**。
2. **紧迫性从「阻塞」降到「低」。** 唯一受害者是 downloader，
   而 downloader 是**第一方子模块**（`.gitmodules` 有它），改它的源码不需要求任何外部作者配合。
   也就是说：**这个缺口完全可以在插件侧关掉，宿主不必为它长出一个新元素。**
3. **但「宿主要不要做点什么」并没有因此归零**，理由是**第三方插件**：
   `<table>` 是极常见的写法，任何第三方作者在 `plugin.json` 里写 `renderEngine: "webf"`
   都会立刻踩到，而这个缺口的表现是 **完全静默的**
   （✅ 已核实 `element_registry.dart:83-85` —— 未知标签的日志只在
   `enableWebFCommandLog` 打开时才 `debugPrint`，产品没开）。
   宿主该出的力是**让它不再静默**，而不是替第三方把表格实现出来（见 §1.8 的推荐 + §1.9）。

⚠️ **需实测/需确认（非源码问题，是产品决策）**：downloader 当前已发布的版本
（`plugin.json:4` 版本号 `2026.8.2`）已经带 `renderEngine: "webf"`。
所以**在 Step 4 落地之前，装了 downloader 的用户看到的就是一张坏掉的表**。
这是本项目当前**唯一一个已知的、已发版的、用户可见的 WebF 回归**。
如果 Step 4 不能马上做完，**最省事的止血是把 downloader 的 `renderEngine` 先改回 `webview`
（或直接删掉那一行）再发一版** —— 这一步零风险、一行改动，
建议在任何方案动工之前先做掉（见 §3 的第 0 步）。

### 1.2 真实命中面（逐字段核实）

两张表都**极其规整**，这是本节最重要的输入：**没有任何一个需要 colspan/rowspan 的地方**，
两张表都是「表头一行 + JS 动态生成的等宽行」，没有 `<tfoot>` / `<caption>` / `<colgroup>` / 嵌套表格。

#### downloader（**唯一真正命中的**）

| 维度 | 事实 | 出处（✅ 已核实） |
|---|---|---|
| 列数 | **6**（复选框 / 标题 / 艺术家 / 专辑 / 来源 / 状态），表头与行**严格等宽 6 格** | `static/index.html:108-117`（表头）、`static/js/app.js:111-118`（行模板） |
| `colspan` / `rowspan` | **零**（整仓库 grep 无命中） | — |
| 行来源 | **JS 动态**，`tbody.innerHTML = list.map(...).join('')` —— **整体重写**，不是增量 append | `app.js:93`（`const tbody = $('#tbody')`）、`:96`（清空分支 `tbody.innerHTML = ''`）、`:103`（重写分支） |
| 行容器 | `<tbody id="tbody">` —— **插件按 id 拿它**，这是 §1.3 死路 ② 的根因 | `index.html:118` |
| sticky 表头 | ✅ 有，走 **CSS** `th { position: sticky; top: 0; background: var(--md-surface) }` | `static/css/style.css:325-327` |
| 滚动容器 | 外层 `.table-wrap` **内联** `max-height:calc(100vh - 380px);overflow-y:auto`，CSS 里另有 `overflow-x:auto` | `index.html:106`、`style.css:307-309` |
| 列宽 | 只有两列写死：`<th style="width:36px">`（复选框列）、`<th style="width:60px">`（状态列），其余 4 列自适应 | `index.html:110`、`:115` |
| 其它 table CSS | `table{width:100%;border-collapse:collapse}`、`th/td{padding:10px 12px;border-bottom:1px solid var(--md-outline-variant)}`、`tr:hover td{background:var(--md-surface-1)}` | `style.css:311-337` |
| `<tr>` 元素被 JS 读过吗？ | ❌ **没有。** 行上写了 `data-id`，但**从未被读** —— 事件绑定走的是 `document.querySelectorAll('.row-cb')`，id 从**复选框自己的** `dataset.id` 取 | `app.js:111`（写 `<tr data-id>`）vs `:120-122`（只读 `.row-cb` 的 `e.target.dataset.id`） |

**最后一行是本节最有价值的发现**：`<tr>` 在 downloader 里**纯粹是结构 + `tr:hover` 样式**，
**没有任何 JS 依赖它的存在**。→ 任何「把行结构换掉」的方案（Grid 展平、换自定义标签）
**都不会打断 downloader 的事件逻辑**，只需要重写 `render()` 的模板字符串 + 相应 CSS。

#### radio（当前**不命中**，仅作 Step 4 的前瞻输入）

| 维度 | 事实 | 出处（✅ 已核实） |
|---|---|---|
| `renderEngine` | **字段不存在** = `webview` → **当前不暴露** | `plugin.json`（无该键） |
| 列数 | **4**（复选框 / 名称 / URL / 分组），表头与行严格等宽 4 格 | `static/index.html:76-81`、`static/js/app.js:179-185` |
| `colspan` / `rowspan` | **零** | — |
| 行来源 | **JS 动态**，但形状与 downloader **不同**：`document.createElement('tr')` + `tr.innerHTML = '<td>…'` + `stationTbody.appendChild(tr)`（清空仍是 `stationTbody.innerHTML = ''`） | `app.js:176`（清空）、`:178-186`（逐行 append） |
| 行容器 | `<tbody id="station-tbody">` | `index.html:83` |
| sticky 表头 | ✅ 有，同样是 **CSS** `th { position:sticky; top:0 }` | `static/css/style.css:265-268` |
| 滚动容器 | `.table-wrap { overflow-x:auto; max-height:50vh; overflow-y:auto }` —— 注意它的 `max-height` 在 **CSS 里**，不像 downloader 写在内联 style | `style.css:245-249` |
| 列宽 | 只有 `<th style="width:36px">`（复选框列） | `index.html:77` |
| `<tr>` 元素被 JS 读过吗？ | ❌ 没有（事件走 `stationTbody.querySelectorAll('.station-cb')` + `dataset.idx`） | `app.js:194-210` |

**radio 与 downloader 的差异只有一处值得记**：radio 是**逐行 `appendChild`**，downloader 是
**一次性 `innerHTML`**。这一点在 §1.4 讨论「WidgetElement 自动重建」时会用到。

#### 命中面小结（写给做决策的人）

- **需要支持的能力上限极低**：2 张表、6 列和 4 列、无跨格、单表头、行全动态、sticky 表头、
  两三列固定宽 + 其余自适应、hover 背景。
- **不需要**：`colspan`/`rowspan`、`<tfoot>`、`<caption>`、`<colgroup>`、嵌套表格、
  可排序表头、列拖拽、行选中范围（shift-click）、虚拟滚动。
- **两张表的 JS 都不依赖 `<tr>`/`<td>` 元素本身**，只依赖：① 行容器能拿到（`#tbody` / `#station-tbody`）；
  ② 单元格里的 `.row-cb` / `.station-cb` 能被 `querySelectorAll` 找到。
  → **换结构的插件侧成本 = 重写一段模板字符串 + 改一段 CSS**，不动事件逻辑，不动数据流。

### 1.3 预研已证伪的死路（**引用，不重复论证**）+ 三条本轮新增的核实

`docs/webf-recon-step456.md` §2 已经把下面这些走完了，**不要重新论证**：

| 死路 | 结论 | 出处 |
|---|---|---|
| 机械改写 `table`→`webf-table` | **得到一张空表**。`<webf-table>` 只看**直接 childNodes**，`<thead>`/`<tbody>` 不拆就是 `rows=[]` / `header=null`，**不报错不打日志** | 预研 §2.6 原因 ①（`webf/lib/src/html/table.dart:188-189`） |
| 拆掉 `<tbody>` | downloader 的 `$('#tbody')` 返回 null → `tbody.innerHTML` 抛 TypeError → 整个 `render()` 中断 | 预研 §2.6 原因 ②（`downloader/static/js/app.js:93/96/103`） |
| 靠 `MutationObserver` 自动接管动态行 | **WebF 没有 `MutationObserver`**（`grep -rn "MutationObserver" lib/` 零命中） | 预研 §2.6 原因 ③ |
| `colspan` / `rowspan` | **零支持，且是 Flutter `Table` widget 的天花板**，上游修也修不了 | 预研 §2.4（`table_cell.dart:70-75`） |
| CSS `width` 定列宽 | **完全无效**。只认表头单元格的 `column-width` **属性**；`<webf-table>` 自己的 `column-widths` 属性是**死属性**（setter 存了，`build()` 从不读） | 预研 §2.3、§2.2（`table.dart:79-85` vs `:194`） |
| sticky 表头 | 能做，但必须用 `<webf-table-header sticky>` **属性**（CSS `position:sticky` 对 Flutter `Table` 内部无效），且 sticky 分支用了 `Expanded` → **必须给 CSS 高度**，且列宽必须逐列写死否则表头/表体两次独立 flex 分配会错位 | 预研 §2.5（`table.dart:191/196-236`，`:214` 的 `Expanded`） |
| `defineCustomElement('table', ...)` 覆盖内建标签 | 标签名校验要求含连字符，`'table'` 直接 `ArgumentError`；内部的 `defineOverrideWidgetElement` 虽无校验但属依赖 `lib/src/` 内部实现，且**一个实际问题都不解决** | 预研 §2.7（`webf.dart:147-192`、`element_registry.dart:64-69`） |
| 原生 `<table>` 家族 | Dart 注册表里**一个都没有**，连常量都不存在 → `_UnknownHTMLElement`，`defaultStyle` 是 `display:block` | 预研 §2.7（`element_registry.dart:32-37`、`:80-96`） |

---

#### ⭐ 本轮新增核实 ①（**决定性，直接改变方案 A / C 的可行性判断**）

**`WidgetElement` 的子节点增删会自动触发 Flutter 侧重建 —— 不需要 `MutationObserver`。**

✅ 已核实，`webf/lib/src/widget/widget_element.dart`：

```dart
// :174-183
@nonVirtual @override
dom.Node appendChild(dom.Node child) {
  super.appendChild(child);
  if (state != null) { state!.requestUpdateState(); }   // ← 这一行
  return child;
}
```

`insertBefore`（`:185-193`）、`replaceChild`（`:195-203`）、`removeChild`（`:205-213`）
**四个方法全都是同一形状**，且都标了 `@nonVirtual`（子类改不掉）。

而**来自 JS 的节点插入确实走这条 Dart 路径**：
✅ `bridge/ui_command.dart:291-293` 的 `UICommandType.insertAdjacentNode` →
`view_controller.dart:812-857` 的 `insertAdjacentNode()`，其 `'beforeend'` 分支
（`:853-855`）就是 `target.appendChild(newNode)`，`'beforebegin'` / `'afterbegin'` 分支
（`:841-851`）是 `insertBefore(...)`。→ **JS 的 `appendChild` / `insertBefore` 一定会命中上面那个
`requestUpdateState()`。**

**对方案的影响**：
- 「往 `<webf-table>` / 未来的 `<songloft-table>` 里动态 append 行，表格会不会刷新」——**会**。
  预研 §2.6 原因 ③ 的措辞（「动态插入的行没有自动重跑机制」）**在「垫片要重新改写标签名」这个语境下完全正确**，
  但**不能推广成「WidgetElement 表格不会自我刷新」** —— 那是两件事。
  自定义元素路线因此**不背「动态行不刷新」这口锅**。
- ⚠️ **需实测**：`innerHTML = '...'` 在 C++ 侧被分解成什么命令序列（是 `createElement` ×N +
  `insertAdjacentNode` ×N，还是先建 `DocumentFragment` 再一次插入）**无法从 pub 包核实**
  —— `grep -rn "innerHTML" lib/` **Dart 层零命中**，它整个在 C++/QuickJS 侧。
  `insertAdjacentNode` 里对 `DocumentFragment` 有专门处理（`view_controller.dart:829-835`），
  两条路最终都调 `target.appendChild/insertBefore`，所以**推断上都会触发重建**，但请实测确认。
- ⚠️ **需实测**：`innerHTML = ''`（清空）走的是 `UICommandType.removeNode`（`ui_command.dart:295`）
  还是别的路径，同样无法核实。

#### ⭐ 本轮新增核实 ②（**这条能省掉一次返工，也直接判死了「伪表格 CSS」的最直觉写法**）

**`display: table` / `table-row` / `table-cell` 在 WebF 里不但不支持，还会把元素变成 `inline`
—— 比什么都不写更糟。**

✅ 已核实 `webf/lib/src/css/display.dart`：

- `enum CSSDisplay`（`:13-26`）只有 `inline / block / inlineBlock / flex / inlineFlex /
  grid / inlineGrid / none` —— **没有任何 table 相关取值**；
- `resolveDisplay(...)`（`:49-86`）的 switch **只认这 8 个字符串**，
  `case 'inline': default: return CSSDisplay.inline;`（`:83-85`）
  → 写 `display: table` 落到 `default` → **`inline`**；
- 上游自己也知道：`:80` 有一行注释 `// Note: inline-table would go here when supported`。

**后果**：`table { display: table }` 会让 `<table>` 从 `_UnknownHTMLElement` 的
`display:block`（`element_registry.dart:32-37`）**退化成 `inline`**。
→ **「给 `<table>` 家族补一套 `display:table/table-row/table-cell` 的 CSS 让它像表格」这条路是死的**，
而且是**负收益**的死路。任何伪表格 CSS 只能用 **grid / flex**（见方案 B / D）。

#### ⭐ 本轮新增核实 ③（**方案 A 的技术底座确实存在，但它同时暴露了方案 A 的悖论**）

**`WidgetElement` 可以把 HTML 子节点当普通 WebF 盒子渲染出来** —— 不必自己 `CustomPaint` 画文字。

✅ 已核实两处用法：
- `webf/lib/src/rendering/widget_element_child.dart:20-32` 的类文档给了标准配方：
  `WebFWidgetElementChild(child: WebFHTMLElement(tagName:, controller:, parentElement:,
  children: widgetElement.childNodes.toWidgetList()))`；
- WebF 自己就这么用：`table_cell.dart:191-194`（单元格内容）、`table_cell.dart:175-180`
  （`TableCell(child: WebFWidgetElementChild(child: toWidget()))`）、
  `router_link.dart:130`、`webf.dart:663/713`。

另外 `WidgetElement` 提供了三个正好对表格有用的钩子（✅ `widget_element.dart:107-121`）：
`disableBoxModelPaint`、`allowsInfiniteWidth`、`allowsInfiniteHeight`
（后两个的文档明说是「给内部包了滚动容器的 widget」用的）。

**但这同时是方案 A 的软肋**（详见 §1.4）：一旦决定「单元格内容交回 WebF 自己排版」，
`<songloft-table>` 剩下要提供的价值就只有**跨行共享的列宽**与**sticky 表头**两件
—— 而这两件 **CSS Grid 已经原生给了**（§1.5）。方案 A 于是变成
「用 500 行 Dart 复刻 WebF 已经实现的 193 KB Grid 布局的一个子集」。

#### 顺带：`display: contents` 也不支持

`grep -rn "'contents'" lib/src/css/` **零命中**，`CSSDisplay` 里也没有。
→ 「让 `<tr>` 变成 `display:contents` 从而让 `<td>` 直接成为 grid item」这条**浏览器里的标准招数在 WebF 下用不了**。
Grid 方案必须**在 DOM 层面就没有行包裹元素**（见 §1.5）。

⚠️ `grid-template-columns: subgrid` 的**解析器存在**（`css/grid.dart:551-567` 的 `GridSubgrid`），
但**布局是否真的实现无法从源码核实**（`rendering/grid.dart` 有 193 KB，没通读）。
**不要**把方案建立在 subgrid 上。

### 1.4 方案 A — 自写 Dart 自定义元素 `<songloft-table>`

**形态**：在 `songloft-player/lib/features/home/presentation/render/elements/` 下新增
`songloft_table.dart`，注册 4 个标签
`<songloft-table>` / `<songloft-table-head>` / `<songloft-table-row>` / `<songloft-table-cell>`，
**不用** WebF 自带的 `<webf-table>`，从而绕开 Flutter `Table` widget 的天花板。

#### A.1 先回答三个被点名的设计问题

**① colspan / rowspan 要不要支持？——❌ 不要。**

- 真实命中面里**一个都没有**（§1.2 已核实两张表都是规整矩阵）。
- 支持它意味着不能用 Flutter `Table`（那是天花板），得自己写一个跨格布局算法：
  列宽求解要在「跨 N 列的单元格的最小宽度如何分摊到 N 列」上做迭代
  （CSS 表格算法真正复杂的就是这一段），还要处理跨格重叠冲突、跨格与 sticky 表头的交互。
  **这是几百行有算法难度的代码，为零个真实用例服务。**
- 折中：**明确不支持 + 显式报错**。若单元格上出现 `colspan` / `rowspan` 属性，
  在 `attributeDidUpdate` 里 `debugPrint` 一条「本元素不支持跨格，属性已忽略」
  （沿用 `songloft_progress_ring.dart:75-99` 的既有做法：**静默夹紧非法值 + 补一条日志**）。

**② 怎么把 HTML 子节点结构喂进去？**

✅ 技术上完全可行，配方见 §1.3 新增核实 ③。具体分工建议：

- `<songloft-table-cell>`：`build()` 返回
  `WebFWidgetElementChild(child: WebFHTMLElement(tagName:'DIV', controller:, parentElement:,
  children: widgetElement.childNodes.toWidgetList()))`。
  ⚠️ **注意用 `'DIV'` 而不是 WebF 自己用的 `'SPAN'`** —— `table_cell.dart:191` 用 SPAN，
  而 SPAN 是 inline，单元格里放 block 子节点（downloader 的 `<span class="song-source">` 还好，
  但复选框 `<input>` 与将来任何 block 内容就不好说）的行为**预研 §2.6 已列为需实测**。
  用 DIV 从一开始避开这个坑。
- `<songloft-table-row>`：**必须自己渲染自己的单元格**（`Row` + 按共享列宽给每个 cell 套
  `SizedBox`/`Flexible`），**不要**学 `<webf-table-row>` 那样 `build()` 返回
  `SizedBox.shrink()` 让父表去 `buildCellChildren()`。
  理由是 §1.3 新增核实 ① 的直接推论：`requestUpdateState` 只通知**发生子节点变化的那个元素自己的** state
  （✅ `css/render_style.dart:908-941` 的 `requestWidgetToRebuild` 只遍历
  `target` 自己的 `_widgetRenderObjects`，**不向上走祖先**）。
  → 往一个**行**里插单元格时，只有**行**的 state 重建；
  若单元格是由**表**的 `build()` 产出的（WebF 自带表的做法），**表不会重建 → 新单元格不显示**。
  这是 `<webf-table>` 的一个真实缺陷，自己写就不要复刻。
- `<songloft-table>`：`build()` 收集直接子节点里的 head/rows，算出共享列宽，
  用 `Column` + 外层滚动组织；**列宽通过 InheritedWidget（或直接构造参数）下发给行**。
- 列宽求解：读表头单元格的 CSS `width`（`renderStyle.width`，`isAuto` 时参与自适应），
  或一个 `column-width` 属性。**不要**复刻 WebF 那个「删属性 → 宽度变 0 → 整表不可见」的陷阱
  （预研 §2.2 第 2 条）。

**③ 动态行怎么办（无 MutationObserver）？**

**这一项不是问题** —— §1.3 新增核实 ① 已证：`WidgetElement.appendChild` 等四个方法
`@nonVirtual` 地调 `state!.requestUpdateState()`，而 JS 侧插入确实走这条 Dart 路径
（`ui_command.dart:291` → `view_controller.dart:853-855`）。
只要**插件直接产出 `<songloft-table-row>` 并 append 到 `<songloft-table>`**，
表格自动重建。**前提是插件必须改源码**（因为标签名必须带连字符，宿主无法顶替 `<table>`
—— `songloft_custom_elements.dart:21-29` 的头注释已经把这条约束写死了）。

⚠️ **需实测**：`innerHTML` 批量重写路径（downloader 就是这个形状）是否同样触发重建，见 §1.3 新增核实 ①。

#### A.2 必须遵守的两条既有约束（**先读这两个先例再动手**）

从 `songloft_progress_ring.dart` 与 `songloft_slider.dart` 抄结论，不要重新发现：

1. **本目录只允许 import `flutter` 与 `webf`**
   （`songloft_progress_ring.dart:36-42` 的头注释）：验证探针
   `songloft-player/scripts/webf-verify/` 是独立 Flutter package，
   `entrypoint.sh` 会把**本目录原样拷进探针的 `lib/elements/`** 后编译
   （源目录缺失直接 `exit 1`，sha1 写到 `out/elements.sha1`）。
   一旦引入产品依赖（riverpod / `../` 上层文件），探针就编不过。
2. **颜色跟随插件页主题只有一条路：CSS `color`（currentColor 语义）**
   （`songloft_progress_ring.dart:121-147`，那段注释是实测结论）：
   - ✅ 可行：读 `renderStyle.color.value`。`color` 可继承，而 `common.css` 把文字色绑到
     `--md-on-surface` → **零配置就跟主题**；实测运行时改变量会重绘。
   - ❌ 属性里写 `var(--md-primary)`：`CSSColor.parseColor` 拿到原始 `var(...)` 串返回 null。
   - ❌ 自己在 build 里 `renderStyle.getCSSVariable()` 展开：首帧对，但主题切换后
     **永久停在旧值**（变量变更通知只驱动登记在 `target.style` 里含 `var(` 的属性）。
   - ❌ `getComputedStyle(el).getPropertyValue('--md-primary')`：WebF **不暴露自定义属性**，返回空串。

   **对表格的直接影响**：表格的分隔线色（`--md-outline-variant`）、
   表头文字色（`--md-on-surface-variant`）、hover 背景（`--md-surface-1`）
   **一个都不能从 CSS 变量读到 Dart 侧**。要么这些视觉全部交给**单元格自己的 CSS**
   （单元格内容走 `WebFHTMLElement` → 它的 border/background/padding 由 WebF 正常渲染 →
   ✅ 这是可行且推荐的），要么就得给元素加一堆 `divider-color=` / `header-bg=` 属性让插件写死色值
   （❌ 那就跟不上主题了）。
   → **结论：`<songloft-table>` 自己应该几乎不画东西，只做布局**（列宽 + sticky 定位 + 滚动），
   视觉全部由单元格 CSS 承担。**而「只做布局」的东西，Grid 已经有了。**
3. **不要用 Material widget**（`songloft_slider.dart` 的教训，交接文档 §2.3 Step 3）：
   Material 组件从宿主 App 的 `Theme` 取色，跟不上插件页的 `--md-*`。
4. **手势**：本方案不需要手势（表格本身无交互，复选框是插件自己的 `<input>`），
   但**滚动**要小心 —— 若表格内部自带滚动容器，会与页面滚动竞争。
   `songloft_slider.dart` 那条教训（裸 `Listener` 不进竞技场 → 滚动与控件同时响应）在这里的对应物是：
   **优先不要在元素内部起滚动**，让外层 `.table-wrap` 的 CSS `overflow` 负责
   （✅ WebF 支持 overflow 滚动，`element.dart:886` 有 `AddScrollerUpdateReason`）。
   WebF 自带表塞了**两层** `SingleChildScrollView`（`table.dart:238-241`）并与外层 CSS overflow 嵌套，
   那正是预研 §2.6 原因 ④ 列出的「双层滚动」问题，**别复刻**。

#### A.3 成本 / 上限 / 风险

| 维度 | 评估 |
|---|---|
| **宿主改动量** | **大。** 新增 `songloft_table.dart` 估 500–700 行（4 个元素 + 列宽求解 + sticky + 属性校验 + 注释密度按本项目标准）；`songloft_custom_elements.dart` 加 4 行注册；验证探针 `probe.html` 加一组检查（纵向预算有限，交接文档 §5 已警告）；插件开发指南中英双语各加一节。**这是本文档四个方案里宿主成本最高的**。 |
| **插件改动量** | **中。** downloader 要重写 `render()` 的模板（`<tr>`→`<songloft-table-row>` 等）、去掉 `#tbody` 这层、改一段 CSS；**且必须按引擎分叉**（`document.documentElement.classList.contains('webf-engine')`，Step 1 已在 `installEarly()` 打了这个 class），因为浏览器/WebView 路径完全不认这些标签 → **两套模板长期并存**。 |
| **能力上限** | 列宽共享 ✅、sticky 表头 ✅、hover ✅（交给 CSS）、border-collapse ⚠️（做不到真正的合并边框，只能靠单元格单边 border 近似）、colspan/rowspan ❌（刻意不做）、`<caption>`/`<tfoot>`/`<colgroup>` ❌。 |
| **未知量 / 需实测** | ① `innerHTML` 批量重写是否触发 `WidgetElement` 重建（§1.3 核实 ①）；② `WebFHTMLElement(tagName:'DIV')` 包裹的单元格内容，其 CSS padding/border/background 是否正常生效（预研 §2.6 原因 ④ 已列为需实测，本方案继承）；③ 元素在无界高度约束下（`allowsInfiniteHeight`）的表现；④ 表头/表体列数不一致时的降级行为（自己写就能自己定，但要测）。 |
| **风险** | **① 悖论性的低性价比**（§1.3 核实 ③）：为了让单元格内容排版正确，必然把内容交回 `WebFHTMLElement`；那 `<songloft-table>` 剩下的价值只有「跨行共享列宽 + sticky」，而 **CSS Grid 已原生提供这两件**（`fr` / `minmax` / `repeat` 已核实存在，`position:sticky` 已核实在渲染器里实现）。等于用几百行 Dart 复刻一个已存在的 193 KB Grid 实现的子集。**② 长期维护负担**：这是宿主对插件作者的一个新公开契约，一旦发布就要向后兼容；而它只服务一个第一方插件的一张表。**③ 第三方作者的接受成本**：他们要为 WebF 单独写一套表格标签，而 Grid 方案他们只需要写标准 CSS。 |

**A 的唯一强场景**：如果将来出现「必须支持 colspan/rowspan 或复杂表格语义」的真实需求，
Grid 做不到而自定义元素能做到。**但那个需求现在不存在**，且真出现时再做也不晚
（届时已有 Grid 方案的插件不受影响）。

### 1.5 方案 B — 放弃表格语义，改 CSS Grid（**推荐**，记作 B'）

**形态**：downloader 把那张表改成一个 CSS Grid 容器；`<table>/<thead>/<tbody>/<tr>` 四层
DOM **全部消失**，只剩「一个 grid 容器 + 一堆单元格 div」。**宿主零改动。**

#### B.1 关键事实：Grid 在 WebF 里是**真的**实现了，而且够用

交接文档 §3.3 的「CSS Grid 已实现（experimental，193 KB 实现，issue 原文写不支持是错的）」
本轮做了针对性的复核，✅ **已核实到具体特性**：

- 实现体量：`webf/lib/src/css/grid.dart` 29 KB（解析/CSSOM）+ `lib/src/rendering/grid.dart`
  **193 KB**（布局）。
- `CSSDisplay` 里**确实有** `grid` / `inlineGrid`（`css/display.dart:21-23`），
  `resolveDisplay` 的 switch 认 `'grid'` / `'inline-grid'`（`:69-72`）。
- 轨道尺寸语法覆盖到（`css/grid.dart`）：
  `<n>fr`（`:377-382`）、`minmax(...)`（`:395`）、`repeat(...)`（`:503`）、
  `auto-fill`（`:445`）、`auto-fit`（`:451`）、`auto` / `span` 关键字（`:255`、`:625`）。
- `grid-auto-flow` 有实现（`css/grid.dart:821-826`，默认 `row`）→ **自动换行放置就是行为的来源**。
- 命名线（line names）也在（`_parseLineNames`）。

**这套语法已经完全覆盖 §1.2 的需求**：downloader 的
`grid-template-columns: 36px minmax(0,2fr) minmax(0,1fr) minmax(0,1fr) minmax(0,1fr) 60px`
一行就等价于它原来那张表的列宽意图（两列写死 + 4 列自适应），
而且**跨行共享列宽是 Grid 的定义性行为**，不需要任何额外机制。

#### B.2 sticky 表头怎么做

✅ **`position: sticky` 在 WebF 里有真实实现**（不只是枚举值）：
- `css/position.dart:17` 有 `sticky`，`:182` 的解析返回 `CSSPositionType.sticky`；
- 渲染侧 `rendering/widget.dart:37-38` 维护 `stickyChildren` 集合，
  `:675-717` 在布局里收集 sticky 子节点并调 `CSSPositionedLayout.applyStickyChildOffset(...)`；
- `rendering/box_model.dart:306-307` 有 `stickyStatus`（relative ↔ fixed 切换），
  `:1413-1414` 与 `:1560` 有 sticky 的专门分支。

写法：表头那 6 个单元格 div 各自 `position: sticky; top: 0; background: var(--md-surface); z-index: 1`
—— **这与 downloader 现在 `th { position:sticky; top:0 }` 的写法几乎一模一样**
（`style.css:325-327`），只是选择器从 `th` 换成 `.tbl-th`。
→ **sticky 表头在 B 方案里几乎是零成本迁移**，因为它原本就不是靠表格语义实现的。

⚠️ **需实测**：sticky 在 **grid 容器的子项**上是否与在 block 流里一样工作
（`rendering/widget.dart:675-717` 的 sticky 收集逻辑在哪些布局路径上跑到，没有通读 193 KB 的
`rendering/grid.dart` 确认）。**这是 B 方案唯一的真实风险点，必须先测**（见 §3 实测清单第 1 项）。
兜底：若 grid 子项的 sticky 不生效，退一步用「表头单独一个 grid 容器 + 表体一个 grid 容器 +
两者用同一份 `grid-template-columns` 写死列宽」的双容器写法
（代价：不能再用 `fr` 自适应，列宽必须写成固定值或百分比；这与 `<webf-table>` sticky 模式
「必须逐列写死 column-width」的约束是同一类问题，见预研 §2.3 最后一条）。

#### B.3 具体改法（给 downloader 的骨架）

**HTML（`index.html:106-119` 那一段）** —— `<table>/<thead>/<tbody>/<tr>` 全删：

```html
<div class="table-wrap" style="max-height:calc(100vh - 380px);overflow-y:auto">
  <div class="tbl" id="tbl">
    <!-- 表头：6 个单元格，静态 -->
    <div class="tbl-th"><input type="checkbox" class="cb" id="cb-all" aria-label="全选"></div>
    <div class="tbl-th">标题</div>
    <div class="tbl-th">艺术家</div>
    <div class="tbl-th">专辑</div>
    <div class="tbl-th">来源</div>
    <div class="tbl-th">状态</div>
    <!-- 行单元格由 JS 追加到本容器末尾 -->
  </div>
  <div class="empty-state" id="empty">…</div>
</div>
```

**CSS** —— **注意不能写 `display:table*`**（§1.3 新增核实 ②，那会退化成 `inline`）：

```css
.tbl {
  display: grid;
  grid-template-columns: 36px minmax(0,2fr) minmax(0,1fr) minmax(0,1fr) minmax(0,1fr) 60px;
  /* border-collapse 无对应物，用单元格单边 border 近似（原本 th/td 也只有 border-bottom） */
}
.tbl-th, .tbl-td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--md-outline-variant);
  min-width: 0;             /* 配合 minmax(0,…) 让长文本能收缩而不是撑爆列 */
  overflow: hidden;
  text-overflow: ellipsis;
}
.tbl-th {
  position: sticky; top: 0; z-index: 1;
  background: var(--md-surface);
  text-align: left; font-size: 11px; font-weight: 500;
  color: var(--md-on-surface-variant);
  text-transform: uppercase; letter-spacing: .5px;
}
.tbl-td { font-size: 13px; color: var(--md-on-surface); }
```

**JS（`app.js:92-127` 的 `render()`）** —— 两处改动：

1. 不再有 `#tbody`。清空要**只删行、保留表头**：
   ```js
   var tbl = $('#tbl');
   // 删掉所有非表头子节点（表头是前 6 个 .tbl-th，静态存在）
   tbl.querySelectorAll('.tbl-td').forEach(function (n) { n.remove(); });
   ```
   ⚠️ 或者更稳的写法：把表头也一起由 JS 渲染，`tbl.innerHTML = headerHtml + rowsHtml`
   —— **一次 `innerHTML`，与现在的代码形状最接近，改动量最小**，而且规避了
   「`querySelectorAll(...).forEach` 在 WebF 下可用吗」的疑问
   （交接文档 §3.3 已核实 `NodeList.forEach` 可用，但少一个依赖总是好的）。
2. 行模板：`<tr>` 与 `<td>` 全换成 6 个 `<div class="tbl-td">`，**顺序不变**。
   `data-id` 直接删掉（§1.2 已核实**从未被读**）；`.row-cb` 的 class 与 `data-id` 保持原样
   → **`app.js:120-127` 的事件绑定一行都不用改**。

**`tr:hover td` 的降级**：Grid 展平后**没有行元素**，纯 CSS 做不到「hover 整行变色」。
三个选项：
- **(a) 接受降级**：改成 `.tbl-td:hover { background: var(--md-surface-1) }`（只高亮单个格子）。
  最省事，观感明显变差。
- **(b) 保留行元素，行本身也是 grid**：`<div class="tbl-tr">` 设 `display:grid` +
  **同一份** `grid-template-columns`，外层容器改成普通 block/flex 纵向堆叠。
  hover 恢复（`.tbl-tr:hover .tbl-td`），但**列宽不再跨行共享**
  （每一行独立算 `fr` → 内容不同就会错位，正是预研 §2.3 指出的 sticky 模式错位问题的同构版本）。
  → 只有把 6 列**全部写死**（`36px 2fr 1fr 1fr 1fr 60px` 里的 `fr` 全换成百分比或固定值）
  才能保证对齐。**这是「hover」与「自适应列宽」的二选一**，取舍要显式做。
- **(c) 用 JS 补 hover**：给每个单元格加 `data-row="<i>"`，`mouseover`/`mouseout` 时批量改同 row 的背景。
  能同时保留自适应列宽与整行 hover，代价是 ~15 行 JS。
  ⚠️ **需实测**：WebF 是否派发 `mouseover` / `mouseout`（Dart 层未核实；
  交接文档 §3.1 提到 WebF 只有一个 tap recognizer，鼠标事件覆盖面不明）。
  **移动端本来就没有 hover，这条在真实使用场景里价值有限** —— 建议先走 (a)，
  真有人抱怨再考虑 (b)/(c)。

#### B.4 成本 / 上限 / 风险

| 维度 | 评估 |
|---|---|
| **宿主改动量** | **零**（这是 B 最大的优势）。不新增 Dart 元素、不动 `common.js` 的垫片链、不动验证探针、不新增公开契约。**唯一建议的宿主动作是 §1.9 那个「只警告不改写」的探测垫片，它与 B 正交、也服务于其它所有方案。** |
| **插件改动量** | **小，且是一次性的。** downloader：`index.html` 删 4 层标签写 6 个表头 div（约 -14/+10 行）、`style.css` 改约 30 行、`app.js` 的 `render()` 改模板字符串（约 10 行）。**事件逻辑、数据流、筛选、下载调用一行不动。** |
| **是否需要按引擎分叉？** | **不需要，这是 B 相对 A / C 的决定性优势。** CSS Grid 是**标准 CSS**，浏览器、WebView、WebF **三条渲染路径共用同一套 HTML + CSS + JS**，观感一致，只维护一份代码。A / C 都必须写「WebF 一套、浏览器一套」并靠 `webf-engine` class 分叉，长期双份维护。 |
| **能力上限** | 列宽共享 ✅（Grid 定义性行为）、sticky 表头 ✅（需实测 grid 子项，有双容器兜底）、自适应列宽 ✅（`fr`/`minmax`）、单元格 padding/border/背景 ✅（普通 CSS 盒模型）、`tr:hover` ⚠️（三个选项见 B.3）、colspan/rowspan ✅**其实能做**（`grid-column: span 2`，`css/grid.dart:625` 有 `span` 解析）—— 这一点**反而比方案 A 强**，尽管当前没人需要。 |
| **未知量 / 需实测** | ① **sticky 是否在 grid 子项上生效**（最重要，见 §3 实测清单第 1 项）；② `minmax(0,1fr)` + `overflow:hidden` + `text-overflow:ellipsis` 在 WebF 的长文本收缩行为；③ `mouseover`/`mouseout` 是否派发（仅选项 (c) 需要）。 |
| **风险** | **低。** 最坏情况是 sticky 在 grid 子项上不работ，退双容器写法 + 写死列宽 —— 那仍然比 A/C/D 任何一个都便宜，且仍是零宿主改动。**另一个要如实说的代价**：`<table>` 有语义/无障碍价值（屏幕阅读器的表格导航、`aria-label="下载列表"`）。Grid 展平后这些丢失；理论上可以补 `role="table"/"row"/"cell"` ARIA 属性，⚠️ **但 WebF 的语义树对 `role` 属性的支持程度无法从源码核实**（`rendering/box_model.dart` 有 `markNeedsSemanticsUpdate`，但没追 role 映射）。**如实记：这是一项确定的降级，且在 WebF 路径下本来就已经丢了**（`_UnknownHTMLElement` 没有任何表格语义），所以 B 并不比现状更差。 |

#### B.5 为什么是 Grid 而不是 flex

flex 做表格必须「每行一个 flex 容器」，于是**列宽在行之间不共享** —— 要对齐只能给每列写死宽度或百分比，
长文本一变就错位。Grid 的整个价值就在于**跨行共享轨道**。
→ **不要退化成 flex 方案**，除非 §3 实测第 1 项发现 grid 在 WebF 下有更严重的问题。

### 1.6 方案 C — 垫片 + 拆 tbody + 改插件 JS（既定方案的补救版）

**形态**：宿主在 `common.js` 的 ready 相加一个 `tableShim`：把 `<table>` 家族改名成
`<webf-table>` 家族、**展平掉 `<thead>`/`<tbody>`**、把 `<th style="width:36px">` 翻译成
`column-width="36"`、把 `th{position:sticky}` 的意图翻译成 `<webf-table-header sticky>` 属性；
同时改 downloader 的 JS 不再依赖 `#tbody`。

#### C.1 它到底要做多少事

预研 §2.8 已经列了「机械改写会丢什么」。补救版必须**每一条都补**：

| 要补的 | 怎么补 | 难度 |
|---|---|---|
| 拆 `<thead>`/`<tbody>`，行提到 `<webf-table>` 直接子节点 | 遍历 + `appendChild` 搬移 | 易 |
| 可能存在的**隐式 `<tbody>`** | ⚠️ **需实测**：静态 HTML 写 `<table><tr>` 时 gumbo 有没有插一层 tbody。C++ 侧的 `html_tag_names.json5` **不随 pub 包发布**（`find . -name "*.json5"` 零结果），无法核实 | 中（取决于实测结果） |
| `<th style="width:Npx">` → `column-width="N"` | 正则/`style.width` 读取 + `setAttribute` | 易，但要**避开 `removeAttribute('column-width')` 陷阱**（预研 §2.2 第 2 条：deleter 会把宽度设成 `0.0`，`_getDefaultColumnWidth` 判 `!= null` → `FixedColumnWidth(0.0)` → **整表不可见**） |
| CSS sticky → `sticky` 属性 | 需要**从 CSS 反推意图** —— 垫片拿不到「`th` 有没有 `position:sticky`」的可靠信号（`getComputedStyle` 能读 `position` 吗？⚠️ 未核实），实际只能靠**插件显式加一个 `data-sl-sticky` 标记** | 中，**且需要插件配合** |
| sticky 模式下**每列都要写死 `column-width`** | 预研 §2.3 最后一条：sticky 是两个独立 `Table`，`FlexColumnWidth()` 会两次独立分配 → 错位。downloader 有 4 列是自适应的 → **必须全部改成固定宽**，这是**可见的产品降级**（长标题不再自适应） | 中，且是降级 |
| sticky 模式的 `Expanded` 需要有界高度 | 预研 §2.5 条件二：downloader 用的是 `max-height`（内联，`index.html:106`）**不是** `height` → ⚠️ 需实测；建议直接写死 `height` | 中，且是降级（写死高度） |
| downloader 的 `tbody.innerHTML` | 必须改插件 JS：去掉 `#tbody` 概念，改成往 `<webf-table>` append 行 | 易，但**改动量与方案 B 相当** |
| 动态行的标签改写 | ✅ **不需要重跑垫片**，只要插件直接产出 `<webf-table-row>`（§1.3 核实 ①：`WidgetElement.appendChild` 自动重建）。**但插件既然要直接产出新标签，垫片对 downloader 就完全没用了** | — |
| 双层滚动 | `<webf-table>` 自带**两层** `SingleChildScrollView`（`table.dart:238-241`），与外层 `.table-wrap` 的 `overflow` 嵌套 → 需要**删掉插件的外层 overflow** | 易，但又一处插件改动 |
| `border-collapse` / `<caption>` / `<tfoot>` / `<colgroup>` | 无对应物，丢弃 | — |
| `tr:hover td` | 行只以 `renderStyle.decoration` 参与（`table.dart:210/225/228/256`），`:hover` 能否触发重建 ⚠️ 需实测（`WebFTableRowState.build()` 返回 `SizedBox.shrink()`） | 未知 |
| 单元格 CSS（padding/border） | `toTableCell()` 把内容包进 `tagName:'SPAN'` 的 `WebFHTMLElement`（`table_cell.dart:191`），**SPAN 是 inline** → 单元格里放 block 子节点的行为 ⚠️ 需实测 | 未知 |
| **`<webf-table-row>` 里动态加单元格不刷新** | §1.4 A.1 ② 已核实的 `<webf-table>` 真实缺陷：`requestWidgetToRebuild` 只通知**发生变化的那个元素自己的** state（`css/render_style.dart:932-940`，不向上走祖先），而单元格 widget 是**表**的 `build()` 产出的 → 往行里插单元格，表不重建。downloader 每次整体重写行所以不踩，但这是个埋着的坑 | 无法在垫片层修（要改上游） |

#### C.2 致命的结构性矛盾

**C 的自我否定在于**：

1. downloader 是**唯一命中的插件**（§1.1），而它**必须改 JS**（`#tbody` 是硬阻塞，预研 §2.6 原因 ②）。
2. 既然要改 JS，插件就会**直接产出 `<webf-table-row>`**（§1.3 核实 ① 已证这样动态行能自动刷新）。
3. → **垫片对 downloader 一行都用不上。** 垫片剩下的服务对象只有「没改过源码的第三方插件」，
   可是对那些插件垫片**只能给出一个残废的表格**（列宽只有写死一种、sticky 要插件配合、
   `tr:hover` 未知、单元格 block 内容未知、colspan 直接错列或 assert）。
4. → **宿主付出「一层脆弱的改写垫片 + 一堆需实测的未知量」，换来的是一个不可靠的兜底**，
   而这个兜底还不如「一行明确的 console.warn 指路」有用（§1.9）。

**另外**：`colspan` 在 C 下**比不改还危险**。不改写时是一坨嵌套 block（丑但能看）；
改写后列数不齐 → Flutter `Table` **assert 失败或错列**（预研 §2.4）。
垫片必须在检测到 `colspan`/`rowspan` 时**主动放弃改写那张表**，
否则 C 会把「丑」变成「崩」。这条**必须写进实现**。

#### C.3 成本 / 上限 / 风险

| 维度 | 评估 |
|---|---|
| **宿主改动量** | **中大。** `common.js` 加一个约 150–250 行的垫片（要处理展平、改名、列宽翻译、sticky 标记、colspan 放弃分支、幂等标记、`SongloftPlugin.applyShims` 重入），加验证探针一组用例，加开发文档双语一节。 |
| **插件改动量** | **与方案 B 相当**（downloader 照样要改 HTML/CSS/JS），**但外加**：必须按引擎分叉（浏览器路径不认 `<webf-table-*>` / `column-width`）、必须写死列宽（降级）、必须写死高度（降级）、必须删外层 overflow。 |
| **能力上限** | **明确低于 B**：列宽只能写死（sticky 模式下）、`border-collapse` 丢、`tr:hover` 未知、colspan/rowspan 必须放弃改写。 |
| **未知量 / 需实测** | 上表里 **6 处**「需实测」（隐式 tbody、`Expanded` + `max-height`、sticky 错位程度、`tr:hover`、单元格 SPAN 包裹的 block 内容、`getComputedStyle` 能否读 `position`）。**这是四个方案里未知量最多的。** |
| **风险** | **高。** ① 全部押在 WebF 自带 `<webf-table>` 上，而它就是 Flutter `Table` 的薄封装，天花板由上游锁死（0.x beta，main 分支自 2026-04-19 静默，交接文档 §6）；② 改写垫片在「插件动态改 DOM」这类场景天生脆弱，`detailsShim` 已经暴露过同类边界（交接文档 §6：垫片跑完后追加的直接子节点会永久留在折叠容器外）；③ 有把「丑」变成「崩」的路径（colspan）。 |

---

### 1.7 方案 D — 接受降级 / 伪表格 CSS

交接文档里 `backdrop-filter` / `mask-image` / `color-mix()` 就是按「接受降级」定的。
D 分两个子方案。

#### D.1 D-纯降级：什么都不做

**实际观感有多糟？** 基于源码可以准确推演（✅ 已核实链条）：

- `<table>` / `<thead>` / `<tbody>` / `<tr>` / `<th>` / `<td>` 全部 → `_UnknownHTMLElement`
  （`element_registry.dart:80-96`，`creator == null`），`defaultStyle` = `{'display': 'block'}`
  （`:32-37`）；
- 于是它们**各自成为一个块级盒**，按文档流**纵向堆叠**；
- 插件的 CSS 仍然生效（选择器 `table`/`th`/`td` 是普通标签选择器，WebF 的 CSS 引擎照常匹配）：
  `th`/`td` 的 `padding:10px 12px` + `border-bottom` 会被应用，
  `table{width:100%}` 会被应用（block 盒子的 width 有效），
  `th{position:sticky;top:0;background:var(--md-surface)}` **也会被应用**（sticky 已实现，§1.5 B.2）。

→ **实际观感**：每一行的 6 个单元格**竖着排成 6 行**，每个占满整行宽度、
各带一条下边框。downloader 那张表本来是 6×N，会变成 **6N 行的一长条纵向列表**，
每条只有一个字段的值、**没有字段名**（表头也竖着排在最上面成 6 行标签）。
30 首歌 = 180 行无标签文本。**这不是「样式差一点」，而是「彻底不可读」**。
另外 6 个表头都是 `position:sticky; top:0` → 它们会**互相叠在同一个位置**。

**结论：D-纯降级不可接受。** 这与 `backdrop-filter`（毛玻璃变纯色，功能完全不受影响）
不是同一个量级——那是**视觉降级**，这是**信息结构丢失**。

#### D.2 D-伪表格：靠插件加几条 CSS 让 block 布局看起来像表格

**⚠️ 最直觉的写法是死路**（§1.3 新增核实 ②）：
`table{display:table} tr{display:table-row} td{display:table-cell}` 会让这些元素
**从 `block` 退化成 `inline`**（`css/display.dart:83-85` 的 `default: return CSSDisplay.inline`），
**比什么都不写更糟**。

**可行的伪表格只能用 flex 或 grid**，于是它有两种形态：

- **D.2a 保留 `<table>` 标签，给它们套 flex/grid**：
  ```css
  table { display: block }              /* 或不写，默认已是 block */
  thead, tbody { display: block }
  tr { display: grid; grid-template-columns: 36px 2fr 1fr 1fr 1fr 60px }
  th, td { min-width: 0; overflow: hidden }
  ```
  ✅ 这**确实可行**（grid 已实现，`<tr>` 是普通 block 盒子，设 `display:grid` 后其 `<td>` 子节点
  成为 grid item）。
  ⚠️ **但列宽不跨行共享** —— 每个 `<tr>` 是独立的 grid 容器，`fr` 各算一次
  → 内容不同就错位。要对齐**必须把 6 列全部写死**（固定 px 或百分比），失去自适应。
  这与 §1.5 B.3 选项 (b) 是同一个权衡，只是多留了一层无用的 `<table>/<thead>/<tbody>` DOM。
  **好处**：浏览器路径下这些 CSS 是「无害的覆盖」（浏览器里 `display:grid` 于 `<tr>` 也生效且行为一致），
  所以**同样不需要按引擎分叉**，且 JS 一行不用改（`#tbody` 还在！）。
  **这是四个方案里插件改动量最小的一个 —— 只改 CSS。**
- **D.2b 保留标签 + 单一 grid 容器**：让 `<tbody>` 成为唯一的 grid 容器、`<tr>` 变
  `display:contents` 从而透明化 —— ❌ **死路**，`display: contents` 在 WebF **不支持**
  （§1.3 末尾已核实，`CSSDisplay` 无该值 → 落 `default` → `inline`）。

#### D.3 成本 / 上限 / 风险

| 维度 | D.1 纯降级 | D.2a 伪表格（只改 CSS） |
|---|---|---|
| **宿主改动量** | 零 | 零 |
| **插件改动量** | 零 | **极小：只改 CSS**（约 10 行），HTML 与 JS 一行不动 |
| **是否需引擎分叉** | — | **不需要**（同一套 CSS 在浏览器里行为一致） |
| **能力上限** | 无（信息结构丢失） | 列宽**必须写死**（不跨行共享）、sticky 表头 ⚠️ 需实测（`<tr>` 变 grid 后表头那一行的 sticky）、`tr:hover` ✅ **保留**（行元素还在！）、colspan/rowspan ❌ |
| **未知量 / 需实测** | 无（结论已由源码确定） | ① `<tr>` 设 `display:grid` 后 `<td>` 是否正确成为 grid item（`css/display.dart:120-140` 的 `effectiveDisplay` 会把 grid item **blockify**，看起来是对的，但没实测）；② 表头行的 sticky |
| **风险** | **不可接受**（见 D.1） | **低**，且**可作为 B 的兜底**：若 §3 实测发现 grid 子项 sticky 不行、或不想动 HTML/JS，D.2a 是一个「只改 CSS、保留全部 JS 与 hover」的次优解 |

**D.2a 值得单独记一笔**：它是**唯一一个「插件只改 CSS」的方案**。
如果实施者时间极紧、或想先出一个止血版本，D.2a 是最快的可用形态
（代价：列宽写死，长标题不再自适应）。**B 与 D.2a 不冲突，可以先 D.2a 再演进到 B。**

### 1.8 方案对比与推荐

#### 总表

| | **A** 自写 `<songloft-table>` | **B'** CSS Grid（改插件 HTML/CSS/JS） | **C** `<webf-table>` 垫片 + 改插件 | **D.1** 纯降级 | **D.2a** 伪表格（只改 CSS） |
|---|---|---|---|---|---|
| 宿主改动量 | **大**（500–700 行 Dart + 探针 + 双语文档 + 新公开契约） | **零** | 中大（150–250 行 JS 垫片 + 探针 + 双语文档） | 零 | 零 |
| 插件改动量 | 中（HTML+CSS+JS，**且双套模板**） | 小（HTML+CSS+JS，**单套**） | 中（HTML+CSS+JS，**且双套模板**） | 零 | **极小（仅 CSS）** |
| 需按引擎分叉？ | ✅ 需要 | ❌ 不需要 | ✅ 需要 | — | ❌ 不需要 |
| 跨行共享列宽 | ✅ | ✅ | 仅写死时 | ❌ | ❌（必须写死） |
| 自适应列宽 | ✅ | ✅ | ❌（sticky 下必须写死） | ❌ | ❌ |
| sticky 表头 | ✅ | ⚠️ 需实测（有兜底） | ✅ 但要插件配合 + 写死高度 | ❌（6 个表头互叠） | ⚠️ 需实测 |
| `tr:hover` 整行 | ✅ | ⚠️ 降级/取舍（B.3 三选项） | ⚠️ 未知 | ✅（但已不成表） | ✅ **保留** |
| colspan/rowspan | ❌（刻意不做） | ✅ `grid-column: span N` | ❌（且有崩溃路径） | ❌ | ❌ |
| 未知量数量 | 4 | **1 个关键 + 2 个次要** | **6** | 0 | 2 |
| 长期维护负担 | **高**（新公开契约，向后兼容义务） | **零** | 中（垫片长期维护 + 押注上游） | 零 | 零 |
| 上游依赖风险 | 只依赖 `WidgetElement`（稳定基座） | 依赖 Grid + sticky（已实现，193 KB） | **押注 `<webf-table>`/Flutter `Table`**，天花板由上游锁死 | — | 依赖 Grid |
| 可读性结果 | 好 | **好** | 中 | **不可接受** | 中上 |

#### 推荐：**B'（CSS Grid） + §1.9 的宿主侧「只警告不改写」垫片**

**理由，按权重排序：**

1. **命中面只有一个第一方插件的一张 6 列规整表**（§1.1 / §1.2 已核实）。
   在这个规模下，宿主长出一个新元素（A）或一层改写垫片（C）**都是为一张表建一座桥**。
   而这张表的所属仓库就在 `.gitmodules` 里，改它不需要求任何人。
2. **B' 是唯一「三条渲染路径共用一套代码」的方案。**
   A 和 C 都必须靠 `webf-engine` class 分叉出两套模板，
   意味着 downloader 从此长期维护两份行渲染逻辑、两份表格 CSS，
   而其中一份（WebF 那份）在本机**永远跑不到**（宿主 glibc 2.35 < 2.38，交接文档 §5），
   只能靠容器验证 —— **双份实现 + 单份可测**是最容易长期腐化的组合。
3. **A 存在结构性悖论**（§1.3 核实 ③ + §1.4 A.3）：为了让单元格内容排版正确，
   必然把内容交回 `WebFHTMLElement`；那 `<songloft-table>` 剩下的价值就只有
   「跨行共享列宽 + sticky」两件，**而这正是 CSS Grid 的定义性能力**，
   且 WebF 里已有 193 KB 的 Grid 布局实现 + 已实现的 sticky。
   用几百行 Dart 复刻它的子集，还要背上一个新的公开契约。
4. **C 自我否定**（§1.2）：唯一命中的插件必须改 JS，改了 JS 之后垫片对它一行都用不上；
   垫片剩下的服务对象只有第三方，而对第三方它只能给出一个残废表格，还有把「丑」变成「崩」的路径（colspan）。
   **6 个需实测项**也是四个方案里最多的。
5. **D.1 不可接受**（§1.7 D.1：6×N 表变 6N 行无标签文本，6 个 sticky 表头互叠），
   这是信息结构丢失，不能类比 `backdrop-filter` 那种视觉降级。
6. **B' 的唯一关键未知量（grid 子项上的 sticky）有明确兜底**（双容器 + 写死列宽），
   兜底之后仍然是零宿主改动。且 **D.2a 可以作为更保守的第二兜底**（只改 CSS）。

**取舍要显式承认的两件事**（不要在实施时才发现）：

- **`tr:hover` 整行高亮会降级**。建议直接选 §1.5 B.3 的 (a)（单元格 hover），
  理由：移动端无 hover，而这三个客户端里插件页的主要使用场景是移动端；
  真要保留就选 (b) 并接受列宽写死。**不要选 (c)**（依赖未核实的 `mouseover` 派发）。
- **表格的无障碍语义会丢**（`role="table"` 在 WebF 的支持度未核实）。
  但**在 WebF 路径下这些语义现在就已经全丢了**（`_UnknownHTMLElement` 无任何表格语义），
  所以 B' 不比现状更差；而**浏览器 / WebView 路径确实是从「有」变成「无」**
  —— 这是 B' 相对现状的唯一真实回退，**必须如实写进 downloader 的 CHANGELOG**。
  缓解：给容器与单元格补 `role="table" / "row" / "columnheader" / "cell"` 与 `aria-label`，
  浏览器路径能恢复大部分语义，WebF 路径不变差。

### 1.9 无论选哪个方案都建议做的一件小事：**把静默变成一行警告**

这一项独立于 A/B/C/D，**成本约 25 行 JS，收益是「第三方作者不再瞎找」**。

**问题**：`<table>` 在 WebF 下的失效是**完全静默**的 ——
未注册标签的日志只在 `enableWebFCommandLog` 打开时才 `debugPrint`
（✅ `element_registry.dart:83-85`），产品没开。
作者看到的是「一坨竖着排的文本」，没有报错、没有可疑元素。
这与 `input[type=range]` 那条教训完全同构（交接文档 §3.2 第 6 条：
「作者看到的是一行莫名空白，既没有报错也没有可疑元素，归因难度比多了个文本框高一个量级」）。

**做法**：在 `internal/jsplugin/assets/common.js` 的 **ready 相**加一个
`tableWarnShim`（只警告、**不改写**），挂进 `readyShims` 数组
（当前是 `common.js:651`：`[emptyImgSrcSweepShim, detailsShim, rangeSliderShim, safeAreaShim]`）：

```
若 collectByTag('table').length > 0：
  console.warn('[songloft] WebF 下原生 <table> 不被支持（会退化成纵向堆叠的 block）。' +
               '请改用 CSS Grid，见插件开发指南 §<n>。命中 N 处。')
  并给每个 <table> 加 data-sl-table-unsupported 标记（幂等 + 便于页面内省定位）
```

**为什么是「只警告不改写」而不是方案 C 的改写垫片**：
改写垫片要背 §1.6 那张表里的全部代价与 6 个未知量，换来一个残废表格；
而一行 warn 把「归因难度高一个量级」的问题**直接降到零**，
并把修复责任明确交给唯一有能力做对的人（插件作者），
这与 §2.4 那次「逐插件声明 `renderEngine`」决策反转的逻辑完全一致。

**必须遵守的既有铁律**（交接文档 §7）：
`common.js` 服务给**所有**客户端版本与普通浏览器 → 改动必须**纯增量 + 特性探测**，
垫片本体要关在 `isWebFEngine()` 分支里（`runShims` 已逐个 try/catch，一个失败只影响它自己）。

---

## 2. Step 6 的 `input[type=file]` 定形状

> **一句话结论**：**当前紧迫性为零** —— 三个命中插件里 radio 声明的是 `webview`，
> ytdlp / lxmusic 拿不到源码且**根本不是本仓库跟踪的子模块、也没有发布过 webf 版本**。
> 但 radio 的用法已经**完全核实清楚**了，桥的形状可以现在定下来：
> **返回 `{name, size, text}`，主载荷是 UTF-8 解码后的文本字符串**，
> 不返回 path、不返回 base64、不需要 `FileList` 垫片。

### 2.1 紧迫性：三个插件的 `renderEngine` 取值

| 插件 | `renderEngine` | 源码 | 当前暴露？ |
|---|---|---|---|
| **radio** | **字段不存在** = `webview`（§2.4 契约：缺失或空串 = webview） | ✅ 已拉到（子模块） | ❌ **否** |
| **ytdlp** | ❓ **无法核实**（仓库拿不到，§0） | ❌ 拿不到 | ❌ 否（不是跟踪子模块，本分支从未给它标过 `webf`） |
| **lxmusic** | ❓ **无法核实**（仓库拿不到） | ❌ 拿不到 | ❌ 否（同上，且交接文档 §6 记它**未构建发布**、在 WebF 下另有布局崩溃） |

→ **`input[type=file]` 这一项当前对任何已发布插件都不构成缺陷。**
它属于「为将来第三方作者准备的能力」，优先级应排在 §1 的 downloader 止血**之后**。
如实说明：**紧迫性零，但设计价值不为零** —— radio 将来若想切 webf，这是它的第一道拦路石。

### 2.2 radio 的实际用法（✅ 已核实，逐行）

`jsplugins-src/songloft-plugin-radio/static/index.html:19`：
```html
<input type="file" id="fileInput" accept=".m3u,.m3u8,.json,.txt" hidden>
```
—— **注意 `hidden` 属性**：这个 input **本来就不可见**，用户点的是旁边那个
`<button id="btn-file">选择文件</button>`（`:20-23`）。

`static/js/app.js:100-120`：
```js
btnFile.addEventListener('click', function () { fileInput.click(); });   // :100

fileInput.addEventListener('change', function () {                        // :102
  var file = fileInput.files[0];                                          // :103
  if (!file) return;
  fileName.textContent = file.name;                                       // :105  ← 只显示文件名
  if (file.size > MAX_FILE_SIZE) {                                        // :107  ← 20 MB 上限
    showSnack('文件超过 20MB 限制', 'error'); return;
  }
  var reader = new FileReader();                                          // :112
  reader.onload = function (e) { parseContent(e.target.result); };         // :113-115
  reader.onerror = function () { showSnack('文件读取失败', 'error'); };
  reader.readAsText(file);                                                // :119  ← 只要文本
});
```

**从这段代码可以确定的全部事实：**

| 问题 | 答案 |
|---|---|
| 拿文件做什么？ | **只读文本内容**，交给 `parseContent()` 解析 M3U / JSON（`src/m3u-parser.ts` 是后端侧解析器；页面侧 `parseContent` 在 app.js 内） |
| 需要二进制吗？ | ❌ **不需要**。`readAsText` 是唯一读取方式 |
| 需要 path 吗？ | ❌ **不需要**。全程没有 `file.path`（浏览器里也没有这个属性） |
| 需要上传到服务端吗？ | ❌ 不需要。解析完全在页面内 |
| 用到 `FileList` 的什么？ | 只有 `files[0]`、`.name`、`.size` 三样 |
| `multiple`？ | ❌ 没有（只取 `files[0]`） |
| `accept`？ | ✅ 有：`.m3u,.m3u8,.json,.txt` —— **扩展名形式**，不是 MIME |
| 触发方式 | **`fileInput.click()` 由另一个按钮的 handler 调用** —— 这一条对垫片设计至关重要，见 §2.4 |
| 消费事件 | `change`（不是 `input`） |
| 大小上限 | 插件自己判 20 MB（`MAX_FILE_SIZE`） |

⚠️ **ytdlp / lxmusic 的用法完全未知**（拿不到源码）。**不要假设它们也只要文本** ——
ytdlp 是下载类插件，完全可能要上传 cookies 文件或二进制；lxmusic 可能要导入配置。
桥的设计**必须留出扩展位**（见 §2.3 的 `as` 参数），但**默认实现只做 radio 需要的那一件事**。

### 2.3 桥 API 的签名与返回值设计

**已核实的前提（引用预研 §3，不重复论证）：**

- `input[type=file]` 在 WebF 下落到 `default` 分支 → 一个 Flutter `TextField`
  （`webf/lib/src/html/form/input.dart:250-266` 的 `build()` switch 无 file 分支；
  `createInput` `:268-278` 只额外处理 `hidden`）。
  ⚠️ **但 radio 的这个 input 带 `hidden` 属性** —— HTML 的 `hidden` 属性会不会被 WebF
  映射成 `type='hidden'`？**不会**：`createInput` 判的是 `widgetElement.type`
  （`base_input.dart:214`：`getAttribute('type') ?? 'text'`），
  而 `type` 属性是 `"file"`。→ ⚠️ **需实测**：radio 那个 input 在 WebF 下会不会
  **变成一个可见的文本框**（HTML `hidden` 属性能否让它 `display:none`）。
  如果会，那是一个额外的视觉 bug：页面上会多出一个空文本框。
  **兜底很便宜**：垫片给 `input[type=file]` 强制 inline `display:none`
  （沿用 `rangeSliderShim` 的 `.sl-range-hidden` + inline `display:none` 双保险做法）。
- `file_picker: ^10.3.10` **已是现有依赖**（`songloft-player/pubspec.yaml:53` +
  `GeneratedPluginRegistrant.java:29` + `.flutter-plugins-dependencies`）→ **契约哈希代价为零**。
  雷区：**不要 bump 它的版本**（`## plugin-versions` 段把版本号纳入哈希）、
  **不要引新原生插件**、**不要改 Kotlin**（走 WebF 自己的 `javascriptChannel`，不经 Kotlin）。
- `url_launcher: ^6.3.1` 同样已在（`pubspec.yaml:59`）。
- Dart 层**没有** `File` / `FileList` / `FileReader` 的任何实现
  （✅ 本轮复核：`grep -rniE "\bFileList\b|\bFileReader\b|'files'"` 在 `webf/lib/` 里
  **零命中**，唯一相关的是 `painting/image_provider_factory.dart:50` 那条讲
  `URL.createObjectURL` 的注释）→ 它们要么在 C++ 侧、要么不存在，
  ⚠️ **无法从源码核实，必须实测**。

#### 推荐形状

**桥调用（JS → Dart，走已有的 `songloft.host` methodChannel，
即 `plugin_render_surface_webf.dart:154-162` 的 `_onMethodCall`）：**

```
method: 'pickFile'
args: {
  accept:   '.m3u,.m3u8,.json,.txt',   // 原样透传 input 的 accept
  multiple: false,
  as:       'text'                      // 'text' | 'bytes' | 'none'，默认 'text'
}
```

**返回值（Dart → JS）：**

```jsonc
// as: 'text'（radio 需要的，也是默认）
{ "ok": true, "files": [ { "name": "radio.m3u", "size": 12345, "text": "#EXTM3U\n…" } ] }

// as: 'bytes'（为 ytdlp/lxmusic 这类未知用途预留）
{ "ok": true, "files": [ { "name": "cookies.txt", "size": 987, "bytesBase64": "…" } ] }

// as: 'none'（只要元信息，例如先给用户确认再决定读不读）
{ "ok": true, "files": [ { "name": "big.zip", "size": 104857600 } ] }

// 用户取消
{ "ok": false, "canceled": true }

// 失败
{ "ok": false, "error": "read_failed", "detail": "…" }
```

**四个设计决定与各自代价：**

| 决定 | 理由 | 代价 |
|---|---|---|
| **主载荷是 `text`（UTF-8 解码后的字符串），不是 base64** | radio 的**唯一**用途是 `readAsText`。返回 text 让插件侧从 `FileReader` 的 3 行异步代码降到 1 行 `resp.files[0].text`，**且完全不需要垫 `FileReader` / `Blob`**（那两个是否存在无法核实，见上）。Dart 侧 `file_picker` 拿到 `bytes` 后 `utf8.decode(bytes, allowMalformed: true)` 即可 | 非 UTF-8 文件（GBK 的 m3u 很常见！）会乱码。**缓解**：Dart 侧照抄后端 `pkg/tag` 的编码修正思路 —— 先按 BOM 判 UTF-16 LE/BE，否则 UTF-8 严格解码，失败则按 GBK 再试。这与后端 `ReadSidecarLyric` 的处理**完全同构**（AGENTS.md「旁挂歌词」一节），可以直接抄那套判定顺序 |
| **`bytesBase64` 只在 `as:'bytes'` 时返回，不无条件带上** | 20 MB 文件的 base64 是 ~27 MB 字符串，要跨 methodChannel 传两次（Dart→C++→JS）。radio 用不到，**默认就不该付这个钱** | 需要二进制的插件多传一个参数 |
| **不返回 `path`** | ① 桌面端 `file_picker` 能给 path，Android SAF 给的是 content URI，**跨平台语义不一致**；② path 对页面侧 JS **毫无用处**（它不能读文件系统）；③ 返回宿主文件系统路径给插件 JS 是**不必要的信息泄露** | 若将来有「把用户选的文件交给后端处理」的需求，那应该走**后端上传端点**而不是把 path 塞给 JS |
| **不实现 `input.files` / `File` / `FileList` 垫片** | 要让 `input.files[0]` 可用，得在 JS 里造一个假的 FileList + File 对象，还得让 `FileReader.readAsText(fakeFile)` 能工作 —— 而 `FileReader` **是否存在无法核实**。假 File 配真 FileReader 必然出问题 | **插件必须改调用点**（见 §2.4）。这与 §3.4 `URL.createObjectURL` 的结论同构：「同步 API 无法用异步实现 → 改调用点，不要垫」 |

**⚠️ 一条容易漏的实测项**：`file_picker` 在 `withData: true` 时返回 `PlatformFile.bytes`，
但**在某些平台/大文件下可能只给 path 不给 bytes**。Dart 侧必须两条都处理
（有 `bytes` 直接用；只有 `path` 就 `File(path).readAsBytes()`）。
另外 Android 需要存储权限 —— 但 Bundle 模式启动流程里**已经申请过存储权限**
（AGENTS.md「Bundle 本地模式」：「申请存储权限 → 启动嵌入后端 → …」），
非 bundle 版走 SAF 不需要权限。**⚠️ 需实测**非 bundle 版 Android 上 `file_picker` 是否弹权限请求。

### 2.4 垫片形状：**不能只拦 click**

**这是本节最容易做错的一处。** radio 的触发链是：

```
用户点 <button id="btn-file">  →  插件的 handler 调 fileInput.click()  →  change 事件
                                   ↑ app.js:100
```

→ **垫片如果只在 `<input type=file>` 上装 click 监听，能不能被 `fileInput.click()` 触发到？**

⚠️ **需实测，且交接文档 §3.1 已经埋了一条相关警告**：
「**`HTMLElement.click()` 是异步的**」（已写进插件开发文档）。
异步意味着调用 `click()` 后 handler 在**稍后**跑，这对本用例其实无害（不需要同步返回值），
但「程序化 `click()` 是否会派发 DOM click 事件到监听器」本身**必须实测**
—— WebF 的 `click()` 实现在 C++ 侧，无法核实。

**因此垫片必须有两条入口，缺一不可：**

1. **拦 `click` 事件**（覆盖「用户直接点可见的 file input」这种常见写法）；
2. **覆写元素实例上的 `click` 方法**（覆盖 radio 这种「隐藏 input + 外部按钮代点」的写法）——
   在垫片扫到的每个 `input[type=file]` 上 `Object.defineProperty(el, 'click', {value: myOpen})`。
   这与 `rangeSliderShim` 在实例上定义 `.value` / `.disabled` 访问器是**同一套手法**
   （`common.js:499-540` 附近），并且要沿用它的 **verified-or-abort** 纪律：
   **装不上就整体放弃**（还原原状、`console.warn`），**不写第二条永远测不到的降级路**。

**垫片还要做的：**

- 强制隐藏原 input（`display:none` inline + class），理由见 §2.3 的第一条实测项；
- **保留原 `<input>` 节点不移除**（Step 1 / Step 3 的同一条教训：插件按 id/标签查节点，
  radio 就是 `getElementById('fileInput')`）；
- 拿到结果后**派发 `change` 事件**（radio 监听的是 `change`，不是 `input`）。
  ✅ Step 3 已经实测通了 `dispatchEvent`（交接文档 §4：「`dispatchEvent` 已实测通，值走 `event.data`」）
  —— **复用那个结论，不要重新踩**；
- 把结果挂到一个**新的、明确的**位置供插件读取，而不是伪造 `input.files`。
  建议 `SongloftPlugin.lastPickedFiles`（数组）+ 在 `change` 事件上带 `event.data`。

**插件侧要改的（radio，若将来切 webf）：**

```js
fileInput.addEventListener('change', function (e) {
  // WebF 路径：垫片把结果放在 e.data / SongloftPlugin.lastPickedFiles
  var picked = (e && e.data && e.data.files && e.data.files[0]) || null;
  if (picked) {                       // ← 新增分支
    fileName.textContent = picked.name;
    if (picked.size > MAX_FILE_SIZE) { showSnack('文件超过 20MB 限制', 'error'); return; }
    parseContent(picked.text);
    return;
  }
  // 浏览器 / WebView 路径：原来的 FileReader 逻辑一字不改
  var file = fileInput.files[0];
  …
});
```

→ **插件改动量：约 10 行，纯增量，原路径零改动。** 这是本方案唯一的插件成本。

### 2.5 Step 6 另两项的现状（顺带对齐，避免实施者按旧文档行动）

本文档不重新设计这两项（预研 §3.3 / §3.4 已经给了完整方案），只标出**状态与优先级**：

- **`window.open`**（仅 miot 一处，`js/auth.js:95`，小米账号二次验证）：
  miot **是** `renderEngine: "webf"`（✅ `plugin.json:13`）→ **这一项是真实暴露的，
  紧迫性高于 `input[type=file]`**。
  预研 §3.3 给了「十分钟决定性实验」（grep `flutter.log` 里那句
  `Attempting to navigate WebF to an external WebF page`）与情形 A 的 8 行代码骨架，
  以及三个必须注意的副作用（也拦 `<a href>`、`#` 锚点要放行、controller 重建要重设 delegate）。
  **直接照那节做。**
- **`URL.createObjectURL`**（仅 miot 两处封面）：miot 是 webf → **也是真实暴露的**。
  预研 §3.4 已定「改 `data:` URL，不要落盘」，并核实 `blob:` 从加载器层面就不可能
  （`bundle.dart:192-194` 直接 throw）。需实测项：`btoa` / `FileReader` /
  `Blob.prototype.arrayBuffer` 是否存在、CSS `background-image: url(data:…)` 能否用。
- **`env(safe-area-inset-*)`**（Step 5）：✅ **看起来已经做完了** ——
  `internal/jsplugin/assets/common.js:651` 的 `readyShims` 数组里已有 `safeAreaShim`
  （另有 `emptyImgSrcSweepShim` / `detailsShim` / `rangeSliderShim`）。
  实施者**先确认这一项的状态**，不要重复做。

**Step 6 内部的建议优先级**（按「真实暴露 × 用户可见度」排）：
`window.open`（miot 登录**走不通**，功能性阻塞） >
`URL.createObjectURL`（miot 封面不显示，视觉缺失） >
`input[type=file]`（**当前零暴露**）。

---

## 3. 给实施者的最短路径

### Step 0（**先做这个，10 分钟，零风险**）：给 downloader 止血

downloader 已经**带着 `renderEngine: "webf"` 发布了**（`plugin.json:4` 版本 `2026.8.2`，`:13`），
而它那张表在 WebF 下是坏的（§1.1 末尾）。**这是当前唯一已发版、用户可见的 WebF 回归。**

- 改 `jsplugins-src/songloft-plugin-downloader/plugin.json`：删掉 `"renderEngine": "webf"` 那一行
  （或改成 `"webview"`），bump 版本号，发一版。
- 等 §1 的 Grid 改造做完并**在容器里验证过**之后，再把 `renderEngine` 加回来。
- ⚠️ 提交规范：子仓库引用父仓库 issue **必须写完整路径** `songloft-org/songloft#341`；
  **禁止** `Co-Authored-By`；改完子模块后回主仓库 `git add jsplugins-src/songloft-plugin-downloader`
  bump 指针再提交。

> 若产品上判断「downloader 的用户量小到可以带着 bug 等一版」，这一步可以跳。
> 但**必须是显式决定**，不要因为没看到这段而默认跳过。

### Step 1（**前置闸门 —— 已由主 agent 从源码判掉，不必再测**）

> **结论已定：grid 子项上的 `position: sticky` 不可用 → B' 直接按「双容器 + 同步列宽」的
> 兜底形态写，不要写单容器 + `fr` 的那一版。** 下面是判据（webf 0.24.27 源码，可复核）。
>
> `rendering/grid.dart:347-351` 的 `_isPositionedGridChild()` 把 **sticky 与 absolute/fixed
> 归成同一类**：
>
> ```dart
> bool _isPositionedGridChild(RenderBox child) {
>   final RenderStyle? style = _unwrapGridChildStyle(child);
>   if (style == null) return false;
>   return style.isSelfPositioned() || style.isSelfStickyPosition();  // ← sticky 也算「定位」
> }
> ```
>
> 而这个判据被用在 **13 处**排除逻辑上（`:385 :471 :872 :913 :2250 :3052 :3077 :3363 :3643
> :3925 :4034 :4477`），其中 `:2250` 的 `.where((child) => !_isPositionedGridChild(child))`
> 就是**构建 grid item 列表**本身，`:3052 / :3363` 是**固有宽度计算**，`:3925` 是 stretch，
> `:4034` 是 gap 调整。也就是说 sticky 子项**既不占格子、也不参与列轨道定宽**。
> 于是 sticky 表头会：① 不给列轨道贡献宽度 → **表头列宽与数据行列宽各算各的，对不齐**（这正是
> B' 唯一的立身之本）；② 不占空间 → **压在第一行数据上**。
>
> `:1948-1972` 的注释写着「their placeholders can reserve correct space」，但 **`placeholder`
> 这个词在整个 `grid.dart` 里只出现在 3 条注释里、没有任何实现**（`grep -i placeholder`
> 只命中 `:1949 :1970 :4583` 三行注释）。别被这句注释误导成「space 会被保留」。
>
> **对照组证明这是 grid 路径独有的缺陷、不是 WebF 全局不支持 sticky**：`rendering/flow.dart`
> 的在流排除判据（`:425 :1212 :1342`）**只判 `isSelfPositioned()`、不含 sticky**，
> 所以块级/流式布局下 sticky 正确留在流内并占据空间（符合 CSS 规范）。
> `applyStickyChildOffset` 在 flow / flex / grid / widget 四条路径上都有调用
> （`flow.dart:881`、`flex.dart:4802`、`grid.dart:4592`、`widget.dart:711,717`），
> 所以 grid 里 sticky 的**偏移**照样会施加 —— 它会「贴住」，只是以脱流方式贴住，
> 这恰恰是最难归因的失败形态：**看起来 sticky 生效了，实际列宽全错、还盖住了一行数据**。
>
> → 这是**第 7 条上游缺陷**，措辞见交接文档 §3.2（task #12 报 bug 时一并提）：
> *grid 把 `position: sticky` 当脱流处理，与 CSS 规范及自身 flow 实现均不一致；
> 注释声称有 placeholder 占位但无实现。*

若仍想在探针里留一组回归用例（非必需，**当前不阻塞实施**），配方如下
（**主 agent 未执行** —— 与并行 agent 抢 `.dart_tool` / 容器）：

```bash
# 1) 在 songloft-player/scripts/webf-verify/probe.html 里加一组检查（注意纵向预算，见交接文档 §5）：
#    一个 max-height + overflow-y:auto 的容器，内含 display:grid（6 列，含 fr 与固定 px 混排），
#    首 6 个子项 position:sticky; top:0; background:<醒目色>，后面塞 40 行子项。
#    判据写在该组自己的注释里：滚动后首行是否仍贴顶、是否与下方列对齐。
# 2) 起容器跑探针
cd songloft-player
./scripts/webf-verify/run.sh --build      # 首次；build 与 run 必须分开跑并检查退出码（见下）
./scripts/webf-verify/run.sh
# 3) 看产出
#    scripts/webf-verify/out/probe.png      截图
#    scripts/webf-verify/out/flutter.log    页面 JS 错误与 console（onJSError/onJSLog 已转发）
#    scripts/webf-verify/out/elements.sha1  确认测的是产品那份 render/elements/
```

**验证环境自身的坑（交接文档 §5，照抄，别重新踩）：**
- **`docker build … && run.sh; echo done`** —— build 失败时 `&&` 短路但 `echo` 照样打印 done，
  然后你读到的是**上一轮的旧截图**。**分开跑，检查 build 退出码。**
- `DIAGNOSE` 落到 Dart 的 `bool.fromEnvironment`，**只认字面 `"true"`**（`run.sh` 已归一化）。
- 探针页是**两列布局，纵向预算有限**，加组之前先确认新行不会把已有行挤出截图。
- 本机 glibc **2.35** < WebF 要求的 2.38 → **宿主上跑不了 WebF，容器是唯一途径。**

### Step 2：改 downloader（方案 B'）

按 §1.5 B.3 的骨架，三个文件：

1. `static/index.html:106-119` —— 删 `<table>/<thead>/<tbody>/<tr>`，写 `<div class="tbl" id="tbl">`
   + 6 个 `<div class="tbl-th">`。**保留** `.table-wrap` 与它的内联 `max-height/overflow-y`。
   **保留** `<div class="empty-state" id="empty">`（它在 `.table-wrap` 内、表格之后，不受影响）。
2. `static/css/style.css:306-337` —— 换成 §1.5 B.3 的 grid CSS。
   **⚠️ 绝对不要写 `display:table` / `table-row` / `table-cell`**（§1.3 核实 ②：会退化成 `inline`）。
   sticky 若 Step 1 判定不可用，改双容器 + 写死列宽。
3. `static/js/app.js:92-127` 的 `render()` —— `$('#tbody')` 换成 `$('#tbl')`，
   一次 `innerHTML = headerHtml + rowsHtml`（形状与现在最接近），行模板的
   `<tr>/<td>` 换成 6 个 `<div class="tbl-td">`，`data-id` 删掉。
   **`:120-127` 的 `.row-cb` 事件绑定一行不动**（§1.2 已核实 `<tr>` 从未被读）。

**同时要做的**：
- CHANGELOG 记两件事：① 表格改 Grid 实现；② **`tr:hover` 整行高亮降级为单元格高亮**
  与 **表格无障碍语义丢失**（§1.8 末尾，这是对现状的真实回退，必须写）。
- ⚠️ **不要跑 `pnpm build` 之外的依赖升级**；尤其**不要**在 player 仓库里跑 `flutter pub upgrade`
  （会把 `file_picker` 从 10.3.10 解析到别的版本 → dart 契约哈希变 → 热更被阻断，预研 §3.1 雷区 1）。

### Step 3：验证 downloader 真实插件页

```bash
# 起本机后端（lite 够用，省掉前端构建）
go build -tags "dev lite" -o /tmp/songloft-webf .
/tmp/songloft-webf -port 58191 -db /tmp/webf/test.db -music <musicdir>
# ⚠️ music 目录不要放 /tmp —— music_path 的默认 exclude_dirs 含 tmp，
#    而 ShouldExcludeDir 按路径任一层级的目录名匹配 → 整个 /tmp 被排除，
#    表现是扫描「成功完成」但 discovered_files=0，不报错不打 warn（AGENTS.md 已记）

# 装上改好的 downloader（手动上传 zip 或本地源），然后渲染它的页面
# ⚠️ URL 必须带尾斜杠（WebF 不采纳 <base href>，交接文档 §2.2 的 76b993a）
HOST_NETWORK=1 \
PROBE_URL='http://127.0.0.1:58191/api/v1/jsplugin/downloader/?embed=&theme=dark&access_token=<t>' \
  ./scripts/webf-verify/run.sh
```

**断言铁律**（与仓库既有无头浏览器验证一致）：
**截图只证明「渲染对了」**。要证明表格真的可用，还得落在后端可观测状态上 ——
例如勾几行点「下载选中」，然后 `curl` 查 `/api/v1/songs?...` 看那些歌的 `type` 有没有变成 `local`；
数 ffmpeg 子进程用 `pgrep -x ffmpeg`，**不要** `ps -ef | grep | wc -l`
（当前 shell 自己的命令行含关键字，会稳定多算 1~2 个）。

### Step 4：宿主侧那一件小事（§1.9）

在 `internal/jsplugin/assets/common.js` 的 `readyShims`（当前 `:651`）里加 `tableWarnShim`
（**只警告不改写**，约 25 行）。
- 必须关在 `isWebFEngine()` 分支里、纯增量 + 特性探测（交接文档 §7 铁律）。
- 顺手在 `docs/js-plugin-development-guide.md` 加一节「WebF 下不要用 `<table>`，用 CSS Grid」
  并附 §1.5 B.3 的可复制骨架 —— ⚠️ **文档双语同步铁律**：`docs/en/js-plugin-development-guide.md`
  必须同步改。
- 改 Go 后根目录 `gofmt -w .`；改 Dart 后 `cd songloft-player && dart format lib/ test/`。
- ⚠️ 这两个文件**当前由并行 agent 持有**（`internal/jsplugin/assets/**`、
  `docs/js-plugin-development-guide.md`、英文版、`CHANGELOG.md`）→ **动之前先确认它已经收工**。

### Step 5：把 downloader 的 `renderEngine` 加回 `webf`，发版

### Step 6（独立，优先级见 §2.5）：Step 6 的三项

顺序：`window.open`（miot 登录阻塞） → `URL.createObjectURL`（miot 封面） →
`input[type=file]`（当前零暴露，按 §2.3 / §2.4 的形状做）。

---

### 必须先做掉的实测项清单

按「不测就会返工」的程度排序。**每一项都很短**，都在 §3 Step 1 的容器里跑。

#### 阻塞方案 B'（必须先测）

1. ~~**`position: sticky` 在 grid 子项上是否生效**（§1.5 B.2）。~~
   → **已由主 agent 从源码判定：不可用，且失败形态是「看着贴住了、实际列宽全错还盖住一行」。**
   B' 直接按「双容器 + 同步列宽」写，**不必再测**。判据与 13 处排除逻辑的行号见 §3 Step 1。
   顺带产出**第 7 条上游缺陷**（grid 把 sticky 当脱流，与自身 flow 实现不一致）。
2. **`minmax(0, 1fr)` + `overflow:hidden` + `text-overflow:ellipsis` 的长文本收缩行为**
   （**现在这是阻塞 B' 的第一项**）。
   → 决定长标题是省略号还是撑爆列。次要，但会影响 CSS 写法。

#### 若改选方案 A（自写元素）才需要

3. `innerHTML = '…'` 批量重写是否触发 `WidgetElement` 的 `requestUpdateState`
   （§1.3 核实 ①：Dart 侧路径已核实，C++ 侧对 `innerHTML` 的分解无法核实）。
   **对 A 是生死项**（downloader 就是这个形状）。
4. `WebFHTMLElement(tagName:'DIV')` 包裹的单元格内容，其 CSS padding / border / background
   是否正常生效；`<input type=checkbox>` 这类替换元素放在里面是否正常。
5. 元素在无界高度约束下（`allowsInfiniteHeight`）的表现。

#### 若改选方案 C（`<webf-table>` 垫片）才需要（**6 项，这就是不推荐 C 的量化理由**）

6. 静态 HTML 写 `<table><tr>` 时 gumbo 有没有插入**隐式 `<tbody>`**（C++ 侧的
   `html_tag_names.json5` 不随 pub 包发布，`find . -name "*.json5"` 零结果）。
7. `<webf-table>` 只有 `max-height`（无 `height`）时，sticky 分支的 `Expanded`
   会不会抛 `RenderFlex … unbounded` / 产出 `Infinity`。
8. sticky 模式下不写 `column-width` 时，表头与表体的实际错位程度。
9. 单元格 CSS（padding / border-bottom / background）是否生效
   —— `toTableCell()` 把内容包进 `tagName:'SPAN'`（`table_cell.dart:191`），**SPAN 是 inline**。
10. `tr:hover` 是否能触发重建（`WebFTableRowState.build()` 返回 `SizedBox.shrink()`，
    行只以 `renderStyle.decoration` 参与）。
11. `getComputedStyle` 能否读到 `position`（垫片要靠它反推 sticky 意图）。
    ⚠️ 已知 `getComputedStyle` **不暴露自定义属性**（`songloft_progress_ring.dart:146` 的实测结论），
    标准属性能不能读没测过。

#### 若改选方案 D.2a（伪表格）才需要

12. `<tr>` 设 `display:grid` 后 `<td>` 是否正确成为 grid item
    （`css/display.dart:120-140` 的 `effectiveDisplay` 会把 grid item blockify，看起来是对的，未实测）。
13. 表头那一行（`<tr>` 内的 `<th>`）的 sticky 表现。

#### Step 6 相关

14. `HTMLElement.click()`（程序化调用）是否会派发 DOM `click` 事件到监听器（§2.4）。
    → 决定垫片是「拦 click 事件」还是「覆写实例 `click` 方法」，**建议两条都做**。
15. `input[type=file]` 带 HTML `hidden` 属性时，在 WebF 下会不会**变成一个可见的空文本框**（§2.3）。
16. `btoa` / `FileReader` / `Blob.prototype.arrayBuffer` 是否存在
    （Dart 层零命中 → 在 C++ 侧或不存在）。影响 `URL.createObjectURL` 那一项。
17. CSS `background-image: url(data:…)` 能否用（miot 全屏播放器的 `bgImage`，预研 §3.4）。
18. `flutter.log` 里有没有 `Attempting to navigate WebF to an external WebF page`
    → 决定 `window.open` 走 8 行 delegate（情形 A）还是 JS 垫片（情形 B）。预研 §3.3 有完整骨架。
19. 非 bundle 版 Android 上 `file_picker` 是否弹存储权限请求（§2.3 末）。

#### 先确认状态，别重复劳动

20. ~~**Step 5（`env(safe-area-inset-*)`）看起来已经做完了**~~ —— **误读，已由主 agent 澄清。**
    看到的 `common.js` 里的 `safeAreaShim` 是**同期并行的 Step 5 agent 尚未提交的工作树改动**
    （当时 `git status` 里 `internal/jsplugin/assets/common.js` 是 ` M` 未提交态），
    不是既有实现。**教训**：多 agent 共享同一个工作树时，「文件里已经有 X」不能推出
    「X 已经做完」—— 判定某项工作是否已落地要看 `git log` / 已提交内容，不能只看工作区。
