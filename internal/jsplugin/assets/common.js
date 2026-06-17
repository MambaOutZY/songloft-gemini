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
        d.dataset.theme = th;
        d.classList.remove('theme-light', 'theme-dark');
        d.classList.add('theme-' + th);
        localStorage.setItem('songloft-theme', th);
        document.dispatchEvent(new CustomEvent('songloft-theme-change', { detail: { theme: th } }));
    }

    applyTheme(initialTheme);

    if (params.has('theme')) {
        params.delete('theme');
        var cleanUrl = window.location.pathname;
        var remaining = params.toString();
        if (remaining) cleanUrl += '?' + remaining;
        history.replaceState(null, '', cleanUrl);
    }

    window.addEventListener('message', function(e) {
        if (e.data && e.data.type === 'songloft-theme' && (e.data.theme === 'light' || e.data.theme === 'dark')) {
            applyTheme(e.data.theme);
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
        return document.documentElement.dataset.theme || 'light';
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
        enhanceClickableElements: enhanceClickableElements
    };
})();
