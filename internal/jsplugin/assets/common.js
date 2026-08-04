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

    // ── 垫片：input[type=range] → <songloft-slider>（ready 段）────────────
    //
    // WebF 的 <input> 实现（html/form/input.dart:251-268 的 createInput switch）
    // 只认 radio/checkbox/button/submit/date/time/hidden，**没有 range 分支**，
    // 于是 input[type=range] 落到 default 走 TextField —— 插件页里的音量条在 WebF
    // 下会变成一个可编辑文本框。而且 min/max/step 在 WebF 里**完全没有实现**，
    // 只是普通 DOM 属性，对渲染与取值零影响。
    //
    // 纯 JS 补不出滑块：mousedown/mousemove/pointerdown/pointermove 在 WebF 里
    // **压根不存在**（全 lib 零命中），touchstart/touchmove 有实现但只对
    // <webf-toucharea> 生效（enableTouchEvent 默认 false）。所以拖动必须由 Dart
    // 侧的自定义元素 <songloft-slider> 提供，这里只做「就地嫁接」。
    //
    // ── 为什么保留原 <input>（不换标签）────────────────────────────────
    //
    // 插件按 id/标签名 querySelector 后**直接读写 .value**（miot 光是音量条就有
    // 9 处，含 5 处写入），还有 `.disabled = x` 与 `addEventListener('input')`。
    // 换掉标签会静默打断插件自己的 JS —— 这是 Step 1 <details> 垫片留下的教训。
    // 所以：原 input 留在 DOM 里当**数据宿主**（插件眼里的唯一真值），
    // <songloft-slider> 只当**视图**，两边双向同步。
    //
    // ── 架构依据（都是 scripts/webf-verify 第 14 组实测出来的，不是推断）──
    //
    //   ① `Object.defineProperty(input, 'value', {...})` **能**遮蔽 WebF 的原生
    //      访问器（实测 shadowV=ok：哨兵写 77 读回 77）。实测还发现 .value 本来就是
    //      **实例上的 own 访问器**且 configurable=true（ownDesc=function/cfg=true）。
    //      所以 JS→滑块方向可以**事件驱动**，不需要轮询 diff .value。
    //      **判据必须是哨兵往返**，不能只看 defineProperty 有没有抛异常：QuickJS 的
    //      exotic get_own_property 处理器优先级高于普通 own property，描述符装上了
    //      也可能读写照旧走原生。下面 install() 里那段自检就是为此。
    //   ② `input.matches` 同样能遮蔽（shadowM=true/true，且能委托回原实现）。
    //      这条用来复原 `el.matches(':active')` 的语义 —— 见下面 dragging 处的注释。
    //   ③ `querySelectorAll('input[type=range]')` 可用（qsa=2）。仍留一条
    //      getElementsByTagName 的退路，与 collectByTag 同一种保守。
    //   ④ Dart→JS 的 dispatchEvent 通得过，且 `event.data` 能带回新值
    //      （15b 的 Hdata/Vdata）。
    //
    // ── 遮蔽失败就整体放弃（verified-or-abort）─────────────────────────
    //
    // 如果 ① 的自检没过，就**把滑块删掉、原 input 还原**，退回 WebF 的原生表现
    // （那个文本框）。刻意**不写**「退化成定时轮询」的第二条路：那条路在本环境里
    // 永远跑不到，也就永远测不到，属于「看起来周全、实际未经验证」的代码。
    // 而「隐藏了 input、又同步不上」比「难看但能用」严重得多 —— 插件会读到永远
    // 不变的值。
    var RANGE_MARK = 'data-sl-range-shim';
    var RANGE_OPT_OUT = 'data-sl-no-slider';
    // 兜底清 dragging 的时间。正常情况下滑块一定会在抬手时派 change，
    // 但万一元素被插件在拖动中途 innerHTML 掉了，change 就永远不来，
    // dragging 卡住 true 之后所有 JS→滑块同步都会被抑制（滑块从此跟不上插件状态）。
    var RANGE_DRAG_TIMEOUT_MS = 1500;

    function collectRangeInputs() {
        var list = [];
        try {
            var found = document.querySelectorAll('input[type="range"]');
            for (var i = 0; i < found.length; i++) list.push(found[i]);
        } catch (e) {
            list = [];
        }
        if (!list.length) {
            var all = collectByTag('input');
            for (var j = 0; j < all.length; j++) {
                var t = all[j].getAttribute('type');
                if (t && t.toLowerCase() === 'range') list.push(all[j]);
            }
        }
        return list;
    }

    // 派发一个「像浏览器那样」的事件给原 input，好让插件既有的监听器照常跑。
    // bubbles:true 是因为浏览器里 input/change 都冒泡，插件可能监听在祖先上；
    // WebF 的 Event 构造若不认第二个参数，退回无参形态（只丢冒泡，不致命）。
    function fireOnInput(input, type) {
        var ev;
        try {
            ev = new Event(type, { bubbles: true });
        } catch (e) {
            try { ev = new Event(type); } catch (e2) { return; }
        }
        input.dispatchEvent(ev);
    }

    function shimOneRange(input) {
        // 幂等：applyShims 可被插件在动态插入 HTML 后重复调用
        if (input.hasAttribute(RANGE_MARK)) return;
        // 插件的正式退出开关：某些 range 就是想保持原生表现（或插件自己已经处理了）
        if (input.hasAttribute(RANGE_OPT_OUT)) return;
        // 游离节点没地方插滑块。**在打标记之前**判掉：打了标记再失败，下一轮
        // applyShims 会跳过它，那个 input 就永久停在 WebF 的文本框形态了。
        if (!input.parentNode) return;
        input.setAttribute(RANGE_MARK, '');

        var slider = document.createElement('songloft-slider');
        slider.className = 'sl-range-slider';

        // 几何：**不拷 class**，只拷 inline style。
        //
        // 新标签匹配不到插件 CSS 里 `input[type="range"]` 那些选择器，所以尺寸
        // （miot 的 110x28 + rotate(-90deg)）拿不到。三条路里选了这条：
        //   ✗ 拷 class + 让插件写 `input[type=range], songloft-slider { ... }`
        //     —— class 上挂的往往是「让文本框长得像滑块」的规则
        //     （-webkit-appearance / ::-webkit-slider-thumb / accent-color），
        //     拷过来只会带进无意义甚至有害的声明。
        //   ✗ 读 computed transform 反推朝向 —— WebF 的 getComputedStyle 支持面
        //     不可靠（实测连自定义属性都不暴露），猜错了是静默的错朝向。
        //   ✓ 拷 inline style（那是作者对**这一个元素**的显式意图，如
        //     miot #scheduleVolume 的 width:100%）+ 朝向由插件用
        //     data-sl-orientation 显式声明 + 其余几何由插件针对
        //     `songloft-slider` / `.sl-range-slider` / `[data-sl-for=...]` 另写一份。
        // 代价：插件必须补几行 CSS。miot 是第一方插件，可接受；第三方插件不补的话
        // 拿到的是本元素的默认尺寸（横向 160x28），能用、只是不合版面。
        var inlineStyle = input.getAttribute('style');
        if (inlineStyle) slider.setAttribute('style', inlineStyle);

        // 朝向：**只认显式声明**，不猜。
        var orientation = input.getAttribute('data-sl-orientation');
        if (orientation) slider.setAttribute('orientation', orientation);

        // min/max/step 必须自己从 attribute 读 —— WebF 没实现这三个属性反射
        // （实测 el.min 是空串，不是 "0"）。
        var passthrough = ['min', 'max', 'step', 'aria-label'];
        for (var p = 0; p < passthrough.length; p++) {
            var v = input.getAttribute(passthrough[p]);
            if (v !== null) slider.setAttribute(passthrough[p], v);
        }
        // 给插件 CSS 一个精确选择器（原 input 的 id 留在 input 上不能挪走）
        if (input.id) slider.setAttribute('data-sl-for', input.id);

        var store = input.value;
        if (store === null || store === undefined || store === '') {
            store = input.getAttribute('value') || '0';
        }
        store = String(store);
        slider.setAttribute('value', store);
        if (input.disabled) slider.setAttribute('disabled', '');

        // 插在原 input 之后：位置、DOM 顺序、flex 里的排布都最接近它替代的那个元素
        if (input.nextSibling) input.parentNode.insertBefore(slider, input.nextSibling);
        else input.parentNode.appendChild(slider);

        var dragging = false;
        var dragTimer = null;

        function pushToSlider() {
            slider.setAttribute('value', store);
        }

        // ── ① .value 遮蔽 + 自检 ──────────────────────────────────────
        var nativeDesc = null;
        var installed = false;
        var origValue = store;
        // 自检期间不回推滑块。**这不是洁癖**：不挡的话哨兵字符串会被
        // setAttribute('value','sl-probe') 写到滑块上，Dart 侧当即打一条
        // 「不是有效数字」的警告，滑块的值也脏了（实测踩过）。
        var booting = true;
        try {
            nativeDesc = Object.getOwnPropertyDescriptor(input, 'value');
            Object.defineProperty(input, 'value', {
                configurable: true,
                get: function () { return store; },
                set: function (v) {
                    store = (v === null || v === undefined) ? '' : String(v);
                    // 拖动期间**不回推**滑块：插件通常会定时轮询设备状态并回写滑块
                    // （miot js/playback.js:400-408），浏览器里它靠
                    // `el.matches(':active')` 判断「用户正在拖，别覆盖」，而隐藏后的
                    // input 永远进不了 :active。这里连同下面的 matches 遮蔽是双保险；
                    // 第三道闸在 Dart 侧（拖动中忽略外部 value），所以哪怕这两层都
                    // 失效，把手也不会在用户手指还没抬起时跳回去。
                    if (!dragging && !booting) pushToSlider();
                }
            });
            // 哨兵往返自检，见本段顶部注释 ①
            var probe = 'sl-probe';
            input.value = probe;
            installed = (input.value === probe);
        } catch (e) {
            installed = false;
        }
        booting = false;
        // 复原用本地变量，而不是读回 slider.getAttribute('value')：
        // 那边只能拿到我们刚才写进去的东西，多一次跨语言往返、多一个失效点。
        store = origValue;
        if (!installed) {
            // verified-or-abort：还原现场，退回 WebF 的原生表现。
            // 注意哨兵那一行在遮蔽没生效时是**真的写进了原生存储**，所以除了
            // 恢复描述符，还必须把原值写回去，否则输入框里会留下哨兵字符串。
            try {
                if (nativeDesc) Object.defineProperty(input, 'value', nativeDesc);
                else delete input.value;
                input.value = origValue;
            } catch (e2) { /* 还原失败也只能到此为止 */ }
            try { slider.parentNode.removeChild(slider); } catch (e3) {}
            input.removeAttribute(RANGE_MARK);
            console.warn('[songloft] range shim aborted: input.value not interceptable');
            return;
        }

        // ── ② matches(':active') 遮蔽（best-effort）───────────────────
        // 语义：拖动中 :active 为真。这是插件用来判断「用户正在操作，别用轮询结果
        // 覆盖」的标准写法，而隐藏后的 input 在 WebF 里永远不会置 isActive
        // （element.dart:245 的 pseudo 标志只由落在它自己身上的指针事件驱动，
        // 而它已经 display:none 了）。装不上只损失这一条便利，不影响主流程。
        try {
            var nativeMatches = input.matches;
            if (typeof nativeMatches === 'function') {
                Object.defineProperty(input, 'matches', {
                    configurable: true,
                    writable: true,
                    value: function (selector) {
                        if (dragging && typeof selector === 'string' &&
                            selector.indexOf(':active') >= 0) return true;
                        return nativeMatches.call(input, selector);
                    }
                });
            }
        } catch (e) {
            console.warn('[songloft] range shim: matches(":active") not interceptable:', e);
        }

        // ── ③ .disabled 遮蔽（best-effort）───────────────────────────
        // 插件写 `input.disabled = !hasDevice`（miot js/utils.js:96）要能传导到滑块。
        // 装不上就只是滑块不变灰、仍可拖，不阻塞主流程。
        try {
            var disabledState = !!input.disabled;
            Object.defineProperty(input, 'disabled', {
                configurable: true,
                get: function () { return disabledState; },
                set: function (v) {
                    disabledState = !!v;
                    if (disabledState) slider.setAttribute('disabled', '');
                    else slider.removeAttribute('disabled');
                }
            });
        } catch (e) {
            console.warn('[songloft] range shim: disabled not interceptable:', e);
        }

        // ── 隐藏原 input ─────────────────────────────────────────────
        // 加 class（样式在 common.css 的 WebF 垫片段）**并**写 inline display：
        // 前者可被插件覆盖调试，后者保证即使 CSS 没加载上也一定隐藏。
        // 隐藏不影响取值：WebF 把 live value 存在**元素**上（base_input.dart:44
        // 的 `String _value`），不在 widget state 里，所以 widget 不建也不丢。
        input.classList.add('sl-range-hidden');
        try { input.style.display = 'none'; } catch (e) {}

        // ── 滑块 → 原 input ──────────────────────────────────────────
        slider.addEventListener('input', function (e) {
            dragging = true;
            if (dragTimer) clearTimeout(dragTimer);
            dragTimer = setTimeout(function () { dragging = false; }, RANGE_DRAG_TIMEOUT_MS);
            // Dart 侧把新值塞在 event.data 里（与 WebF 自己的 <input> 同一种写法，
            // 实测通：探针 15b 的 Hdata/Vdata）。
            //
            // 刻意**不做**「拿不到 data 就退回读 slider.getAttribute('value')」的兜底：
            // 那个属性只反映我们上一次**推给**滑块的值，拖动期间压根不会更新，
            // 拿它当兜底等于把把手往回拽。宁可这一次不同步（并留一条日志），
            // 也不要写一个方向相反的"兜底"。
            var next = (e && e.data !== undefined && e.data !== null && e.data !== '')
                ? e.data : null;
            if (next === null) {
                console.warn('[songloft] range shim: slider input event carried no data');
                return;
            }
            // 直接改 store 而不是走 input.value = ...：后者会触发上面的 setter，
            // 把值再回推给滑块，绕一圈没意义。
            store = String(next);
            fireOnInput(input, 'input');
        });
        slider.addEventListener('change', function () {
            dragging = false;
            if (dragTimer) { clearTimeout(dragTimer); dragTimer = null; }
            fireOnInput(input, 'change');
        });
    }

    var rangeSliderShim = {
        name: 'range-slider',
        apply: function () {
            var list = collectRangeInputs();
            for (var i = 0; i < list.length; i++) {
                // 逐元素兜底：页面里某个畸形 range 不该让其余的一起失去滑块
                try {
                    shimOneRange(list[i]);
                } catch (e) {
                    console.warn('[songloft] range shim skipped one element:', e);
                }
            }
        }
    };

    // ── 垫片：安全区内边距 --sl-safe-*（宿主注入，songloft-org/songloft#341）───
    //
    // WebF 不实现 `env(safe-area-inset-*)`（`css/keywords.dart` 里那 6 个
    // SAFE_AREA_INSET* / ENV 常量全库无引用，连解析入口都不存在），于是刘海屏 /
    // 圆角屏 / 手势条上插件页会顶到状态栏或被下巴切掉。
    //
    // **刻意不做「把 CSS 里的 env() 自动改写成 var()」**，那条路已核实不可行：
    //   ① 写入面不存在 —— CSSOM 的 `cssText` 只有 getter，`CSSStyleRule` 既不暴露
    //      `selectorText` 也不暴露 `.style`；唯一写入面是 insertRule/deleteRule/
    //      replaceSync 这种「整条规则进出」，而 `cssText` 是从解析结果**重建**的
    //      （简写被展开、WebF 不认识的属性已丢），往返即有损；`@media` 里的规则
    //      `cssText` 是空串且没有 `.cssRules`，delete+insert 会不可逆地摧毁它。
    //   ② 即便能改写也救不了命中面 —— 真实插件的 env() 全都套在 calc() / max()
    //      里，而 WebF **没有实现 CSS max() / min()**，换成 var() 照样是死的。
    // 所以改成「宿主只注入变量，插件作者直接写 var(--sl-safe-*)」：默认值由
    // common.css 承担（三种环境各自的取值见那边的注释），这里只负责把宿主推来的
    // 真实 inset 写成 documentElement 的**内联**自定义属性（内联优先级最高，
    // 覆盖 common.css 的默认值）。
    //
    // 关在 isWebFEngine() 里：浏览器与系统 WebView 下 env() 是原生支持的，
    // common.css 的 :root 默认值已经把变量绑到 env() 上，宿主也不会推这条消息。
    var SAFE_AREA_SIDES = ['top', 'right', 'bottom', 'left'];
    // 记住最后一次收到的值：消息可能早于 DOM 就绪到达（宿主在 onLoad 回调里推，
    // 而 WebF 的 documentElement.style 在 <head> 阻塞脚本期的可用性不敢假定 ——
    // dataset 在那时就是 null，见 applyTheme 的注释）。存下来由 ready 相补一次。
    var lastSafeArea = null;

    function applySafeAreaInsets(insets) {
        if (!insets || typeof insets !== 'object') return;
        lastSafeArea = insets;
        var de = document.documentElement;
        // 特性探测而不是假定：setProperty 拿不到就静默留给 ready 相重试。
        if (!de || !de.style || typeof de.style.setProperty !== 'function') return;
        for (var i = 0; i < SAFE_AREA_SIDES.length; i++) {
            var side = SAFE_AREA_SIDES[i];
            var v = insets[side];
            // 只接受有限非负数字。宿主推的是 MediaQuery.viewPadding 的逻辑像素，
            // 单位固定 px；拿到别的形态（字符串 / NaN / 负数）一律跳过而不是写进去，
            // 免得把一个非法值盖在 common.css 的合法默认值上。
            if (typeof v !== 'number' || !isFinite(v) || v < 0) continue;
            de.style.setProperty('--sl-safe-' + side, v + 'px');
        }
    }

    var safeAreaShim = {
        name: 'safe-area',
        apply: function () {
            // DOM 就绪后补写一次。宿主的推送时机（onLoad）与本函数的先后没有保证，
            // 没收到过值时什么都不做 —— 默认值在 common.css 里，不需要 JS 兜。
            if (lastSafeArea) applySafeAreaInsets(lastSafeArea);
        }
    };

    // ── 垫片：input[type=file] → 宿主原生选择器（ready 段）─────────────────
    //
    // WebF 的 <input> 实现（html/form/input.dart:250-266 的 build switch）只认
    // radio/checkbox/button/submit/date/time，`hidden` 在 createInput 里另有一支；
    // **file 落到 default → 一个 Flutter TextField**，点了什么都不会发生。
    //
    // 验证容器实测到的三条事实（scripts/webf-verify 第 18 组，都不是推断）：
    //   ① `FileReader` **不存在**（`typeof` 为 undefined），`FileList` 也不存在。
    //      → **绝不能**去伪造 `input.files` + `FileReader` 那一套：假 File 配不上
    //      真 FileReader，而真 FileReader 压根没有。这条直接决定了「插件必须改
    //      调用点」，与 URL.createObjectURL 那一项是同一个结论。
    //   ② WebF **不认 HTML `hidden` 属性**：带 hidden 与不带 hidden 的 file input
    //      盒子都是 170x24。所以垫片必须自己强制 display:none —— 否则插件刻意
    //      隐藏的那个 input 会在页面上占掉一行（还是一个点不动的空文本框）。
    //   ③ 程序化 `el.click()` **确实会**派发 DOM click 到监听器（progClick=dispatched）。
    //
    // ── 为什么两条入口都要装（③ 已经通了也要装第二条）──────────────────
    //
    // 真实插件的形状是「隐藏 input + 外部按钮代点」（radio 的 app.js:100
    // `btnFile.addEventListener('click', () => fileInput.click())`）。③ 说明只拦
    // click 事件在**当前** WebF 版本下够用，但那是 C++ 侧的实现细节、不随 pub 包
    // 发布、也没有任何契约保证；而覆写实例 click 方法的成本只有 5 行。
    // 两条入口互不冲突：覆写版**不调**原生 click，所以不会派发事件，
    // 不存在「一次点击弹两个选择器」。
    //
    // ── 结果怎么交给插件 ──────────────────────────────────────────────
    //
    // 主通道是 `SongloftPlugin.lastPickedFiles`（一个普通 JS 数组，一定可读）。
    // 派发的 `change` 事件上**尝试**挂 `event.data`，但那是 WebF 的 binding
    // object，能不能挂自定义属性没有契约 —— 所以文档要求插件读
    // `SongloftPlugin.lastPickedFiles`，`event.data` 只当锦上添花。
    // 刻意不派发 `input` 事件：浏览器里 file input 只派 change。
    var FILE_MARK = 'data-sl-file-shim';
    var FILE_OPT_OUT = 'data-sl-no-file-picker';
    // 载荷形态：'text'（默认，UTF-8 解码后的字符串）/ 'bytes'（base64）/ 'none'（只要元信息）。
    // 插件用 data-sl-file-as 声明；不声明就是 text —— 真实用例（radio 导入 m3u/json）
    // 只要文本，而 base64 让 20 MB 文件变成 ~27 MB 字符串跨两次桥，默认不该付这个钱。
    var FILE_AS_ATTR = 'data-sl-file-as';

    function collectFileInputs() {
        var list = [];
        try {
            var found = document.querySelectorAll('input[type="file"]');
            for (var i = 0; i < found.length; i++) list.push(found[i]);
        } catch (e) {
            list = [];
        }
        if (!list.length) {
            var all = collectByTag('input');
            for (var j = 0; j < all.length; j++) {
                var t = all[j].getAttribute('type');
                if (t && t.toLowerCase() === 'file') list.push(all[j]);
            }
        }
        return list;
    }

    function shimOneFileInput(input) {
        // 幂等：applyShims 可被插件在动态插入 HTML 后重复调用
        if (input.hasAttribute(FILE_MARK)) return;
        // 插件的正式退出开关（想保留 WebF 原生表现，或自己已经处理了）
        if (input.hasAttribute(FILE_OPT_OUT)) return;
        input.setAttribute(FILE_MARK, '');

        // 隐藏原 input：见本段顶部实测事实 ②。class + inline 双保险，
        // 与 rangeSliderShim 同一种保守。
        input.classList.add('sl-file-hidden');
        try { input.style.display = 'none'; } catch (e) {}

        // 一次只允许一个选择器在飞。没有这道闸时，插件那种「按钮 handler 里
        // 调 click()」的写法配上用户连点，会同时挂起两次宿主调用，
        // 回来的两个 change 里后到的那个未必是用户最后选的文件。
        var pending = false;

        function openPicker() {
            if (pending) return;
            if (input.disabled) return;
            pending = true;
            var as = (input.getAttribute(FILE_AS_ATTR) || 'text').toLowerCase();
            invokeHost('files', 'pickFile', {
                // accept 原样透传（radio 写的是扩展名形式 '.m3u,.m3u8,.json,.txt'，
                // 不是 MIME）—— 由 Dart 侧决定怎么翻译给 file_picker。
                accept: input.getAttribute('accept') || '',
                multiple: input.hasAttribute('multiple'),
                as: as
            }).then(function (res) {
                pending = false;
                var files = (res && res.files) || null;
                // 用户取消：**不派发 change**（浏览器语义也是如此）。
                // 派发一个空 change 会让插件走进「读不到文件」的错误分支，
                // 弹一个用户没做错任何事的报错。
                if (!files || !files.length) return;
                try {
                    if (window.SongloftPlugin) {
                        window.SongloftPlugin.lastPickedFiles = files;
                    }
                } catch (e) { /* 主通道写不进去也还有下面的 event.data */ }
                var ev;
                try {
                    ev = new Event('change', { bubbles: true });
                } catch (e) {
                    try { ev = new Event('change'); } catch (e2) { ev = null; }
                }
                if (!ev) {
                    console.warn('[songloft] file shim: cannot construct change event');
                    return;
                }
                // best-effort：WebF 的 Event 是 binding object，挂自定义属性
                // 没有契约。挂不上不影响主通道。
                try { ev.data = { files: files }; } catch (e) {}
                input.dispatchEvent(ev);
            }, function (err) {
                pending = false;
                console.warn('[songloft] file shim: host pickFile failed:', err);
            });
        }

        // 入口①：拦 click 事件（覆盖「用户直接点可见的 file input」）。
        // preventDefault 只是形式上的对齐 —— WebF 的 file input 本来就没有
        // 默认行为可阻止（它是个 TextField）。
        input.addEventListener('click', function (e) {
            try { e.preventDefault(); } catch (e2) {}
            openPicker();
        });

        // 入口②：覆写实例 click 方法（覆盖「隐藏 input + 外部按钮 fileInput.click()」）。
        // 装不上不致命 —— 入口① 已实测可承接程序化 click（progClick=dispatched），
        // 所以这里**不做** verified-or-abort：两条入口是冗余而非串联，
        // 为「其中一条没装上」就整体放弃反而是把能用的功能丢掉。
        try {
            Object.defineProperty(input, 'click', {
                configurable: true,
                writable: true,
                value: function () { openPicker(); }
            });
        } catch (e) {
            console.warn('[songloft] file shim: click() not interceptable:', e);
        }
    }

    var filePickerShim = {
        name: 'file-picker',
        apply: function () {
            var list = collectFileInputs();
            for (var i = 0; i < list.length; i++) {
                // 逐元素兜底：页面里某个畸形 input 不该让其余的一起失去选择器
                try {
                    shimOneFileInput(list[i]);
                } catch (e) {
                    console.warn('[songloft] file shim skipped one element:', e);
                }
            }
        }
    };

    // ── 垫片：<table> 只警告不改写（ready 段）──────────────────────────────
    //
    // WebF 的两份标签注册表里**一个表格标签都没有**（Dart 侧
    // element_registry.dart 连 `const String TABLE` 这样的常量都不存在），于是
    // <table>/<thead>/<tbody>/<tr>/<th>/<td> 全部降级为 _UnknownHTMLElement，
    // 默认样式 `display:block` → 6 列的表变成「6N 行无标签文本」，
    // 6 个 position:sticky 的表头还会互相叠在同一个位置。
    //
    // **这个失败是完全静默的**：未知标签的日志只在 `enableWebFCommandLog` 打开时
    // 才 debugPrint（element_registry.dart:83-85），产品没开。插件作者看到的是
    // 「一坨竖着排的文本」，既没有报错也没有可疑元素 —— 与 input[type=range]
    // 那条教训完全同构（那边是「一行莫名空白」）。
    //
    // ── 为什么只警告、不改写 ──────────────────────────────────────────
    //
    // 「把 <table> 改写成 WebF 自带的 <webf-table> 家族」这条路已经证伪：
    //   ① <webf-table> 只看**直接 childNodes**（table.dart:188-189），
    //      <thead>/<tbody> 不拆掉就是 rows=[] / header=null → **一张空表，
    //      而且不报错不打日志**；
    //   ② 拆掉 <tbody> 会让插件的 `$('#tbody').innerHTML = ...` 抛 TypeError，
    //      整个渲染函数中断；
    //   ③ 列宽只认表头单元格的 `column-width` 属性（CSS width 完全无效），
    //      sticky 表头必须逐列写死宽度否则表头与表体两次独立 flex 分配会错位，
    //      sticky 分支还用了 Expanded（要求有界高度）；
    //   ④ colspan/rowspan 零支持，且列数不齐会让 Flutter Table assert
    //      —— 也就是说改写能把「丑」变成「崩」。
    // 换来的是一个残废表格；而一行明确的 warn 把「归因难度高一个量级」的问题
    // 直接降到零，并把修复责任交给唯一有能力做对的人（插件作者）。
    // 推荐替代写法是 CSS Grid（WebF 的 Grid 有 193 KB 的真实实现，
    // `fr` / `minmax` / `repeat` 都在），见插件开发指南。
    //
    // ⚠️ **不要**顺手给 <table> 补 `display:table` 之类的 CSS：CSSDisplay 枚举里
    // 没有任何 table 取值，`resolveDisplay` 落到 `default: return inline`
    // （css/display.dart:83-85）→ 从 block 退化成 inline，**比什么都不写更糟**。
    var TABLE_MARK = 'data-sl-table-unsupported';

    var tableWarnShim = {
        name: 'table-warn',
        apply: function () {
            var list = collectByTag('table');
            var fresh = 0;
            for (var i = 0; i < list.length; i++) {
                // 幂等：applyShims 可被重复调用，不该每次刷一遍同样的 warn
                if (list[i].hasAttribute(TABLE_MARK)) continue;
                // 打标记不只是幂等用：它同时是页面内省的定位手段
                // （DIAGNOSE 脚本 / 插件自己都能 querySelectorAll 出来）
                list[i].setAttribute(TABLE_MARK, '');
                fresh++;
            }
            if (!fresh) return;
            console.warn('[songloft] WebF 不支持原生 <table>（会退化成纵向堆叠的 ' +
                'block，且完全静默）。请改用 CSS Grid，见插件开发指南「WebF 渲染' +
                '引擎」章节。本页命中 ' + fresh + ' 处，已标记 ' + TABLE_MARK + '。');
        }
    };

    // ── 垫片注册表 ─────────────────────────────────────────────────────────
    var earlyShims = [emptyImgSrcAccessorShim];
    var readyShims = [
        emptyImgSrcSweepShim, detailsShim, rangeSliderShim, safeAreaShim,
        filePickerShim, tableWarnShim
    ];

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
        } else if (e.data.type === 'songloft-safe-area') {
            // WebF-only（浏览器 / 系统 WebView 下 env() 原生可用，宿主也不推这条）。
            // try/catch 是必须的：本监听器里抛出会吞掉同一条消息的后续处理，
            // 而安全区失效只该是「少几像素内边距」，不该连带打掉别的消息通道。
            if (!isWebFEngine()) return;
            try {
                applySafeAreaInsets(e.data.insets);
            } catch (err) {
                console.warn('[songloft] safe-area apply failed:', err);
            }
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
     * Blob → `data:` URL（songloft-org/songloft#341）。
     *
     * 存在的理由：**WebF 没有 `URL.createObjectURL`**（验证容器实测
     * `typeof URL.createObjectURL === 'undefined'`；`Blob` 本身有，但没有任何
     * 入口能产出 `blob:` URL），而「带鉴权头 fetch 一张图 → 显示」是插件的常见
     * 写法（fetch 拿到的是 Blob，`<img src>` 不能直接吃 Blob）。
     *
     * 也**不可能**给 WebF 垫一个返回 `blob:` 的 createObjectURL：它的资源加载器
     * `WebFBundle.fromUrl` 只认 http/https/assets/file/`data:`，其余分支直接
     * `throw FlutterError('Unsupported url')`（foundation/bundle.dart:192-194）
     * —— 就算 JS 侧造出 `blob:xxx`，加载那一步必然失败。
     *
     * 而 `data:` URL 是原生支持的（同文件 `_isDataScheme` 分支 → `DataBundle`），
     * 且**确实能画出来**：探针第 18 组用「抓屏后统计整页里出现了多少该颜色的像素」
     * 这种**绘制层**判据验过 4 项全通过 —— `<img src="data:…">`、CSS
     * `background-image: url(data:…)`、以及 1×1 PNG 放大到 16×16。
     *
     * ⚠️ 这条结论我们反复翻转过两次，记下判据的演进以免第三次：
     *   ① 最初只看 `getBoundingClientRect`（`bgRect=16x16`）就断言「能出图」——
     *      **盒子有尺寸不等于图被画出来**，这个判据不成立；
     *   ② 随后一次目视观察说「1×1 红点 PNG 的方块是灰的」，据此改成「没有可信证据」——
     *      那是**按坐标取色**引入的假阴性：报坐标与抓屏之间隔了几秒，其间上面几组的
     *      异步读数还在写 DOM，把下面的行整体推下约 96px，于是取到了空白处的颜色；
     *   ③ 现在的判据与坐标无关（只问「这个颜色在整页出现了多少像素」），4 项全过。
     * **别再用「盒子尺寸」或「按坐标取色」当出图判据。**
     *
     * ⚠️ **本函数是异步的，而 `createObjectURL` 是同步的** —— 这不是实现懒惰，
     * 是无法弥合的形状差异（blob → base64 只能经 `arrayBuffer()` / `FileReader`，
     * 两者都是异步的）。所以插件**必须改调用点**，不能指望宿主垫一个假的同步
     * `createObjectURL`。
     *
     * 实现选 `blob.arrayBuffer()` 而不是 `FileReader.readAsDataURL`：
     * 验证容器实测 **WebF 里 `FileReader` 不存在**，而 `Blob.prototype.arrayBuffer` 在。
     * 浏览器与系统 WebView 下同样存在，所以三条渲染路径共用这一份实现，不分叉。
     *
     * ⚠️⚠️ **刻意不用 `btoa`，自带 base64 编码表。** WebF 的 `btoa` **不是二进制安全的**：
     * 它把 > 0x7F 的码点当字符先做了一次 UTF-8 编码，而不是按 latin1 取字节。实测
     *
     *     btoa('\x89')                       → "wg=="       正确应为 "iQ=="
     *     btoa('\x89PNG')                    → "wolQTg=="   正确应为 "iVBORw=="
     *
     * 而 `"wolQTg=="` 解码是 `0xC2 0x89 0x50 0x4E` —— `0xC2 0x89` 正是 U+0089 的
     * UTF-8 编码，签名一目了然。`atob` 方向是**正确**的
     * （`atob("iQ==").charCodeAt(0) === 137`），所以只需自己实现 encode 方向。
     *
     * ⚠️ **而且它不只是「值错」，还会静默丢数据。** 拿全部 256 种字节值过一遍
     * `btoa`：输出**长度是正确的 344**，但从第 170 个字符起就对不上，解出来是
     * `0x7E 0x7F 0xC2 …` —— 它按 UTF-8 展开字节流、却按**字符数**算输出长度，
     * 于是原始的 `0xC1..0xFF` 共 **63 个字节被直接丢掉**。所以「长度对得上」
     * 同样不能作为 base64 正确的判据。
     *
     * 对照：同一个 256 字节的 blob 走本函数（自带编码表）产出的 data URL
     * 与预期**逐字符相等**，顺带说明 `Blob.arrayBuffer()` 拿到的字节是准确的，
     * 锅只在 `btoa`。
     *
     * 这个坑此前被漏掉，是因为探针第 18 组的往返用例是 `new Blob(['hi'])` ——
     * **纯 ASCII，全程没有一个字节 > 0x7F**，结构上抓不到这个 bug。教训写在这里：
     * 二进制编解码的回归用例**必须**包含高位字节。
     *
     * 分块处理（3 字节一组、每 8 KB 拼一次）而不是一次拼完整串：
     * 几百 KB 的图上单次 `String.fromCharCode.apply` 会因参数个数超限抛 RangeError，
     * 而一个字符一个字符 `+=` 在 QuickJS 上又太慢。
     *
     * @param {Blob} blob
     * @param {string} [mimeType] 覆盖 blob.type（WebF 下 blob.type 恒为空串，
     *   且 `Response.headers.get('content-type')` 返回 null，所以调用方往往拿不到
     *   mime —— 不传就落到 application/octet-stream，能显示但语义不精确）
     * @returns {Promise<string>} 形如 `data:image/jpeg;base64,...`
     */
    var B64_CHARS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';

    /** Uint8Array → base64。不依赖 btoa，见 blobToDataURL 的注释。 */
    function bytesToBase64(bytes) {
        var out = '';
        var buf = [];
        var i = 0;
        var len = bytes.length;
        // 每次吃 3 字节产 4 个 base64 字符；余数在循环外单独补 padding
        var limit = len - (len % 3);
        for (i = 0; i < limit; i += 3) {
            var n = (bytes[i] << 16) | (bytes[i + 1] << 8) | bytes[i + 2];
            buf.push(
                B64_CHARS.charAt((n >> 18) & 63),
                B64_CHARS.charAt((n >> 12) & 63),
                B64_CHARS.charAt((n >> 6) & 63),
                B64_CHARS.charAt(n & 63)
            );
            // 攒够一批再 join，避免 O(n²) 的字符串拼接
            if (buf.length >= 8192) {
                out += buf.join('');
                buf.length = 0;
            }
        }
        var rem = len % 3;
        if (rem === 1) {
            var a = bytes[len - 1];
            buf.push(
                B64_CHARS.charAt((a >> 2) & 63),
                B64_CHARS.charAt((a << 4) & 63),
                '=', '='
            );
        } else if (rem === 2) {
            var b0 = bytes[len - 2], b1 = bytes[len - 1];
            buf.push(
                B64_CHARS.charAt((b0 >> 2) & 63),
                B64_CHARS.charAt(((b0 << 4) | (b1 >> 4)) & 63),
                B64_CHARS.charAt((b1 << 2) & 63),
                '='
            );
        }
        out += buf.join('');
        return out;
    }

    function blobToDataURL(blob, mimeType) {
        if (!blob) return Promise.reject(new Error('blobToDataURL: no blob'));
        if (typeof blob.arrayBuffer !== 'function') {
            return Promise.reject(new Error('blobToDataURL: Blob.arrayBuffer unavailable'));
        }
        var mime = mimeType || blob.type || 'application/octet-stream';
        return blob.arrayBuffer().then(function(buf) {
            return 'data:' + mime + ';base64,' + bytesToBase64(new Uint8Array(buf));
        });
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

    /**
     * 读取指定 origin 的 Cookie（仅原生客户端可用，Web 不支持）。
     * @param {string} origin - 目标站点 origin，如 "https://example.com"
     * @returns {Promise<Record<string, string>>} name→value 映射
     */
    function getCookies(origin) {
        return invokeHost('cookies', 'get', { origin: origin });
    }

    window.SongloftPlugin = {
        getAuthToken: getAuthToken,
        apiGet: apiGet,
        apiPost: apiPost,
        apiPut: apiPut,
        apiDelete: apiDelete,
        getTheme: getTheme,
        onThemeChange: onThemeChange,
        // Blob → data: URL。WebF 没有 URL.createObjectURL，见函数上方注释。
        // 三条渲染路径共用同一份实现（浏览器/WebView 下同样可用），插件不必分叉。
        blobToDataURL: blobToDataURL,
        // WebF 下 input[type=file] 垫片最近一次选到的文件数组
        // （每项 {name, size, text?, bytesBase64?}）。**这是主通道** ——
        // 派发的 change 事件上那个 event.data 是 best-effort（WebF 的 Event 是
        // binding object，挂自定义属性没有契约）。未选过时为 null。
        lastPickedFiles: null,
        announce: announce,
        hideDecorationIcons: hideDecorationIcons,
        enhanceClickableElements: enhanceClickableElements,
        // 重跑 WebF 垫片的 ready 段。插件用 innerHTML 动态插入内容（新的 <details>
        // 等）后调用即可；幂等，且在浏览器 / 系统 WebView 下是彻底的 no-op。
        applyShims: applyOnReady,
        host: host,
        player: player,
        getCookies: getCookies
    };
})();
