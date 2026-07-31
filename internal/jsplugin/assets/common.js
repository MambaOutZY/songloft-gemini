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
        host: host,
        player: player
    };
})();
