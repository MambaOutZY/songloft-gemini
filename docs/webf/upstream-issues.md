# WebF 上游缺陷报告草稿（临时件）

> **这是 `webf-plugin-render` 分支的临时工作文件，不是产品文档。**
>
> 内容为面向 WebF 上游仓库（`openwebf/webf`）的缺陷报告**草稿**，正文刻意用英文书写（上游为英文仓库）。
> 本文件**刻意不做中英双语同步**，也已从文档站（VitePress）的构建中排除。
> `songloft-org/songloft#341` 落地后，本文件将连同其他临时件一并删除。
>
> 草稿尚未提交到上游。是否提交、何时提交、以何种形式拆分合并，由维护者决定。

**核实基准**：`webf 0.24.27`（`~/.pub-cache/hosted/pub.dev/webf-0.24.27/`）。
注意 WebF 的 C++ binding 层不随 pub 包发布，因此部分事实只能从 Dart 侧核实，或只能标注为
「容器实测观察到、源码层面未能完全解释」。

---

## 1. `@font-face` remote URL branch passes the whole underlying `ByteBuffer` to `FontLoader`, so HTTP-loaded fonts render as tofu

**Title (one line):**
`@font-face` src url() fonts fail to render (tofu) — `_loadFont` uses `bundle.data!.buffer.asByteData()` instead of `ByteData.sublistView(bundle.data!)`

### Environment

- webf `0.24.27` (pub.dev)
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0` (exact toolchain patch version not recorded in this report)
- Platform: Linux x86-64, inside a container (WebF Linux embedder requires glibc >= 2.38)

### Description

A `@font-face` rule whose `src` is a remote `url(...)` pointing at a `.ttf` file never renders with that
font. The network request succeeds (HTTP 200, correct `Content-Length`), `_loadFont` returns `true`,
and no warning is logged — yet every glyph that should come from the custom family paints as a
tofu box / falls back to the default family.

The same font embedded as a `data:` URI in the very same stylesheet renders correctly. That
asymmetry is the tell: the two branches build their `ByteData` differently.

### Root cause

`lib/src/css/font_face.dart`, `FontFaceRegistry._loadFont` (two branches in the same method):

```dart
// lib/src/css/font_face.dart:373-378  — data: / in-memory branch (CORRECT)
if (descriptor.font.content.isNotEmpty) {
  Uint8List content = descriptor.font.content;
  Future<ByteData> bytes = Future.value(ByteData.sublistView(content));
  FontLoader loader = FontLoader(descriptor.fontFamily);
  loader.addFont(bytes);
  await loader.load();
```

```dart
// lib/src/css/font_face.dart:394-397  — url() branch (BUG)
if (bundle.data == null || bundle.data!.isEmpty) return false;

FontLoader loader = FontLoader(descriptor.fontFamily);
Future<ByteData> bytes = Future.value(bundle.data!.buffer.asByteData());
loader.addFont(bytes);
```

`bundle.data` is a `Uint8List` (`lib/src/foundation/bundle.dart:109`) and is **not guaranteed to
span its whole backing `ByteBuffer`**. `.buffer.asByteData()` discards the view's `offsetInBytes`
and `lengthInBytes` and hands `FontLoader` the entire underlying allocation.

Two concrete ways `bundle.data` ends up as a partial view:

- `lib/src/foundation/bundle.dart:430`/`:448` (Dio path): `Uint8List bytes = resp.data ?? Uint8List(0); ... data = bytes;`
  With `ResponseType.bytes`, Dio accumulates into a `BytesBuilder`; `takeBytes()` on the copying
  builder returns a view over a power-of-two over-allocated buffer, i.e. `lengthInBytes >= length`.
  The extra bytes are zero padding **appended after the TTF**.
- `lib/src/foundation/bundle.dart:557`: `data = bytes.buffer.asUint8List();` — same mistake one layer
  down, already widening whatever view `bytes` was.

Flutter's `FontLoader.load()` forwards each `ByteData` to `loadFontFromList`, which parses it as a
complete font file. Trailing padding past the TTF's declared tables makes the parse fail (or produce
an unusable typeface). Since the failure happens inside the engine rather than as a thrown Dart
exception, the `catch` at `font_face.dart:400` never fires and nothing is logged.

### Expected vs actual

| | |
|---|---|
| **Expected** | `url(...)`-sourced `@font-face` renders identically to the same font supplied as a `data:` URI. |
| **Actual** | Glyphs render as tofu / fall back to the default family. Zero diagnostics: request is 200, `_loadFont` returns `true`, no log line. |

### Minimal reproduction

```html
<!doctype html>
<style>
  @font-face { font-family: 'MyFont'; src: url('/fonts/MyFont.ttf'); }
  body { font-family: 'MyFont', sans-serif; font-size: 40px; }
</style>
<body>Hello &#x4F60;&#x597D;</body>
```

Serve `MyFont.ttf` over HTTP from any static server and load the page in WebF. Compare against a
second page where the identical font bytes are inlined as `src: url(data:font/ttf;base64,...)` — the
inlined one renders, the HTTP one does not.

Note on minimality: whether the bug reproduces depends on the HTTP stack handing back an
over-allocated / offset view, which in turn depends on response size and chunking. A font whose size
is not an exact power of two is the reliable case. This is why the report leads with the code-level
asymmetry rather than the repro.

### Suggested fix

One line, mirroring the branch directly above it:

```diff
-        Future<ByteData> bytes = Future.value(bundle.data!.buffer.asByteData());
+        Future<ByteData> bytes = Future.value(ByteData.sublistView(bundle.data!));
```

`lib/src/foundation/bundle.dart:557` (`data = bytes.buffer.asUint8List();`) should likewise become
`data = bytes;` (or `Uint8List.sublistView(bytes)`), since it currently widens a view for every
consumer of `bundle.data`, not just fonts.

### Two adjacent defects in the same file (may warrant splitting out)

**(a) `woff2` / `woff` are not in `supportedFonts`.**

```dart
// lib/src/css/font_face.dart:20-25
final List<String> supportedFonts = [
  'ttc',
  'ttf',
  'otf',
  'data'
];
```

`woff2` is the dominant web font format. Flutter's `loadFontFromList` cannot decompress it, so
excluding it is arguably correct — but combined with (b) it means a spec-correct stylesheet is
silently dropped.

**(b) The `format` is inferred from the URL's file extension, and `format()` in the CSS is ignored;
if nothing matches, the method returns without any diagnostic.**

```dart
// lib/src/css/font_face.dart:154-155 (resolveFontFaceRules)
String formatFromExt = tmpSrc.split('.').last;
fonts.add(FontSource(tmpSrc, formatFromExt));
```

```dart
// lib/src/css/font_face.dart:238-241
FontSource? targetFont = fonts.firstWhereOrNull((f) => supportedFonts.contains(f.format));
if (targetFont == null) {
  return;
}
```

(Identical logic at `:232-233` in `registerFromBridge` and `:160-167` in `resolveFontFaceRules`.)

Consequences:

- A CSS-conformant `src: url('/f/MyFont') format('truetype')` is dropped, because `split('.').last`
  on a extensionless URL yields the whole path, which is not in `supportedFonts`. The
  `format('truetype')` token the spec says to use is never read.
- Query strings and cache-busters break detection: `MyFont.ttf?v=3` → format `ttf?v=3` → dropped.
- The drop is **completely silent**: no request is issued and nothing is logged, so from the page
  author's side the `@font-face` rule simply has no effect. A `cssLogger.warning` naming the family,
  the rejected source and the supported list would turn a multi-hour investigation into one log line.

### Confidence

**Source-verified.** All line numbers above were re-checked against `webf-0.24.27` in pub-cache. The
buffer-vs-view asymmetry between the two branches is unambiguous in the source. The precise
mechanism by which the engine rejects the padded buffer is inferred (the engine does not surface a
Dart-visible error), but the fix is correct regardless of that detail.

---

## 2. `<base href>` is not implemented — every relative URL resolves against `controller.url` instead of the document base URL

**Title (one line):**
`<base href>` has no effect: relative URLs in `<a>`, `<img>`, `<script>`, `<link>` and CSS all resolve against `controller.url`

### Environment

- webf `0.24.27` (pub.dev)
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0`
- Platform: Linux x86-64 container (glibc >= 2.38)

### Description

`<base href="...">` in `<head>` is silently ignored. A page served from a nested path, or served
under a reverse-proxy path prefix, cannot use `<base href>` to make its root-relative and
document-relative URLs resolve correctly — the standard mechanism for exactly that.

This matters for any content that is authored once and mounted under a configurable path prefix, and
it is the one HTML-level knob that normally makes such content path-agnostic without rewriting every
URL at build time.

### Root cause

There is no `BaseElement` in WebF at all, and no document-level base URL concept.

`lib/src/html/head.dart` defines exactly these head elements — note the absence of `base`:

```
lib/src/html/head.dart:54    class HeadElement extends Element
lib/src/html/head.dart:160   class LinkElement extends Element
lib/src/html/head.dart:632   class MetaElement extends Element
lib/src/html/head.dart:639   class TitleElement extends Element
lib/src/html/head.dart:646   class NoScriptElement extends Element
lib/src/html/head.dart:849   class StyleElement extends Element with StyleElementMixin
```

There is also **no `document.baseURI`** anywhere in `lib/src` (`grep -rn baseURI lib/src` → no hits).

Instead, every consumer of a relative URL independently hardcodes `controller.url` as its base:

```dart
// lib/src/html/a.dart:290-292
String base = ownerDocument.controller.url;
...
_resolvedHyperlink = ownerDocument.controller.uriParser!.resolve(Uri.parse(base), Uri.parse(href));
```

```dart
// lib/src/html/img.dart:1092  +  :1121
String base = ownerDocument.controller.url;
return ownerDocument.controller.uriParser!.resolve(Uri.parse(base), srcUri);
```

```dart
// lib/src/html/script.dart:356-358
String base = ownerDocument.controller.url;
_resolvedSource = ownerDocument.controller.uriParser!.resolve(Uri.parse(base), Uri.parse(source));
```

```dart
// lib/src/html/head.dart:335 (LinkElement)   and   lib/src/html/head.dart:87-88 (@import)
String base = ownerDocument.controller.url;
// Base URL to resolve relative imports
String base = sheet.href ?? document.controller.url;
```

CSS goes through the same path (`lib/src/css/background.dart:366` — `final String base = baseHref ?? controller.url;`).

Terminology note for reviewers: the identifier `baseHref` appears widely in WebF's Dart source
(`css/font_face.dart`, `css/background.dart`, `dom/style_node_manager.dart`,
`bridge/from_native.dart:824`), but in every one of those places it means *"the URL of the stylesheet
that contained this declaration"*, used to resolve `@import` / `url()` relative to the sheet. It is
never the HTML `<base href>` element. Searching for `baseHref` therefore gives a misleading
impression that the feature exists.

Because `base` is not a registered element, `<base href="/x/">` is parsed as an unknown element and
simply contributes nothing.

### Expected vs actual

| | |
|---|---|
| **Expected** | Per HTML spec, the first `<base href>` in the document sets the document base URL; all document-relative URLs (`<a href>`, `<img src>`, `<script src>`, `<link href>`, and CSS `url()` in inline `<style>`) resolve against it. |
| **Actual** | `<base href>` has no effect whatsoever. Everything resolves against `controller.url`. No warning is logged. |

### Minimal reproduction

```html
<!doctype html>
<html>
<head>
  <base href="/prefix/">
</head>
<body>
  <img src="logo.png">
  <a id="a" href="page.html">link</a>
  <script>
    // In a browser: "/prefix/page.html". In WebF: resolved against controller.url.
    console.log(document.getElementById('a').href);
  </script>
</body>
</html>
```

Serve this from `http://host/index.html` and place the real asset at `/prefix/logo.png`. In a
browser the image loads and the logged href is `http://host/prefix/page.html`. In WebF the image
404s and the href resolves to `http://host/page.html`.

### Suggested fix

This is a feature gap rather than a one-line bug, so a full fix is nontrivial. A staged approach:

1. Add a `BaseElement` in `lib/src/html/head.dart` that records its `href` attribute on the document.
2. Add `Document.baseURI` returning the first `<base href>` resolved against `controller.url`,
   falling back to `controller.url` when absent.
3. Replace the ~6 hardcoded `String base = ownerDocument.controller.url;` sites
   (`a.dart:290`, `a.dart:34`, `img.dart:1092`, `script.dart:356`, `head.dart:335`, `head.dart:88`,
   `css/background.dart:366`) with `ownerDocument.baseURI` (keeping the stylesheet-specific
   `sheet.href` / `baseHref` precedence where it already exists — a stylesheet's own URL correctly
   takes priority over the document base for `url()` inside that sheet).

Even step 1+2 alone, plus a `console.warn`-equivalent when a `<base href>` is present but not yet
honored by a given call site, would remove the silent-failure aspect.

### Confidence

**Source-verified.** The absence of `BaseElement` and of `document.baseURI` was confirmed by grep
over the whole of `lib/src` in `0.24.27`. All quoted line numbers re-checked. The behavioural
consequence (relative URLs resolving against `controller.url`) follows directly from the quoted code
and does not depend on runtime observation.

---

## 3. `<script>` silently does not execute when QuickJS bytecode evaluation fails — the `isBytecode` branch has no fallback to source JS, unlike `to_native.dart`

**Title (one line):**
`_evaluateScriptBundle`'s `isBytecode` branch throws instead of falling back to raw JS, so a bytecode failure silently drops the whole `<script>`

### Environment

- webf `0.24.27` (pub.dev)
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0`
- Platform: Linux x86-64 container (glibc >= 2.38)

### Description

When QuickJS bytecode evaluation fails for a `<script>`, the script does not run at all. There is no
retry with the original JavaScript source. The page continues loading as if the script had never
existed: `DOMContentLoaded` still fires, no `error` event is dispatched on the element, and the only
trace is a `debugPrint` of `Bytecode are not valid to execute.` (see issue #4 about that message
carrying no attribution).

WebF already has the correct self-healing behaviour for this exact failure — one layer down, in
`to_native.dart`. The `script.dart` path is missing it.

### Root cause

**The failing path** — `lib/src/html/script.dart:65-69`:

```dart
} else if (bundle.isBytecode) {
  bool result = await evaluateQuickjsByteCode(contextId, bundle.data!, url: bundle.url, scriptElement: scriptElement);
  if (!result) {
    throw FlutterError('Bytecode are not valid to execute.');
  }
}
```

**The correct path, for comparison** — `lib/src/bridge/to_native.dart:392-402`:

```dart
if (QuickJSByteCodeCacheObject.cacheMode == ByteCodeCacheMode.DEFAULT &&
    cacheObject.valid &&
    cacheObject.bytes != null) {
  bool result =
      await evaluateQuickjsByteCode(contextId, cacheObject.bytes!, url: url, scriptElement: scriptElement);
  // If the bytecode evaluate failed, remove the cached file and fallback to raw javascript mode.
  if (!result) {
    await cacheObject.remove();
    // Fallback to normal script mode.
    return evaluateScripts(contextId, codeBytes, url: url, scriptElement: scriptElement);
  }
  return result;
}
```

**Why `script.dart` cannot simply copy that fallback: the source has already been destroyed.**

`lib/src/foundation/bundle.dart:135-143` rewrites the bundle in place:

```dart
Future<void> preProcessing(double contextId, {bool isModule = false}) async {
  if (isJavascript && data != null && isPageAlive(contextId)) {
    assert(isValidUTF8String(data!), 'JavaScript code is not UTF-8 encoded.');

    data = await dumpQuickjsByteCode(contextId, data!, url: _uri.toString(), isModule: isModule);

    _contentType = webfBc1ContentType;
  }
}
```

`data` is **overwritten** with the bytecode and `_contentType` is **overwritten** to
`webfBc1ContentType`. Since `isBytecode` is derived from `contentType`
(`bundle.dart:230`), a bundle that was fetched as ordinary `.js` now reports `isBytecode == true` and
its JavaScript source is unrecoverable.

`preProcessing` is invoked for `<script src>` on the preload path at `lib/src/html/script.dart:247-250`:

```dart
} else {
  if (!isModule) {
    await bundle.preProcessing(_contextId);
  }
```

So this is not a hypothetical: an ordinary external `<script src="app.js">` taking the preload path
reaches the `isBytecode` branch, and at that point the only copy of the source is gone. That is
structurally why the fallback is absent, and it means a fix must change the data flow rather than
just add an `if`.

**What the current error handling does instead** — `lib/src/html/script.dart:150-158`:

```dart
try {
  await _evaluateScriptBundle(_contextId, bundle, element, async: async, isModule: isModule);
} catch (err, stack) {
  debugPrint('$err\n$stack');
  _document.decrementDOMContentLoadedEventDelayCount();
  await WebFBundle.invalidateCache(bundle.url);
  dumper.recordScriptElementError(scriptSource, err.toString());
  return;
}
```

The HTTP cache *is* invalidated, so a subsequent page load may recover. But for **this** load the
script is simply skipped, and no `error` event is dispatched on the `<script>` element (contrast the
network-error path, which does `Timer.run` an event). A page whose bootstrap logic lives in that
script comes up blank-but-error-free — the hardest failure shape to attribute.

### Expected vs actual

| | |
|---|---|
| **Expected** | A bytecode failure transparently falls back to evaluating the original JavaScript source, matching `to_native.dart`. Failing that, an `error` event fires on the `<script>` element so the page can detect it. |
| **Actual** | The script never executes. `DOMContentLoaded` fires normally, no `error` event, only a `debugPrint`. |

### Minimal reproduction

Not reproducible from page content alone — this fires only when QuickJS rejects bytecode it
previously emitted (version skew in the cache, truncated/corrupted cache entry, or an
`evaluateQuickjsByteCode` failure inside the C++ binding layer). The C++ binding layer is not
published with the pub package, so we could not construct a deterministic trigger.

The code-level defect does not depend on reproducing it: the asymmetry against
`to_native.dart:392-402` is visible in the source, and the destroyed-source problem in
`bundle.preProcessing` is unconditional.

A maintainer with the full tree could force it by having `evaluateQuickjsByteCode` return `false`
unconditionally and loading any page with an external `<script src>` on the preload path.

### Suggested fix

Two options, in increasing order of invasiveness:

1. **Preserve the source across `preProcessing`.** Keep the pre-bytecode bytes in a new field (e.g.
   `Uint8List? _sourceBeforeBytecode`) in `bundle.dart:135-143`, then in
   `script.dart:65-69` fall back to `evaluateScripts(contextId, bundle.sourceBeforeBytecode!, ...)`
   when `evaluateQuickjsByteCode` returns `false`. Memory cost is one extra copy of the script
   source, freed in `dispose()`.
2. **Do not transform the bundle in place.** Have `preProcessing` return the bytecode as a separate
   value rather than overwriting `data` and `_contentType`, so `isBytecode` keeps meaning "this
   resource *was served as* bytecode" and the JS path stays available.

Independently, and cheaply: dispatch an `error` event on the `<script>` element in the
`script.dart:151` catch block, so failures are observable from page script instead of only in
`debugPrint`.

### Confidence

**Source-verified.** All quoted code and line numbers re-checked against `0.24.27`. The trigger
condition (what makes `evaluateQuickjsByteCode` return `false`) lives in the unpublished C++ layer
and is **not** verified — but the missing fallback and the in-place source destruction are both plain
in the Dart source.

Note on line numbers: an earlier internal note cited `script.dart:66` and `to_native.dart:397`; those
are the `evaluateQuickjsByteCode` call line and the explanatory comment line respectively. The full
blocks are `script.dart:65-69` and `to_native.dart:392-402`.

---

## 4. `Bytecode are not valid to execute.` carries no attribution — no URL, no underlying QuickJS exception, and it conflates "page torn down" with "bytecode invalid"

**Title (one line):**
Improve diagnostics for `Bytecode are not valid to execute.` — include the script URL and propagate the QuickJS exception

### Environment

- webf `0.24.27` (pub.dev)
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0`
- Platform: Linux x86-64 container (glibc >= 2.38)

### Description

When bytecode evaluation fails, the only diagnostic a WebF embedder sees is:

```
Bytecode are not valid to execute.
```

…followed by a Dart stack trace of WebF's own internals. It does not say **which script** failed, and
it does not carry the **QuickJS-level reason**. On a page with a dozen `<script src>` tags this is
close to unactionable: the embedder cannot even determine which file to look at.

This is filed separately from issue #3 (the missing fallback), because it is worth fixing on its own:
even after a fallback exists, the diagnostic will still be the only signal when the fallback is
skipped or also fails, and the fix here is small and self-contained.

### Root cause

**The URL is in scope and simply not interpolated** — `lib/src/html/script.dart:65-69`:

```dart
} else if (bundle.isBytecode) {
  bool result = await evaluateQuickjsByteCode(contextId, bundle.data!, url: bundle.url, scriptElement: scriptElement);
  if (!result) {
    throw FlutterError('Bytecode are not valid to execute.');
  }
}
```

`bundle.url` is passed into `evaluateQuickjsByteCode` on the line directly above, yet the error
message omits it. Note the sibling `else` branch three lines down *does* interpolate a URL —
`throw FlutterError('Unknown type for <script> to execute. $url');` (`script.dart:71`) — so the
inconsistency is local and obvious.

**The QuickJS exception never reaches Dart.** `evaluateQuickjsByteCode` collapses everything to a
`bool`, and the native callback receives only an `int` — `lib/src/bridge/to_native.dart:470-477`:

```dart
void handleEvaluateQuickjsByteCodeResult(Object handle, int result) {
  _EvaluateQuickjsByteCodeContext context = handle as _EvaluateQuickjsByteCodeContext;
  malloc.free(context.bytes);
  if (context.url != nullptr) {
    malloc.free(context.url);
  }
  context.completer.complete(result == 1);
}
```

The `url` pointer is carried all the way into native code and then freed on the way back without ever
being used for reporting.

**`false` is overloaded** — `lib/src/bridge/to_native.dart:479-483`:

```dart
Future<bool> evaluateQuickjsByteCode(double contextId, Uint8List bytes,
    {String? url, ScriptElement? scriptElement}) async {
  if (WebFController.getControllerOfJSContextId(contextId) == null) {
    return false;
  }
```

A torn-down page returns the same `false` as genuinely invalid bytecode, and therefore produces the
same `Bytecode are not valid to execute.` message. During navigation or widget disposal this is a
routine, benign condition being reported with alarming and misleading wording — which in turn trains
embedders to ignore the message, so the real failures get ignored too.

The same overloading exists at `to_native.dart:505-507` for `evaluateModule`.

### Expected vs actual

| | |
|---|---|
| **Expected** | An error identifying the script URL, distinguishing "page no longer alive" from "bytecode rejected", and including the QuickJS exception message where one exists. |
| **Actual** | A fixed 6-word string with no URL, no reason, and no way to tell a benign teardown from a real failure. |

### Minimal reproduction

Same limitation as issue #3: triggering a genuine bytecode rejection requires the unpublished C++
binding layer. The *teardown* variant, however, is reachable from Dart alone — dispose a
`WebFController` while an external `<script src>` is mid-execution and the same message appears.

The defect here is a code-quality/diagnostics one and is fully visible in the quoted source without
any reproduction.

### Suggested fix

Cheap, in three parts:

1. Interpolate the URL, matching the sibling branch at `script.dart:71`:

```diff
-        throw FlutterError('Bytecode are not valid to execute.');
+        throw FlutterError('Bytecode is not valid to execute. url=${bundle.url}, bytes=${bundle.data!.length}');
```

2. Distinguish teardown from failure. Return a small result type (or throw a distinct sentinel) from
   `evaluateQuickjsByteCode` for the `getControllerOfJSContextId(contextId) == null` case, so callers
   can log it at `fine`/`debug` level instead of as an error.

3. Propagate the QuickJS reason. Widen the native callback so the C++ side can pass back an error
   string (the pattern already exists for other bridge calls that surface exception text), and include
   it in the thrown `FlutterError`.

Part 1 alone would have saved substantial investigation time on our side and is a one-line change.

### Confidence

**Source-verified** for everything quoted (all line numbers re-checked against `0.24.27`). The claim
that the C++ side *has* an exception message available to propagate is an assumption — that layer is
not published with the pub package — but parts 1 and 2 of the suggested fix are entirely within the
Dart tree and do not depend on it.

---

## 5. `document.documentElement.dataset` is `null` inside a blocking `<head>` script, so any `dataset` write there throws `TypeError` and aborts the whole script

**Title (one line):**
`document.documentElement.dataset` is `null` during blocking `<head>` script execution — `TypeError` aborts the script before any of it runs

### Environment

- webf `0.24.27` (pub.dev)
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0`
- Platform: Linux x86-64 container (glibc >= 2.38)

### Description

A synchronous, non-`defer` `<script>` in `<head>` that touches
`document.documentElement.dataset` throws:

```
TypeError: cannot read property 'xxx' of null
```

Because the idiomatic shape for such a bootstrap script is a single IIFE, the throw aborts **the
entire script**. Everything the IIFE was supposed to install — in our case a `window.<Namespace>`
global that the rest of the page depends on — never comes into existence.

The resulting failure mode is deceptive: the page loads, later scripts run, and the only symptom is
that a global is mysteriously `undefined`. Nothing points at `dataset`. On our side this cost several
rounds of investigation; four separate hypotheses were tested and all passed before page-level
introspection finally located it.

`dataset` is not broken in general in WebF — our page code reads `dataset` off event targets
(`e.target.dataset.id`) successfully once the DOM is interactive. The problem is specific to
`documentElement` during the `<head>`-blocking phase.

The classic write to signal a theme before first paint is exactly this pattern, which is why it is
worth fixing rather than documenting:

```js
document.documentElement.dataset.theme = 'dark';
```

### Root cause

**Not determined.** `dataset` does not appear anywhere in the published Dart source
(`grep -rn dataset lib/src --include=*.dart` in `webf-0.24.27` → **zero hits**), so it is implemented
entirely in the C++ binding layer, which is not shipped with the pub package. We could not inspect it.

What we can say from the Dart side is that `documentElement` is created as part of a distinct
initialisation step and is nullable throughout WebF's own code — e.g.
`lib/src/launcher/controller.dart:1115-1124`:

```dart
// Initialize document, window and the documentElement.
flushUICommand(view, nullptr);
...
// Manually initialize the root element and create renderObjects for each elements.
view.document.documentElement!
    .applyStyle(view.document.documentElement!.style);
```

(the same block again at `controller.dart:1261-1269`, and `controller.dart:1172` on a third path).
Every access is a force-unwrap `documentElement!`, and WebF's own DevTools modules guard with
`document.documentElement == null` checks (`devtools/cdp_service/modules/dom.dart:365`,
`modules/page.dart:318`, `modules/css.dart:313`).

This is consistent with — but does **not** prove — a hypothesis that the root element's
`dataset`-backing structure is not yet materialised at the point blocking `<head>` scripts execute,
while `documentElement` itself is already a live object. A maintainer with the C++ tree should be able
to confirm or refute quickly.

### Expected vs actual

| | |
|---|---|
| **Expected** | Per spec, `document.documentElement.dataset` is a live `DOMStringMap` and is never `null`. Reads return `undefined` for absent attributes; writes set `data-*`. |
| **Actual** | It is `null` during blocking `<head>` script execution, so both reads and writes throw `TypeError` and abort the script. |

### Minimal reproduction

```html
<!doctype html>
<html>
<head>
<script>
  (function () {
    // Throws in WebF 0.24.27: TypeError, cannot read property of null.
    console.log('dataset is:', document.documentElement.dataset);
    document.documentElement.dataset.theme = 'dark';
    window.MARKER = 'installed';
  })();
</script>
</head>
<body>
  <script>
    // In a browser: "installed". In WebF: undefined — the head IIFE aborted.
    console.log('MARKER =', window.MARKER);
  </script>
</body>
</html>
```

This is genuinely minimal and self-contained. Note that seeing the `TypeError` at all requires the
embedder to forward page JS errors to the host log (`onJSError` / equivalent); without that
forwarding the script simply vanishes with no output, which is how this stayed hidden for so long.

For comparison, both of these work in the same position, which is the workaround we adopted:

```js
document.documentElement.setAttribute('data-theme', 'dark');
document.documentElement.getAttribute('data-theme');
```

### Suggested fix

Make `documentElement.dataset` return an empty-but-valid `DOMStringMap` rather than `null` at every
point where `documentElement` is exposed to script. If there is a genuine ordering constraint that
makes the backing store unavailable during the `<head>` phase, the lazy-initialisation should happen
on first `dataset` access rather than leaving the property `null`.

Secondary, and useful independent of the fix: as a general invariant, no DOM property that is
non-nullable per the HTML spec should be observable as `null` from script.

### Confidence

**Container-observed; not explained at the source level.** The behaviour was reproduced reliably in
our Linux container and the workaround (`setAttribute`/`getAttribute`) confirms the diagnosis of
`dataset` specifically. The root cause is **not** verified — `dataset` has no presence in the
published Dart source, so this report deliberately stops at "here is a minimal repro" rather than
proposing a patch location. The `documentElement` initialisation code quoted above is offered as a
lead, not as the confirmed cause.

---

## 6. `<input type="range">` causes its entire line to paint nothing — sibling text and the line's own `background` disappear, while the box model is correct

**Title (one line):**
`<input type="range">` is unimplemented AND paints nothing at all: the whole containing line renders blank (siblings + background) despite correct layout geometry

### Environment

- webf `0.24.27` (pub.dev)
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0`
- Platform: Linux x86-64 container (glibc >= 2.38)

### Description

Two distinct problems, one of which we can fully explain and one of which we cannot. They are filed
together because the second is only reachable through the first.

**(a) `type="range"` is not implemented** and silently degrades to a text input. This part is
explained by the source.

**(b) The actual observed behaviour is far worse than (a): nothing on that line paints at all.**
There is no slider, no text field — and critically, **the sibling text on the same line and the
line's own `background` also disappear.** The line renders as blank space.

Meanwhile **the box model is entirely correct.** Measured in our container:

```
tagRect = 118 x 40      (the <input type=range> element)
rawRect = 120 x 24
```

Layout produced correct, non-zero geometry. So this is **not** a layout collapse — it is purely a
painting failure. We could not explain it, and say so plainly below.

Why (b) matters much more than (a): the author's symptom is *"one row of my UI is mysteriously
blank."* There is no error, no console output, and no suspicious element in the tree — the geometry
inspects as correct. Attributing this is an order of magnitude harder than noticing "my slider turned
into a text box." A missing slider is a cosmetic gap; a silently blanked row of content is a
correctness bug that will burn hours for anyone who hits it.

### Root cause

**For (a) — fully verified.** `lib/src/html/form/input.dart:250-268`, `FlutterInputElementState.build`:

```dart
@override
Widget build(BuildContext context) {
  switch (widgetElement.type) {
    case 'radio':
      return createRadio(context);
    case 'checkbox':
      return createCheckBox(context);
    case 'button':
    case 'submit':
      return createButton(context);
    case 'date':
    // case 'month':
    // case 'week':
    case 'time':
      return createTime(context);
    default:
      return createInput(context);
  }
}
```

There is no `case 'range'`, so it falls to `default` → `createInput` (`input.dart:270-279`), which —
after the `'hidden'` special case — returns `createInputWidget(context)`, i.e. a Flutter `TextField`.

Confirming the gap is total: `grep -rn "'range'" lib/src --include=*.dart` in `0.24.27` returns only
`devtools/cdp_service/modules/css.dart:867`, `:935` and `bridge/code_gen/blink_css_ids.dart:271` —
all unrelated to input types. There is no range slider implementation anywhere, and `initState`
(`input.dart:236-247`) likewise has no `range` case.

**For (b) — OPEN QUESTION. We could not explain it.**

The source above says a `TextField` should be built and painted. Our container observation is that
**zero pixels are painted on that line**. We do not know why, and we are deliberately not guessing in
this report. Two leads a maintainer may want to check first:

1. `lib/src/rendering/widget.dart:762-775` — `RenderWidget.paint` has a `disableBoxModelPaint`
   escape hatch that skips `paintBoxModel` entirely:

   ```dart
   void paint(PaintingContext context, Offset offset) {
     if (renderStyle.target is WidgetElement) {
       final widgetElement = renderStyle.target as WidgetElement;
       if (widgetElement.disableBoxModelPaint) {
         performPaint(context, offset);
         return;
       }
     }
     paintBoxModel(context, offset);
   }
   ```

   This could account for the element's own decoration going missing — but **not** for the sibling
   text vanishing, so it is at best a partial lead.

2. The sibling text and the line background are painted by the **containing line box**, not by the
   input. Their simultaneous disappearance is more consistent with an exception thrown partway
   through the enclosing inline formatting context's paint, aborting the remainder of that line's
   paint. If Flutter swallows or only `debugPrint`s that exception, it would match the observed
   total absence of diagnostics. Relevant machinery:
   `lib/src/rendering/inline_items_builder.dart:198-200` treats widget-backed children as atomic
   inline placeholders (`isSelfRenderWidget()`), and `lib/src/rendering/inline_formatting_context.dart`
   drives the line paint.

We did not have the C++ binding layer or a debug WebF build available to confirm either, and are not
claiming a root cause for (b).

### Expected vs actual

| | |
|---|---|
| **Expected** | `<input type="range">` renders a slider. Failing that, it at minimum renders *something* and does not affect any other element. |
| **Actual (a)** | Not implemented; source path leads to a `TextField`. |
| **Actual (b)** | Nothing paints on that line at all — no input, and sibling text plus the line's own `background` are also gone. Box model geometry is correct (`tagRect=118x40`, `rawRect=120x24`). No error, no log output. |

### Minimal reproduction

The yellow-background test isolates (b) from (a) cleanly — if any yellow appears, (b) is not
reproducing; a fully blank row means it is:

```html
<!doctype html>
<div style="background: yellow; padding: 8px;">
  before
  <input type="range" min="0" max="100" value="50">
  after
</div>
<div style="background: #cfc; padding: 8px;">control row with no range input</div>
```

In a browser: a yellow row reading `before [slider] after`, then a green row.
In WebF 0.24.27: the **entire yellow row renders blank** — no yellow, no `before`, no `after` — while
the green control row renders normally.

This is minimal and self-contained. Our own detection of it is pinned as probe group 14b in an
internal verification harness using exactly this dye-the-row technique.

### Suggested fix

For (a): add a `case 'range':` to the switch at `input.dart:252` returning a Flutter `Slider` wired to
the element's `min` / `max` / `step` / `value` attributes, plus a matching `initState` case at
`input.dart:237`.

For (b): we cannot responsibly propose a patch without knowing the cause. However, **(b) is the more
important of the two to triage**, because it is not confined to `range` — if the mechanism is "an
exception during an atomic-inline child's paint aborts the whole line's paint", then any future
unimplemented or throwing widget-backed inline element will silently blank a row of unrelated content.
That is worth understanding independently of range sliders.

A defensive guard — wrapping the per-child paint of atomic inline widget children so that one child's
paint failure cannot abort its siblings, and logging when it trips — would contain the blast radius
regardless of the specific trigger.

### Confidence

- **(a): source-verified.** `input.dart:250-268` re-checked against `0.24.27`; the absence of any
  `range` handling in `lib/src` confirmed by grep. (An earlier internal note cited `input.dart:251-268`
  for "the `createInput` switch"; the switch is at `:252-267` inside `build` at `:251`, and
  `createInput` itself is at `:270-279`. Substance unchanged.)
- **(b): container-observed only; NOT explained at the source level.** Reproduced reliably, with
  geometry measured, but the published Dart source says a `TextField` should paint and we cannot
  reconcile that with the observation. The two leads above are explicitly hypotheses, not findings.
  **This gap is an open question we are asking the maintainers to help close, not a diagnosis.**

---

## 7. Grid layout treats `position: sticky` children as out-of-flow — they occupy no grid cell and do not contribute to track sizing (flow layout gets this right)

**Title (one line):**
`position: sticky` grid items are excluded from placement and track sizing: `_isPositionedGridChild()` lumps `sticky` together with `absolute`/`fixed`

### Environment

- webf `0.24.27` (pub.dev)
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0`
- Platform: Linux x86-64 container (glibc >= 2.38)

### Description

A `position: sticky` child of a `display: grid` container is treated as out-of-flow. Consequently it:

- **occupies no grid cell** — the item that would have followed it takes its place, shifting the
  entire remaining layout by one cell;
- **does not contribute to intrinsic track sizing** — a sticky item in a column whose track is `auto`
  or `min-content`/`max-content` sized is invisible to the sizing algorithm, so the column comes out
  too narrow.

Per CSS, `position: sticky` is explicitly **in-flow**: a sticky box participates in layout exactly
like `position: relative` and is merely offset at paint/scroll time
([CSS Position 3, §3.5](https://www.w3.org/TR/css-position-3/#sticky-pos)). Only `absolute` and
`fixed` are out-of-flow.

**This is a grid-specific defect, not a general lack of sticky support in WebF** — WebF's own flow
layout gets it right (contrast below). And because sticky *offsetting* is still applied on the grid
path, the failure takes the most confusing possible shape: **the element visibly sticks, so sticky
appears to be working**, while the column widths are silently wrong and the stuck element overlaps a
row of content. An author debugging this has no reason to suspect grid participation.

### Root cause

**The predicate conflates sticky with absolute/fixed** — `lib/src/rendering/grid.dart:347-351`:

```dart
bool _isPositionedGridChild(RenderBox child) {
  final RenderStyle? style = _unwrapGridChildStyle(child);
  if (style == null) return false;
  return style.isSelfPositioned() || style.isSelfStickyPosition();
}
```

where (`lib/src/css/render_style.dart:1160-1168`):

```dart
@pragma('vm:prefer-inline')
bool isSelfPositioned() {
  return position == CSSPositionType.absolute || position == CSSPositionType.fixed;
}

@pragma('vm:prefer-inline')
bool isSelfStickyPosition() {
  return position == CSSPositionType.sticky;
}
```

Note that `isSelfPositioned()` is already correct on its own — it means exactly "out-of-flow". The bug
is the `|| style.isSelfStickyPosition()` disjunction added on top of it.

**That predicate drives 12 exclusion sites across `grid.dart`:**

```
grid.dart:385    grid.dart:471    grid.dart:872    grid.dart:913
grid.dart:2250   grid.dart:3052   grid.dart:3077   grid.dart:3363
grid.dart:3643   grid.dart:3925   grid.dart:4034   grid.dart:4477
```

Two of these are the load-bearing ones:

- **`grid.dart:2248-2251` — building the grid item list itself**, i.e. the placement pass:

  ```dart
  // Pass 1: resolve placements and grow implicit track lists.
  final List<RenderBox> placementChildren = _collectChildren()
      .where((RenderBox child) => !_isPositionedGridChild(child))
      .toList(growable: false);
  ```

  A sticky child is never a placement candidate, hence occupies no cell.

- **`grid.dart:3052` and `grid.dart:3363` — intrinsic width computation**, hence sticky children do
  not size `auto` / `min-content` / `max-content` column tracks.

**A 13th site inlines the same predicate rather than calling it** — `grid.dart:1947-1968`:

```dart
// Out-of-flow positioned children (absolute/fixed/sticky) do not participate in
// grid item placement or track sizing. They are laid out and positioned after
// the grid container size and placeholder static positions are known.
final List<RenderBoxModel> positionedChildren = <RenderBoxModel>[];
final List<RenderBoxModel> stickyChildren = <RenderBoxModel>[];
final List<RenderBoxModel> absFixedChildren = <RenderBoxModel>[];
RenderBox? positionedScan = firstChild;
while (positionedScan != null) {
  final GridLayoutParentData pd = positionedScan.parentData as GridLayoutParentData;
  if (positionedScan is RenderBoxModel) {
    final RenderStyle? style = _unwrapGridChildStyle(positionedScan);
    if (style != null && (style.isSelfPositioned() || style.isSelfStickyPosition())) {
      positionedChildren.add(positionedScan);
      if (style.isSelfStickyPosition()) {
        stickyChildren.add(positionedScan);
      } else {
        absFixedChildren.add(positionedScan);
      }
    }
  }
  positionedScan = pd.nextSibling;
}
```

The comment on `:1947` states the intent explicitly, so this is a deliberate design decision rather
than an oversight — but the decision is not what CSS specifies.

**The comment's stated mitigation does not exist.** `grid.dart:1970-1974`:

```dart
// Pre-layout sticky positioned children so their placeholders can reserve
// correct space during the subsequent grid layout pass.
for (final RenderBoxModel sticky in stickyChildren) {
  CSSPositionedLayout.layoutPositionedChild(this, sticky);
}
```

The word `placeholder` occurs in `grid.dart` at **exactly three lines — 1949, 1970 and 4583 — all of
them comments.** `grep -in placeholder lib/src/rendering/grid.dart` returns nothing else. There is no
placeholder mechanism in the grid path at all, so nothing reserves the space the comment promises.
`layoutPositionedChild` lays the child out as an out-of-flow box; it does not insert it into the grid.

**Why the failure looks like success** — `grid.dart:4586-4593`:

```dart
for (final RenderBoxModel child in absFixedChildren) {
  CSSPositionedLayout.layoutPositionedChild(this, child);
}

for (final RenderBoxModel child in positionedChildren) {
  CSSPositionedLayout.applyPositionedChildOffset(this, child);
  CSSPositionedLayout.applyStickyChildOffset(this, child);
}
```

`applyStickyChildOffset` is still invoked on the grid path, so the element **does** stick to the
scroll container. The author sees sticky behaviour working and therefore has no reason to suspect that
the same element has been removed from grid placement and track sizing.

**Contrast: WebF's flow layout is correct.** `lib/src/rendering/flow.dart` excludes children from flow
using `isSelfPositioned()` **only**, with no sticky term:

```dart
// lib/src/rendering/flow.dart:425
if (child is RenderBoxModel && child.renderStyle.isSelfPositioned()) {

// lib/src/rendering/flow.dart:1212
if (childRenderBoxModel.renderStyle.isSelfPositioned()) {

// lib/src/rendering/flow.dart:1342
if (qualifies && !rs.isParentRenderFlexLayout() && !rs.isSelfPositioned()) {
```

So under block/flow layout, sticky children correctly remain in flow and occupy space — matching CSS.
The two layout engines disagree with each other, which is itself strong evidence for which one is
wrong.

### Expected vs actual

| | |
|---|---|
| **Expected** | A `position: sticky` grid child is an in-flow grid item: it is placed in a cell by the auto-placement algorithm, contributes to intrinsic track sizing, and is merely offset at scroll time. |
| **Actual** | Excluded from placement (`grid.dart:2250`) and from intrinsic sizing (`grid.dart:3052`, `:3363`). Subsequent items shift up by one cell; `auto`/`min-content`/`max-content` tracks come out too narrow. Sticky offsetting is still applied (`grid.dart:4592`), so it looks like sticky works. |

### Minimal reproduction

```html
<!doctype html>
<div style="display: grid; grid-template-columns: auto auto; gap: 4px; border: 2px solid black;">
  <div style="background:#fcc">A</div>
  <div style="background:#cfc; position: sticky; top: 0;">B-sticky-wide-content</div>
  <div style="background:#ccf">C</div>
  <div style="background:#ffc">D</div>
</div>
```

Expected (browser): a 2x2 grid, `A B / C D`, and the second column is wide enough for
`B-sticky-wide-content` because that item sizes the `auto` track.

Actual (WebF 0.24.27): `B` is not placed, so `C` lands in row 1 column 2 and `D` in row 2 column 1 —
the grid reads `A C / D` with `B` overlaid out-of-flow. The second column is sized only by `C`, so it
is far too narrow for `B`'s text.

This is minimal and self-contained; no scrolling is required to observe the placement and sizing
errors (scrolling is only needed to observe that the offsetting still works, which is the part that
masks the bug).

### Suggested fix

The core change is one line — `grid.dart:347-351`:

```diff
 bool _isPositionedGridChild(RenderBox child) {
   final RenderStyle? style = _unwrapGridChildStyle(child);
   if (style == null) return false;
-  return style.isSelfPositioned() || style.isSelfStickyPosition();
+  return style.isSelfPositioned();
 }
```

This aligns the grid path with both CSS and `flow.dart`, and fixes all 12 call sites at once.

The inlined duplicate at `grid.dart:1958` must be updated in the same commit, otherwise sticky
children remain in `positionedChildren` / `stickyChildren` and get laid out twice — once as grid items
and once as out-of-flow boxes:

```diff
-        if (style != null && (style.isSelfPositioned() || style.isSelfStickyPosition())) {
+        if (style != null && style.isSelfPositioned()) {
```

That also makes the `stickyChildren` list and the pre-layout loop at `:1970-1974` dead code, and they
should be removed along with the now-inaccurate comments at `:1947-1949` and `:1970-1971`. Sticky
offsetting at `:4590-4593` should then be applied to in-flow grid items rather than to
`positionedChildren` — i.e. `applyStickyChildOffset` needs to iterate the placed grid items. (Note that
the current loop calls `applyStickyChildOffset` on abs/fixed children too, which is already a
no-op-at-best.)

Ideally the reduced predicate would also be renamed (e.g. `_isOutOfFlowGridChild`) so the next reader
does not re-add the sticky term.

### Confidence

**Source-verified, highest confidence of the set.** Every line number in this report was
re-checked against `webf-0.24.27` in pub-cache and all matched exactly: the predicate at `:347-351`,
all 12 call sites, the inlined duplicate at `:1958`, the three `placeholder`-only comment lines
(`:1949`, `:1970`, `:4583`), the sticky offsetting at `:4592`, the `render_style.dart:1160-1168`
definitions, and the three `flow.dart` contrast sites (`:425`, `:1212`, `:1342`). The
`flow.dart`-vs-`grid.dart` divergence and the nonexistent placeholder mechanism are both directly
demonstrable from the source and need no runtime observation.

---

## 8. Flex: base size of `RenderWidget` / nested-flex items measures as the container width — items wrap onto their own line, and nested-flex subtrees can end up never painted at all (the existing §9.2 workaround only covers `RenderFlowLayout`)

**Title (one line):**
`flex-wrap: wrap`: flex base size of WidgetElement and nested-flex children resolves to the container width — each item wraps onto its own line, and nested-flex subtrees are silently left unpainted

### Environment

- webf `0.24.27` (pub.dev), `webf_cupertino_ui` `0.4.1`
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0`
- Platform: macOS 15 (arm64), observed in a release-mode host app

### Description

In a `display: flex; flex-wrap: wrap` row container, children whose render object is a
`RenderWidget` (any `WidgetElement` — e.g. `<flutter-cupertino-button>`, or WebF's own
`<webf-list-view>`) or a nested flex container (`RenderFlexLayout`) get a **flex base size equal to
the flex container's available width**. The wrap algorithm therefore decides that the second child
does not fit, and **every child ends up on a line of its own, stretched to the full line width**.

Authors see a horizontal toolbar collapse into a vertical stack of full-width buttons. Nothing is
logged, and the same markup is correct in every browser.

WebF **already recognises this class of bug** and carries a targeted workaround, whose comment
states the problem precisely (`lib/src/rendering/flex.dart:5331-5340`):

```dart
// CSS Flexbox §9.2: For flex-basis:auto with an auto main-size, the flex base size
// should come from the item's max-content contribution in the main axis, not from
// the block formatting context's "fill-available" used size. Our intrinsic pass can
// mistakenly inherit a container-bounded width for block-level items that establish
// an inline formatting context (IFC), causing the base size to equal the container
// width. Detect that case and prefer the IFC's max-intrinsic width instead.
if (isHorizontal && child is RenderFlowLayout) {
```

The guard is `child is RenderFlowLayout`, so the correction never applies to `RenderWidget` or
`RenderFlexLayout` children, even though they reach the same wrong base size by the same route.

### Root cause

The base size is not a max-content measurement at all — it is **the width the child happens to take
when laid out once with relaxed constraints** (`flex.dart:5470-5474`):

```dart
childSize = Size.copy(_getChildSize(child)!);
_transientChildSizeOverrides![child] = childSize;
intrinsicMain = isHorizontal ? childSize.width : childSize.height;
```

For a `WidgetElement` child the chain closes as follows:

1. `_getIntrinsicConstraints` relaxes `maxWidth` to infinity for non-replaced auto-width items
   (`flex.dart:2461-2482`). A `WidgetElement` is **not** replaced — `isSelfRenderReplaced()` tests
   `attachedRenderBoxModel is RenderReplaced` (`css/render_style.dart:1131-1133`) while a
   `WidgetElement` attaches a `RenderWidget` (`rendering/widget.dart:17`).
2. `RenderWidget._layoutChild` refuses to pass that unbounded width down and **clamps it back to the
   viewport width** (`rendering/widget.dart:169-187`); `allowsInfiniteWidth` defaults to `false`
   (`widget/widget_element.dart:119`). The clamp is deliberate — the comment notes an unbounded
   width would crash a hosted `Row` + `Expanded`.
3. The hosted content element is block-level (e.g. `display: flex`), so with `width: auto` it
   fill-availables to that clamped width; only `inline-block` / `inline-flex` / `inline-grid` /
   `inline` leave `logicalWidth == null` and shrink-to-fit
   (`css/render_style.dart:3425-3432`).
4. `RenderWidget` then shrink-wraps to that child: `size = getBoxSize(childSize)`
   (`rendering/widget.dart:268`).
5. Back in flex, `intrinsicMain == viewport width`, so
   `isExceedFlexLineLimit` (`flex.dart:5476-5480`) is true from the second child onward.

A nested flex container (`RenderFlexLayout`) reaches step 3's fill-available result directly, without
the `RenderWidget` detour — and is likewise skipped by the `RenderFlowLayout` guard.

### Expected vs actual

Given a 900px-wide container:

```html
<div style="display:flex; flex-wrap:wrap; gap:8px">
  <flutter-cupertino-button><span style="display:flex">Select all</span></flutter-cupertino-button>
  <flutter-cupertino-button><span style="display:flex">Refresh</span></flutter-cupertino-button>
</div>
```

- **Expected** (and what every browser does with a `<button>` in place of the custom element): both
  buttons on one line, each as wide as its own content.
- **Actual**: two lines, each button stretched to ~900px.

### Minimal reproduction

1. Register any `WidgetElement` (the bug needs no third-party package — WebF's own
   `<webf-list-view>` reproduces it, as does a two-line custom element).
2. Put two of them in a `display:flex; flex-wrap:wrap` container, each wrapping one block-level
   child (`<span style="display:flex">text</span>`).
3. Observe one item per line, each at full container width.
4. Change the inner `display: flex` to `inline-flex` — the layout becomes correct, confirming step 3
   of the chain.
5. Remove `flex-wrap: wrap` — items stay on one line but are now sized by `flex-shrink` from the
   inflated base size, i.e. all equal width rather than content width. This confirms the base size
   itself is wrong, independently of the wrap decision.

### Second, more severe symptom: the nested-flex subtree is never painted at all

The same measurement path has a worse consequence than bad wrapping. When a flex container's child
is itself a flex container, that child's subtree can end up **permanently unpainted** — the
container draws its own border/background, and none of its descendants appear.

There is no exception and no log line, and the DOM side looks perfectly healthy:
`getBoundingClientRect()` returns correct, non-zero, correctly-positioned boxes for every element in
the invisible subtree. The only way we could detect it was by counting non-background pixels in a
screenshot: the region occupied by one such container held **184** non-background pixels (all of them
its own 1px `border-bottom`) and rose to several thousand once the nested flex was removed.

The mechanism on the Flutter side is `RenderObject._paintWithContext`:

```dart
// If we still need layout, then that means that we were skipped in the layout
// phase and therefore don't need painting.
if (_needsLayout) {
  return;
}
```

So any render object left in `needsLayout` is silently skipped during paint. Independent
corroboration from the same session: a single `editable.dart:2018` assertion,
`RenderEditable.handleEvent` → `assert(!debugNeedsLayout)`, i.e. a pointer-down was dispatched to a
hosted text field whose layout had not completed — proving that render objects really are stuck in
`needsLayout`, not merely mispositioned.

Correlation observed across one page (every container tested, no exceptions):

| container | child | painted? |
|---|---|---|
| `display:flex; flex-direction:column` | block `div` | **no** |
| `display:flex; flex-direction:column` | `display:flex; flex-direction:column` | **no** |
| `display:flex` (row, `wrap`) | `display:flex; flex-direction:column` | **no** |
| `display:flex` (row, `wrap`) | `WidgetElement` / `span` | yes |
| block | `<webf-list-view>` | yes |

Our workaround was to remove nested flex entirely — every vertical stack became block flow
(`display:block` + `margin`) instead of `flex-direction:column` + `gap`. Visually identical, and it
never enters the trial-layout path.

### Suggested fix

The narrow fix is to extend the §9.2 correction beyond `RenderFlowLayout`:

```diff
-          if (isHorizontal && child is RenderFlowLayout) {
+          if (isHorizontal && child is RenderBoxModel) {
```

with the candidate width taken from `child.getMaxIntrinsicWidth(double.infinity)` for children that
have no `inlineFormattingContext` — the fallback branch at `flex.dart:5356-5362` already does exactly
this, so the change is mostly about widening the type test. `RenderWidget` implements
`computeMaxIntrinsicWidth` by forwarding to its primary child (`rendering/widget.dart:596-618`), so a
usable max-content value is available.

The structural fix is to stop deriving flex base size from a trial layout under relaxed constraints,
and use the intrinsic (`getMaxIntrinsicWidth`) contribution as the spec requires. That is a larger
change, but the current approach is why the same defect keeps reappearing per render-object type.

**Workaround for authors** (what we shipped): give the inner content layer of a `WidgetElement`
`display: inline-flex` so it shrink-to-fits, and give nested flex items an explicit `width`. Note
that `flex-basis` does **not** work as a substitute — the relaxation test at `flex.dart:2461` reads
`width.isAuto` only. And `max-width` must be avoided on a `WidgetElement`: it flips
`hasExplicitInlineWidth` (`rendering/widget.dart:103-104`), which routes `_layoutChild` through the
inline-block branch and forwards the unbounded `maxWidth` into the hosted Flutter subtree — the very
crash the clamp in step 2 exists to prevent.

### Confidence

**Source-verified end to end, plus a real-device before/after.** Every step of the chain was read in
`webf-0.24.27` in pub-cache. The fix was verified on macOS by pixel-measuring screenshots (the host's
JS-console forwarding is unavailable on the cached-controller path, so no runtime numbers could be
logged from inside the page): after switching the content layer to `inline-flex`, three filter items
measured 170/170/169 CSS px against a declared `width: 170px`, the grow item 244 px, and three
buttons 51/76/107 px — their own content widths — all on one line.

Not verified: whether the same defect occurs with `flex-direction: column` + `wrap` (the relaxation
code at `flex.dart:2484+` is symmetric, so it likely does), and whether `RenderReplaced` children
(e.g. `<img>`) are affected — they are explicitly excluded from the relaxation, so probably not.

---

## 9. Hit test reaches a hosted text field whose text layout was invalidated in the same frame — `Text layout not available`, then `!_debugDuringDeviceUpdate` floods every frame

**Title (one line):**
`MouseTracker` hit test on a `WidgetElement`-hosted `TextField` asserts `Text layout not available` after the text was set in the same frame

### Environment

- webf `0.24.27` (pub.dev), `webf_cupertino_ui` `0.4.1`
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0`
- Platform: macOS 15 (arm64), **debug build**

### Description

When a hosted text field's text is set programmatically (a controlled input whose value arrives
from an async load), and the pointer happens to rest over the WebF viewport in that same frame,
`MouseTracker`'s hit test walks into the field's `RenderEditable` before its `TextPainter` has been
re-laid-out, and trips `assert(_debugAssertTextLayoutIsValid)`.

The first exception is survivable, but it aborts the hit-test pass **while
`MouseTracker._debugDuringDeviceUpdate` is still set**, so every subsequent frame throws
`'!_debugDuringDeviceUpdate': is not true` — hundreds of lines per second — and the app renders
blank from that point on.

This reproduces with both `<flutter-cupertino-input>` (`webf_cupertino_ui`) and WebF's own
`<input>`, because the latter is also a `WidgetElement` hosting a Flutter `TextField`
(`lib/src/html/form/base_input.dart:615`) — i.e. the same `RenderEditable` underneath.

### Root cause

The invalidation and the hit test happen in the same frame, in this order:

```
_Editable.updateRenderObject (editable_text.dart:6133)
  → RenderEditable.text=            (editable.dart:789)
    → TextPainter.markNeedsLayout   (text_painter.dart:787)   // layout cache dropped
...
MouseTracker.updateAllDevices       (mouse_tracker.dart:367)
  → RendererBinding.hitTestInView   (binding.dart:676)
    → RenderWidget.hitTestChildren  (webf/src/rendering/widget.dart:907)
      → RenderEventListener.hitTest (webf/src/rendering/event_listener.dart:152)
        → _RenderBaselineAlignedStack.hitTestChildren (cupertino/text_field.dart:1947)
          → RenderEditable.hitTestChildren            (editable.dart:1990)
            → TextPainter.getClosestGlyphForOffset    (text_painter.dart:1688)
              → assert(_debugAssertTextLayoutIsValid) // throws
```

In a plain Flutter app the hosted subtree would have been laid out before the post-frame mouse
tracker pass. Under WebF the `RenderWidget` subtree's layout is driven by WebF's own pipeline, so a
`markNeedsLayout` issued during the build phase is not necessarily flushed before
`hitTestInView` runs.

### Expected vs actual

- **Expected**: setting a hosted text field's value while the pointer rests over the view is a
  no-op for hit testing; the field re-lays-out and the hit test either misses or resolves normally.
- **Actual**: assertion in debug, then a per-frame assertion flood and a blank view. Nothing in the
  log points at the embedding app, which makes this very hard to attribute.

### Minimal reproduction

1. Host any text field via a `WidgetElement` (WebF's own `<input>` suffices).
2. Bind its value to something that arrives asynchronously (e.g. set it from a `fetch` callback a
   few hundred ms after load), so the text is assigned *after* first paint.
3. **Leave the mouse cursor sitting over the view** — this is the part that makes it look flaky;
   with the pointer outside the window `MouseTracker` has no device position and never hit-tests.
4. Observe `Text layout not available`, followed by an unbounded flood of
   `'!_debugDuringDeviceUpdate': is not true`, and a blank view.

### Suggested fix

Ensure the hosted Flutter subtree's layout is flushed before WebF participates in a hit test, or
make `RenderWidget.hitTestChildren` skip children that currently need layout:

```dart
// lib/src/rendering/widget.dart, in hitTestChildren
if (child.debugNeedsLayout) return false;   // or: skip this child
```

Skipping is safe here — a box that needs layout has no trustworthy geometry to hit-test against,
so the correct answer is "no hit" rather than "read stale glyph data". A narrower variant is to
guard only the `RenderEventListener` path, but the general check is cheaper to reason about.

Note that the debug asserts are hiding a real hazard rather than being the whole problem: the line
right after the assert is `final _TextPainterLayoutCacheWithOffset cachedLayout = _layoutCache!;`,
so in release the same state yields a null-check exception instead.

### Confidence

**Source-verified for the chain and the fix location; reproduction is condition-dependent.**
The full stack above is verbatim from a real run. The `assert`-only nature of both throw sites was
checked against the local Flutter SDK (`text_painter.dart:1688`). Not verified: whether
`debugNeedsLayout` is the cleanest predicate available inside WebF's hit-test path, and whether the
same crash occurs on other platforms (only macOS was exercised).

---

## 10. `onLoad` never fires on the second load in the same process — `checkCompleted()` bails on one of four guards and is never re-entered

**Title (one line):**
`WebFController.onLoad` is not called for a second page load in the same process, so embedders that gate their loading UI on it hang forever

### Environment

- webf `0.24.27` (pub.dev)
- Flutter: host app constrains `flutter: '>=3.29.0'`, `sdk: ^3.7.0`
- Platform: macOS 15 (arm64), debug build
- Loading mode: `WebFControllerManager` preloading (`WebF.fromControllerName` with a `bundle`)

### Description

An embedder that shows a spinner until `onLoad` and gives up after a timeout works on the first
load of a page and **hangs on the second load of the same URL in the same process** — even though
the second load is, by every other measure, completely successful:

- a fresh `WebFController` is created (the previous one was removed and disposed),
- the bundle is fetched and `load success` / `load complete with bundle` are logged,
- the page's own JavaScript runs (its `console.log` reaches `onJSLog`),
- every `fetch()` the page issues returns `200`,
- and the content is visibly rendered.

Only the Dart-side `onLoad` callback never arrives.

`onLoad` is invoked exclusively from `dispatchWindowLoadEvent()`
(`lib/src/launcher/controller.dart:1821`), which is invoked exclusively from `checkCompleted()`
(`lib/src/launcher/controller.dart:1718`). `checkCompleted()` has four early returns:

```dart
void checkCompleted() {
  if (_view!.document.parsing) return;
  if (_view!.document.isDelayingDOMContentLoadedEvent) return;
  if (_isDOMComplete) return;
  ...
  if (_view!.document.hasPendingRequest) return;
  if (_view!.document.isDelayingLoadEvent) return;
  ...
  dispatchWindowLoadEvent();
}
```

Whichever guard is hit, `checkCompleted()` simply returns — and **nothing guarantees it will be
called again** once the blocking condition clears. The call sites are the end of each loading-mode
routine plus resource-completion paths, so if the last blocker clears outside those paths, the load
event is lost permanently.

### Why it only reproduces on the second load

It looks like a straightforward race that the second load loses because everything is already warm.
From one run, first load vs second load of the same URL:

| | first load | second load |
|---|---|---|
| bundle `load complete` | 266 ms | 55 ms |
| the page's four `fetch()` calls | 279 / 277 / 277 / 52 ms | 6 / 6 / 6 / 10 ms |

Same document, same scripts, same requests — only the timing differs, and only the fast one fails
to emit `load`.

### Impact

`onLoad` is the natural "page is ready" signal for an embedder, and it is the one the public API
advertises. When it silently never fires, the embedder cannot distinguish "still loading" from
"loaded fine" and will eventually show a spurious error to the user over a perfectly good page.

### Suggested fix

Either of these would be sufficient, and they are complementary:

1. **Re-run `checkCompleted()` when a blocker clears.** Each guard corresponds to a counter or flag
   (`parsing`, `isDelayingDOMContentLoadedEvent`, `hasPendingRequest`, `isDelayingLoadEvent`);
   have the code that decrements/clears each one call `checkCompleted()` again, so the check is
   level-triggered rather than edge-triggered.
2. **Document `onBuildSuccess` as the render-readiness signal** (or add an explicit
   `onFirstContentfulPaint`). `onBuildSuccess` is currently reliable — it is called post-frame from
   the two success paths of `buildRootView()` (`lib/src/widget/webf.dart:673` and `:723`) and never
   from the error paths — but it is undocumented and reads like an internal hook, so embedders
   naturally reach for `onLoad` instead.

### Workaround (what we shipped)

Treat `onBuildSuccess` as the primary readiness signal and `onLoad` as a secondary one, both
funnelling into one idempotent handler. **The idempotency is mandatory**: `onBuildSuccess` fires on
every `buildRootView()`, so if the handler triggers a `setState` on an ancestor of the `WebF`
widget, the resulting rebuild re-enters `buildRootView()` and the callback loops forever.

### Confidence

**Symptom reproducible and the call chain source-verified; the specific guard is not isolated.**
We confirmed that `onLoad` → `dispatchWindowLoadEvent()` → `checkCompleted()` is the only path, and
that `checkCompleted()` returns early without any re-entry guarantee. We did **not** instrument
which of the four guards is hit on the failing run — that is the first thing to nail down before
filing, ideally with a minimal repro that loads the same URL twice with a warm HTTP cache.

---

## 提交前检查清单

> 这一节由主 agent 补写（起草 agent 在写到这里之前因账号当日额度用尽被中断，
> 7 条正文都已完成）。
>
> 📌 **2026-08-05 追加第 8 条**（flex `wrap` 下 base size 被测成容器宽度）。下面凡是写「7 条」
> 的地方都请按 **8 条**读 —— 排重、模板确认、用户批准这三件事对第 8 条同样适用，它一条都没查过重。
> 与前 7 条不同的是，**第 8 条有真机 before/after 佐证**（截图量像素），而且上游源码里已经带着
> 一个只覆盖 `RenderFlowLayout` 的同类修正，`Suggested fix` 因此是「把已有修正的类型判据放宽」，
> 属于最容易被接受的形状 —— 值得与第 1、7 条一起优先提。

### 必须由人做的三件事

1. **搜一遍上游已有 issue，逐条排重。** 起草时刻意禁用了联网（此前有 agent 因
   WebSearch / WebFetch 反复超时死掉），所以**这 10 条一条都没查过重**。
   已知至少存在一处重复风险：`env(safe-area-inset-*)` 那条对应上游 **#907（open）**，
   说明这类缺陷上游是有人报过的。搜索建议按报错原文而不是我们的措辞来搜
   （例如 `Bytecode are not valid to execute`、`Attempting to navigate WebF to an external`），
   我们的标题是重写过的、跟上游用词未必一致。
2. **确认上游 issue 模板**。`openwebf/webf` 的 `.github/ISSUE_TEMPLATE/` 可能要求特定
   字段顺序或 checkbox，本文档的结构（Environment / Description / Root cause /
   Expected vs actual / Minimal reproduction / Suggested fix / Confidence）是按开源惯例
   拟的，可能需要重排。
3. **用户批准后再提。** 向第三方仓库提 issue 是**外发动作**，且这 10 条里带着我们的
   源码分析与项目信息。**未经用户明确同意不要提交任何一条。**

### 提交时的注意事项

- **第 6 条（`input[type=range]` 整行不绘制）带一个未解释的 gap**：源码说该画一个
  Flutter `TextField`，实测一个像素都没画，而盒模型正常。正文已把它写成 open question。
  提这条之前最好在容器里再补一轮（比如把 range 换成 `type=text` 做对照、
  或打开 WebF 的 paint 调试），否则上游第一个回复很可能就是「你的复现里还有别的问题」。
- **第 3 与第 4 条可以合并**（同属字节码执行路径：一个缺回退、一个缺归因），
  但默认保持独立 —— 上游更容易分别 triage 与分别修。若上游偏好少而大的 issue，
  合并时把第 4 条降级为第 3 条里的一个小节。
- **第 1 与第 7 条的 `Suggested fix` 最具体**（前者是 `asByteData()` →
  `ByteData.sublistView()` 的一行改动，后者是把 sticky 从 `_isPositionedGridChild()`
  的判据里拆出来），这两条最可能被直接接受，建议优先提。
- 贴的行号全部基于 **webf 0.24.27**（pub 版本）。上游 main 分支自 2026-04-19 起
  静默至今，所以行号大概率仍然对得上，但提交时**在标题或 Environment 里写明版本号**。

### 复核状态

第 1–7 条的 `file:line` 都在 0.24.27 源码里逐条核对过（起草 agent 的报告未能送出，
但正文里每条都带了核对结论）。**第 7 条的行号由主 agent 独立复核过一遍**，
第 8–10 条为 2026-08-05 追加，各自的 `Confidence` 一节写明了核实到哪一步（第 10 条**未隔离出具体 guard**，提交前必须补最小复现）。
`_isPositionedGridChild` 的 12 处调用点 + `:1958` 的内联重复、3 条 placeholder
注释、`:4592` 的 sticky 偏移、以及 `flow.dart` 的 3 处对照全部对上。
