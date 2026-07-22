# Songloft Frontend Gotchas

This document captures pitfalls in the Flutter client (`songloft-player`) that have been **diagnosed and fixed, but whose root causes are broadly applicable** — mostly in Flutter Web (CanvasKit) rendering, platform views (plugin iframes), and the audio backend. These burned through many rounds of investigation; archiving them here prevents repeating that.

> Backend/business-side pitfalls are in the repo-root `AGENTS.md` under "Business gotchas summary" and "Platform adaptation gotchas".

---

## 1. Cover / image rendering (hard rule)

### Rule: all covers must go through `CoverImage` / `NetworkCoverImage`; never use bare `CachedNetworkImage` / `Image.network`

- Unified wrappers: `lib/shared/widgets/cover_image.dart` (`CoverImage`) and `lib/shared/widgets/network_cover_image.dart` (`NetworkCoverImage`). On Web they force `imageRenderMethodForWeb: HttpGet` + `memCacheWidth` (decode a thumbnail scaled to the displayed physical pixels).
- **Every new cover render site** must use one of these two components; don't write `CachedNetworkImage(...)` or `Image.network(...)` directly.

### Why (the Web CanvasKit GPU-memory trap)

On Web, covers occasionally turn **solid black** (pure black, not the load-failure placeholder icon) or degrade to a placeholder icon, triggered by **in-app navigation / tab switch / filter and back**:

- A bare `CachedNetworkImage` on Web defaults to `ImageRenderMethodForWeb.HtmlImage`, uploading the image at **full resolution** (up to ~1.5 MB each) as a GPU texture.
- CanvasKit has a single WebGL context (Chrome uses `OffscreenCanvasRasterizer`). Platform views like plugin iframes / video each consume GPU resources; **full-resolution cover textures accumulate → exhaust GPU memory → the WebGL context is dropped** (`ImageCodecException: Failed to create image` / `MakeLazyImageFromTextureSourceWithInfo` returns null). After the context dies, reused `ui.Image` textures paint black; new decodes during the dead window can't get a texture and degrade to the placeholder icon.
- `memCacheWidth` **only takes effect on the `HttpGet` path**; on the default `HtmlImage` path it does not downscale — so you must explicitly set `imageRenderMethodForWeb: HttpGet`, shrinking each cover from ~1.5 MB to tens/hundreds of KB and drastically reducing memory pressure.

### Investigation lessons (avoid the 9-round detour)

- **Pure black ≠ load failure**: pure black is at the GPU texture layer and `errorWidget` can't catch it; the placeholder icon is a load/decode failure. Distinguish first.
- `imageCache.evict` / changing the key / attaching a `NavigatorObserver` to force rebuilds are **only probabilistic mitigations** and can even worsen the re-decode storm. GPU texture lifetime issues are only truly fixed by: shrinking decode size (`HttpGet`+`memCacheWidth`), reducing concurrent platform views, or upgrading the engine.
- `canvasKitForceCpuOnly` (CPU software rendering) removes GPU-texture black but is **very laggy** and covers still get lost at the loading layer — net negative, ruled out.
- To fix a class of image problem, **grep all render sites** (`CachedNetworkImage` / `Image.network`); don't assume fixing one shared component covers everything — cover rendering is spread across `CoverImage` plus several card components.
- The decisive breakthrough came from **having the user paste real-device console logs** (the exact error string + `[Cover] OK yet renders black` + cache not evicted), not from many rounds of pure inference.

---

## 2. Web plugin-tab iframe reloading / jitter

The `<iframe>` (`HtmlElementView` platform view) hosting a plugin page suffered repeated reloads (the entry page requested 25–40 times/sec) and visual jitter. Fixed; there were three layers of root cause:

### 1. Widget tree: iframe reload ⟺ its hosting widget element is disposed+rebuilt

On Flutter Web the iframe DOM element is cached once per platform view `viewId` and projected through a shadow-DOM `<slot>` — a normal rebuild / relocation **won't** touch the iframe; only **destroying and rebuilding the viewId** re-fetches it.

- **Rule**: in the plugin keep-alive Stack (`shell_layout.dart`), anything that disposes+rebuilds the `PluginTabPage` element triggers an iframe reload — reordering, reparenting (a local `ValueKey` can't move across parents), or a provider that momentarily goes empty dropping the active tab.
- Fix: (1) `PluginTabPage` uses a **stable `GlobalKey` cached by entryPath** (moves across the tree instead of rebuilding); (2) keep-alive pruning keys only off the stable `tabConfig.pluginTabs` and only when `tabConfigAsync.hasValue`, never off the momentarily-empty `jsPluginsProvider` snapshot. Caching the iframe inside the factory is **useless** (when the engine rebuilds the viewId it moves the iframe to a new wrapper, and the move still forces a browser reload).

### 2. CSS: the document-level scrollbar add/remove reflow loop

When content height ≈ viewport, "scrollbar appears → content narrows → reflow → height changes → scrollbar disappears → …" flips every frame.

- Fix: in `internal/jsplugin/assets/common.css`, give `html` `overflow-y: scroll` (scrollbar always present, constant width, loop broken).
- Note: `scrollbar-gutter: stable` **only works on scroll containers (`overflow:auto/scroll`)**; it's a no-op when `html` is `visible` — which is why an earlier attempt didn't fix it. Linux/headless with overlay (0-width) scrollbars can never reproduce it; only Windows Chrome (classic ~15px scrollbar) does.

### 3. Cache layer (hard rule): immutable long-cache assets must use versioned URLs

Even after the CSS was correct, users still saw jitter — the real culprit was cache headers. `jsplugin-assets/*` (`common.css`/`common.js`) previously used a **fixed, unversioned URL** + `Cache-Control: immutable`, so browsers didn't even revalidate and cached the old file for a year; the fix could never reach users.

- **General rule**: any `immutable` long-cache asset **must use a content-hash versioned URL** (e.g. `?v=<first 8 of sha256>`), or any later change is unreachable for existing users.
- Implementation: `injectHTMLHead` appends `?v=<hash>` to the injected `common.css/js` URLs; the hosting HTML is `no-cache` so it always carries the latest version. When content is unchanged the URL is stable and the long cache still applies.

---

## 3. Web plugin iframe blocked by semantics nodes (unclickable)

After the Web player bar appears, `flt-semantics` nodes (`pointer-events:auto`) cover the plugin iframe and steal clicks.

- **Root cause**: `main.dart` forces a persistent semantics tree on Web (`ensureSemantics`, an accessibility improvement) and hits a lingering engine bug [flutter/flutter#175119]: `ensureSemantics` + go_router + platform-view iframe — after closing a dialog with a barrier, a semantics node stays at `pointer-events:auto` layered over the platform view.
- **Fix (option A)**: temporarily release the semantics handle to close the semantics tree when entering a plugin iframe page, and restore it on leave. `lib/core/a11y/web_semantics_controller.dart` is a singleton holding a `SemanticsHandle`; `shell_layout.dart` suspends/resumes on the `isPluginTab` edge. When a screen reader is active the platform holds its own handle, so releasing ours doesn't close the tree — AT users are unaffected.

---

## 4. Web mobile black screen after returning from background (status: accepted, awaiting engine fix)

Android Chrome shows a black screen after returning from background.

- **Root cause**: while backgrounded the browser discards CanvasKit's WebGL context; the engine sets `_forceNewContext` then **passively waits for `webglcontextrestored`** — which Android Chrome often never fires. Deeper still is engine bug [flutter/flutter#184683]: in the new surface architecture `onContextLost` fires before a `late` field is assigned → `LateInitializationError` → render frame crashes → white screen. The fix PR [#185116] is not yet in the latest stable.
- **Current decision**: all platforms (including Web) pin to stable Flutter 3.44.6 with GPU rendering; 3.44.6 lacks #185116, so mobile Web background black screen is **accepted as-is**, to be truly fixed by upgrading once the fix lands in stable.
- **Don't retry** these directions: source backport is useless (the web engine is precompiled into `.dill`; release dart2js doesn't read `lib/_engine` sources); CDN vs offline wasm is unrelated to the black screen (it only changes the initial load source, not the GPU context lifetime); pure JS can't force Flutter to produce a frame — recovery must be on the Dart side.

---

## 5. Audio playback backend

### All native platforms unified on media_kit/libmpv, no kill-switch

- All native platforms (Win/Linux/macOS/Android/iOS) always use media_kit (libmpv); the native ExoPlayer/AVPlayer fallback and the `SONGLOFT_MEDIAKIT_*` switch have been **removed** (`AudioBackend.usesMediaKit => !kIsWeb`). EQ always goes through mpv `af` (`MpvEqualizerService`).
- ⚠️ There is no kill-switch anymore; if media_kit breaks on a platform you **can no longer fall back via `--dart-define` — you must change code**.
- Historical pitfalls (`SongloftMediaKitPlayer` implementation notes):
  - Must override `setAndroidAudioAttributes` as a safe no-op — the base class throws by default, just_audio calls it only on Android, otherwise every track on Android crashes right at `setAudioSource` (`UnimplementedError`).
  - On mobile, **do not pre-create** a `VideoController` — pre-creating makes every `open()/play()/seek()` `await` video-texture readiness, and on Android without a Video widget that Future never completes → every track hangs. Build it lazily only when a video source is detected.

### Self-signed / incomplete-chain HTTPS playback: `AudioSource.uri(headers=null)` bypasses the local proxy

- **Root cause**: SSL bypass is only done at the Dart layer (`HttpOverrides` trust-all + mpv `tls-verify`). But when `AudioSource.uri` is passed `headers=null`, just_audio **skips the local plaintext loopback proxy** and hands the URL straight to the native player, whose TLS handshake is outside `dart:io` where `HttpOverrides` can't reach → self-signed fails. (just_audio only engages the local proxy when `headers!=null || userAgent!=null`.)
- **Fix**: when `insecureTls==true`, attach a non-empty header to `AudioSource.uri` to force the local proxy — the native player only connects to `127.0.0.1` in plaintext while upstream HTTPS is fetched by just_audio's Dart `HttpClient` (trust-all). Mobile HLS with self-signed certs needs a custom HLS-aware trust-all local proxy (`lib/core/network/insecure_media_proxy_native.dart`) that recursively rewrites m3u8 sub-resources.

### Desktop HLS radio failures

- **Live HLS segment expiry**: desktop mpv has no CORS, but when the user has the global HLS proxy on, timestamped short-window segments have already rolled out of the window after a round trip through the local proxy → `404`. Fix: the backend `serveRadio` supports `?hls=direct` to force a 302 bypass of the proxy, and the desktop frontend appends it for live playback (desktop carries its own Referer/UA); mobile does not (its native player sends no Referer/UA, so it keeps the proxy).
- **HE-AAC client decode crash**: HE-AAC stations like `streamtheworld` crash after ~3.7s on both Web and mpv with `aac: decode_band_types: Input buffer exhausted`. **Confirmed by testing that the backend is not at fault** (both direct ffmpeg and a faithful proxy replica decode cleanly); the real cause is the client decoder failing on HE-AAC. Partial mitigation: backend maps `audio/aacp`→`audio/aac`, mpv adds `network-timeout`; a full fix needs backend transcoding fallback (not done).

### Windows plugin page requires explicit WebView2 initialization

- The Windows portable build never initializes `WebViewEnvironment`; the default user-data folder lands in a read-only directory → `Cannot create the InAppWebView instance!`, and on failure the controller stays null and `reload()` is a no-op that never self-heals.
- Fix: `core/utils/webview_environment.dart` builds a global singleton using `getApplicationSupportDirectory()/webview2` as a writable `userDataFolder`, passed to both `InAppWebView` sites, with retry changed to rebuild the widget via a new `ValueKey`.
- Related: in cleanup that hard-kills the process before exit (`TerminateProcess`), **tray-icon/resource removal must come before the hard-kill call**, or `destroy()` becomes dead code (the hard-kill never returns).

---

## 6. Playing a single song from a paginated list (hard rule)

### Rule: any "tap a single song to play" in a paginated list must `playPlaylist(loaded, startIndex)` + backfill in the background with the same filter

- **Pitfall**: a paginated list (playlist >100 songs, category pages, etc.) calling `playPlaylist(songs, startIndex)` directly, where `songs` is only the currently-loaded page → the play queue is truncated to the loaded count (e.g. max 100) and only loops within it.
- **Fix**: use `PlayerNotifier.playPlaylistFromLoaded` — play immediately with the loaded page + startIndex, and when `total > loaded count` backfill in the background with the same sort/order/keyword (reusing `_loadRemainingSongsById`).
- **Any new "tap single song in a paginated list" page** must include the backfill logic; category/facet filter pages must keep the field→getSongs parameter mapping consistent (`categorySongsFilter`).

---

## Related module references

- Plugin common assets and theme bridge: `internal/jsplugin/assets/` (`common.css`/`common.js`), `injectHTMLHead`.
- Play activity / prefetch transcoding: prefetch transcoding must not be canceled by activating the current song via `playactivity.Activate` (`Activate` skips `CatPrefetch`); otherwise the next track's prefetch ffmpeg is SIGKILLed and playback still transcodes in real time.
- Sources without duration (e.g. WebDAV): `song.duration=0` makes miot speakers not advance tracks (advancing relies on server-side duration); probe and backfill on import (`RefreshSong` with bounded concurrency after `AddRemoteSongs`).
