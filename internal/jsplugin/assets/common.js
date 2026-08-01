/**
 * Songloft Plugin Common JS — 由主程序自动注入到所有插件 HTML 页面
 * 职责：embed 检测、主题桥接、API 工具（window.SongloftPlugin）
 */
(function() {
    'use strict';

    // ── Embed 检测 ──
    if (new URLSearchParams(window.location.search).has('embed')) {
        document.documentElement.classList.add('embed');
    }

    // ── 主题桥接 ──
    var params = new URLSearchParams(window.location.search);
    var initialTheme = params.get('theme') || localStorage.getItem('songloft-theme') || 'light';

    function applyTheme(th) {
        var d = document.documentElement;
        // 用 setAttribute 而不是 d.dataset.theme：本脚本是 <head> 内的阻塞脚本，
        // 在 WebF 里此刻 documentElement.dataset 还是 null，赋值会抛
        // TypeError，而它会**中断整个 IIFE** —— window.SongloftPlugin 压根不会
        // 定义，宿主桥连带全废（songloft-org/songloft#341 实测）。
        // setAttribute 语义等价，且不依赖 dataset 何时就绪。
        d.setAttribute('data-theme', th);
        d.classList.remove('theme-light', 'theme-dark');
        d.classList.add('theme-' + th);
        localStorage.setItem('songloft-theme', th);
        document.dispatchEvent(new CustomEvent('songloft-theme-change', { detail: { theme: th } }));
    }

    // 刻意包 try/catch：本文件是一个 IIFE，任何早期异常都会中断其余全部代码，
    // 包括最后那段 window.SongloftPlugin 的定义 —— 表现是宿主桥整体静默失效，
    // 极难归因（songloft-org/songloft#341 就踩过：dataset 在 WebF 里为 null）。
    // 主题失效只是外观问题，不该连带打掉插件的宿主能力。
    try {
        applyTheme(initialTheme);
    } catch (e) {
        console.warn('[songloft] applyTheme failed, continuing:', e);
    }

    if (params.has('theme')) {
        params.delete('theme');
        var cleanUrl = window.location.pathname;
        var remaining = params.toString();
        if (remaining) cleanUrl += '?' + remaining;
        history.replaceState(null, '', cleanUrl);
    }

    // ════════════════════════════════════════════════════════════════════════
    // WebF 兼容垫片层（songloft-org/songloft#341）
    // ════════════════════════════════════════════════════════════════════════
    //
    // WebF 是自研 W3C 运行时，有一批 HTML/CSS 能力缺失。本文件由后端注入到**每个**
    // 插件页，而插件是独立仓库、第三方可自由发布 —— 我们改不了别人的插件，所以这里
    // 是统一垫掉这些缺口的唯一位置。
    //
    // 三条铁律（common.js 由后端服务给**所有**客户端版本和普通浏览器）：
    //   ① 纯增量 ② 特性探测 ③ 全部关在 isWebFEngine() 分支里。
    // 绝不能改变浏览器与系统 WebView 下的既有行为。
    //
    // ── 两个时机，刻意分开 —— 不是代码风格，是能力边界 ────────────────────
    //
    //   installEarly()   立即执行。common.js 是 <head> 内的阻塞脚本，此刻 <body>
    //                    还没解析。**原型 / 属性访问器级的拦截只能放这里**：它必须
    //                    早于插件自己的脚本安装，晚一步就漏掉别人已经跑过的赋值。
    //                    代价是不能碰 DOM —— 那时候还没有 DOM。
    //
    //   applyOnReady()   DOMContentLoaded 后执行。**就地替换 / 改造元素只能放这里**：
    //                    要等解析器把节点建出来才有东西可改。
    //
    // ── 两类垫片的分界线（已经踩过的坑）────────────────────────────────────
    //
    // 判据是「这个缺口会不会在**解析期**就产生不可撤销的副作用（发请求 / 起解码）」：
    //
    //   会 → applyOnReady **根本来不及**。静态 HTML 里 <img src=""> 的加载在解析期
    //        就已发起，早于任何脚本能跑的时机，所以那一项最终是在服务端 injectHTMLHead
    //        里把属性剥掉的（internal/jsplugin/routes.go 的 stripEmptySrcAttrs）；
    //        这里只剩「运行时 img.src = ''」那条访问器路径和一道 DOM 兜底扫描。
    //   不会 → <details>、<table> 这类纯渲染 / 交互缺口不触发任何资源加载，
    //        晚到 DOMContentLoaded 再改造完全没有副作用，就地替换是最省事的做法。
    //
    // 加新垫片时先回答这个问题，再决定挂到哪个数组里。

    function isWebFEngine() {
        return !!window.webf;
    }

    // 每个垫片各自包 try/catch，一个失败不能拖垮其它垫片。
    // 理由与上面 applyTheme 处相同、但后果更严重：本文件是一个 IIFE，任一垫片抛出
    // 都会中断其余**全部**代码，包括文件末尾 window.SongloftPlugin 的定义 —— 表现是
    // 宿主桥整体静默失效，极难归因（#341 就踩过：dataset 在 WebF 里为 null 抛
    // TypeError）。垫片失效只意味着「某个元素退回 WebF 的原生表现」，绝不该连带
    // 打掉插件的宿主能力。
    function runShims(shims, phase) {
        for (var i = 0; i < shims.length; i++) {
            try {
                shims[i].apply();
            } catch (e) {
                console.warn('[songloft] shim "' + shims[i].name + '" failed (' + phase + '):', e);
            }
        }
    }

    // 按标签名收集元素并快照成数组。
    //
    // 两点讲究：① 优先 querySelectorAll，但对 WebF 里**未注册的标签**
    // （details/summary 落到 _UnknownHTMLElement）不敢假定类型选择器一定能匹配，
    // 拿到空结果时退回 getElementsByTagName；② 一律拷成普通数组 ——
    // getElementsByTagName 返回 live 集合，垫片会改 DOM，边改边遍历不安全。
    function collectByTag(tag) {
        var list = null;
        try {
            list = document.querySelectorAll(tag);
        } catch (e) {
            list = null;
        }
        if (!list || !list.length) {
            try {
                list = document.getElementsByTagName(tag);
            } catch (e) {
                list = null;
            }
        }
        var out = [];
        if (list) {
            for (var i = 0; i < list.length; i++) out.push(list[i]);
        }
        return out;
    }

    // ── 垫片：空 img src（early 段 —— 属性访问器）──────────────────────────
    //
    // 按 HTML 规范 src="" 是无效值，浏览器不会为它发请求。WebF 却把空 src
    // **解析成当前文档 URL**，于是把插件页自己的 HTML 抓回来当图片解码，报
    // 「Failed to decode image ... (mime=text/html)」。实测命中 miot / stats /
    // music-feed 等多个插件，所以在宿主侧统一挡掉，而不是逐个插件改。
    var emptyImgSrcAccessorShim = {
        name: 'img-src-accessor',
        apply: function() {
            var imgProto = window.HTMLImageElement && window.HTMLImageElement.prototype;
            var srcDesc = imgProto && Object.getOwnPropertyDescriptor(imgProto, 'src');
            if (!srcDesc || !srcDesc.set || !srcDesc.configurable) return;
            Object.defineProperty(imgProto, 'src', {
                configurable: true,
                enumerable: srcDesc.enumerable,
                get: srcDesc.get,
                set: function(value) {
                    // 改为移除属性，语义上等价于「没有图」
                    if (value === '' || value === null || value === undefined) {
                        this.removeAttribute('src');
                        return;
                    }
                    srcDesc.set.call(this, value);
                }
            });
        }
    };

    // ── 垫片：空 img src（ready 段 —— DOM 兜底扫描）─────────────────────────
    //
    // 服务端 stripEmptySrcAttrs 已经剥掉插件页里写死的 src=""，这道扫描是兜底：
    // 覆盖「插件运行时用 innerHTML 插进来的 <img src="">」——那条路径既不过服务端
    // 正则，也不过上面的属性访问器（innerHTML 走的是解析器，不是 setter）。
    var emptyImgSrcSweepShim = {
        name: 'img-src-sweep',
        apply: function() {
            var imgs = collectByTag('img');
            for (var i = 0; i < imgs.length; i++) {
                if (imgs[i].getAttribute('src') === '') imgs[i].removeAttribute('src');
            }
        }
    };

    // ── 垫片：<details> / <summary>（ready 段 —— 就地改造）─────────────────
    //
    // WebF 的两份标签注册表（bridge/core/html/html_tag_names.json5 与 Dart 的
    // element_registry.dart）里都没有 details/summary，它们降级为 _UnknownHTMLElement
    // （display:block）：子内容照常渲染成块，但**没有折叠交互、没有 open 属性语义、
    // 没有三角标记** —— 也就是「详情」永远是摊开的。
    //
    // 刻意**不**把 details/summary 换成 div：真实插件按标签名选样式（lxmusic 的
    // `.import-results-details summary { cursor:pointer; color:var(--md-primary) }`），
    // 换标签会静默丢掉这些样式，插件里的 querySelector('summary') 也会失配。
    // 所以保留原元素（连带 class / style / id 都不动），只做四件事：
    //   ① 把非 summary 的子节点收进一个可整体折叠的容器
    //   ② 给 summary 挂 click / 键盘切换
    //   ③ 补 open 属性访问器，让 el.open = true 真的能展开
    //   ④ 插一个 Material Symbols 连字做三角标记（实测在 WebF 下可用）
    //
    // 纯 CSS/JS，不写任何 Dart：样式走 common.css 里既有的 --md-* 变量，因此自动
    // 跟随主题切换 —— 做成 Flutter 组件就拿不到插件页的主题变量了，这是不做成
    // 原生组件的主要理由。
    var DETAILS_MARK = 'data-sl-details-shim';

    function shimOneDetails(el) {
        // 幂等：applyShims 可被插件在动态插入 HTML 后重复调用
        if (el.hasAttribute(DETAILS_MARK)) return;
        el.setAttribute(DETAILS_MARK, '');

        var content = document.createElement('div');
        content.className = 'sl-details-content';

        // 第一个 <summary> 当触发器，其余子节点（含文本节点）全进 content。
        // 先把 childNodes 快照成数组：appendChild 会改动这个 live 集合。
        var kids = [];
        for (var i = 0; i < el.childNodes.length; i++) kids.push(el.childNodes[i]);

        var summary = null;
        for (var j = 0; j < kids.length; j++) {
            var node = kids[j];
            if (!summary && node.nodeType === 1 && node.tagName &&
                node.tagName.toLowerCase() === 'summary') {
                summary = node;
                continue;
            }
            content.appendChild(node);
        }
        el.appendChild(content);

        // 没写 <summary> 时规范要求显示 "Details"。这里补一个同名元素，否则折叠后
        // 内容永远打不开 —— 宁可多一行占位文字，也不能做出打不开的黑洞。
        if (!summary) {
            summary = document.createElement('summary');
            summary.textContent = 'Details';
            el.insertBefore(summary, content);
        }
        el.classList.add('sl-details');
        summary.classList.add('sl-details-summary');
        summary.setAttribute('role', 'button');
        if (!summary.getAttribute('tabindex')) summary.setAttribute('tabindex', '0');

        var marker = document.createElement('span');
        marker.className = 'material-symbols-outlined sl-details-marker';
        marker.setAttribute('aria-hidden', 'true');
        summary.insertBefore(marker, summary.firstChild);

        // 尊重原有 open 属性的初始状态
        var open = el.hasAttribute('open');

        function render() {
            // 刻意写死 'block' 而不是复位成 ''：WebF 对「把 inline style 置空
            // 是否等于撤销声明」没有稳定表现，content 是我们自己建的 div，
            // block 就是它的正确显示值。
            content.style.display = open ? 'block' : 'none';
            // 三角用连字文本切换，而不是 CSS transform 旋转：WebF 的 transform /
            // 伪元素 content 支持面不确定，换字是最稳的一条路。
            marker.textContent = open ? 'arrow_drop_down' : 'arrow_right';
            summary.setAttribute('aria-expanded', open ? 'true' : 'false');
            if (open) el.setAttribute('open', '');
            else el.removeAttribute('open');
        }

        function toggle() {
            open = !open;
            render();
            // 与规范对齐：状态变化派发 toggle（浏览器里插件本来就能收到这个事件，
            // 垫片不补的话 WebF 下监听 toggle 的插件会静默失灵）
            try {
                el.dispatchEvent(new CustomEvent('toggle', { bubbles: false }));
            } catch (e) { /* 事件构造不可用不影响折叠本身 */ }
        }

        render();
        summary.addEventListener('click', toggle);
        summary.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                toggle();
            }
        });

        // open 在浏览器里是反射属性（details.open = true 即展开），WebF 的未知元素
        // 没有这层语义。补个访问器，让插件既有的 el.open 读写照常工作。
        try {
            Object.defineProperty(el, 'open', {
                configurable: true,
                get: function() { return open; },
                set: function(v) {
                    var next = !!v;
                    if (next === open) return;
                    open = next;
                    render();
                }
            });
        } catch (e) {
            // 访问器装不上时仍可点击展开，不致命
            console.warn('[songloft] details.open accessor unavailable:', e);
        }
    }

    var detailsShim = {
        name: 'details',
        apply: function() {
            var list = collectByTag('details');
            for (var i = 0; i < list.length; i++) {
                // 逐元素兜底：页面里某个畸形 details 不该让其余的一起失去折叠
                try {
                    shimOneDetails(list[i]);
                } catch (e) {
                    console.warn('[songloft] details shim skipped one element:', e);
                }
            }
        }
    };

    // ── 垫片注册表 ─────────────────────────────────────────────────────────
    var earlyShims = [emptyImgSrcAccessorShim];
    var readyShims = [emptyImgSrcSweepShim, detailsShim];

    function installEarly() {
        if (!isWebFEngine()) return;
        // 根 class：给 common.css 与插件 CSS 一个 WebF-only 的作用域钩子。
        // 只在这里加，所以 `html.webf-engine` 在浏览器 / 系统 WebView 下永不出现。
        try {
            document.documentElement.classList.add('webf-engine');
        } catch (e) {
            console.warn('[songloft] webf-engine root class unavailable:', e);
        }
        runShims(earlyShims, 'early');
    }

    function applyOnReady() {
        if (!isWebFEngine()) return;
        runShims(readyShims, 'ready');
    }

    installEarly();
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', applyOnReady);
    } else {
        applyOnReady();
    }

    window.addEventListener('message', function(e) {
        if (!e.data || !e.data.type) return;
        if (e.data.type === 'songloft-theme' && (e.data.theme === 'light' || e.data.theme === 'dark')) {
            applyTheme(e.data.theme);
        } else if (e.data.type === 'songloft-player-state') {
            dispatchPlayerState(e.data.state);
        } else if (e.data.type === 'songloft-host-reply') {
            // 安全：host 回执只接受来自父窗口的消息（native 顶层 parent===self 亦成立）。
            if (e.source && e.source !== window.parent) return;
            resolveHostReply(e.data);
        }
    });

    // ── API 工具 ──
    var API_BASE = '.';

    /**
     * 从 localStorage 获取 Songloft 认证 Token
     * @returns {string}
     */
    function getAuthToken() {
        try {
            var authData = localStorage.getItem('songloft-auth');
            if (authData) {
                var auth = JSON.parse(authData);
                return auth.accessToken || '';
            }
        } catch (e) {
            // ignore
        }
        return '';
    }

    function buildHeaders() {
        var headers = { 'Content-Type': 'application/json' };
        var token = getAuthToken();
        if (token) {
            headers['Authorization'] = 'Bearer ' + token;
        }
        return headers;
    }

    function parseResponse(response) {
        if (!response.ok) {
            return response.text().then(function(text) {
                var msg = response.statusText || ('HTTP ' + response.status);
                try {
                    var body = JSON.parse(text);
                    if (body && (body.message || body.error)) {
                        msg = body.message || body.error;
                    }
                } catch (_) {}
                throw new Error(msg);
            });
        }
        return response.text().then(function(text) {
            if (!text) return null;
            return JSON.parse(text);
        });
    }

    /**
     * 发送 GET 请求并返回 JSON
     * @param {string} path
     * @returns {Promise<any>}
     */
    function apiGet(path) {
        return fetch(API_BASE + path, {
            method: 'GET',
            headers: buildHeaders()
        }).then(parseResponse);
    }

    /**
     * 发送 POST 请求并返回 JSON
     * @param {string} path
     * @param {any} body
     * @returns {Promise<any>}
     */
    function apiPost(path, body) {
        return fetch(API_BASE + path, {
            method: 'POST',
            headers: buildHeaders(),
            body: JSON.stringify(body)
        }).then(parseResponse);
    }

    /**
     * 发送 PUT 请求并返回 JSON
     * @param {string} path
     * @param {any} body
     * @returns {Promise<any>}
     */
    function apiPut(path, body) {
        return fetch(API_BASE + path, {
            method: 'PUT',
            headers: buildHeaders(),
            body: JSON.stringify(body)
        }).then(parseResponse);
    }

    /**
     * 发送 DELETE 请求并返回 JSON
     * @param {string} path
     * @returns {Promise<any>}
     */
    function apiDelete(path) {
        return fetch(API_BASE + path, {
            method: 'DELETE',
            headers: buildHeaders()
        }).then(parseResponse);
    }

    /**
     * 获取当前主题
     * @returns {'light' | 'dark'}
     */
    function getTheme() {
        // 与 applyTheme 对称，不走 dataset（WebF 早期为 null，见 applyTheme 注释）
        return document.documentElement.getAttribute('data-theme') || 'light';
    }

    /**
     * 监听主题变化
     * @param {(theme: 'light' | 'dark') => void} callback
     */
    function onThemeChange(callback) {
        document.addEventListener('songloft-theme-change', function(e) {
            callback(e.detail.theme);
        });
    }

    // ── Accessibility ──

    function hideDecorationIcons() {
        document.querySelectorAll('.material-symbols-outlined, .mi').forEach(function(el) {
            if (!el.getAttribute('aria-hidden')) {
                el.setAttribute('aria-hidden', 'true');
            }
        });
    }

    function enhanceClickableElements() {
        document.querySelectorAll('[onclick]').forEach(function(el) {
            var tag = el.tagName.toLowerCase();
            if (tag !== 'button' && tag !== 'a' && tag !== 'input' && tag !== 'select') {
                if (!el.getAttribute('role')) el.setAttribute('role', 'button');
                if (!el.getAttribute('tabindex')) el.setAttribute('tabindex', '0');
                el.addEventListener('keydown', function(e) {
                    if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        el.click();
                    }
                });
            }
        });
    }

    function announce(message, priority) {
        var region = document.getElementById('songloft-a11y-live');
        if (!region) {
            region = document.createElement('div');
            region.id = 'songloft-a11y-live';
            region.className = 'sr-only';
            region.setAttribute('aria-live', priority || 'polite');
            region.setAttribute('aria-atomic', 'true');
            document.body.appendChild(region);
        }
        region.textContent = '';
        setTimeout(function() { region.textContent = message; }, 100);
    }

    function initAccessibility() {
        hideDecorationIcons();
        enhanceClickableElements();
        var snackbar = document.getElementById('snackbar');
        if (snackbar && !snackbar.getAttribute('role')) {
            snackbar.setAttribute('role', 'status');
            snackbar.setAttribute('aria-live', 'polite');
        }
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initAccessibility);
    } else {
        initAccessibility();
    }

    // ── 宿主客户端桥接（仅 Flutter 客户端 webview 有效）──
    //
    // 让 webview 打开的插件页调用 Flutter 宿主能力（改写正在播放队列、播放控制、
    // 状态订阅等）。请求走 flutter_inappwebview 的 callHandler（原生 Promise 返回值），
    // 事件（播放状态变更）复用上面的 postMessage 通道。
    // Web/iframe 或无原生桥接时优雅降级：isHostAvailable() 返回 false，调用会 reject。

    var HOST_HANDLER = 'songloftHost';
    var HOST_CALL_TIMEOUT_MS = 10000;

    // native（Android/iOS/桌面）webview：flutter_inappwebview 提供请求/响应式 callHandler。
    function isNativeHost() {
        return !!(window.flutter_inappwebview &&
            typeof window.flutter_inappwebview.callHandler === 'function');
    }

    // WebF 渲染引擎（songloft-org/songloft#341）：既没有 flutter_inappwebview，
    // 也不是 iframe（WebF 无 iframe 实现，window.parent === window），所以必须
    // 单独探测。走 WebF 自带的 methodChannel，语义与 callHandler 近乎一对一。
    function isWebFHost() {
        return !!(window.webf && window.webf.methodChannel &&
            typeof window.webf.methodChannel.invokeMethod === 'function');
    }

    // Web：插件页运行在宿主 iframe 内，走 postMessage 与父窗口通信。
    // 独立浏览器标签（parent === self）没有宿主，返回 false。
    function isIframeHost() {
        try {
            return !!window.parent && window.parent !== window;
        } catch (e) {
            return true; // 跨域访问 parent 抛错 → 视为嵌入
        }
    }

    function isHostAvailable() {
        return isWebFHost() || isNativeHost() || isIframeHost();
    }

    // ── Web/iframe postMessage 传输：请求/响应关联 ──
    var hostPending = {};
    var hostCallSeq = 0;

    function invokeViaPostMessage(ns, method, params) {
        return new Promise(function(resolve, reject) {
            var id = 'c' + (++hostCallSeq) + '_' + Date.now();
            var timer = setTimeout(function() {
                delete hostPending[id];
                reject(new Error('songloft host call timeout: ' + ns + '.' + method));
            }, HOST_CALL_TIMEOUT_MS);
            hostPending[id] = { resolve: resolve, reject: reject, timer: timer };
            window.parent.postMessage(
                { type: 'songloft-host-call', id: id, ns: ns, method: method, params: params || null },
                '*'
            );
        });
    }

    function resolveHostReply(msg) {
        var p = hostPending[msg.id];
        if (!p) return;
        clearTimeout(p.timer);
        delete hostPending[msg.id];
        if (msg.ok) p.resolve(msg.data);
        else p.reject(new Error(msg.error || 'songloft host call failed'));
    }

    // ── WebF methodChannel 传输 ──
    //
    // 请求体与响应体两端都是 JSON 字符串：WebF 的 method channel 对复杂对象的
    // 序列化形态没有稳定契约，字符串是唯一两端都确定的载体。响应侧对 string
    // 与 object 都做兼容，不假定其中一种。
    function invokeViaWebF(ns, method, params) {
        return window.webf.methodChannel
            .invokeMethod(HOST_HANDLER, JSON.stringify({ ns: ns, method: method, params: params || null }))
            .then(function(res) {
                var parsed = res;
                if (typeof parsed === 'string') {
                    try { parsed = JSON.parse(parsed); } catch (e) { parsed = null; }
                }
                if (parsed && parsed.ok) return parsed.data;
                throw new Error((parsed && parsed.error) || 'songloft host call failed');
            });
    }

    // 宿主请求页面回退（songloft-org/songloft#341）。
    //
    // WebF 侧没有 canGoBack，宿主无法自行判断页面内还有没有历史，只能问页面。
    // 返回 true 表示「已消费」，宿主就不再退出路由 / 退出应用。
    // 只在 WebF 下注册：另外两条链路的宿主用各自 webview 的 canGoBack。
    function registerWebFBackHandler() {
        if (!isWebFHost()) return;
        var mc = window.webf.methodChannel;
        if (typeof mc.addMethodCallHandler !== 'function') return;
        mc.addMethodCallHandler('requestBack', function() {
            if (window.history && window.history.length > 1) {
                window.history.back();
                return true;
            }
            return false;
        });
    }

    registerWebFBackHandler();

    /**
     * 调用宿主能力。约定返回 { ok, data } 或 { ok:false, error }。
     * WebF 走 methodChannel，native 走 callHandler，Web/iframe 走 postMessage 关联。
     * @returns {Promise<any>}
     */
    function invokeHost(ns, method, params) {
        if (isWebFHost()) {
            return invokeViaWebF(ns, method, params);
        }
        if (isNativeHost()) {
            return window.flutter_inappwebview
                .callHandler(HOST_HANDLER, { ns: ns, method: method, params: params || null })
                .then(function(res) {
                    if (res && res.ok) return res.data;
                    throw new Error((res && res.error) || 'songloft host call failed');
                });
        }
        if (isIframeHost()) {
            return invokeViaPostMessage(ns, method, params);
        }
        return Promise.reject(new Error('songloft host bridge unavailable (not running in a Songloft client webview)'));
    }

    // 播放状态订阅
    var playerStateListeners = [];

    function dispatchPlayerState(state) {
        for (var i = 0; i < playerStateListeners.length; i++) {
            try { playerStateListeners[i](state); } catch (e) { /* ignore */ }
        }
        document.dispatchEvent(new CustomEvent('songloft-player-state-change', { detail: state }));
    }

    var host = {
        isAvailable: isHostAvailable,
        getInfo: function() { return invokeHost('host', 'getInfo'); }
    };

    var player = {
        getState: function() { return invokeHost('player', 'getState'); },
        setQueue: function(ids, options) {
            options = options || {};
            return invokeHost('player', 'setQueue', {
                ids: ids,
                startIndex: options.startIndex,
                sourcePlaylistId: options.sourcePlaylistId
            });
        },
        addToQueue: function(ids) { return invokeHost('player', 'addToQueue', { ids: ids }); },
        insertToQueue: function(index, id) { return invokeHost('player', 'insertToQueue', { index: index, id: id }); },
        removeFromQueue: function(index) { return invokeHost('player', 'removeFromQueue', { index: index }); },
        reorderQueue: function(oldIndex, newIndex) { return invokeHost('player', 'reorderQueue', { oldIndex: oldIndex, newIndex: newIndex }); },
        clearQueue: function() { return invokeHost('player', 'clearQueue'); },
        play: function(id) { return invokeHost('player', 'play', { id: id }); },
        pause: function() { return invokeHost('player', 'pause'); },
        togglePlay: function() { return invokeHost('player', 'togglePlay'); },
        next: function() { return invokeHost('player', 'next'); },
        prev: function() { return invokeHost('player', 'prev'); },
        seek: function(seconds) { return invokeHost('player', 'seek', { seconds: seconds }); },
        setVolume: function(volume) { return invokeHost('player', 'setVolume', { volume: volume }); },
        setPlayMode: function(mode) { return invokeHost('player', 'setPlayMode', { mode: mode }); },
        playPlaylistById: function(playlistId) { return invokeHost('player', 'playPlaylistById', { playlistId: playlistId }); },
        onStateChange: function(handler) {
            playerStateListeners.push(handler);
            return function() {
                var idx = playerStateListeners.indexOf(handler);
                if (idx >= 0) playerStateListeners.splice(idx, 1);
            };
        }
    };

    window.SongloftPlugin = {
        getAuthToken: getAuthToken,
        apiGet: apiGet,
        apiPost: apiPost,
        apiPut: apiPut,
        apiDelete: apiDelete,
        getTheme: getTheme,
        onThemeChange: onThemeChange,
        announce: announce,
        hideDecorationIcons: hideDecorationIcons,
        enhanceClickableElements: enhanceClickableElements,
        // 重跑 WebF 垫片的 ready 段。插件用 innerHTML 动态插入内容（新的 <details>
        // 等）后调用即可；幂等，且在浏览器 / 系统 WebView 下是彻底的 no-op。
        applyShims: applyOnReady,
        host: host,
        player: player
    };
})();
