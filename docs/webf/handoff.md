# WebF 插件页渲染迁移 — 交接文档（songloft-org/songloft#341）

> **这是分支临时交接件，不是产品文档。** 它是 `docs/webf/` 目录的**主文档**。
> 整个目录刻意**不做**中英双语同步、也整体从文档站排除
> （`docs/.vitepress/config.mts` 的 `srcExclude: ['webf/**']`）。
> **#341 落地后连同整个 `docs/webf/` 目录一起删除。**
>
> 面向对象：接手这条分支继续做的下一个 agent / 开发者。
> 最后更新：2026-08-05（**新增第 29 条 —— flex 套 flex 会让整个子树静默不绘制**，这是第 21 条
> 那个 base size 缺陷的另一副面孔，downloader 已修复并像素验证；同日：**第 26 条被推翻、
> 新增第 27、28 条** —— 「过一会白屏」的真根因是
> 「渲染面重新挂载 + controller 命中缓存」，不是两个渲染面抢 controller；同一个白屏我连错
> 三次归因，教训记在第 26 条末。同日此前：新增第 21 条 flex `wrap` 下 base size 被测成
> 容器宽度、第 22 条 `[plugin][console]` 转发在缓存命中时静默失效并给第 16 条补上后果 ④。
> 更早：2026-08-03 Step 4/6 实测落地并发版、撤回 data URL 的过度更正、
> 缺陷台账扩到 11 条、文档搬进 `docs/webf/`）。
>
> **同目录的配套件** —— 先读 [README.md](README.md)，它有导航、当前状态速览、
> 以及**被推翻结论的完整清单**：
> [recon-step456.md](recon-step456.md)（Step 4/5/6 预研，证伪了两条既定方案）、
> [step4-design.md](step4-design.md)（Step 4 四方案对比与选型）、
> [upstream-issues.md](upstream-issues.md)（10 条上游缺陷草稿，英文正文）。
>
> ⚠️ **本文档反复发生过「先写下结论、后被实测推翻」**，每处都保留原文并划掉。
> 只搜关键词、不读上下文的话可能捡到废弃结论 —— 动手前先扫一眼 README 里那张推翻清单。

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

### 🆕 2026-08-04 之后：webf-ui 路线（downloader 已落地）

> **这一节晚于本文档其余部分，与它们冲突时以本节为准。**

2026-08-04 三个插件的 `renderEngine` 曾被 `397f4bd` 全部回退成 `webview`。之后 downloader
**换了思路重做**：不再拿浏览器语义的 HTML/CSS 硬凑，而是用 **webf-ui 原生组件**
（`webf_cupertino_ui` 的 31 个 `<flutter-cupertino-*>` + `webf` 包内建的 `<webf-list-view>`），
前端改为 Vue 3 + Vite，并已重新声明 `renderEngine: "webf"`（随下一次发版生效）。

> ⚠️ **v2026.8.4 不是这一版。** 那个 tag 打在 `a9ed05b`（回退成 webview 的那次）上、当天已由
> 发版 bot 发布，内容是**旧的 webview 前端**。webf-ui 这一版要等下一次发版才有版本号 ——
> 引用它时**别写 v2026.8.4**。顺带记住这个仓库的约定：`plugin.json` 的 `version` /
> `download_url` **只由发版 bot 改**，功能提交一律不碰（历史上 `02b9b11` / `4f747bd` /
> `d629153` / `a9ed05b` 都没碰过）。手写版本号会撞上已发布的 tag，而商店的 `has_update` 用的是
> `CompareVersion(...) > 0`，撞号的结果是**用户永远收不到更新推送**。

**这条路线让 §3.2 里最难缠的几条不再命中**：第 7 条（sticky）、第 9 条（grid `auto` 行高）、
§3.3 的 `<table>` 那一行 —— 因为根本不再走 CSS grid / table 布局，行高由 Flutter 排版决定。
**§4 的 Step 4（`<table>` → CSS Grid）因此降级为「真正的二维表格才用」的备选方案**，
列表类内容一律优先 `<webf-list-view>`。

宿主侧改动：`plugin_render_surface_webf.dart` 的 `_ensureWebFProcessSetup()` 新增第 ③ 步
`installWebFCupertinoUI()`。**刻意不放进 `elements/`** —— 那个目录有「只能 import flutter 与
webf」的铁律（§7），而它需要 import `webf_cupertino_ui`。探针侧同步（`probe_pubspec.yaml`
加依赖 + `probe_main.dart` 调 install），否则探针里 cupertino 元素全是未知标签、验不了。

#### 本轮新核实的事实（都读了包内源码，不是推断）

1. **`installWebFCupertinoUI()` 内部是 31 条连续的 `WebF.defineCustomElement(...)`，
   没有逐元素 try/catch。** 而 `defineCustomElement` 对重复注册是**抛异常**（热重启后进程级
   registry 仍在）。任何一条抛出，后面的元素全部注册不上 → 必须整体包 try/catch。
   这与 `SongloftCustomElements._define` 逐个包的理由完全一致。
2. **cupertino 元素的 HTML 属性是 kebab-case，JS 属性才是 camelCase。**
   `switch_bindings_generated.dart` 里 `attributes['active-color']` 对应
   `'activeColor': StaticDefinedBindingProperty`。模板里写 camelCase 会变成全小写属性、
   静默不匹配任何注册项。
3. **同一个布尔属性的两个入口语义相反，这是本轮最危险的一条。**
   HTML 属性 setter 是 `value == 'true' || value == ''`（认字符串），
   JS 属性 setter 是 `value == true`（认真布尔，见 `switch.dart:26-32`）。
   于是字符串 `'true'` 走 JS 属性入口得到 **false**。而 Vue / React 对自定义元素是**启发式**
   决定走 prop 还是 attr 的 → 插件侧无法确定，选错就是「开关点了没反应」。
   **结论：布尔属性一律绕开模板绑定，命令式赋 JS 属性并传真布尔。**
4. **`<flutter-cupertino-input>` 的值属性叫 `val` 不是 `value`，且它是受控的**：
   `build()` 里 `if (_controller.text != val)` 就整段替换文本并把光标收到末尾
   （`input.dart:265-278`）。所以回写路径上**不能做类型转换**，数字字段要存字符串、
   提交时才 `parseInt`。
5. **`<flutter-cupertino-input>` 没有 `change` 事件**，只有 `input` / `submit` / `focus` /
   `blur` / `clear`。「改完即存」用 `blur`。
6. **`webf-list-view` 的 `shrink-wrap` 默认是 `true`**，不显式关掉就不在内部滚动。
   关掉后必须给显式高度（无界约束会触发 §3.2 第 11 条的 `Infinity or NaN toInt`）。
   标签名 `WEBF-LISTVIEW` 与 `WEBF-LIST-VIEW` 两个别名都注册了
   （`webf-0.24.27/lib/src/html/listview.dart:52-53`）。
7. **Cupertino 31 个元素里没有任意选项的下拉选择器**（只有 `date-picker`）。下拉继续用原生
   `<select>`（§3.3 已记录它在 WebF 下可用）。
8. **`common.css` 自己的 `.snackbar` 在 WebF 下是坏的**：它靠
   `position: fixed` + `transform: translateX(-50%)` 定位，而 WebF 不对 fixed 元素应用
   transform（miot `9196fe6` 已实测过这条）。插件要自绘 snackbar 就不能复用它 ——
   用「外层 left/right 拉满 + 内层 margin auto」居中。**这条也意味着宿主的 `.snackbar`
   本身值得修**，但那会影响所有插件，未动。
9. **`webf_cupertino_ui` 是 Apache-2.0**（以包内 `LICENSE` 为准）。npm 上同名包的元数据写
   ISC 是错的。**不引入新的 copyleft** —— 客户端整体按 GPL-3.0 分发的原因仍只有 webf 本身。
   已记进 `songloft-player/NOTICE` 的第 1 项与第 6 项。
10. **`@songloft/plugin-builder` 的 `frontend/` 构建钩子只有 ≥ 2.13.1 才有。**
    downloader 原先锁在 2.5.0（`package-lock.json`），更早的版本会**静默跳过**整个前端构建、
    产出一个没有前端资源的包。要用 `frontend/` 的插件必须 bump 这个 devDependency。

#### 真机跑出来的十条（2026-08-04，用户在 macOS 客户端上看到的）

前 10 条是读源码得来的；这几条是**页面真的画出来之后**才暴露的，且**全部是「画出一个东西但
不对」而非报错**，所以只看日志抓不到。

> ⚠️ **第 14 条（进程内缓存的旧 bundle）请先读。** 它是这一批里唯一会让你**误判其它所有条**
> 的坑：本轮为它烧掉了整整一轮返工，还害得第 11 条的机理被错记了一次。

11. **cupertino button 只渲染 `childNodes.first`。** `button.dart:154` 就是
    `childNodes.isEmpty ? SizedBox() : childNodes.first.toWidget()`，官方 `button.md` 也明写
    "The first child is used as the primary content"，所以「图标 + 文字」两个并列子节点时文字被
    整段丢弃 —— 无报错、无日志。落地做法：**内容包成恰好一个子元素**，包装组件里别开放插槽、
    把文字与图标做成 prop，让调用方无法违反。文字那层要 `white-space: nowrap`，否则按钮的固有
    宽度会按 CJK「一字一行」量出来（这条与第 9 条同源）。
    **裸文本子节点是可以用的**（upstream `button.md` 的快速上手示例就是
    `<FlutterCupertinoButton>Tap me</...>`）；包一层元素只是本项目的统一写法 —— 反正有图标时
    本来就必须包。
    > 📌 **一段值得留着的错误归因。** 首轮观察到三个按钮**全是空盒子**，包括第一个子节点是
    > `<flutter-cupertino-icon>`（元素，不是文本）的那两个。我当时把它归因成「裸文本画不出来」，
    > 并据此写进了四份文档。真相是那次复测**看的是旧 bundle**（第 14 条）—— 重启客户端后按钮
    > 正常画出图标 + 文字，**空盒子从来就只是旧页面的样子**。教训不是「要更仔细读源码」（我为此
    > 读了 `RenderWidgetElementChild.performLayout`、`Align(widthFactor:1)` 一整圈，全是白费），
    > 而是**在解释一个现象之前，先确认被测的页面就是被改的代码**。这一步花 30 秒。
12. **cupertino input 的装饰来自 CSS，于是 CSS 上写 `border` 会画出两个框。**
    `CupertinoTextField` 直接把 `renderStyle.decoration` 当自己的 `decoration`
    （`input.dart:301`），而 WebF 自己的 render box **也**画同一份 —— 外框在 border box 上，
    widget 那份落在 content box 上被 `padding` 内缩一圈。**这类元素在 CSS 里只写尺寸**：
    背景/边框/圆角/阴影都不写时 `renderStyle.decoration` 返回 null
    （`css/box.dart:46`），widget 回落到自带的 `systemGrey6` + 8px 圆角并跟随
    CupertinoTheme 亮/暗。`font-size` / `color` 对它同样无效（widget 不读 renderStyle 的
    文本样式）。
13. **HTML 回落分支同样要真机看一遍。** 本轮 M3 开关的选中态滑块位置算错了 2px 并溢出轨道，
    原因是滑块的 `top/left` 是相对轨道的 **padding box** 算的，而轨道有 `border: 2px`
    —— 可用区是 48×28 而不是 52×32。这条与 WebF 无关，纯 CSS，但说明「非 WebF 分支不用测」
    是错的假设。
14. **⚠️ 重装插件后不重启客户端 = 你看的还是旧 bundle，而且毫无提示。** 这条会让你把正确的修改
    当成没生效，然后去重写本来是对的代码 —— 本轮就为它烧掉了一整轮返工。
    机制在 `plugin_render_surface_webf.dart:469`：渲染面用
    `WebF.fromControllerName(controllerName: 'plugin:<去掉 query 的 URL>')`，controller 由
    `WebFControllerManager` 按名字**缓存到进程结束**（本项目 `maxAliveInstances: 8`）。
    命中缓存时**不会重新取 bundle**，日志里那句
    `WebF: loading with controller: WebFController#xxxxx (disposed: false, evaluated: true,
    status: PreloadingStatus.done)` 就是它 —— `evaluated: true` 的意思是「这一页的 JS 早就跑完
    了，我只是把它重新挂上来」。退出页面再进、切 Tab、甚至换主题都不会重取（缓存键刻意去掉了
    query，就是为了让切主题不整页重载）。
    **不是 HTTP 缓存**：`networkOptions: enableHttpCache: false`（第 5 条），后端也已经把
    `index.html` 设成 `Cache-Control: no-cache`、静态资源走内容哈希文件名 + `immutable`。
    所以**完全退出应用再启动**就够了，不用清任何缓存目录。
    **怎么在只有一张截图时确认自己看的是旧的**（不用装任何工具，这套取证本轮就是这么定案的）：
    截图是 2x 的，按像素量几何再跟**服务端当前真的在 serve 的那份**对：
    ```bash
    curl -s http://<host>/api/v1/jsplugin/<entry>/ | grep -o 'static/[^"]*'   # 拿到当前哈希文件名
    curl -s http://<host>/api/v1/jsplugin/<entry>/static/css/style.<hash>.css # 再读内容
    ```
    然后用 PIL 逐点取色：`--md-outline` 是 `#79747E` = `(121,116,126)`，量到的边框颜色一模一样
    就说明那条线来自 CSS 的 `border`（而不是 widget 自绘的装饰）；再量它相对外框的内缩量，
    对得上 `padding` + `border` 就能反推出**页面用的是哪一版 CSS**。更省事的判据是找一个
    「只在旧版里可能出现的东西」：本轮是空状态图标画的是 `question_square`，而服务端 bundle 里
    `grep -c question_square` 是 **0** —— 一条就定案，不需要几何推理。
    要不要在宿主侧修（装完插件自动 `WebFControllerManager.removeAndDisposeController(name)`）
    是另一个议题：对**真实用户**同样成立 —— 从插件商店更新完，已经打开过的插件 Tab 仍是旧页面。
    该 API 存在（`launcher/controller_manager.dart:1251`），但改动落在客户端、要另测，本轮未做。
15. **`webf_cupertino_ui` 没有依赖 `cupertino_icons`，于是所有 `<flutter-cupertino-icon>` 画成
    `?` 方框。** 这是那个包的打包遗漏，**每个用它的 App 都会中**，必须在**产品 pubspec.yaml 里
    自己补** `cupertino_icons`（本仓库已补，probe_pubspec.yaml 同步）。
    因果链：`cupertino_icons_map_generated.dart` 的取值全是 `CupertinoIcons.xxx`，而那些常量是
    `IconData(0xf43c, fontFamily: 'CupertinoIcons', fontPackage: 'cupertino_icons')` —— 字体文件
    由 `cupertino_icons` 包自己的 pubspec 声明（`family: CupertinoIcons`，`assets/CupertinoIcons.ttf`，
    257 KB）。而 `webf_cupertino_ui` 的 pubspec 只有 flutter/webf/logger/collection，`flutter:` 段
    是空的，所以不列这条依赖那个 ttf 就不进 assets。
    **症状会把你带到完全错误的方向**：不是「图标不显示」，而是**画出一个 `?` 方框** ——
    那些码点在私用区（0xF000+），缺字体时由系统兜底字体渲染成「未知字符」占位符
    （macOS 是 Apple LastResort，正是圆角框里一个 `?`）。所以第一反应必然是「图标名写错了」，
    而那是错的：`icon.dart` 对查不到的 `type` 返回 `SizedBox.shrink()`，**名字写错是看不见，
    不是问号**。判据一句话：**看得见问号 = 名字对、字体缺；什么都看不见 = 名字错。**
    自查：`grep -A4 '^  cupertino_icons:' pubspec.lock` —— 不在 lock 里就是没打包。
16. **⚠️ `onControllerCreated` 只在「新建 controller」时回调，于是二次进入插件页 100% 走到
    「页面加载超时」。** 这条是真 bug、不是性能问题，且**必然复现**（同一进程第二次进入）。
    机制：`AutoManagedWebFState._getOrCreateController()` 里那句
    `widget.onControllerCreated!(newController)` 位于
    `if (bundle != null && (controller == null || forceLoad))` **之内**。命中进程内缓存
    （第 14 条）时它、`createController`、`onLoad` **一个都不跑**。后果三连：
    ① `onLoadStop` 永远不被调用 → `PluginRenderView` 的 20s 超时定时器必然烧到，把**已经画好的
    页面**整个换成「页面加载失败 · 页面加载超时」（`_errorMessage != null` 时 build 里根本不挂
    渲染面）—— 用户的观感是「用着好好的，过一会突然卡死了」；
    ② `_controller` 恒为 null → 安全区与播放器状态推不下去、返回键问不到页面；
    ③ `onControllerReady` 不回调 → 宿主拿不到渲染面引用；
    ④ **`onJSLog` 也没被赋值 → 页面 console 与 JS 侧诊断输出全部丢失**（2026-08-05 实测补记，
    详见第 22 条）——排错手段本身被这条 bug 打掉，是它最隐蔽的代价。
    修法在 `_adoptPreloadedController()`（`initState` 里同步 `getControllerSync(name)`，
    命中且 `evaluated` 就自己补上桥、delegate、主题、安全区、`onLoadStop`）。
    **不要**改用 `forceLoad: true` 绕开 —— 那等于每次进页面都重取 bundle 重放整页，
    缓存意义全无、页面内 JS 状态归零。注意 `onControllerCreated` 里那句「被淘汰后重建时兜住
    引用与桥」的原注释**建立在一个错误假设上**（以为它总会回调），两条路径现在要一起改。
17. **`<flutter-cupertino-input>` 的 `blur` 事件没有去重，「blur == 用户改完了」这个前提不成立。**
    `input.dart` 的 `initState` 里就是
    `_focusNode!.addListener(() { hasFocus ? dispatch('focus') : dispatch('blur') })` ——
    **没有记住上一次的焦点态**，FocusNode 只要在未聚焦状态下发出任何一次通知就再派发一个
    `blur`。实测表现：日志里十几条内容**完全相同**的 `POST /api/settings`，界面同时发卡。
    所以「文本框改完即存」这类语义**不能只靠 blur**，要在业务侧加「值真的变了才提交」的守卫
    （downloader 的做法：`store.js` 里按提交后的语义算指纹，与已保存的一致就直接
    `Promise.resolve(null)`）。指纹必须用**和提交体同一个规范化函数**，否则会出现
    「指纹按 `-4` 算、提交的是 `0`」从而永远判定为「有改动」的死循环。
18. **`WebF` 的 `build()` 里现造 future，所以每次祖先 rebuild 都会多跑一次异步 controller 查找
    并刷一行日志。** `AutoManagedWebFState.build()` 是
    `FutureBuilder(future: _getOrCreateController())` —— future 在 build 里现造。而插件渲染面会随
    **任何祖先 rebuild** 一起重建（播放中迷你播放器的进度更新就够了），于是日志里
    `WebF: loading with controller: ...` 会刷屏。解法是让渲染面**返回同一个 widget 实例**
    （`_webfChild ??= WebF.fromControllerName(...)`，`url` 变时置空），`Element.updateChild` 会
    直接短路整棵子树。
    > **别顺着它推出「子树被反复拆掉重建」——那是错的**，我推过一次。`FutureBuilder` 换 future
    > 时用的是 `_snapshot.inState(waiting)`，**data 会被保留**，而 `webf.dart:357` 的判断是
    > `connectionState == waiting && snapshot.data == null` —— data 非空就不会回落到
    > `loadingWidget`，已挂载的 WebF 子树不会被拆。这里省掉的是重复的异步查找与日志噪音，
    > 不是修「拆树」。同理，外层那个 `FutureBuilder(future: PluginRenderFonts.ensureLoaded())`
    > 也是安全的：`ensureLoaded()` 返回 `_pending ??= _load()`，**每次是同一个 Future 实例**，
    > `didUpdateWidget` 里 `oldWidget.future != widget.future` 不成立，压根不会重订阅。
19. **⛔ `<select>` 在 WebF 下不能用来做双向绑定 —— 换 `<flutter-cupertino-action-sheet>`。**
    这条把 §3.3「已排除的伪阻塞」里那句「`select` 在 WebF 下可用」**收窄**：它只是能画出来、
    能弹出菜单，**选中值传不回 JS**。症状是「歌单/艺术家/专辑筛选都不重筛列表，
    只有关键字搜索正常」，不报错、不打日志。
    - **根因（读源码确证）**：`webf/lib/src/html/form/select.dart` 的
      `HTMLSelectElement.initializeDynamicProperties` 只暴露 `value` / `selectedIndex` /
      `disabled` / `multiple` / `required`，**没有 `options`**，也没有 `selectedOptions`
      （Dart 侧 `grep "\['options'\]" lib/` 零命中；`macos/libwebf.dylib` 的绑定表里
      `HTMLSelectElement` 附近也只有类名注册，没有属性名）。于是 Vue 的 `vModelSelect`
      指令直接抛 TypeError —— 它的 change 监听器是
      `Array.prototype.filter.call(el.options, o => o.selected)`，`mounted`/`updated` 调的
      `setSelected()` 里是 `el.options.length`。**任何框架**的 `<select>` 双向绑定都会踩这个。
    - **绕开 v-model 也不够。** 我改成显式 `@change` 读 `el.value` + `flush:'post'` 的
      `watchEffect` 命令式回写，**真机实测仍然不通**。剩下的可疑点全在 Dart 侧且从 JS
      观测不到：`_openOptionsMenu()` 用 Material `showMenu`；而元素一旦挂上 JS 监听，
      `Element.hasEvent` 翻真 → `requestWidgetToRebuild(AddEventUpdateReason())` →
      `RenderEventListener.enableEventCapture()` 起自己的 `GestureDispatcher`，与 select
      自带的 `GestureDetector.onTap` 同处一个手势竞技场。**没有继续深挖**——换实现更便宜。
    - **官方的 `<flutter-cupertino-action-sheet>` 是 webf-ui 给「从 N 个里选一个」的正解**
      （`webf_cupertino_ui.dart:60` 已注册；31 个元素里**没有**任意选项的 picker，
      `flutter-cupertino-picker` 是注释掉的，只有 `date-picker` 注册了）。契约在
      `action_sheet.dart`，**不是**那份 React 口径的 `.md`：
      * 命令式 `el.show(config)`，config 可以是对象**或 JSON 字符串**（`args[0] is String`
        → `jsonDecode`）。**传字符串**，别依赖 JS 对象跨桥 marshal 成 Dart Map。
      * 选中派发 `CustomEvent('select', detail: {text, event, isDefault, isDestructive, index})`。
        `index` 是在 `actions` 里的下标，**`cancelButton` 不带 index**；点遮罩关闭则不派发。
      * 内部 `showCupertinoModalPopup(useRootNavigator: true)`，不依赖局部 Overlay。
      * 宿主元素自己 build 的是 `SizedBox.shrink()`，**必须常驻且不要用 `display:none` 藏**。
      **但 downloader 最后没用它**，理由是它有一个**从 JS 侧无法观测的静默失败模式**：
      `show()` 的实现是 `state?._showActionSheetImpl(args)` —— state 还没建立时直接
      no-op，不抛异常、不打日志，于是「点了什么都不发生」与「正常工作」在代码里无法区分，
      只能靠人肉试。同类不确定还有：方法能否被 `typeof` 探到（属性与方法在 Dart 侧是
      `_getBindingObjectProperty` 与 `_getBindingObjectMethodType` **两条独立查找路径**），
      以及 `CustomEvent.detail` 过桥后是对象还是字符串（Dart 侧 Map 走 `tagJson`，
      但同一通道在 cupertino input 的 `input` 事件上传的是裸字符串）。
      这些本来一次容器探针就能全验完，但**探针在 Apple Silicon 上跑不了**（见下条），
      于是选择了确定性优先的自绘方案。
    - **downloader 现在的实现**：cupertino button 触发 + **常规流里的内联面板**（普通 `div` 行，
      `v-if` 展开）。只用「已在同页跑通」的原语：button 的 `click`、块盒布局、
      普通元素的 `click`（DOM click 由 WebF 唯一那个全局 tap recognizer 派发，见第 420 行附近）。
      刻意**不用**浮层（要赌层叠与命中测试 —— 面板得盖在歌曲列表那个 Flutter widget 上）、
      **不嵌** `<webf-list-view>`（要赌 tap 穿过 Flutter ListView 的手势竞技场）、
      **不靠 overflow 滚动**。值只在自己的 JS 里流动，完全不碰 WebF 元素的属性读写。
    - **判据陷阱：「下拉显示更新了」不是「数据通了」的证据。** WebF 的 select 是
      `WidgetElement`，`_openOptionsMenu()` 先 `widgetElement.selectedIndex = result` 改自己的
      内部态、再派发 `change`，显示的标签由 Flutter 侧维护，与 JS 收不收到值完全无关。
    - 反过来说，`v-model` 用在**组件**上是安全的（编译成 `:modelValue` + `@update:modelValue`，
      纯 Vue 逻辑，不碰原生指令）；`<input type=checkbox>` 的 `vModelCheckbox` 也安全
      （只依赖 `el.checked` / `el.value`，WebF 两个都有）。要提防的只有原生 `<select>`。
    > **过程复盘（返工三轮）**：第一轮只查到「`options` 缺失 → v-model 挂了」就收工，把
    > 「改成显式事件」当成了修复 —— 诊断本身没错，但**它不是唯一的断点**，而我只验证了
    > 「编译产物里有新代码」就下了「修好了」的结论。第二轮换成官方 action sheet，仍然
    > 只在产物里 grep 到新标记就交付，而那个组件恰好有个静默 no-op 的失败模式。
    > 两条教训：**① 一个静默失效的元素上，找到一个足以解释症状的断点，不等于找到了全部断点**；
    > **② 在无法运行期验证的环境里，要按「失败模式能不能从代码侧观测」来选实现**——
    > 官方组件不等于确定可用，能打日志、能定位断点的自绘方案反而更省往返。

20. **⚠️ `scripts/webf-verify` 在 Apple Silicon 上跑出来的结论不可信（会静默跑旧二进制）。**
    `webf` 包只提供 **x86-64** 的 `libwebf.so`（`file linux/libwebf.so` 确认；这也是
    §F「Linux 仅 x86-64」的同一个根因）。在 arm64 宿主上 Docker 跑 linux/arm64 镜像，
    容器里 `flutter build linux` 产出 `build/linux/**arm64**/release/bundle`，那份产物
    **加载不到原生库**，运行期报
    `Failed to load dynamic library 'libwebf.so'` → `_contextId has not been initialized`
    → 注入的诊断脚本一行都不输出。
    而 `entrypoint.sh` 的 `BUNDLE=` 硬编码 `build/linux/**x64**/release/bundle`，
    于是它**启动的是镜像里烘进去的那份旧 x64 二进制**，而不是刚构建的产物 ——
    脚本照常打印「构建探针」「运行探针」「截图已落」，`run.sh` 退出码 0，截图也有内容。
    也就是说：**你改的探针代码根本没跑，但一切看起来都成功了。**
    判据：`out/build.log` 里那行 `✓ Built build/linux/<arch>/release/bundle/...`
    的 arch 与 `entrypoint.sh` 的 `BUNDLE=` 不一致 → 本次结论作废。
    要在 arm64 机器上用它，得 `--platform linux/amd64` 走 QEMU 模拟（Flutter 构建 +
    软渲染 GL，慢到不实用），或者换 x86-64 宿主。
    - 顺带修掉一个让这条路径**一次都跑不起来**的 bug：`run.sh` 里
      `echo "…：$DIAGNOSE_JS（…"` 的裸变量引用后面紧跟全角「（」，bash 5.3 会把那几个
      高位字节当成标识符的一部分，变量名成了 `DIAGNOSE_JS（`，`set -u` 下直接
      unbound variable 退出，而报错里的变量名是乱码。已改成 `${DIAGNOSE_JS}`。

21. **⚠️ `flex-wrap: wrap` 的容器里，子项不能靠「内容固有宽度」定宽 —— base size 会被测成
    容器宽度，于是每个子项独占一行并铺满。** 2026-08-05 在 downloader 上复现并修掉
    （症状：工具栏「全选 / 刷新 / 下载选中」三个按钮各占一整行、筛选栏四项各占一整行，
    而 webview 分支完全正常）。
    - **机理（读 `webf-0.24.27` 源码确证，逐环闭合）**：
      ① `RenderFlexLayout._computeRunMetrics`（`rendering/flex.dart:5199+`）算 flex base size
      用的是**「以放松后的约束实际 layout 一遍子项」得到的宽度**（局部变量 `intrinsicMain`），
      不是真正的 max-content 贡献；
      ② `_getIntrinsicConstraints`（`flex.dart:2461-2482`）在 row 方向 + `width:auto` + 非 replaced
      时把子项 `maxWidth` 放松成 `double.infinity`。**WidgetElement 不是 replaced**——
      `isSelfRenderReplaced()` 判的是 `is RenderReplaced`（`css/render_style.dart:1131`），
      而 cupertino button 是 `RenderWidget`；
      ③ `RenderWidget._layoutChild`（`rendering/widget.dart:166-188`）把那个无界宽度
      **clamp 回 viewport 宽度**（注释明写是为了防 hosted Flutter subtree 拿无界约束崩溃，
      `allowsInfiniteWidth` 默认 false，见 `widget/widget_element.dart:119`）；
      ④ 按钮内容层 `.dl-btn-inner` 当时是 `display:flex` 即 **block-level** → `width:auto`
      → **fill-available** = 上一步的 viewport 宽度（`css/render_style.dart:3425-3432` 显示只有
      `inline-block` / `inline-flex` / `inline-grid` / `inline` 才在 `width:auto` 时 shrink-to-fit）；
      ⑤ `RenderWidget` 完全 shrink-wrap 到唯一子节点（`size = getBoxSize(childSize)`，
      `rendering/widget.dart:268`）→ 按钮宽度 = viewport 宽度；
      ⑥ 回到 flex：换行判定 `isExceedFlexLineLimit`（`flex.dart:5476-5480`）从第二个子项起恒真。
    - **WebF 自己知道这个 bug 并打了补丁，但补丁覆盖不到这两类子项。**
      `flex.dart:5331-5369` 的注释直接引 CSS Flexbox §9.2：*"Our intrinsic pass can mistakenly
      inherit a container-bounded width for block-level items … causing the base size to equal the
      container width."* 可它的入口条件是 `if (isHorizontal && child is RenderFlowLayout)` ——
      **`RenderWidget`（WidgetElement）与 `RenderFlexLayout`（嵌套 flex 容器）都不在内**。
    - **两条落地规则**（downloader `style.css` 的约束 ⑦ 已固化）：
      · WidgetElement 内部的内容层用 **`inline-flex`**（shrink-to-fit）。blockify 不会打掉它：
        `css/display.dart:86-140` 只在**父**是 flex/grid 容器时 blockify，而 WidgetElement 默认
        `display: block`（`widget/widget_element.dart:17-19`）。
      · 嵌套 flex 容器给**显式 `width`**。必须是 `width` —— 放松条件只看 `width.isAuto`，
        写 `flex-basis` 无效；改 `inline-flex` 也无效（它自己是 flex item，会被 blockify 回 flex）。
    - **⛔ 不要给 WidgetElement 加 `max-width` 来夹住 base size**：`hasExplicitInlineWidth`
      （`rendering/widget.dart:103-104`）会让 `_layoutChild` 改走 inline-block 分支（`:157-165`），
      把 intrinsic pass 下那个 **∞ 直接透给 hosted Flutter subtree** → 撞 §3.2 第 11 条的
      `Infinity or NaN toInt`。
    - **⛔ 也不要只把 `wrap` 改成 `nowrap`**：换行是没了，但 base size 仍是 viewport 宽度，
      单行内按 flex-shrink 均分 → 按钮变成等宽铺开，仍然不对。
    - **为什么列表行 `.dl-row` 从来没出过这个问题**（对照组，也是修法的范本）：它 `flex-wrap`
      取默认 `nowrap`（压根不进换行判定），且每列都是**显式 flex 比例**
      （`.dl-col-title{flex:2 1 0}`、`.dl-col-cb{flex:0 0 48px}`），不依赖 intrinsic 测量。
      出问题的两处恰好是全项目唯一的 **wrap + auto basis** 组合。
    - 修复后实测（截图量像素，换算成 CSS px）：三个筛选项 170/170/169（设定 `width:170px`）、
      搜索项 244（`200px` + `flex:1` 吃余量）、三个按钮 51/76/107（各自内容宽度）。

22. **⚠️ 插件页的 `console.log` 转发（`[plugin][console]`）经常整体失效 —— 不要把「日志里没有」
    当成「代码没跑」。** 这是第 16 条的一个**未被记下的后果**：`onJSLog` 是在 `createController`
    里赋值的（`plugin_render_surface_webf.dart:531`），而 controller 命中 `WebFControllerManager`
    的预加载 / 进程内缓存时 `createController` **一个都不跑**。
    2026-08-05 实测：整份客户端日志里 `[plugin][console]` **零命中**，而 `engine.js` 里那条
    `console.log` 是无条件执行的、页面也确实画出来了（`WebF: start for loading …` 有 17 条）。
    - 于是**第 16 条列的三连后果要加第 ④ 条**：页面 console 与 JS 侧的一切自定义诊断输出全部丢失
      —— 这等于把 §0 说的「排错闭环已建立」在**最常见的那条路径上**打掉了。
    - 落地影响：布局类问题改用**截图量像素**取证（第 14 条记了那套方法，本轮第 21 条就是这么定案的）。
      想靠 `console.log` 打诊断前，先确认日志里真能看到它 —— 否则会像本轮一样白等一轮。
    - 日志文件位置（macOS，沙盒容器内）：
      `~/Library/Containers/com.songloft.songloftFlutter/Data/Library/Application Support/com.songloft.songloftFlutter/logs/songloft_<date>.log`
      （`onJSLog` → `debugPrint` → `FileLogger`，`lib/main.dart:98-105`；`debugPrint` 同时进 stdout，
      所以直接跑 Debug 二进制并重定向输出也能拿到同一份）。

23. **cupertino button 的 `variant` 配色对不上宿主主题，而且 `filled` 的底色压根改不了 ——
    统一走 `plain` 分支、外观全交给 CSS。** 2026-08-05 在 downloader 上落地。
    - `button.dart` 三个分支里只有 `default`（plain）与 `tinted` 会把 CSS 的 `background-color`
      当自己的底色（`color: backgroundColor`）；**`CupertinoButton.filled` 的构造器不接受
      color**，固定用 `CupertinoTheme.primaryColor` → 画出来是 iOS 配色，与 M3 色板无关。
    - 圆角同理：filled 分支在没有 CSS `border-radius` 时是 `BorderRadius.zero`（**直角**），
      plain / tinted 是固定 `circular(8)` —— 都不是 M3 的胶囊。
    - 落地写法：模板里**不传** `variant`（默认即 plain），底色 / 圆角 / 边框 / 前景色全部由 CSS 给，
      与宿主 `common.css` 的 `.btn-filled` / `.btn-outlined` 语义一一对齐。前景色本来就只能由
      CSS 给（第 11 条：文字与图标读 renderStyle，widget 侧的 DefaultTextStyle 到不了）。
    - **`disabled` 的变灰也只能由 CSS 表达**：plain 分支的 `disabledColor` 是
      `Colors.transparent`，挡不住 WebF 按 CSS 画的那层底色，于是 disabled 按钮看起来跟可用的
      一模一样。要用**class** 而不是 `[disabled]` 属性选择器 —— disabled 是命令式赋的 JS 属性
      （第 3 条），不反映到 HTML 属性上。
    - 顺带修了宿主侧的一个对称缺陷：`common.css` **没有任何 `:disabled` 规则**，所以 webview
      分支的 disabled 按钮一直是「看起来完全可点」的实心主色。downloader 在自己的 CSS 里补了
      `.btn:disabled`，**没动 common.css**（那会影响所有插件，属独立议题）。

    **两条 CSS 属性在 WidgetElement 上语义不同，都是实测校准出来的**：
    - **`padding` 会被应用两次**（写目标值的一半）：WebF 的 `RenderWidget` 按 CSS padding 内缩
      content box（`_setChildrenOffset` 里 `borderLeftWidth + paddingLeftWidth`），而
      `button.dart` 在 `hasPadding` 为真时又把**同一份** `renderStyle.padding` 交给
      CupertinoButton。M3 要左右 24px → CSS 写 `0 12px`，实测按钮宽 76.7px 对目标 76px。
    - **高度必须用 `height`，`min-height` 不管用**（写 `min-height: 40px` 实测量出 33px）：
      `min-height` 只经 `hasMinHeight` 喂给 `minimumSize`，那是 Flutter widget **自绘**的下限；
      而 WebF 盒子的高度是 `size = getBoxSize(childSize)`（第 21 条同一处，
      `rendering/widget.dart:268`）—— 跟着子节点内容走，**CSS 边框也画在这个盒子上**。
      写 `height` 才会经 `renderStyle.height.isNotAuto` 把子树 tighten 到指定高度。
      改完实测边框高度 = 40px（按 `--md-primary` 的 RGB 扫像素定位边框行，不靠目视）。

24. **下拉浮层（`position: absolute`）在 WebF 下可行 —— 能盖住 `<webf-list-view>`。**
    这推翻了 downloader 原先「刻意不用浮层」的设计决策（理由曾是「浮层要赌 WebF 的层叠与命中
    测试，面板得盖在歌曲列表那个 Flutter widget 上」）。改的动机是内联块盒的代价更难接受：
    一展开就把工具栏和整张列表往下顶。
    - **已实测**：`.dl-select-wrap{position:relative}` + 面板 `absolute` + `z-index` 后，面板正常
      悬浮，**盖住了表头、歌曲行与 `<webf-list-view>`**（截图确证），下方内容不再被顶开。
      即「WidgetElement 的绘制不遵守 CSS 层叠顺序」这个担心**不成立**。
    - ⚠️ **未实测**：点面板里的选项能否选中（命中测试是否被 Flutter ListView 的手势竞技场抢走）。
      判据与退路写在 `style.css` 的 `.dl-select-panel` 注释里。
    - **踩到一个连带问题**：浮层脱离常规流后，**祖先链上任何 `overflow: hidden` 都会把它整段
      切掉**。downloader 的 `.dl-card` 原有 `overflow: hidden`（内联时代无害，因为面板会把卡片
      撑高），改浮层后艺术家一多面板底部就消失 → 已去掉，并留空规则 + 注释防止被加回来。
      面板另加 `max-height` + `overflow-y` 作为**降级保护**（不是依赖：WebF 下 CSS overflow 滚动
      仍未验证，但至少面板高度有界、不会盖满整页）。

25. **⚠️ 受控文本框被外部改写会让整页白屏（debug 构建）—— `Text layout not available` +
    `!_debugDuringDeviceUpdate` 无限刷屏。** 2026-08-05 在 downloader 上实测到，
    表现是「页面画得好好的，数据一加载完就整页白掉」，且**日志里没有任何插件侧的错误**。
    - **崩溃链（栈已确证）**：受控输入的 `val` 变化 → Flutter 的
      `_Editable.updateRenderObject` → `RenderEditable.text=` → `TextPainter.markNeedsLayout`
      → 同一帧 `MouseTracker.updateAllDevices` 做 hit test →
      `RenderWidget.hitTestChildren`（`webf/src/rendering/widget.dart:907`）→
      `RenderEventListener.hitTest` → `_RenderBaselineAlignedStack.hitTestChildren`
      （`flutter/src/cupertino/text_field.dart:1947`）→ `RenderEditable.hitTestChildren`
      → `TextPainter.getClosestGlyphForOffset` → 撞 `assert(_debugAssertTextLayoutIsValid)`。
      随后 `mouse_tracker.dart:199` 的 `!_debugDuringDeviceUpdate` 每帧刷屏，帧循环烂掉。
    - **触发条件是「那一帧鼠标正停在插件页上」** —— 所以它极容易被误判成偶发/与改动无关：
      本轮前几次验证都没撞到，只因为截图时鼠标不在客户端窗口里（MouseTracker 无设备位置就
      不做 hit test）。**用截图脚本验证时永远撞不到这条。**
    - **换 HTML `<input>` 绕不开**：WebF 自己的 `<input>` 也是 WidgetElement + Flutter
      `TextField`（`webf/lib/src/html/form/base_input.dart:615`），底下是同一个 `RenderEditable`。
    - **debug-only**：两个抛出点都是 `assert`（`text_painter.dart:1688` 的
      `assert(_debugAssertTextLayoutIsValid)`）。release 构建会整条剥掉 → 不会烂成白屏。
      但**不能因此说生产无事**：紧接的下一行是 `_layoutCache!`，release 下若真为 null 会抛
      NPE，只是不会连锁刷屏。
    - **规避（downloader 已落地）**：让受控输入**等值到齐后再挂载**（`v-if="settingsLoaded"`），
      首次赋值走 mount 而不是 update，消掉最主要那条路径。**无法根除** —— 任何「外部纠正
      显示值」的场景（downloader 里是 `-4 → 0` 的间隔规范化）仍会走 update。
    - 这是 WebF 侧的时序问题（hit test 打到了 layout 未就绪的 hosted Flutter 子树），
      已写成 `upstream-issues.md` 第 9 条。

26. ~~**同一个 WebF controller 被两个渲染面持有 = 其中一个整页白屏。**~~
    **⛔ 本条作为「白屏根因」已被推翻，见第 27 条。** 保留原文是因为它记录的
    *机制* 仍然成立（`attachToFlutter` 无重入防护是真的），而且它引出的 redirect
    仍然值得保留 —— 只是**理由变了**，且它不是白屏的原因。
    - **仍然成立的部分**：`WebFController.attachToFlutter`（webf `controller.dart:1518`）
      确实**没有重入防护** —— 直接覆写 `_ownerFlutterView`、重新
      `view.attachToFlutter(context)`、`pushNewBuildContext(...)`。
      两个渲染面同时持有一个 controller 是真的危险。
    - **被推翻的部分**：~~「Tab 页靠 shell 层 Offstage 保活、永不释放，所以只要插件在
      Tab 里，独立路由就必然与它冲突」~~ —— **桌面端根本不保活**。
      `shared/layouts/shell_layout.dart` 里 Offstage 保活**只对 Web 与移动端**生效；
      Windows/macOS/Linux 走的是「只渲染当前激活的插件 tab，**切走即销毁**」
      （为规避 #246 的 WebView2 残留灰块）。所以在 macOS 上两个渲染面**从不共存**，
      而是**先后**挂载 —— 探针（下一条）实测一次都没触发，正是这个事实的证据。
    - **redirect 保留，理由改为**：让同一插件只有一个入口，避免「独立页 push/pop 一次
      = Tab 页渲染面被销毁再重建」这种白丢页面状态的抖动。
    - **`_liveSurfacesByController` 从诊断探针升级为归属表**：现在它决定「谁有权在
      dispose 时销毁缓存的 controller」（第 27 条的修法需要它来处理同帧交接）。
    - **教训**：这是我在同一个白屏上第**三**次归因错误（hot reload → 文本框断言 →
      controller 抢占）。第三次错在**把一个真实存在的危险机制当成了本案的成因**，
      而没有先去核对「两个渲染面真的同时活着吗」——探针本来就是为回答这个问题加的，
      我却在它给出否定答案之前就把结论写进了文档。**判据要先跑，再下结论。**

27. **⚠️⚠️ 渲染面重新挂载 + controller 命中缓存 = 静默白屏。这才是「过一会白屏」的根因。**
    2026-08-05 定案（songloft-org/songloft#341）。与第 25 条（文本框断言）**彼此独立、
    都真实存在**；本次那一轮的白屏是本条，日志里**零异常**可证。
    - **因果链**：桌面端插件 Tab **切走即销毁**（上一条）→ 每次离开再回来都是一次完整的
      `dispose` + 重新挂载 → 重新挂载时 controller 命中进程内缓存（`evaluated: true`）→
      `createController` / `onLoad` / `onLoadError` / `onJSLog` **一个都不跑**（第 16、22 条）
      → `_adoptPreloadedController()` 又**无条件**上报 `onLoadStop`
      → 于是这条路径上**任何**失败都必然表现为「整页白屏 + 日志一个字都没有」。
    - **日志判据（不用猜，直接看）**：白屏总是紧随**第二条**
      `WebF: start for loading ...`，且该行前面那条 `WebF: loading with controller: ...`
      带 `evaluated: true`。只 mount 过一次（`evaluated: false`）的会话从不白屏 ——
      此前所有「页面正常」的截图都来自那种会话。
      两条打印分别出自 webf `widget/webf.dart:384` 与 `:813`，后者在 element 的
      `mount()` 里，所以**一条 `start for loading` = 一次重新 mount**，可直接用来数挂载次数。
    - **修法（已落地）**：渲染面 `dispose()` 时**连带销毁缓存里的 controller**
      （`_dropCachedController` → `WebFControllerManager.removeAndDisposeController`），
      即「不跨渲染面生命周期复用 controller」。这样每次挂载都退回**正常路径**：
      有 loading、有 20s 超时、有错误 UI、有 console 转发。
      **代价**：页面 JS 状态（筛选项、滚动位置）归零、bundle 要重取 —— 与 webview 分支在
      桌面端的行为**一致**（原生 WebView 被销毁后同样重载），所以不会再出现
      「两条渲染路径行为不对称」这种更难排查的情形。
    - **同帧交接的坑（必须防）**：同一 URL 的新旧渲染面在**同一帧内**交接时，新面的
      `initState` **早于**旧面的 `dispose` —— Flutter 先在 build 阶段 inflate 新子树，
      到帧末 `BuildOwner.finalizeTree()` 才拆 inactive 元素。若旧面无条件销毁 controller，
      销毁掉的正是新面刚认领的那个。故用 `_liveSurfacesByController` 判「登记的还是自己吗」，
      新面 `initState` 会改写归属，旧面据此让权。
    - **顺带修掉两条长期拖慢排查的坑**：重装插件后不必再完全退出客户端才能看到新 bundle
      （第 14 条）；`[plugin][console]` 转发不再因命中缓存而静默失效（第 22 条）。
    - **不要**改用 `forceLoad: true` 达到同样效果：那只是每次重新取 bundle，缓存里那个
      controller 仍然留着不放，内存与 `maxAliveInstances` 配额照旧被占。
    - **上游侧刻意没写成草稿**：我们绕开了这条路径，但**从未隔离出它到底哪一步坏了**
      （`detachFromFlutter` 把 `viewport` / `_isFrameBindingAttached` /
      `_frameFlushLoopEnabled` 全置否，而 `attachToFlutter` 并不逐一恢复，这只是
      **候选**，未证实）。没有最小复现和确定机制就去提 issue，只会得到
      「你的复现里还有别的问题」。真要提，先做一个「加载同一 URL、挂载→卸载→再挂载」
      的最小 Flutter 例子。第 28 条那个 `onLoad` 问题相反 —— 调用链已确证，
      已写成 `upstream-issues.md` 第 10 条。

28. **⚠️ `onLoad` 在「同一进程内第二次加载」时不来 —— 不能把它当唯一的就绪信号。**
    2026-08-05 实测（第 27 条的修复落地后暴露出来的下一层）。
    - **现象**：丢掉缓存后第二次挂载会**新建** controller、bundle 完整加载、JS 真的执行
      （页面自己那行 `[downloader] engine: ...` 打出来了）、四个 API 全部 200 ——
      然后页面被 `PluginRenderView` 的 20s 超时定时器换成「页面加载失败·页面加载超时」。
      即**页面是好的，只是没人报告「好了」**。
    - **机理**：`onLoad` 只由 `dispatchWindowLoadEvent()` 调，而后者只由
      `checkCompleted()`（webf `controller.dart:1718`）调。`checkCompleted()` 有**四道
      early-return**：`document.parsing`、`isDelayingDOMContentLoadedEvent`、
      `hasPendingRequest`、`isDelayingLoadEvent`。任一条命中就直接 return，
      而**没有任何东西保证它之后会被再调一次**。
    - **为什么第二次才犯**：这是个竞态。第二次挂载所有资源都已在本机热着 ——
      日志里 bundle `266ms → 55ms`、四个 API `279/277/277/52ms → 6/6/6/10ms`。
      时序一变，`checkCompleted()` 就撞在了不同的 guard 上。
    - **修法（已落地）**：改用 **`onBuildSuccess`** 作主信号（`onLoad` 降为次要信号，
      两条都指向幂等的 `_markPageReady`，谁先到算谁）。`onBuildSuccess` 在
      `buildRootView()` 真把根视图建出来之后 post-frame 回调，且**只在成功分支**调
      （webf `widget/webf.dart:673` / `:723`，所有 error 分支都不调）——
      语义正是我们要的「页面画出来了」。
    - **幂等守卫不是可选的**：`onBuildSuccess` 每次 `buildRootView` 都回调，而
      `onLoadStop` 会 setState 祖先 → 重建 → WebF 重建 → 又一次 `buildRootView`
      → 又一次回调。**不守卫就是无限重建循环。**
    - **一般化的教训**：这一整轮（第 16、22、27、28 条）都是同一个形状 ——
      **把「宿主自己假定的成功」当成了「引擎报告的成功」**。第 16 条补调 `onLoadStop`、
      第 27 条无条件上报成功，都是在替引擎打包票。正确做法是找一个**引擎真的做完那件事
      才会发**的信号。

29. **⚠️⚠️ flex 容器里再套 flex 容器 = 整个子树一个像素都不画。无异常、无日志、结构全对。**
    2026-08-05 实测并已修复（downloader）。**这是第 21 条那个 base size 缺陷的另一副面孔，
    但后果严重得多** —— 第 21 条只是排版难看，这条是内容整片消失。
    - **机制（Flutter 侧，确证）**：`RenderObject._paintWithContext` 开头就是
      `if (_needsLayout) return;`，源码注释写明「说明我们在 layout 阶段被跳过了，
      因此不需要绘制」。**停在 `needsLayout` 的 render object 被静默跳过绘制。**
      独立佐证：日志里 `editable.dart:2018` 的 `assert(!debugNeedsLayout)`
      （`RenderEditable.handleEvent` 里，点击命中了 layout 未完成的文本框）。
    - **触发条件（相关性 100%）**：容器的**子节点自己也是 `display:flex`**。
      `.dl-container`(flex col)→`.card`、`.dl-card-body`(flex col)→`.dl-field`(flex col)、
      `.dl-filter-bar`(flex row)→`.dl-filter-item`(flex col) **全不绘制**；
      `.dl-toolbar`(flex row)→按钮/span、SongList→`<webf-list-view>` **全正常**。
      根因同第 21 条：WebF 用「以放松约束试排一遍」测 base size，而它自己那个 §9.2 补丁的
      入口条件是 `child is RenderFlowLayout`，**嵌套 flex 容器不在覆盖范围内**。
    - **修法**：**竖向堆叠一律用块流**（`display:block` + `margin`），不用
      `flex-direction:column` + `gap`。视觉等价，但块流完全不碰那套测量。
      横向 flex 行保留 —— 问题从来不在它们自己，在**子节点**也是 flex 容器。
      写成了 `style.css` 的约束 ⑧。
    - **为什么第 21 条的修复只对了一半**：那轮我给 `.dl-filter-item` 补 `width` 修好了
      **主轴** base size（排版正常了、像素也量过），就此结案。但同一个试排层还会把子树留在
      `needsLayout`，而那个症状在当时的窗口尺寸/时序下没暴露。**「排版对了」不等于
      「这条缺陷绕开了」。**
    - **取证方法（本轮唯一有效的那条）**：`getBoundingClientRect()` 与 DOM 计数**全部正常**，
      靠结构完全查不出来。定案靠**按区域数非背景像素**——FilterBar 应在的区域内只有 184 个
      非背景像素（全是它自己那条 `border-bottom`），修复后同区域涨到数千。
      这正是本目录反复强调的那条判据：**「盒子有尺寸」≠「图被画出来了」。**

#### 仍然要主动规避的（webf-ui 救不了的）

`<base href>` 不被采纳（第 2 条）、`display: none` 仍占位、内联 `style.display=''` 不可靠、
layout 异步、`resize` 不一定派发、`max()`/`min()` 未实现、无界 flex 触发
`Infinity or NaN toInt`、`btoa` 不是二进制安全、**`flex-wrap: wrap` 下子项 base size 被测成
容器宽度**（第 21 条）、**`[plugin][console]` 转发在缓存命中时静默失效**（第 22 条）。

---

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
   `docs/webf/step4-design.md` §3 Step 1 的完整判据）。`rendering/grid.dart:347-351` 的
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
| `<table>` 元素**根本不存在**（退化成嵌套 `display:block`；且 `display:table/table-row/table-cell` 也救不了 —— `css/display.dart` 的 `CSSDisplay` 枚举**没有任何 table 值**，`resolveDisplay` 的 `default` 返回 `CSSDisplay.inline`，比 block **更糟**） | **实际只有** downloader（`webf`）；radio 虽有表格但未声明 = `webview`，**进不到 WebF** | **Step 4 实施中**。原定的「垫片改写成 `<webf-table>`」**已证伪**，改走 CSS Grid，方案见 `docs/webf/step4-design.md` |
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
`textarea` / `input[type=time]` 全部可用。

> ⚠️ **`select` 从这份名单里撤回了。** 它只是「能画出来、能弹菜单」，**选中值传不回 JS**
> （没有 `options` 属性，任何框架的双向绑定都挂）。要做下拉请用
> `<flutter-cupertino-action-sheet>`，详见上面第 19 条。

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
> 已被证伪。看到那句话时以本节与 `docs/webf/step4-design.md` 为准。**

**证伪理由**：`<webf-table>` 的 `build` 只读**直接 `childNodes`**（`html/table.dart:188-189` 的
`firstWhereOrNull` / `whereType`），`<thead>`/`<tbody>` 不拆就是**一张空表，且不报错不打日志**；
而 downloader 恰好靠 `#tbody` + `innerHTML` 渲染行 —— 保留 `<tbody>` 得空表，拆掉插件 JS 抛
`TypeError`。WebF 又**没有 `MutationObserver`**。更要命的是 `colspan`/`rowspan` 零支持是
**Flutter `Table` widget 的天花板，不是 WebF 没做，上游修也修不了**。

**现在的方案**：CSS Grid（`docs/webf/step4-design.md` 的方案 B'）。WebF 的 Grid 是**已实现**的
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
   从未标 `webf`）。桥形状与垫片形状已定形，见 `docs/webf/step4-design.md` §2.3 / §2.4：
   桥返回 `{name, size, text}`、**主载荷是 UTF-8 解码后的字符串**（不返回 path），
   垫片**必须同时拦 `click` 事件与覆写实例 `click` 方法**（radio 是「隐藏 input + 外部按钮代点」）

`file_picker: ^10.3.10` **早就在 `pubspec.yaml` 里了**，所以用它的原生契约哈希代价是**零**。
⚠️ 但**绝不要 bump 它的版本**：版本号进哈希，而它是 caret 约束，**一次 `pub upgrade` 就会静默改变**。

顺带还有一件小事（`docs/webf/step4-design.md` §1.9）：给 `common.js` 加一个**只警告不改写**的
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

#### ⛔ 缺陷台账：许可步骤炸掉了 player 的第一次 release（2026-08-03 已修）

上面那条挂账项「首次真实 release 后确认版本号没回落」**已经兑现，而且比预想的更糟**：
版本号不是回落成占位符，而是**整个 `Create Release` job 失败**。
player release run `30784825599`：8 个构建 job 全绿，只有 `Create Release` 挂掉，报
`awk: fatal: cannot open file 'pubspec.lock' for reading: No such file or directory`，
位置 `songloft-player/.github/workflows/build-and-release.yml` 的
`Attach GPL-3.0 license text + corresponding source notice`。两层缺陷叠加：

1. **`pubspec.lock` 不入库**（`songloft-player/.gitignore:49`），而 `Create Release` 是**独立
   job 的全新 checkout**、**不跑 `flutter pub get`** → 那个文件在 CI 里**必然不存在**。
   本机能跑通纯粹因为本地有 lock。
2. **兜底那行是死代码**（关键）：紧跟其后的
   `[ -n "$WEBF_VERSION" ] || WEBF_VERSION="see pubspec.lock"` 永远执行不到。
   判据：GitHub Actions `run:` 的默认 shell 是 `bash -e`（失败日志里就写着
   `shell: /usr/bin/bash -e {0}`），而 `VAR=$(失败命令)` 里命令的非零退出码会成为**赋值语句
   本身**的退出码，`-e` 立刻中止整步。本意是「取不到版本号就填占位符」，
   实际是「取不到就炸掉整个 release」。**这是个会反复踩的通用陷阱，不限于许可步骤。**

**修法**（不动 `.gitignore`，那是仓库策略、未获授权）：版本号主来源改成**入库的
`pubspec.yaml`**，`pubspec.lock` 恰好存在时才升级成精确解析版本；所有命令替换加 `|| true`、
文件存在性用 `[ -f ]` 守卫；三条分支都保证有值，末路兜底指向本次 run 的构建日志 URL。
产出文本**必须区分两种语义**：lock 给的是「本次构建真正链接的解析版本」，yaml 给的是
「版本**约束** `^0.24.27`」—— GPL §6 的对应源码声明里把约束当版本写是不诚实的。

同时修掉父仓库 `.github/workflows/release.yml` 里的**死指针**：原文写
`version : see songloft-player/pubspec.lock at the commit above`，
而该路径在那个 commit 上**根本不存在**（同样被 gitignore），等于让 GPL 接收者去找一个
不存在的文件。改为指向子模块里**入库的** `pubspec.yaml` 的 `webf:` 约束 + 构建日志 URL，
并写明那是约束而非解析版本。

**根因不是这两处文本，是验证方式**：许可相关的 workflow 步骤此前只做过 YAML 语法校验，
**从没在 `bash -e` 下真跑过步骤正文**，所以 1 和 2 都没被发现。本次修复的验证做法（照抄即可）：
用 YAML parser 把该 step 的 `run:` 正文抽成脚本、把 `${{ }}` 表达式换成测试值，
在仓库外的 `mktemp -d` 里造最小环境（`LICENSES/GPL-3.0.txt` + `pubspec.yaml` + 一个假的
`final-assets/` 产物），`bash -e` 跑「有 lock」「无 lock」「两者都无」「零产物」四条路径。
反向验证也做了：同一 harness 跑**旧脚本**在无 lock 树上 `exit=2`、`CORRESPONDING-SOURCE.txt`
根本没生成，与 CI 日志一致 —— 证明 harness 忠实，不是「换个说法就算修好」。

**待用户决策的后续项**：把 `pubspec.lock` 入库能从根上消掉这两处（Flutter 官方对
application package 的建议就是入库；它也确实属于 GPL 意义上「完整对应源码」的一部分；
已确认两个 workflow 里**没有任何 `hashFiles` 依赖它**）。但那是仓库策略变更、
有人是刻意加进 `.gitignore` 的，**未获授权不要自行改**。

**未修但同类的写法**（本次不扩大范围）：player
`build-and-release.yml:161/169/177/301` 的 `DEB_FILE/RPM_FILE/APPIMAGE_FILE/DMG_FILE=$(find dist ...)`
同样是 `-e` 下的脆弱赋值（`dist/` 不存在时 `find` 退出 1 → 整步中止），
但这四步都带 `continue-on-error: true`、打包本就是 best-effort，**失败不阻塞 release**，
故只记录不改。每平台的 `cp LICENSES/GPL-3.0.txt ...`（`:152` Linux、`:230` Windows pwsh、
`:294` macOS）路径都入库且必然存在，无此隐患；父仓库没有逐平台的许可拷贝步骤，
只有 `release.yml:1079` 那一处 release 附件步骤。

### task #12 — 给 WebF 上游报 §3.2 里的 **7** 条

草稿在 `docs/webf/upstream-issues.md`（同为分支临时件）。
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
- workflow 的 `run:` 默认 shell 是 `bash -e` → `VAR=$(可能失败的命令)` 会**吃掉紧跟其后的兜底逻辑**
  （赋值语句本身非零退出，整步中止）。必须写 `|| true`，文件存在性用 `[ -f ]` 守卫。
  已在 §4 task #3 台账里兑现过一次真实事故
- 改任何 workflow 的 `run:` 正文，**YAML 语法校验不算验证** → 把正文抽出来在
  `bash -e` + 仓库外临时目录里真跑一遍，且要跑「依赖文件缺失」那条分支
