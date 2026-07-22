# Songloft 前端踩坑与铁律（Frontend Gotchas）

本文沉淀 Flutter 客户端（`songloft-player`）里**已定位并修复、但根因具有普适性**的坑，主要集中在 Flutter Web（CanvasKit）渲染、平台视图（插件 iframe）、音频后端。这些坑反复消耗过大量排查回合，归档在此避免重蹈。

> 后端/业务侧踩坑见项目根 `AGENTS.md` 的「业务踩坑总结」与「平台适配踩坑」章节。

---

## 一、封面 / 图片渲染（铁律）

### 铁律：所有封面必须走 `CoverImage` / `NetworkCoverImage`，禁止裸 `CachedNetworkImage` / `Image.network`

- 统一封装：`lib/shared/widgets/cover_image.dart`（`CoverImage`）和 `lib/shared/widgets/network_cover_image.dart`（`NetworkCoverImage`），Web 下强制 `imageRenderMethodForWeb: HttpGet` + `memCacheWidth`（按显示物理像素缩略解码）。
- **新增任何封面渲染点**都必须用这两个组件之一；不要直接写 `CachedNetworkImage(...)` 或 `Image.network(...)`。

### 为什么（Web CanvasKit 的 GPU 显存陷阱）

Web 端封面偶发**变黑**（纯黑，不是加载失败占位图标）或退化成占位图标，触发场景是**应用内切页/切 tab/筛选后回来**：

- 裸 `CachedNetworkImage` 在 Web 默认走 `ImageRenderMethodForWeb.HtmlImage`，**全分辨率**（单张可达 ~1.5MB）上传为 GPU 纹理。
- CanvasKit 只有单一 WebGL context（Chrome 走 `OffscreenCanvasRasterizer`）。插件 iframe / 视频等 platform view 各占用 GPU 资源，**满分辨率封面纹理累积 → 挤爆 GPU 显存 → WebGL context 被丢弃**（`ImageCodecException: Failed to create image` / `MakeLazyImageFromTextureSourceWithInfo` 返回 null）。context 死亡后复用的 `ui.Image` 纹理直接绘制成黑；context 失效期新解码拿不到纹理则退化成占位图标。
- `memCacheWidth` **只在 `HttpGet` 路径生效**，在默认 `HtmlImage` 路径不缩略——所以必须显式设 `imageRenderMethodForWeb: HttpGet`，把每张封面从 ~1.5MB 缩到数十~百 KB，大幅降低显存压力。

### 排查教训（避免重蹈 9 轮弯路）

- **纯黑 ≠ 加载失败**：纯黑是 GPU 纹理层，`errorWidget` 捕获不到；占位图标才是加载/解码失败。先分清。
- `imageCache.evict` / 换 key / 挂 `NavigatorObserver` 主动重建都**只是概率缓解**，甚至会加剧重解码风暴。GPU 纹理生命周期问题的确定根治只有：缩小解码尺寸（`HttpGet`+`memCacheWidth`）、减少并存 platform view、或升级引擎。
- `canvasKitForceCpuOnly`（CPU 软件渲染）能消掉 GPU 纹理黑，但**非常卡**且封面仍会因加载层问题丢失，是负收益，已排除。
- 修某类图片问题要 **grep 全部渲染点**（`CachedNetworkImage` / `Image.network`），别只改一个共享组件就以为覆盖全——本项目封面渲染散落在 `CoverImage` + 多处卡片组件。
- 诊断决定性突破来自**让真机贴 console 日志**（具体错误串 + `[Cover] OK 但显示黑` + cache 未淘汰），别连续多轮纯推断改代码。

---

## 二、Web 插件 Tab 的 iframe 反复重载 / 抖动

承载插件页的 `<iframe>`（`HtmlElementView` 平台视图）曾出现反复重新加载（入口页被 25~40 次/秒重复请求）与视觉抖动。已修复，涉及三层根因：

### 1. widget 树层：iframe 重载 ⟺ 承载它的 widget 元素 dispose+重建

Flutter Web 里 iframe DOM 元素按 platform view 的 `viewId` 缓存一次，经 shadow-DOM `<slot>` 投影——普通 rebuild / 重定位**不会**动 iframe，只有 **viewId 被销毁并重建**才重新拉取。

- **铁律**：插件保活 Stack（`shell_layout.dart`）里，凡让 `PluginTabPage` 元素 dispose+重建的都会触发 iframe 重载——重排、换父（local `ValueKey` 不能跨父移动）、依赖会瞬时为空的 provider 误删激活 tab。
- 修复：① `PluginTabPage` 用**按 entryPath 缓存的稳定 `GlobalKey`**（跨树移动而非重建）；② 保活裁剪只按稳定的 `tabConfig.pluginTabs` 且仅 `tabConfigAsync.hasValue` 时执行，不依赖会瞬时为空的 `jsPluginsProvider` 快照。工厂内缓存 iframe **无效**（引擎重建 viewId 时会把 iframe 挪到新 wrapper，move 仍触发浏览器重载）。

### 2. CSS 层：文档级滚动条增删的重排回路

内容高度 ≈ 视口时「滚动条出现→内容变窄→重排→高度变化→滚动条消失→…」每帧翻转。

- 修复：`internal/jsplugin/assets/common.css` 给 `html` 加 `overflow-y: scroll`（滚动条常驻、宽度恒定，打断回路）。
- 注意：`scrollbar-gutter: stable` **只对滚动容器（`overflow:auto/scroll`）生效**，`html` 是 `visible` 时无效——这是它一度没修好的原因。Linux/headless 用覆盖式（0 宽度）滚动条永远复现不了，Windows Chrome（经典 ~15px 滚动条）才复现。

### 3. 缓存层（铁律）：immutable 长缓存资源必须走版本化 URL

CSS 改对了用户却仍抖——真凶是缓存头。`jsplugin-assets/*`（`common.css`/`common.js`）原用**固定无版本 URL** + `Cache-Control: immutable`，浏览器连重新验证都不做，把旧文件缓存一年，修复永远到不了用户。

- **通用铁律**：凡 `immutable` 长缓存的资源，**必须走内容哈希版本化 URL**（如 `?v=<sha256前8位>`），否则任何后续修改对老用户都不可达。
- 实现：`injectHTMLHead` 给注入的 `common.css/js` URL 加 `?v=<hash>`，承载页 HTML 为 `no-cache` 每次带出最新版本号；内容不变时 URL 恒定，长缓存照旧。

---

## 三、Web 插件 iframe 被语义节点遮挡（无法点击）

Web 端播放栏出现后，`flt-semantics` 节点（`pointer-events:auto`）盖住插件 iframe 抢走点击。

- **根因**：`main.dart` 在 Web 强制常驻语义树（`ensureSemantics`，无障碍改进）命中 Flutter 引擎残留 bug [flutter/flutter#175119]：`ensureSemantics` + go_router + 平台视图 iframe，关闭带 barrier 的对话框后语义节点卡在 `pointer-events:auto` 叠在平台视图之上。
- **修复（方案 A）**：进入插件 iframe 页时临时释放语义句柄关闭语义树、离开时恢复。`lib/core/a11y/web_semantics_controller.dart` 单例持 `SemanticsHandle`；`shell_layout.dart` 按 `isPluginTab` 边沿 suspend/resume。读屏器激活时平台持独立句柄，释放我们的句柄不关语义树，AT 用户不受损。

---

## 四、Web 移动端切后台回来黑屏（现状：接受，待引擎修复）

Android Chrome 切后台回来黑屏。

- **根因**：后台时浏览器丢弃 CanvasKit 的 WebGL context，引擎置 `_forceNewContext` 后**被动等 `webglcontextrestored`**——Android Chrome 常不 fire 该事件。更深一层是引擎 bug [flutter/flutter#184683]：新 surface 架构里 `onContextLost` 在 `late` 字段赋值前 fire → `LateInitializationError` → 渲染帧崩 → 白屏。修复 PR [#185116] 至今未进最新 stable。
- **现状决策**：全平台（含 Web）统一 stable Flutter 3.44.6、保持 GPU 渲染；3.44.6 不含 #185116，移动端切后台白屏**按现状接受**，待修复进 stable 后升级根治。
- **不要再试**的方向：源码 backport 无效（web 引擎预编进 `.dill`，release dart2js 不读 `lib/_engine` 源码）；CDN vs 离线 wasm 与黑屏无关（只改初始加载来源，不改 GPU context 生命周期）；纯 JS 侧无法强制 Flutter 产帧，恢复必须在 Dart 侧。

---

## 五、音频播放后端

### 所有原生平台统一 media_kit/libmpv，无 kill-switch

- 所有原生平台（Win/Linux/macOS/Android/iOS）音频后端恒用 media_kit（libmpv），已**删除**原生 ExoPlayer/AVPlayer 回退与 `SONGLOFT_MEDIAKIT_*` 开关（`AudioBackend.usesMediaKit => !kIsWeb`）。EQ 一律走 mpv `af`（`MpvEqualizerService`）。
- ⚠️ 已无 kill-switch，某平台 media_kit 出问题时**不能再靠 `--dart-define` 回退，需改代码**。
- 历史坑（`SongloftMediaKitPlayer` 实现要点）：
  - 必须重写 `setAndroidAudioAttributes` 为安全 no-op——基类默认 throw，just_audio 仅 Android 调它，否则 Android 全曲在 `setAudioSource` 阶段就炸（`UnimplementedError`）。
  - 移动端构造时**不要预建** `VideoController`——预建会让每个 `open()/play()/seek()` 都 `await` 视频纹理就绪，Android 无 Video widget 时该 Future 永不完成 → 全曲挂住。改为判定视频源时才惰性建。

### 自签 / 不完整链 HTTPS 播放：`AudioSource.uri(headers=null)` 会绕过本地代理

- **根因**：SSL 忽略只做在 Dart 层（`HttpOverrides` trust-all + mpv `tls-verify`）。但 `AudioSource.uri` 传 `headers=null` 时 just_audio **跳过本地明文回环代理**，把 URL 直接交给原生播放器，其 TLS 握手在 `dart:io` 之外，`HttpOverrides` 管不到 → 自签失败。（just_audio 仅当 `headers!=null || userAgent!=null` 才 engage 本地代理。）
- **修复**：`insecureTls==true` 时给 `AudioSource.uri` 附非空 header，强制走本地代理——原生只连 `127.0.0.1` 明文，上游 HTTPS 由 just_audio 的 Dart `HttpClient`（trust-all）拉取。移动端 HLS 自签需自研 HLS-aware trust-all 本地代理（`lib/core/network/insecure_media_proxy_native.dart`）递归改写 m3u8 子资源。

### 桌面端 HLS 电台失败

- **HLS 直播切片过期**：桌面端 mpv 无 CORS，但用户全局开着 HLS 代理时，带时间戳的短窗口切片经本机反代往返后已滚出窗口 → `404`。修复：后端 `serveRadio` 支持 `?hls=direct` 强制 302 绕过反代，前端桌面 live 播放追加该参数（桌面自带 Referer/UA）；移动端不加（其原生 player 不发 Referer/UA，保留反代）。
- **HE-AAC 客户端解码崩**：`streamtheworld` 等 HE-AAC 电台 web+mpv 都播 ~3.7s 后 `aac: decode_band_types: Input buffer exhausted` 崩。**已实测坐实后端无责**（ffmpeg 直连 + 复刻代理都干净解码），真因是客户端解码器扛不住 HE-AAC。部分缓解：后端 `audio/aacp`→`audio/aac`、mpv 加 `network-timeout`；彻底解决需后端转码兜底（未做）。

### Windows 插件页 WebView2 需显式初始化

- Windows 免安装版从不初始化 `WebViewEnvironment`，默认用户数据目录落在只读目录 → `Cannot create the InAppWebView instance!`，且实例创建失败时 controller 恒 null、`reload()` 是 no-op 永不自愈。
- 修复：`core/utils/webview_environment.dart` 用 `getApplicationSupportDirectory()/webview2` 作可写 `userDataFolder` 建全局单例，两处 `InAppWebView` 都传入，重试改为换 `ValueKey` 重建部件。
- 相关：退出前要硬杀进程（`TerminateProcess`）的清理里，**托盘图标/资源移除必须排在硬杀调用之前**，否则 `destroy()` 成死代码（硬杀永不返回）。

---

## 六、分页列表点击单曲播放（铁律）

### 铁律：任何分页列表「点击单曲播放」都要 `playPlaylist(已加载, startIndex)` + 按同筛选后台补齐

- **坑**：分页列表（歌单 >100 首、分类页等）直接 `playPlaylist(songs, startIndex)`，`songs` 只是当前已加载页 → 播放队列被截断到已加载数（如最长 100），只在其中循环。
- **修复**：用 `PlayerNotifier.playPlaylistFromLoaded`——以已加载分页 + startIndex 立即播放，`total > 已加载数` 时按相同 sort/order/keyword 后台补齐（复用 `_loadRemainingSongsById`）。
- **新增任何「分页列表点单曲」页面**务必带补齐逻辑；分类/facet 过滤页要保证 field→getSongs 参数映射一致（`categorySongsFilter`）。

---

## 相关模块参考

- 插件公共资源与主题桥接：`internal/jsplugin/assets/`（`common.css`/`common.js`）、`injectHTMLHead`
- 播放活动 / 预热转码：预热（prefetch）转码不应被切当前歌的 `playactivity.Activate` 取消（`Activate` 跳过 `CatPrefetch`），否则下一首预热 ffmpeg 被 SIGKILL、播放仍实时转码。
- WebDAV 等无时长源：`song.duration=0` 会导致 miot 音箱不切歌（切歌依赖服务端 duration），导入时应探测补齐（`AddRemoteSongs` 后台限并发 `RefreshSong`）。
