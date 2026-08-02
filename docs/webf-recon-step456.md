# WebF 迁移 — Step 4 / 5 / 6 源码调研（songloft-org/songloft#341）

> **这是分支临时件**，与 `docs/webf-migration-handoff.md` 同期存在，#341 落地后**连同交接文档一起删除**。
> 刻意**不做**中英双语同步、也不进文档站导航。
>
> 调研范围：WebF 包源码 `/home/ejoydev/.pub-cache/hosted/pub.dev/webf-0.24.27`（只读）+ 本仓库源码。
> **全部结论只从本地源码得出，没有查网络。**
> 每条结论都标注了 **`已核实（文件:行号）`** 或 **`无法核实，需实测`** —— 请严格按这个区分行事。
>
> 最后更新：2026-08-02。

---

## 第 1 节 — Step 5：`env(safe-area-inset-*)` 的 CSSOM 可行性

> **一句话结论**：CSSOM 比交接文档假设的**完整得多**（`styleSheets` / `cssRules` / `insertRule` /
> `deleteRule` / `replaceSync` 全都有），但 **`cssText` 只读、`CSSStyleRule` 不暴露 `selectorText`
> 也不暴露 `.style`**，所以「原地改写规则」这条路 **已核实不可行**。
> 而且**就算 CSSOM 完整，改写也解决不了 miot 的 3 处实际写法**——它们全都套在 `calc()` / `max()` 里，
> 其中 `max()` 在 WebF 里**根本没实现**。
> **建议采纳「宿主只注入变量 + 插件作者自己写 `var(--sl-safe-*)`」方案**（下面方案 D），
> 理由见「结论与方案对比」。

### 1.1 先纠正一个前提：`env()` 确实不求值，交接文档的行号**准确但会误导**

- **已核实**：`style_declaration.dart:736` 确实存在，内容是
  `lowerCase = _replacePattern(string, lowerCase, 'env(', ')');`
  —— 但它**不是求值逻辑**，只是 `_toLowerCase()` 里「把 `env(...)` 内部的大小写原样保留」的
  字符串保护，与 `url(` / `var(` 并列（`style_declaration.dart:735-738`）。
  引用这个行号说「env() 在这里不求值」容易让下一个人以为「这里改一下就能求值了」，实际不是。
- **已核实（真正的证据）**：`env(` 在整个 WebF Dart 层**只出现这一处**。
  `grep -rn "env(" lib/ --include=*.dart` 只有 `style_declaration.dart:736`。
- **已核实**：`css/keywords.dart:239-243` 定义了 `SAFE_AREA_INSET` / `SAFE_AREA_INSET_TOP` /
  `_LEFT` / `_RIGHT` / `_BOTTOM`，`keywords.dart:382` 定义了 `const String ENV = 'env'`
  —— 这 6 个常量**在 lib/ 里没有任何其他引用点**（死常量）。
  这才是「env() 完全没实现」的硬证据：连解析入口都不存在。
- **已核实**：`bridge/code_gen/blink_css_ids.dart:1392` 里有 `'env'`，那是 **Blink CSS 模式**
  （C++ 侧）的 property/function id 表。但 `enableBlink` 默认 `false`
  （`launcher/controller.dart:801`、`launcher/view_controller.dart:67`），
  且产品侧 `plugin_render_surface_webf.dart` 的 `_createController()`（:219）**没有传 `enableBlink`**
  → 走 Dart CSS 引擎。**不要**指望 Blink 模式救 env()：见 §1.3 的 ⚠️。

### 1.2 CSSOM 到底有什么（逐条核实）

全部在 `lib/src/css/css_style_sheet_binding.dart`（该文件头注释自称
「A Dart-side CSSOM wrapper … exposed to JavaScript via WebF's binding object mechanism」）。

| API | 状态 | 证据 |
|---|---|---|
| `document.styleSheets` | ✅ **已实现** | `dom/document.dart:469-472` 注册 `StaticDefinedBindingProperty`，getter → `_styleSheetsForBinding`（`:502-527`） |
| 是不是 live 集合？ | ❌ **不是 live 对象，但每次访问重算** | `_styleSheetsForBinding` 每次调用都 `updateStyleIfNeeded()` 然后**新建一个 `List`** 遍历 `styleNodeManager.styleSheetCandidateNodes`（`document.dart:508-526`）。经 `bridge/native_value.dart:326-334` 的 `value is List` 分支编码为 `tagList` → **JS 侧拿到的是一个普通 JS Array**（不是 `StyleSheetList`）。所以 `.length` / `[i]` / `forEach` / `map` / 展开都能用，但**对象身份每次不同**，别缓存它、别 `===` 比较 |
| ⚠️ Blink 模式下返回空 | ✅ 已核实 | `document.dart:503-506`：`if (ownerView.enableBlink) return const <CSSStyleSheetBinding>[];`。**任何基于 CSSOM 的垫片都会在 `enableBlink: true` 下静默失效**（返回空数组、不报错） |
| `sheet.cssRules` | ✅ **可枚举**，但**不是数组** | `css_style_sheet_binding.dart:49-61` 返回 `CSSRuleListBinding`（一个 binding object）。它暴露 `length`（`:190-192`）、`item(i)`（`:197-201`），并在 `refresh()`（`:179-188`）里把 `0..length-1` 写进 `dynamicProperties` 使 `sheet.cssRules[0]` 可用（上游注释明说「so JS indexed access works」）。**没有 `Symbol.iterator`** → `for...of` / `[...]` / `Array.from` / `forEach` **都不能用**，只能 `for (let i = 0; i < list.length; i++)` |
| `CSSRule.cssText` | ⚠️ **只读** | `css_style_sheet_binding.dart:213-216` 只注册了 `getter`，**没有 `setter`**。JS 里 `rule.cssText = '...'` 是静默无效赋值（binding object 上写一个不存在的属性） |
| `CSSStyleRule.selectorText` | ❌ **不暴露** | `grep -rn selectorText lib/` 只命中 `css_rule.dart:215`（Dart 内部拼 cssText 用）与 `parser/selector.dart`。binding 层**完全没有** `selectorText` 属性 |
| `CSSStyleRule.style`（`CSSStyleDeclaration`） | ❌ **不暴露** | `CSSRuleBinding`（`:222-238`）只有 `cssText` + `type`。`style_declaration.dart:112-114` 的文档注释提到 `document.styleSheets[0].cssRules[0].style`，但那只是**注释里抄的 MDN 描述**，binding 层没实现 |
| `sheet.insertRule(text, index)` | ✅ **已实现** | `css_style_sheet_binding.dart:63-79`（binding）→ `style_sheet.dart:28-49`（真正插入 + `CSSParser` 解析）。插完会 `cssRules.refresh()` + `_scheduleStyleUpdate()`（`:121-126`，走 `appendPendingStyleSheet` + `updateStyleIfNeeded`），**样式立即生效**。⚠️ `index` 是**必填**（`StaticDefinedSyncBindingObjectMethod` 直接 `args[1]`，`:142-145`），标准里的可选默认 0 **不能省** |
| `sheet.deleteRule(index)` | ✅ **已实现** | `:81-85` → `style_sheet.dart:52-54` |
| `sheet.replaceSync(text)` | ✅ **已实现**（非标准位置但可用） | `:87-94` → `style_sheet.dart:57-71`，`cssRules.clear()` 后整体重新解析 |
| `sheet.disabled` | ✅ **可读可写** | `:129-132`，setter 会触发样式刷新。**这是一个被低估的能力**，见方案 C |
| `sheet.href` / `sheet.type` | ✅ 只读 | `:133-134` |
| `@media` 里的规则 | ❌ **完全不可达** | 见 §1.3 |

### 1.3 三个会让人踩坑的硬边界（都已核实）

**① `@media` / `@keyframes` / `@font-face` / `@import` 规则的 `cssText` 是空串，且没有 `.cssRules`**

- `CSSRule.cssText` 基类默认 `=> ''`（`css/rule.dart:13`）。
- 只有 **3 个**子类 override 了 `cssText`：`CSSStyleRule`（`css_rule.dart:213-215`，
  拼成 `'${selectorGroup.selectorText} {${declaration.cssText}}'`）、
  `CSSLayerStatementRule`、`CSSLayerBlockRule`。
- `CSSMediaDirective`（`css_rule.dart:383-394`）、`CSSKeyframesRule`（`:302`）、
  `CSSFontFaceRule`（`:370`）、`CSSImportRule`（`:233`）**都没 override** → `cssText === ''`。
- binding 层只给 **`CSSLayerBlockRule`** 加了嵌套 `cssRules`（`:289-301`）；
  media rule 走通用 `CSSRuleBinding`（`:101-117` 的 else 分支）→ 只有 `cssText`（空）+ `type`（4）。
- **后果**：`@media` 块里的 `env()` **既读不到、也改不了**。而且
  **对一条 media rule 做 `deleteRule` + `insertRule` 会不可逆地摧毁它**（cssText 是空串，
  重新插入等于插入空规则）。任何遍历都必须 `if (rule.type !== 1) continue;`。
- ✅ **好消息**：miot 现有的 3 处 `env()` **都不在 `@media` 里**（已核实，见 §1.4），
  所以这条边界本轮不致命，但它把「通用自动改写」这个想法直接判死。

**② `CSSStyleDeclaration.cssText` 是从解析结果**重建**的，不是原文**

`style_declaration.dart:206-215`：遍历 `_properties` map 拼
`'${_kebabize(property)}: $value ${!important};'`。含义：

- **简写会被展开成 longhand**（`_expandShorthand`，`style_declaration.dart:540`、`:923`）
  → `padding: 6px 16px calc(...)` 读回来是 4 条 `padding-*`
- **WebF 不认识的属性在解析期就被丢弃**，`cssText` 里不会出现（所以读到的 cssText
  ≠ 作者写的 CSS，`backdrop-filter` / `mask-image` 这类会消失）
- 值被规范化（`_toLowerCase`，但 `env(` / `url(` / `var(` 内部大小写保留）
- 有个小瑕疵：`!important` 那段无论有没有都拼了一个空格，产出形如 `color: red ;`
  （合法但难看）
- **推论**：`cssText` 适合**读出来做检测**，**不适合**当作「原样搬运」的载体。
  「delete + insert 改写」= 让整条规则过一轮 lossy 往返，可能丢掉本来还能渲染的声明。

**③ Blink 模式是一颗定时炸弹**

`document.styleSheets` 在 `enableBlink: true` 时返回空数组（`document.dart:503-506`）。
当前产品是 Dart CSS 模式（未传 `enableBlink`，`plugin_render_surface_webf.dart:219`），
所以 CSSOM 可用。但如果将来为了别的原因打开 Blink，**所有 CSSOM 垫片会静默变成 no-op**。
写垫片时**必须**加「读到 0 张表就 `console.warn` 并放弃」的自检，不要静默失败。

### 1.4 决定性发现：miot 的 3 处写法，改写救不了

**已核实**，`jsplugins-src/songloft-plugin-miot/static/css/style.css` 里全部 3 处
（也是唯一能在本工作树核实的 3 处，见 §1.6）：

| 位置 | 声明 | 花括号深度（已核实=1，即**不在 `@media` 内**） |
|---|---|---|
| `:133`（`#tab-settings`） | `padding-bottom: calc(24px + env(safe-area-inset-bottom));` | 1 |
| `:968`（`.player-bar`） | `padding: 6px 16px calc(4px + env(safe-area-inset-bottom, 0px));` | 1 |
| `:3740`（`.fp-controls`） | `padding-bottom: max(24px, env(safe-area-inset-bottom));` | 1 |

对每一处，把 `env(...)` 机械替换成 `var(--sl-safe-bottom)` 之后会发生什么：

- **`:133` → `calc(24px + var(--sl-safe-bottom))`：能work。** ✅ 已核实 `calc()` 支持 `var()`
  —— `css/values/calc.dart` 有专门的 `CalcVariableNode`（`calc.dart:104-120`），
  会 `value.computedValue('')` 拿到变量文本再 `CSSLength.parseLength`。
- **`:968` → `padding: 6px 16px calc(4px + var(--sl-safe-bottom))`：能work**，但注意
  `_isFunctionNotation`（`values/function.dart:28-44`）会因为首字符 `6` 不是 ASCII 字母而
  返回 false → 这个值走**简写展开**路径，`padding-bottom` longhand 拿到 `calc(...)` 字符串。
  行为上没问题，但**读回 cssText 时它已经是 4 条 longhand 了**（见边界②）。
- **`:3740` → `max(24px, var(--sl-safe-bottom))`：⚠️ 仍然无效。**
  **已核实 WebF 没有实现 CSS `max()` / `min()`**：`css/values/calc.dart` 的
  `CSSCalcValue.tryParse`（`:24-53`）**只**处理 `calc`（`:26`）和 `clamp`（`:39`），
  `grep` 全库也找不到 `max(` / `min(` 的 CSS 函数求值。
  `_isFunctionNotation('max(...)')` 返回 true → `_isValidValue` 放行 → 值被原样存下 →
  后续 `CSSLength.parseLength('max(...)')` 认不出 → 这条声明**现在就是死的，和 env() 无关**。
  要修它必须**改写整个值的形状**（例如换成 `clamp(24px, var(--sl-safe-bottom), 999px)`
  或干脆 `max(24px, ...)` → 由 Dart 侧把变量本身就算成 `max(24px, inset)` 的结果），
  而这已经**不是「env→var 的机械替换」**，是逐处的语义改写。

**这一条基本上单独决定了结论**：命中面里 1/3 的写法无法通过机械替换修复，
而能修的 2/3 又都是低风险的、插件作者一行就能改的东西。

### 1.5 变量怎么从 Dart 注入页面（已核实）

**注入通道有现成的，不需要新造。** 产品已有的 Dart→JS 推送是
`plugin_render_surface_webf.dart:196-205` 的 `_pushToPage()`：

```dart
_controller?.view.evaluateJavaScripts('window.postMessage($messageLiteral,"*")');
```

`evaluateJavaScripts` 返回 `Future<void>`、拿不到返回值（交接文档 §3.1 已记，本次复核成立），
但**注入变量不需要返回值**，所以这不是障碍。三种可行形态：

1. **直接改 documentElement 内联样式**（最简单，推荐）：
   `evaluateJavaScripts("document.documentElement.style.setProperty('--sl-safe-bottom','34px')")`。
   ✅ **已核实自定义属性会沿渲染树继承**：`css/variable.dart` 的 `getCSSVariable`
   在本节点没有该 identifier 时会 `getAttachedRenderParentRenderStyle()?.getCSSVariable(...)`
   一路向上找（`variable.dart:40-62`），所以挂在 `<html>` 上全页可见。
   `CSSVariableMixin` 还维护 `_propertyDependencies`（`variable.dart:20-37`），
   说明变量变更**会反向通知依赖它的属性**（转屏后重新注入应能触发重排，
   但**这条属于「需实测」**——我只核实了依赖表存在，没核实通知链路端到端）。
2. **复用既有 `window.postMessage` 协议**：Dart 推 `{type:'songloft-safe-area',insets:{...}}`，
   `common.js` 侧接收后自己 `setProperty`。**推荐用这个**——和主题推送（`:100`）、
   播放状态推送（`:207-216`）同构，一致性最好，且 JS 侧能顺手做「变量名归一 + 兜底」。
3. **反向 methodChannel**：`controller.javascriptChannel.invokeMethod(...)`
   （产品已在 `:114` 用它做 `requestBack`）。能拿返回值，但注入变量不需要，**没必要**。

**`MediaQuery.padding` 变化时怎么重新注入**：`PluginRenderSurfaceWebF` 是
`StatefulWidget`（有 `didUpdateWidget`，`:96-102`），所以标准 Flutter 做法就够：

- 在 `build()` 里读 `MediaQuery.paddingOf(context)`（或 `viewPaddingOf`），
  与上次推送的值比较，变了就 `_pushToPage(...)`。`MediaQuery` 依赖变化会自动触发 rebuild，
  转屏 / 进退全屏 / 键盘弹出都会走到。
- 已有的 `_lastPushedStateSig` 去重模式（`:209-212`）可以照抄一个 `_lastPushedInsets`，
  避免每帧重复 `evaluateJavaScripts`。
- **必须**在 `_pageReady` 之后才推（同 `:99` / `:212` 的判断），并在 `_pageReady` 变 true 时
  补推一次首屏值——否则首屏拿不到安全区。
- ⚠️ **`WebFControllerManager` 的 detach/dispose 会清掉 JS 状态**（交接文档 §3.1：
  超 `maxAliveInstances` 会 dispose，重挂 JS 状态归零）。所以「注入过一次就不管了」是错的，
  **重新挂载后必须重推**。挂在 `_pageReady` 转 true 的那条路径上即可。
- 用 `padding` 还是 `viewPadding`：`padding` 在键盘弹出时底部会变 0，`viewPadding` 不变。
  插件页的 `.player-bar` 是 `position:fixed;bottom:0`，键盘弹出时行为**需实测**再定，
  建议先用 `viewPadding`（更接近浏览器 `env(safe-area-inset-*)` 的语义）。

### 1.6 无法核实的部分（不要当成已知）

- **交接文档 §3.3 的命中面「miot 3、dav 3、subsonic 1、cloudflared 1、hostc 1、ytdlp 1」
  我只能核实 miot 的 3 处。** 本工作树里 `jsplugins-src/songloft-plugin-{cloudflared,dav,hostc,
  radio,registry,subsonic}/` **都是空目录**（子模块未 checkout），`ytdlp` 根本不在 `jsplugins-src/` 下。
  → 做 Step 5 之前**必须**先把这些子模块 clone 出来重新统计一遍，
  特别是**确认它们的 `env()` 是否在 `@media` 内**（在 `@media` 内 = CSSOM 方案对它们直接失效）。
- **`sheet.cssRules[0]` 的数字下标在 QuickJS 侧是否真的路由到 `getProperty('0')`**：
  上游注释（`css_style_sheet_binding.dart:180-182`）明说是这个意图，但 binding object 的
  数字属性拦截在 C++ 侧（不随 pub 包发布）。**建议先在验证容器里跑一条 3 行探针确认**，
  别在上面压整个方案。
- **变量变更后的重排通知链路是否端到端可靠**（转屏后页面是否真的重新算 padding）：
  依赖表存在（`variable.dart`），但通知 → dirty → relayout 的完整链路我没追完。**需实测**。
- **`insertRule` 插入的规则在层叠里的位置**是否严格按「同选择器同优先级、后来者胜」：
  `style_sheet.dart:44` 只是 `cssRules.insertAll(index, rules)`，
  层叠排序由 `rule_set.dart` / `element_rule_collector.dart` 的 `position` 决定。
  方案 C 依赖「文档顺序靠后的表胜出」，**这一点需实测**（一条探针即可：两张表同选择器不同颜色）。

### 1.7 结论与方案对比

**「垫片自动把 CSS 里的 `env()` 改写成 `var()`」 → 已核实不可行。**
两个独立的致命原因，任一成立即否决：

1. 改写所需的写入面**不存在**：`cssText` 只读、无 `selectorText`、无 `rule.style`
   （`css_style_sheet_binding.dart:213-238`）。唯一的写入面是 `insertRule` / `deleteRule` /
   `replaceSync`，而它们只能整条规则进出，配合 lossy 的 `cssText` 往返（边界②）
   会丢掉本来还能渲染的声明。
2. 就算能改写，**`max(24px, env(...))` 这一处（1/3 的命中面）改写后仍然无效**，
   因为 WebF 没实现 `max()`（`values/calc.dart:24-53`）。

下面是替代方案，按推荐度排序。

---

**方案 D（推荐）— 放弃自动改写：宿主只注入变量，插件作者直接写 `var(--sl-safe-*)`**

- **代价**：改插件源码。本工作树可核实的只有 miot 3 处；交接文档另称 dav 3、subsonic 1、
  cloudflared 1、hostc 1、ytdlp 1（**未核实，见 §1.6**）。
- **归属问题（用户提到的顾虑，已核实）**：`jsplugins-src/` 下的 9 个子模块目录
  （cloudflared / dav / downloader / hostc / lyrics / miot / radio / registry / subsonic）
  都在本仓库的 `.gitmodules` 里，即**都是 songloft-org 自己的仓库**（第一方）。
  `ytdlp` **不在** `jsplugins-src/` 下（交接文档 §6 也说 lxmusic / bili / ytdlp 不是跟踪的子模块）
  → 那 1 处**无法由我们直接改**。
- **关键杠杆：只有声明了 `renderEngine: "webf"` 的插件才会暴露在这个缺口下**（§2.4 契约）。
  目前只有 **miot / downloader / lyrics** 三个标记为 `webf`。
  **downloader / lyrics 没有 `env()`**（已核实：`grep -rn "env(" jsplugins-src/*/static jsplugins-src/*/src`
  只命中 miot 3 处）。
  → **实际必须改的就是 miot 的 3 处，而 miot 是第一方。**
  dav / subsonic / cloudflared / hostc / ytdlp **现在都是 `webview`**，它们的 `env()`
  一行都不用动；等哪天作者要切 `webf`，改 `env()` 本来就是他自己的验证工作
  （这正是 §2.4「逐插件声明把决定权交给唯一有能力验证的人」的设计意图）。
- **写法建议**（兼容浏览器 + WebF 同一份 CSS，无需分叉）：
  ```css
  /* 宿主在 WebF 下注入 --sl-safe-bottom；浏览器下不注入，回落到 env() */
  padding-bottom: calc(24px + var(--sl-safe-bottom, env(safe-area-inset-bottom, 0px)));
  ```
  ⚠️ **这一写法需实测**：WebF 的 `var()` fallback 里嵌 `env()` 时，
  `env()` 是否会被当成「无效值」而使整条声明失效（浏览器语义是 fallback 只在变量未定义时用）。
  **保守写法**：`common.css` 里给 `:root` 预置 `--sl-safe-*: 0px` 默认值，
  插件只写 `var(--sl-safe-bottom)`，两个引擎都拿到确定值，WebF 下由宿主覆写。
  这样连 fallback 语义都不依赖，**强烈建议走这条**。
- **优点**：零 CSSOM 依赖（不受 Blink 模式、`@media` 不可达、`cssText` lossy 影响）；
  `max()` 那处顺手改成 `calc()` 或 `clamp()` 就真的能生效；
  改动量小、可读、可被浏览器路径共享；不引入运行期开销。
- **缺点**：不是「宿主自动兜住」，未来第三方插件切 `webf` 时要自己改。
  但这本来就是 §2.4 契约下插件作者的责任。

---

**方案 C（备选）— 垫片注入一段新 `<style>`，用后置规则覆盖含 `env()` 的声明**

- **可行性**：读取面存在（CSSOM 能遍历 `type===1` 的规则、读 `cssText`），
  写入面用 `document.head.appendChild(styleEl)`（不依赖 CSSOM 写）或 `sheet.insertRule`。
- **但要付的代价，逐个说清**：
  1. **必须能读到「哪些选择器用了 env()」，仍然依赖 CSSOM 读取** → 继承全部读取侧限制：
     `@media` 内的规则**读不到**（`cssText === ''`，边界①）→ 那些 `env()` 覆盖不了；
     Blink 模式下 `styleSheets` 为空（边界③）→ 整个垫片静默 no-op。
  2. **`selectorText` 不暴露** → 只能从 `cssText` 里**字符串切第一个 `{`** 反推选择器。
     这在 `content: "{"`、属性选择器 `[data-x="{"]` 这类值里会切错。
     实际风险低（miot 没有这种写法），但要写成「切不出来就跳过 + warn」。
  3. **优先级处理**：同选择器 + 同 specificity 时靠**文档顺序**取胜，所以覆盖表必须是
     `<head>` 里**最后**一张（或 `document.body` 末尾）。⚠️ **「后置表胜出」需实测**（§1.6）。
     插件后续动态 `appendChild` 新 `<style>` 会重新盖过我们 → 需要
     `SongloftPlugin.applyShims()` 重跑并把覆盖表**移到末尾**，幂等逻辑更复杂。
  4. **`!important` 无解**：原声明带 `!important` 时（miot 的 `.fp-*` 一堆 `!important`，
     虽然目前不在 `padding-bottom` 上）后置同优先级规则**赢不了**。
     唯一办法是覆盖规则也加 `!important` —— 但那会把**作者本来靠 `!important`
     压别的规则的意图**一起压掉，且我们无法区分「原来有没有 important」
     （`cssText` 里能看到 `!important` 字样，可以镜像，但两条都 important 时又回到文档顺序）。
     结论：**要写成「原声明带 important → 我们也带 important 并放最后」**，
     属于能做但要格外小心的一档。
  5. **`max()` 那处依然救不了**（同 §1.4）：覆盖成 `max(24px, var(--sl-safe-bottom))`
     照样死。要救就得把值换成 `clamp()`，那已经是语义改写，垫片不该自作主张。
  6. 运行期开销：每次 ready + 每次 `applyShims()` 都要遍历所有 sheet × 所有 rule 读 `cssText`
     （`CSSStyleRule.cssText` 是**每次访问都重新拼字符串**，`css_rule.dart:213-215` 无缓存），
     miot 的 style.css 有 500+ 条规则 → 每次全表拼串。可接受但不是零成本。
- **总评**：**技术上可行，但复杂度与脆弱性远高于方案 D，且仍留下 `@media` 与 `max()` 两个洞。**
  只有在「插件源码不可改」这个前提成立时才值得——而 §1.6 已核实那个前提**不成立**
  （需要改的 3 处都在第一方 miot）。

---

**方案 B（不推荐）— `deleteRule` + `insertRule` 原地改写规则**

- 唯一优点：改写后位置不变，层叠语义最干净。
- 致命代价：`cssText` 是 lossy 重建（边界②）→ 往返会丢掉 WebF 解析期已丢弃之外的
  **格式与简写信息**，且**一旦对 `type !== 1` 的规则误操作就是不可逆销毁**
  （media rule 的 cssText 是空串）。为 3 处 padding 承担「摧毁整张样式表」的风险，
  性价比极差。**不要做。**

---

**方案 A（不推荐）— `replaceSync` 整表重写**

- 把整张表的 `cssText` 拼起来做字符串替换再 `replaceSync`。
- 致命：`@media` / `@keyframes` / `@font-face` 的 `cssText` 全是空串
  → **整表重写会把所有 at-rule 抹掉**。miot 有 `@media (min-width:600px)` 等多处。
  **绝对不要做。**

---

### 1.8 给 Step 5 实施者的最短路径

1. 先在验证容器里跑一条 **3 行探针**确认两件「需实测」的事（十分钟的事，能省掉整轮返工）：
   - `document.documentElement.style.setProperty('--sl-x','40px')` 之后，
     一个 `padding-bottom: calc(10px + var(--sl-x))` 的元素高度是否变化（变量注入 + calc+var 生效）
   - 转屏 / 改注入值之后是否**重新布局**（变量变更通知链路）
2. `common.css` 的 `:root` 加 `--sl-safe-top/right/bottom/left: 0px` 默认值
   （两个引擎共享，浏览器下也无害）。
3. `common.js` 加 `songloft-safe-area` 消息处理，`setProperty` 到 `documentElement`。
   放 **early** 还是 **ready** 都行（不是解析期不可撤销副作用），
   但接收侧要能处理「宿主在 `_pageReady` 之前就推」的情况——建议 early 装监听。
4. `plugin_render_surface_webf.dart` 在 `build()` 读 `MediaQuery.viewPaddingOf(context)`，
   与 `_lastPushedInsets` 比较后 `_pushToPage`，并在 `_pageReady` 转 true 时补推一次。
5. 改 miot 的 3 处 CSS：前两处 `env(...)` → `var(--sl-safe-bottom)`；
   第三处 `max(24px, env(...))` → 换成 `calc(24px + var(--sl-safe-bottom))`
   或 `clamp(24px, var(--sl-safe-bottom), 96px)`（`clamp` 已核实被支持，`calc.dart:39-51`）。
   **注意 miot 的 `static/` 目录本轮归 Step 3 的 agent 所有，动之前先确认它已交付。**
6. **不要**写 CSSOM 遍历垫片。若未来真的需要（第三方插件切 `webf` 且作者不配合），
   再回来看方案 C 的 6 条代价。

---

## 第 2 节 — Step 4：`<webf-table>` 的实际能力

> **一句话结论**：四个标签**确实存在**（交接文档这点没错），但**「机械改写 table→webf-table、
> tr→webf-table-row、td→webf-table-cell」会得到一张空表**，因为
> `<thead>` / `<tbody>` 这层包裹必须**被拆掉**（行必须是 `<webf-table>` 的**直接子节点**），
> 而 downloader 恰好靠 `#tbody` + `innerHTML` 渲染行 —— **拆掉它插件的 JS 就断了**。
> 另外 **`colspan` / `rowspan` 零支持**、**列宽只认表头单元格的 `column-width` 属性**（CSS `width` 无效）、
> **表格自带两层 `SingleChildScrollView`**、**sticky 模式用了 `Expanded`（需要外部给定高度）**。
> **建议**：Step 4 不要做成「通用 table 垫片」，做成「downloader / radio 各自改源码用 `<webf-table>`」
> + 一个只做**结构展平**的窄垫片。理由见 §2.6。

### 2.1 四个标签确实存在，注册在哪（已核实）

| 标签（JS 里写小写即可） | Dart 类 | 实现文件 | 注册点 |
|---|---|---|---|
| `<webf-table>` | `WebFTable` | `lib/src/html/table.dart`（262 行） | `lib/src/dom/element_registry.dart:231` |
| `<webf-table-header>` | `WebFTableHeader` | `lib/src/html/table_header.dart`（68 行） | `element_registry.dart:232` |
| `<webf-table-cell>` | `WebFTableCell` | `lib/src/html/table_cell.dart`（93 行） | `element_registry.dart:233` |
| `<webf-table-row>` | `WebFTableRow` | `lib/src/html/table_row.dart`（35 行） | `element_registry.dart:234` |

- 常量分别是 `WEBF_TABLE = 'WEBF-TABLE'`（`table.dart:13`）、`'WEBF-TABLE-HEADER'`（`table_header.dart:11`）、
  `'WEBF-TABLE-CELL'`（`table_cell.dart:12`）、`'WEBF-TABLE-ROW'`（`table_row.dart:11`）。
- 都是通过 `defineWidgetElement` 注册的 **WidgetElement**（不是普通 Element），
  即**由 Flutter widget 树渲染**，走 `WebFWidgetElementState.build()`。
- 底层就是 **Flutter 的 `Table` widget**（`table.dart:203` / `:219` / `:242` 三处 `Table(...)`）。
  **所有能力上限 = Flutter `Table` 的能力上限**，这句话是理解本节全部限制的钥匙。
- 生成的绑定在 `lib/src/html/table*_bindings_generated.dart`（`webf codegen` 产物）。

### 2.2 完整属性清单（从生成的绑定里逐条抄出，已核实）

**`<webf-table>`**（`table_bindings_generated.dart:71-104`）

| 属性 | 取值 | 效果 |
|---|---|---|
| `text-direction` | `ltr` / `rtl` | → `Table.textDirection`（`table.dart:145-153`） |
| `default-vertical-alignment` | `top`/`middle`/`bottom`/`baseline`/`fill` | → `Table.defaultVerticalAlignment`，默认 `middle`（`table.dart:115-131`） |
| `default-column-width` | 数字（px） | 非 null → **所有列**用 `FixedColumnWidth(值)`；否则 `FlexColumnWidth()`（`table.dart:136-145`） |
| `column-widths` | 字符串 | ⚠️ **死属性**。setter 存进 `_columnWidths`（`table.dart:79-85`），**`build()` 从头到尾没读它**（`build()` 用的是 `headerColumnWidths`，`table.dart:194`）。**别指望它** |
| `border` | 任意非空字符串 | 非空 → 硬编码 `TableBorder.all()`（`table.dart:169-175`）。**不解析值**，`border="1px solid red"` 和 `border="x"` 效果完全一样：默认黑色 1px 全框线 |
| `text-baseline` | `alphabetic` / `ideographic` | → `Table.textBaseline` |

**`<webf-table-header>`**（`table_header_bindings_generated.dart:15-21`）

| 属性 | 取值 | 效果 |
|---|---|---|
| `sticky` | `sticky` / `sticky=""` / `sticky="true"` | 走 sticky 布局分支（`table.dart:191`、`table_header.dart:47-56`）。其他值（含 `sticky="false"`）= false |

**`<webf-table-row>`**：**一个属性都没有**（`table_row_bindings_generated.dart` 整个只有一个空的
`abstract class WebFTableRowBindings extends WidgetElement`，12 行）。

**`<webf-table-cell>`**（`table_cell_bindings_generated.dart:31-43`）

| 属性 | 取值 | 效果 |
|---|---|---|
| `vertical-alignment` | `top`/`middle`/`bottom`/`baseline`/`fill` | → `TableCell.verticalAlignment`（`table_cell.dart:73-78`） |
| `column-width` | 数字（px） | **只有表头行的单元格上有效**，见 §2.3 |

**⚠️ 两个属性层面的坑（已核实，写垫片时必踩）**

1. **枚举属性传非法值会抛 `ArgumentError`**：`WebFTableTextDirection.parse` 等用
   `firstWhere(..., orElse: () => throw ArgumentError(...))`（`table_bindings_generated.dart:17-22`），
   而属性 setter 直接调 `parse`（`:74-78`）。所以 `text-direction="auto"` / `vertical-alignment="center"`
   **会抛异常**（不是静默忽略）。垫片必须只写白名单里的值。
2. **删属性会把宽度设成 0，不是「恢复默认」**：
   `default-column-width` 的 `deleter: () => defaultColumnWidth = 0.0`（`table_bindings_generated.dart:88`），
   `column-width` 的 `deleter: () => columnWidth = 0.0`（`table_cell_bindings_generated.dart:42`）。
   而 `_getDefaultColumnWidth` 判的是 `!= null`（`table.dart:138`）→ `0.0 != null` 成立 →
   `FixedColumnWidth(0.0)` → **整表列宽 0，内容全部不可见**。
   `removeAttribute('column-width')` 是个陷阱，**要清就设成空/不管，别 remove**。

### 2.3 列宽到底怎么定（这是最容易返工的一条，已核实）

**列宽只有两个来源，CSS 完全不参与：**

1. **`<webf-table-header>` 里各 `<webf-table-cell>` 的 `column-width` 属性** →
   `WebFTableHeader.getColumnWidths()`（`table_header.dart:29-43`）按**表头单元格的序号**
   生成 `{index: FixedColumnWidth(value)}`；没写 `column-width` 的列**不进这个 map**，
   由 `Table` 回落到 `defaultColumnWidth`。
2. **`<webf-table>` 的 `default-column-width`**（所有未指定列共用一个固定宽），
   不写则 `FlexColumnWidth()`（按内容/剩余空间分配）。

**推论（都要记住）：**

- **没有 `<webf-table-header>` = 没有任何列宽 map**：`headerColumnWidths` 直接是 null
  （`table.dart:194` 的 `header?.getColumnWidths()`），所有列走 `FlexColumnWidth()`。
- **`<td style="width:36px">` / `<th style="width:36px">` 的 CSS 宽度被完全忽略**。
  downloader 的 `index.html:110` 和 `:115` 正是 `<th style="width:36px">` / `<th style="width:60px">`
  → 垫片要把它翻译成 `column-width="36"` / `column-width="60"`，**否则那两列宽度全乱**。
- **body 行的单元格数必须与表头单元格数一致**，否则列宽 map 的下标会对错列
  （Flutter `Table` 本身要求所有 `TableRow` 子节点数相同，不一致会 assert）。
- sticky 模式下表头与表体是**两个独立的 `Table`**（`table.dart:203` 与 `:219`），
  靠**传同一份 `headerColumnWidths`** 保持对齐。所以
  **sticky 模式下如果列宽走 `FlexColumnWidth()`（没写 `column-width`），
  表头和表体是两次独立的 flex 分配 → 内容宽度不同就会错位。**
  ⚠️ **想要 sticky 表头对齐，必须给每一列都写死 `column-width`。** 这条是 radio / downloader
  两个 sticky 表头场景的硬要求，**需实测确认错位程度**，但从源码看几乎必然发生。

### 2.4 `colspan` / `rowspan`：**零支持**（已核实）

- `grep -rni "colspan|rowspan" lib/` 只命中 `css/grid.dart` 与 `rendering/grid.dart`
  （CSS Grid 的 `grid-row-span`，与表格无关）。
- 根因是 Flutter 的 `Table` widget **本身不支持单元格跨行/跨列**，
  `WebFTableCell.toTableCell()`（`table_cell.dart:70-75`）产出的就是一个裸 `TableCell`。
- 后果：`colspan` / `rowspan` 属性**被当作未知属性静默丢弃**，
  对应的单元格只占 1 格 → **整行列数与表头不符 → Flutter `Table` 会 assert 失败或错列**。
- ✅ 好消息：**已核实 downloader 的表格没有用 colspan/rowspan**
  （`jsplugins-src/songloft-plugin-downloader/static/index.html:107-119` 与
  `static/js/app.js:103-119` 的行模板，都是整齐的 6 列）。
- ⚠️ **radio 无法核实**：`jsplugins-src/songloft-plugin-radio/` 在本工作树是**空目录**
  （子模块未 checkout）。做 Step 4 前**必须**先 clone radio 确认它有没有 colspan、
  确认它的表头列数与行列数是否一致。

### 2.5 sticky 表头：**能做，但有三个附加条件**（源码已核实 + 一条需实测）

- **能做**：`table.dart:196-236` 有专门的 sticky 分支 —— 表头单独一个 `Table` 固定在
  `Column` 顶部，表体放进 `Expanded` 里的滚动区。触发条件是
  `header != null && header.sticky`（`table.dart:191`）。
- ⚠️ **条件一：`sticky` 是 `<webf-table-header>` 的属性，CSS `position: sticky` 不起作用。**
  downloader 现在的写法是 CSS：`css/style.css:316-328` 的 `th { position: sticky; top: 0; ... }`。
  改成 `<webf-table>` 之后表头由 Flutter `Table` 渲染、不走正常流，
  这段 CSS **必须换成 `<webf-table-header sticky>` 属性**。
  （WebF 本身是支持 `position:sticky` 的 —— `css/position.dart:17`、
  `rendering/box_model.dart:306-307`、`rendering/widget.dart:37-38` —— 但那对 Flutter `Table`
  内部的 widget 无效。）
- ⚠️ **条件二（重要）：sticky 分支用了 `Expanded`（`table.dart:214`），
  它在 `Column` 里要求父级给出「有界高度」。**
  `<webf-table>` 是 WidgetElement，高度来自它自己的 CSS 盒。
  → **必须给 `<webf-table>` 一个确定的 CSS 高度**（`height` 或 `max-height` + 明确的
  flex/grid 约束），否则 `Expanded` 拿到无穷高约束会抛
  `RenderFlex ... unbounded` 或产出 `Infinity`（**这跟交接文档 §6 记的 lxmusic
  `Unsupported operation: Infinity or NaN toInt` 是同一类症状，但我没有证据说那就是同一个 bug，
  别把两件事当成一件**）。
  downloader 现在把表格包在 `.table-wrap`（`index.html:106`：
  `style="max-height:calc(100vh - 380px);overflow-y:auto"`）里，`max-height` **不是** `height`
  → **需实测**这个约束能不能让 `Expanded` 满足。**建议直接给 `<webf-table>` 写死 `height`。**
- ⚠️ **条件三：sticky 模式下列宽必须写死**，见 §2.3 最后一条。

### 2.6 「机械改写能得到可用表格吗？」—— **不能。四个原因，逐条给证据**

**原因 ①（决定性）：行/表头必须是 `<webf-table>` 的直接子节点，`<thead>` / `<tbody>` 必须被拆掉**

- `table.dart:188`：`header = widgetElement.childNodes.firstWhereOrNull((n) => n is WebFTableHeader)`
- `table.dart:189`：`rows = widgetElement.childNodes.whereType<WebFTableRow>()`
- 两处都只看 **`childNodes`（直接子节点）**，不递归。
- 单元格同理：`table_header.dart:24-27` 与 `table_row.dart:21-25` 的 `buildCellChildren()`
  都是 `childNodes.whereType<WebFTableCell>()`。
- → 只把标签名一一对应地换掉（`table`→`webf-table`、`tr`→`webf-table-row`、`td`→`webf-table-cell`），
  **`<thead>` / `<tbody>` 会原样留下**，于是 `<webf-table>` 的直接子节点是两个
  `_UnknownHTMLElement`（display:block）→ **`rows` 是空列表、`header` 是 null →
  渲染出一张空表**（不报错、不打日志）。
- 还要注意 **只取第一个 header**（`firstWhereOrNull`），且 `<tfoot>` / `<caption>` /
  `<colgroup>` **没有任何对应物**，只能丢弃或自己另外渲染。

**原因 ②（对 downloader 是硬阻塞）：拆掉 `<tbody>` 会打断插件自己的 JS**

- `downloader/static/index.html:118` 是 `<tbody id="tbody"></tbody>`，
  `static/js/app.js:93` 拿 `const tbody = $('#tbody')`，
  `:96` 和 `:103` 用 **`tbody.innerHTML = ...`** 整体重写所有行。
- 这正是 Step 1 `detailsShim` 学到的教训的加强版：**插件按 id/标签查节点并直接写 `innerHTML`**。
  - 保留 `<tbody id="tbody">` → 行不是直接子节点 → 空表（原因 ①）
  - 拆掉它 → `$('#tbody')` 返回 null → **`tbody.innerHTML` 抛 TypeError → 整个渲染函数中断**
  - 把 `id="tbody"` 挪到 `<webf-table>` 上 → `innerHTML` 会**连表头一起抹掉**
    （downloader 每次刷新都整体重写），且写进去的是原始 `<tr>/<td>` → 又得重跑改写
- → **downloader 只能改源码**（它是第一方仓库，`.gitmodules` 里跟踪的
  `jsplugins-src/songloft-plugin-downloader`），让它直接产出 `<webf-table-row>` /
  `<webf-table-cell>`，并把「行容器」这个概念去掉（直接往 `<webf-table>` 里 append 行）。

**原因 ③：动态插入的行没有自动重跑机制 —— WebF 没有 `MutationObserver`**

- **已核实**：`grep -rn "MutationObserver" lib/` **零命中**。
- 所以垫片只能在 `applyOnReady` 跑一次；downloader 每次刷新都 `innerHTML = ...` 重建行，
  重建出来的是原始 `<tr>/<td>`。必须由插件在每次写完后调 `SongloftPlugin.applyShims()`。
- 结合原因 ②，「插件要改代码」这件事已经不可避免 —— **既然要改，不如直接改成写 `<webf-table-*>`，
  别再维护一层脆弱的改写垫片。**

**原因 ④：CSS 语义大面积丢失（各条都已核实）**

| downloader 现有 CSS | 改写后的命运 |
|---|---|
| `table { width:100%; border-collapse:collapse }`（`style.css:311-314`） | `border-collapse` 无对应物；`<webf-table>` 自带两层 `SingleChildScrollView`（`table.dart:238-241`），`width:100%` 的实际效果**需实测** |
| `th { position:sticky; top:0 }`（`:325-326`） | 无效，必须改用 `sticky` 属性（§2.5） |
| `th { padding:10px 12px; border-bottom:... }`（`:322-323`） | 单元格自己的 CSS **应该有效**（`toTableCell()` 把 `toWidget()` 包进 `TableCell`），但**需实测**——`WebFTableCellState.build()` 把内容包在一个 `tagName:'SPAN'` 的 `WebFHTMLElement` 里（`table_cell.dart:81-88`），**SPAN 是 inline**，单元格里放 block 子节点的布局行为需实测 |
| `tr:hover td { background:... }`（`:337`） | 行的 CSS 只以 `renderStyle.decoration` 形式参与（`table.dart:210`/`:225`/`:249` 的 `TableRow(decoration: row.renderStyle.decoration)`），即**只有背景/边框**这类装饰生效；`:hover` 能不能触发重建**需实测**（`WebFTableRowState.build()` 返回 `SizedBox.shrink()`，行本身不渲染） |
| `<th style="width:36px">`（`index.html:110`） | **完全无效**，必须翻译成 `column-width="36"`（§2.3） |
| `.table-wrap { overflow-x:auto }` + 内联 `max-height/overflow-y:auto`（`index.html:106`） | 与 `<webf-table>` 自带的两层 `SingleChildScrollView` **嵌套**，会出现「双层滚动」；sticky 模式还叠一个 `Expanded`（§2.5 条件二） |

### 2.7 原生 `<table>` 家族在两份注册表里的状态

**Dart 侧 `lib/src/dom/element_registry.dart` —— 已核实：一个都没注册。**

- `grep -n "defineElement\|defineWidgetElement" lib/src/dom/element_registry.dart` 的完整清单
  （`:143-247`）里**没有** `TABLE` / `TR` / `TD` / `TH` / `THEAD` / `TBODY` / `TFOOT` /
  `CAPTION` / `COLGROUP` / `COL`。
- 连**常量都不存在**：`grep -rn "const String TABLE\|const String TR \|const String TD \|const String TH \|const String THEAD\|const String TBODY\|const String CAPTION\|const String TFOOT\|const String COLGROUP" lib/`
  **零命中**。
- 未注册标签的归宿：`createElement`（`element_registry.dart:80-96`）→ `creator == null` →
  `_UnknownHTMLElement`，而
  **`_UnknownHTMLElement.defaultStyle => {'display': 'block'}`（`element_registry.dart:32-37`）**。
  → ✅ **交接文档 §3.3 说的「退化成嵌套 `display:block`」完全准确。**
- 顺带：未知标签的日志**只在 `enableWebFCommandLog` 打开时才 `debugPrint`**
  （`element_registry.dart:83-85`），默认**静默** —— 别指望日志告诉你用了不支持的标签。

**C++ 侧 `bridge/core/html/html_tag_names.json5` —— ❌ 无法从源码核实。**

- 该文件**不随 pub 包发布**。`webf-0.24.27/` 顶层只有
  `android/ example/ include/ lib/ linux/ macos/ scripts/ test/ tool/ windows/`，
  `include/` 里只有一个 `webf_bridge.h`；`find . -name "*.json5"` **零结果**。
- **需实测/需查上游仓库**。这条很重要，因为 HTML5 的表格解析规则是特例化最重的一块
  （隐式插入 `<tbody>`、foster parenting），而 WebF 用 gumbo 做 HTML 解析
  （包内有 `lib/src/bridge/native_gumbo.dart`，但那只是 FFI 声明）。
  → **实测时要专门确认：静态 HTML 里写 `<table><tr>...` 之后，DOM 里到底有没有被
  自动插入一层 `<tbody>`。** 这直接决定展平垫片要不要处理隐式 tbody。

**另一个可能有用的发现（`overrideCustomElement` 也救不了）**

- `WebF.defineCustomElement` / `WebF.overrideCustomElement`（`lib/src/widget/webf.dart:180-192`）
  都先过 `_isValidCustomElementName`（`:147-177`），该函数**要求首字符 a-z 且必须含至少一个 `-`**
  （`return hasDash`，`:176`）→ ✅ 交接文档 §3.1 的说法准确，**`'table'` 会抛 `ArgumentError`**。
- ⚠️ 但注意：内部函数 `defineOverrideWidgetElement(name, creator)`（`element_registry.dart:64-69`）
  **本身没有连字符检查**，而 `createElement` 的查表顺序是
  `_htmlRegistry[name] ?? _overrideWidgetElements[name] ?? _widgetElements[name]`
  （`element_registry.dart:81`）—— `TABLE` **不在** `_htmlRegistry` 里，
  所以理论上 `defineOverrideWidgetElement('TABLE', ...)` 是能生效的。
  **但不要走这条路**，三个理由：
  ① 它绕过了公开 API 的校验，属于依赖内部实现（`lib/src/` 下，随时可变）；
  ② C++ 侧的标签表与 widget element 形状表在 controller 创建前就同步过去了
  （`bridge/to_native.dart:802-807` 传 `widgetElementCreators.length` 与 shapes），
  Dart 侧偷偷加一个 `TABLE` 不保证 C++ 那边认；
  ③ 就算认了，`<thead>`/`<tbody>`/`colspan`/CSS 宽度的问题**一个都没解决**（§2.6 原因 ①③④）。
  **记录在这里只是为了让下一个人不用再去发现它、也不要被它诱惑。**

### 2.8 结论与建议

**回答那个问题：「机械地 table→webf-table、tr→webf-table-row、td→webf-table-cell 改写，
能得到可用的表格吗？会丢什么？」**

**不能。** 直接机械改写得到的是**一张空表**（`<thead>`/`<tbody>` 没拆 → 行不是直接子节点）。
即使补上「展平 thead/tbody」这一步，还会丢：

- `colspan` / `rowspan`（零支持，且列数不齐会让 Flutter `Table` assert）
- 所有列宽（CSS `width` 无效，必须翻译成表头单元格的 `column-width`）
- CSS sticky 表头（必须改成 `sticky` 属性）
- `border-collapse`、`<caption>`、`<tfoot>`、`<colgroup>`
- 插件自己按 `#tbody` / `tr` / `td` 做的 DOM 操作（downloader 的 `tbody.innerHTML` 是硬阻塞）
- 动态新增行的自动接管（WebF 无 `MutationObserver`）

**建议的 Step 4 形态（收益/成本最优）：**

1. **主路径：改插件源码**。downloader 与 radio 都是第一方子模块，直接让它们输出
   `<webf-table>` / `<webf-table-header sticky>` / `<webf-table-row>` / `<webf-table-cell>`，
   列宽写成表头单元格的 `column-width`，去掉 `<tbody>` 这层（行直接 append 到 `<webf-table>`）。
   为了同时兼容浏览器/WebView 路径，两种做法：
   - **(a) 按 `document.documentElement.classList.contains('webf-engine')` 分叉渲染函数**
     （Step 1 已经在 `installEarly()` 里给 `<html>` 打了 `webf-engine` class，直接用），
     两套模板各自最优；
   - **(b) 统一只写 `<webf-table-*>`，另在 `common.css` 里给这些标签补一套
     `display:table/table-row/table-cell` 的 CSS**，让普通浏览器也能渲染。
     **(b) 更省事，但浏览器路径的 sticky 表头得靠 CSS 另做，且 `column-width` 属性对浏览器无意义**
     → 建议 **(a)**。
2. **垫片只做一件窄事**：把**遗漏的**原生 `<table>` 结构**展平并改名**（含处理可能的隐式
   `<tbody>`），并把 `<th style="width:Npx">` 翻译成 `column-width="N"`、
   把 `th { position:sticky }` 的意图翻译成 `sticky` 属性。
   **明确写进注释：这是尽力而为的兜底，不承诺 colspan/rowspan/border-collapse**，
   并在改写时 `console.warn` 提示插件作者改源码。
3. **实测清单**（做之前先跑，都很短）：
   - `<webf-table>` 不给 CSS 高度时，sticky 分支的 `Expanded` 会不会崩 → 决定要不要强制写 `height`
   - 静态 HTML 的 `<table><tr>` 是否被 gumbo 插了隐式 `<tbody>`
   - sticky 模式下不写 `column-width` 时表头与表体错位的实际程度
   - 单元格 CSS（padding / border-bottom / background）是否生效；`tr:hover` 是否生效
   - 表头/表体单元格数不一致时是崩还是错列（决定垫片要不要补齐空单元格）

---

## 第 3 节 — Step 6：三项经桥下沉

> **一句话结论**：这一节的三项**风险都比交接文档假设的低**，而且有两个能省掉大量工作的发现：
> ① **`file_picker` 和 `url_launcher` 都已经是 player 的现有依赖**，用它们**不会改动原生契约哈希**
> —— 代价不是「顺便一起破」，而是**零**；
> ② **`window.open` 在 Dart 层不是 no-op**，它走 `handleNavigationAction` → `WebFNavigationDelegate`，
> 而产品**没有设置 delegate** → 落到默认的「一律 cancel」处理器。
> 如果 C++ 侧确实转发到 Dart（**这一步无法从源码核实，但有一个决定性的廉价实验**），
> 那 `window.open` 只需在 `_createController()` 里加约 8 行、**完全不需要 JS 垫片**。

### 3.1 `file_picker` 与原生契约哈希：**已核实，零额外代价**（比推断更好）

**先把哈希算法讲清楚**（`songloft-player/scripts/compute_native_contract.sh`，已通读）。
`dart` 哈希 = 下面 4 段文本按固定顺序拼接、`LC_ALL=C` 排序后 `sha256`：

| 段 | 内容 | 脚本行号 |
|---|---|---|
| `## channels` | Kotlin 源里所有 `"com.songloft/[a-z_]+"` 字面量 | `:60-64` |
| `## methods` | Kotlin 源里 `call.method == "x"` 与 `"x" ->` 两种模式抽出的方法名 | `:67-75` |
| `## plugins` | `android/app/src/main/java/io/flutter/plugins/GeneratedPluginRegistrant.java` 里所有 `add(new X())` 行 | `:78-81` |
| `## plugin-versions` | `.flutter-plugins-dependencies` 里 android 段每个插件的 `name<TAB>version`（version 从 pub 缓存目录名 `name-version` 里切出来） | `:84-101` |

`go` 哈希 = `sha256(sort(mobile/export_surface.txt))`，仅 bundle 版（`:107-110`），与本节无关。

**核实用户的推断**：

- ✅ **「引入 webf 破坏了一次哈希」——正确。** 已核实
  `GeneratedPluginRegistrant.java:119` 有
  `flutterEngine.getPlugins().add(new com.openwebf.webf.WebFPlugin());`，
  且 `.flutter-plugins-dependencies` 的 android 段里有 `webf | webf-0.24.27`。
  → 同时命中 `## plugins` 与 `## plugin-versions` 两段 → dart 哈希必变 → 本轮**确实**是整包发版事件。
- ❌ **「`file_picker` 属于顺便一起破」——推断的前提不成立，但结论比推断更好。**
  **`file_picker` 早就在依赖里了**：
  - `songloft-player/pubspec.yaml:53`：`file_picker: ^10.3.10`
  - `GeneratedPluginRegistrant.java:29`：`add(new com.mr.flutter.plugin.filepicker.FilePickerPlugin())`
  - `.flutter-plugins-dependencies` android 段：`file_picker | file_picker-10.3.10`
  - `pubspec.lock` 里 `dependency: "direct main"`
  - 实际使用点（说明不是残留依赖）：`lib/features/playlist/presentation/widgets/playlist_cover_edit_mixin.dart`、
    `lib/features/jsplugin/presentation/widgets/jsplugin_manager.dart`、
    `lib/features/playlist/presentation/widgets/playlist_browse_view.dart`、
    `lib/features/settings/presentation/widgets/settings_category_content.dart`、
    `lib/core/backend/embedded_backend_service.dart`
  → **给 `input[type=file]` 用 `file_picker` 不需要动 `pubspec.yaml`，
  `## plugins` 与 `## plugin-versions` 两段一个字符都不变，dart 哈希不变。代价是零。**

**所以准确的说法是**：`file_picker` 这一条**根本不产生契约哈希代价**，
既不需要「借 webf 已经破了」当理由，也不需要在发版计划上做任何安排。

**⚠️ 但有三条真实的哈希雷区，做 Step 6 时要避开：**

1. **不要 bump `file_picker` 的版本号。** `## plugin-versions` 段把**版本号**也纳入了哈希
   （`compute_native_contract.sh:96` 从目录名 `file_picker-10.3.10` 切出 `10.3.10`）。
   `pubspec.yaml` 写的是 caret `^10.3.10` → **一次 `flutter pub upgrade` 就可能解析到 10.3.11，
   在没人改一行代码的情况下把哈希改掉**。`pubspec.lock` 已入库（已核实存在），
   所以只要 CI 走 `flutter pub get` 就稳定；**别在这条分支里跑 `pub upgrade`**。
2. **不要为了「拿到文件字节」而引入新的原生插件**（例如另装一个 `image_picker`）。
   那才会真的改 `## plugins` + `## plugin-versions`。`file_picker` 已能返回 `bytes` / `path`，够用。
3. **不要在 Kotlin 里加新 MethodChannel 方法**（player `AGENTS.md`「Kotlin 层冻结规则」第 1 条）。
   Step 6 的三项都应该走**WebF 自己的 `javascriptChannel`**（产品已在
   `plugin_render_surface_webf.dart:154-162` 用 `_onMethodCall` 接 `songloft.host` 调用），
   那条通道**不经过 Kotlin**、不进哈希。

**顺带核实**：`.flutter-plugins-dependencies` 是入库文件（`ls` 可见，且脚本直接读仓库根的这份），
所以「本地 `pub get` 后忘记提交它」会造成 CI 与本地哈希不一致。改依赖时记得一起提交。

### 3.2 `url_launcher`：**已核实，早就在了**

- `songloft-player/pubspec.yaml:59`：`url_launcher: ^6.3.1`
- `GeneratedPluginRegistrant.java:104`：`add(new io.flutter.plugins.urllauncher.UrlLauncherPlugin())`
- `.flutter-plugins-dependencies`：`url_launcher_android | url_launcher_android-6.3.32`
- 现有使用点：`client_download_page.dart`、`upgrade_dialog.dart`、`frontend_upgrade_dialog.dart`、
  `settings_category_content.dart`、`plugin_webview_page_native.dart`、`plugin_webview_page_stub.dart`、
  `jsplugin_manager.dart`

→ **零新增依赖、零哈希影响。** 注意 `plugin_webview_page_native.dart` **已经**在
webview 路径上用 `url_launcher` 处理外链，Step 6 只是要在 WebF 路径上补齐同样的行为
——**去抄那份实现**，别重新设计。

### 3.3 `window.open`：交接文档的「no-op」是**现象正确、归因可能错**，且有一条更便宜的解法

**已核实（Dart 侧）：`window.open` 有完整实现，不是空函数。**

- `lib/src/dom/window.dart:43-44` 把 `'open'` 注册成
  `StaticDefinedSyncBindingObjectMethod` → `Window.open(String url)`
- `lib/src/dom/window.dart:71-74`：
  ```dart
  void open(String url) {
    String? sourceUrl = document.controller.view.rootController.url;
    document.controller.view.handleNavigationAction(sourceUrl, url, WebFNavigationType.navigate);
  }
  ```
- `handleNavigationAction`（`lib/src/launcher/view_controller.dart:1107-1158`）第一件事就是
  `await delegate.dispatchDecisionHandler(action)`，
  **`policy == cancel` 时直接 `return`**（`:1117-1118`）。
- 默认处理器 `defaultDecisionHandler`（`lib/src/module/navigation.dart:85-106`）
  **无条件 `return WebFNavigationActionPolicy.cancel`**，
  并且在 `kDebugMode || kProfileMode` 下 `debugPrint` 一段很显眼的提示：
  > `Attempting to navigate WebF to an external WebF page: **<target>** from **<source>**. This behavior is disabled by default.`
- **已核实产品没有设置 delegate**：`plugin_render_surface_webf.dart:219-262` 的
  `_createController()` 传的是 `networkOptions` / `onLoad` / `onLoadError` / `onJSError`，
  之后只赋了 `onJSLog` 和 `javascriptChannel.onMethodCall`，**没有 `navigationDelegate`**。
  `WebFController.navigationDelegate` 的 getter 是
  `_navigationDelegate ?? WebFNavigationDelegate()`（`controller.dart:285-286`）
  → 拿到的是一个**全新的、用默认 cancel 处理器的** delegate。

**结论**：从页面视角看，`window.open(url)` 确实什么都没发生 —— **交接文档描述的现象是对的**。
但归因很可能不是「C++ 的两个重载都 `return this`」，而是
**「Dart 侧被默认导航策略 cancel 掉了」**。这两个归因导向**完全不同的修法**。

**❗ 无法核实的一环（老实说）**：`window.cc:157-168` 是 C++ 桥的源码，
**不随 pub 包发布**（`webf-0.24.27/` 顶层只有 `android/ example/ include/ lib/ linux/ macos/
scripts/ test/ tool/ windows/`，`include/` 里只有 `webf_bridge.h`）。
所以我**无法判断 C++ 的 `Window::open` 是否把调用转发到 Dart 的 binding**。
两种可能：

- **A. C++ 转发到 Dart** → 现在的行为 = 被默认策略 cancel →
  **修法：设一个 `navigationDelegate`，约 8 行搞定，无需 JS 垫片。**
- **B. C++ 自己 `return this` 就结束了** → Dart 的 `Window.open` 永远不会被调到 →
  **修法：JS 垫片覆写 `window.open`，经 `songloft.host` 桥调 Dart 的 `url_launcher`。**

**决定性实验（十分钟，务必先做这个再动手）**：

在验证容器里（debug 模式，`debugPrint` 已被转发到 `out/flutter.log`，见交接文档 §5）
渲染任意页面并执行 `window.open('https://example.com')`，然后 grep 日志：

```
grep -F "Attempting to navigate WebF to an external WebF page" \
  songloft-player/scripts/webf-verify/out/flutter.log
```

- **有这行** → 情形 A 成立 → 走 delegate 方案（下面的代码骨架）
- **没有这行** → 情形 B 成立（或探针没触发到），退回 JS 垫片方案

**情形 A 的实现骨架**（写在这里省下下一个人的摸索时间；注意
**`navigationDelegate` 不是 `WebFController` 的构造参数** —— 上游
`navigation.dart:96-103` 的文档示例写成 `WebFController(navigationDelegate: ...)` 是**过时/错误的文档**，
`controller.dart:799-828` 的构造参数列表里**没有**它。它是构造后可写的属性，
setter 会顺带同步到 view：`controller.dart:288-294`）：

```dart
// _createController() 内，controller 构造完之后：
final delegate = WebFNavigationDelegate();
delegate.setDecisionHandler((action) async {
  final target = action.target;
  // 外链交给系统浏览器；同源/相对路径按需自行决定放不放行。
  if (target.startsWith('http://') || target.startsWith('https://')) {
    await launchUrl(Uri.parse(target), mode: LaunchMode.externalApplication);
  }
  // 一律 cancel：绝不让 WebF 自己 load 走，否则插件页会被整页替换掉。
  return WebFNavigationActionPolicy.cancel;
});
controller.navigationDelegate = delegate;
```

**⚠️ 这个方案的三个必须注意的副作用（都已从源码核实）：**

1. **它不只拦 `window.open`，也拦 `<a href>` 点击**：`lib/src/html/a.dart:62-73` 的
   `_getNavigationType` 走的是同一个 `handleNavigationAction`。
   所以**返回 `allow` 会让插件页被整页导航替换掉**（`view_controller.dart:1138-1143` 的
   `navigate` 分支是 `rootController.load(...)`）。**一律 `cancel` 是正确的默认**。
2. **锚点跳转（`#xxx`）在 cancel 之后就不工作了**：`handleNavigationAction` 里
   `#` 开头的处理（`view_controller.dart:1126-1134`，pushState + `HashChangeEvent`）
   **排在 cancel 检查之后**。如果插件用了 `<a href="#tab-player">` 这类页内跳转，
   decision handler 必须对 `target.trim().startsWith('#')` 返回 **`allow`**。
   **先 grep 一遍插件里有没有 `href="#..."`**，别一刀切 cancel。
3. **`WebFControllerManager` 的 detach/dispose 重建 controller 时要重新设 delegate**
   —— 与 §1.5 提到的重挂问题同源。产品的 `WebF.fromControllerName(createController:)`
   在被淘汰后重建时 `createController` 不一定再跑（`plugin_render_surface_webf.dart:281`
   的注释已经记了这个坑），所以 delegate 的赋值要和那里的「兜住引用与桥」逻辑放在一起。

**情形 B 的实现要点**：JS 垫片覆写 `window.open`。注意
**必须放 `applyOnReady` 之前的 `installEarly()`**（页面脚本可能在解析期就调 `open`），
且**要保留返回值形状**（返回一个假的 window-like 对象或 null，
miot `js/auth.js:95` 拿返回值做什么**需去读那段代码确认**——本工作树里 miot 的 `static/`
归 Step 3 的 agent 所有，我按要求只读不改，但没有把 auth.js 纳入本次核实范围）。

### 3.4 `URL.createObjectURL`：**改 `data:` URL，不要落盘**（已核实用途）

**两处的实际用途已核实，都是「拉带鉴权头的封面图 → 显示」，不是下载。**

**① `jsplugins-src/songloft-plugin-miot/static/js/playback.js:419-424`**（播放栏封面缩略图）
```js
fetchWithAuth(coverUrl, COVER_FETCH_TIMEOUT_MS).then(blob => {
    if (coverUrl !== currentPlayerBarCoverUrl) return;
    playerBarCoverObjectUrl = URL.createObjectURL(blob);
    const img = document.getElementById('playerBarCover');
    if (img) img.src = playerBarCoverObjectUrl;
})
```
→ 只塞 `img.src`。前面还有 `URL.revokeObjectURL` 的清理（`:415-417`、`:434-435`）。

**② `jsplugins-src/songloft-plugin-miot/static/js/fullscreen-player.js:199-205`**（全屏播放器封面）
```js
fetchWithAuth(url, COVER_FETCH_TIMEOUT_MS).then(blob => {
    if (url !== coverUrl) return;
    if (coverObjectUrl) URL.revokeObjectURL(coverObjectUrl);
    coverObjectUrl = URL.createObjectURL(blob);
    if (coverImg) coverImg.src = coverObjectUrl;
    if (bgImage) bgImage.style.backgroundImage = `url(${coverObjectUrl})`;
})
```
→ 塞 `img.src` **和** CSS `background-image: url(...)`。**注意这第二个消费点**
（交接文档只说了「带鉴权头拉封面 ×2」，没说其中一处还要走 CSS `url()`）。

**结论：`data:` URL 是正解，理由有源码支撑：**

- ✅ **`data:` URL 被 WebF 的资源加载器原生支持**：`lib/src/foundation/bundle.dart:82-84`
  的 `_isDataScheme(path) => path.startsWith('data:')`，
  `WebFBundle.fromUrl`（`:175-195`）在 `_isDataScheme` 分支返回
  `DataBundle.fromDataUrl(url)`（`:246-275`，用 `UriData.parse`，失败还有
  `_parseDataUrlFallback` 手工解析 `:276-300`）。
- ✅ **`<img src>` 走的就是这条路**：`lib/src/html/img.dart:1332` 是 `WebFBundle.fromUrl(...)`。
- ✅ **`blob:` 走不通，即使 JS 侧能产出也一样**：`WebFBundle.fromUrl` 的分支只有
  http/https、assets、file、`data:`、DEFAULT_URL，**else 直接
  `throw FlutterError('Unsupported url. $url')`**（`bundle.dart:192-194`）。
  → 「给 WebF 加个 `createObjectURL` 垫片返回 `blob:xxx`」这条路**从加载器层面就是死的**，
  别去尝试。（`lib/src/painting/image_provider_factory.dart:16-56` 的 `ImageType` 枚举里
  确实有个 `blob` 值、注释还写着 "created by URL.createObjectURL()"，
  但 `grep -rn "ImageType\." lib/` **零命中** → 那是从 Kraken 继承下来的**死代码**，
  别被它误导。）
- ⚠️ **CSS `background-image: url(data:...)` 能不能用：需实测。**
  CSS 背景图走 `lib/src/css/background.dart` 而不是 `img.dart`，
  我没有追完它到 `WebFBundle.fromUrl` 的完整链路。`data:` URL 里含逗号和分号，
  **CSS `url()` 的词法解析可能出问题**（WebF 自己的 `CSSFunction.parseFunction`
  在 `values/function.dart:88-92` 对 `url(` 做了「不按逗号切参数」的特殊处理，
  这是个好兆头，但不等于 base64 长串一定能过）。
  **实测这一条**；不行的话第 ② 处的 `bgImage` 只能退而求其次（比如改成叠一个 `<img>` 元素）。

**「落盘」为什么不选**：这两处是**每次切歌都会变的封面缩略图**，
落盘要处理写入目录、并发、清理、`revokeObjectURL` 对应的删除，而封面本来就是几十 KB 的图。
`data:` URL 在内存里、随 `img.src` 替换自动回收，语义上和 objectURL 一一对应
（`revokeObjectURL` 变成 no-op 即可）。**只有当封面大到 data URL 影响性能时才考虑落盘**，
现在没有证据说会。

**实现路径建议**（垫片，无需 Dart 侧介入）：

- 在 `common.js` 的 **early** 垫片里补一个 `URL.createObjectURL` / `URL.revokeObjectURL`
  的**同步 API 无法用异步实现**的问题：`blob → base64` 本质是异步的
  （`FileReader` 或 `blob.arrayBuffer()`），而 `createObjectURL` 是**同步返回字符串**。
  → **不要**试图垫 `createObjectURL` 本身。**改插件的调用点**：
  让 `fetchWithAuth` 在 WebF 下直接返回 `data:` 字符串（内部 `await blob.arrayBuffer()` +
  base64），调用点从 `URL.createObjectURL(blob)` 改成直接用返回值。
  这也顺带解释了为什么 Step 6 这一项**必须改 miot 源码**，垫片救不了。
- ⚠️ **需实测**：WebF 的 JS 运行时里有没有 `btoa` / `FileReader` / `Blob.prototype.arrayBuffer`。
  这三个都在 C++/QuickJS polyfill 层，**不在 pub 包里，无法从源码核实**。
  `grep -rn "createObjectURL|FileReader|btoa|blob:" lib/` 在 Dart 层**零命中**（只有那条注释），
  说明它们要么在 C++ 侧、要么不存在。**先探针确认再动手。**
  兜底方案：`arrayBuffer` + 手写 base64 编码表（纯 JS，不依赖任何宿主 API），
  几十行，性能对几十 KB 的封面完全够。

### 3.5 `input[type=file]`：现象已核实，但命中面无法核实

- ✅ **「静默变文本框」已核实**：`lib/src/html/form/input.dart:250-266` 的
  `build()` 只 `switch` 了 `radio` / `checkbox` / `button` / `submit` / `date` / `time`，
  **`default: return createInput(context)`**；`createInput`（`:268-278`）只额外处理
  `hidden`（返回 `SizedBox(0,0)`），其余一律 `createInputWidget` = 文本框。
  → `type=file` 和 `type=range` 都落到 `default`，**都是文本框**，与交接文档一致。
  （`input.type` 的取值见 `form/base_input.dart:214`：`getAttribute('type') ?? 'text'`。）
- ❌ **命中面无法核实**：交接文档说「ytdlp、radio、lxmusic 各 1」。
  本工作树里 `jsplugins-src/songloft-plugin-radio/` 是**空目录**（子模块未 checkout），
  `ytdlp` / `lxmusic` **根本不在 `jsplugins-src/` 下**（交接文档 §6 也说它们不是跟踪的子模块）。
  → 做这一项前**必须先 clone 出来**，确认：
  ① 文件选完之后插件拿它做什么（读文本内容？上传 multipart？只取文件名？）——
  这决定桥的返回值形状是 `path` / `bytes(base64)` / `text`；
  ② 有没有 `multiple` / `accept`；
  ③ 有没有读 `input.files[0].name` 这类 `FileList` API（**WebF 有没有 `File`/`FileList`
  同样无法从源码核实，在 C++ 侧**）。
- **建议的形态**（与 §3.1 雷区 3 一致，不碰 Kotlin）：
  JS 垫片给 `input[type=file]` 挂 click 拦截 → 经 `songloft.host` 桥调 Dart →
  Dart 用**现有的** `file_picker` 弹选择器 → 返回 `{name, path, bytesBase64}` →
  垫片在 input 实例上定义 `files` 访问器并派发 `change` 事件。
  **保留原 `<input>` 节点不要移除**（Step 1 / Step 3 的同一条教训：插件按标签/id 查节点）。
  ⚠️ `dispatchEvent` 在本项目是 Step 3 才第一次真用，**JS 侧能否真的收到需实测**
  （交接文档 §4 Step 3 已经把这一条列为必须实测项）——
  **Step 6 应该复用 Step 3 得出的结论，不要重复踩。**

### 3.6 Step 6 小结

| 项 | 风险 | 结论 |
|---|---|---|
| `file_picker` 的契约哈希代价 | **无** | **已核实**：`file_picker` 已是现有依赖（pubspec:53 / Registrant:29 / plugin-deps），dart 哈希不变。唯一雷区是别 bump 它的版本、别引新原生插件、别改 Kotlin |
| `url_launcher` 是否已在 pubspec | **无** | **已核实**：`pubspec.yaml:59` + Registrant:104，且 webview 路径已有现成实现可抄 |
| `window.open` | **低**（但归因待定） | Dart 侧**有实现**（`window.dart:71-74`）→ `handleNavigationAction` → 默认 delegate **cancel**。C++ 是否转发**无法核实，需实测**（grep 日志里那句 `Attempting to navigate WebF to an external WebF page`）。若转发，约 8 行 delegate 解决；注意别把 `#` 锚点一起 cancel |
| `URL.createObjectURL` | **低** | 两处都只是显示封面 → **改 `data:` URL**（`bundle.dart:82-84`/`:175-195` 已核实支持）。`blob:` 从加载器层面就不可能（`bundle.dart:192-194` 直接 throw）。CSS `background-image: url(data:...)` **需实测**；`btoa`/`FileReader` 是否存在**需实测** |
| `input[type=file]` | **中**（因为命中面不明） | 「变文本框」已核实（`form/input.dart:250-278`）。但 3 个命中插件**一个都没 checkout**，实现形状取决于它们拿文件做什么 → **先 clone 再设计** |
