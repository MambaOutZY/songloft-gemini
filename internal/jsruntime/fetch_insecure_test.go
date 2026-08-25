package jsruntime

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// 本文件覆盖 songloft-org/songloft#401 的第二个根因：X-Fetch-Insecure。
//
// 背景：飞牛 fnOS 这类自建设备默认自签证书，且插件是按【裸 IP】访问它们的
// （fnID 中继返回 ipv4/publicIpv4 + 5667 端口）——即便设备装了 CA 签发的正式
// 证书，证书主体也是域名，按 IP 连必然 hostname mismatch，没有"让用户换个正规
// 证书"的出路。
//
// pcyear-bridge 插件为此给每个请求注入 X-Fetch-Insecure: 1，并在注释里断言
// "当前宿主若未识别该头会自动忽略，零副作用"。而修复前宿主根本不认识这个头：
// 既没跳过校验（所有候选网关在 TLS 握手就失败，报"访问码验证失败，全部候选网关
// 均不可用"），又把这个内部控制头当普通头转发给了上游。
//
// 这些测试锁住修复后的契约：权限门控生效、内部头不外泄、无权限时行为不变。

// newSelfSignedGateway 返回一台自签 HTTPS 的模拟访问码网关：
// GET /access_code_verify → 204 + Set-Cookie: os-access-code=...
// 同时记录上游实际收到的 X-Fetch-Insecure 头，用于断言内部控制头没有外泄。
func newSelfSignedGateway(t *testing.T) (*httptest.Server, func() string) {
	t.Helper()
	var seen string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("X-Fetch-Insecure")
		w.Header().Add("Set-Cookie", "os-access-code=abc123; Path=/; HttpOnly")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	return srv, func() string { return seen }
}

// TestSelfSignedGatewayIsActuallyUnverifiable 是其余断言的地基：
// 确认这台测试服务器确实"默认校验会失败、跳过校验才能连通"。
// 若哪天 httptest 的证书变成可校验的，下面的测试会变成假绿，这条会先炸。
func TestSelfSignedGatewayIsActuallyUnverifiable(t *testing.T) {
	srv, _ := newSelfSignedGateway(t)

	if _, err := (&http.Client{}).Get(srv.URL); err == nil {
		t.Fatal("默认客户端居然校验通过了：测试服务器不再是自签，本文件的前提失效")
	}
	skip := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402 — 测试对照组
	}}
	resp, err := skip.Get(srv.URL)
	if err != nil {
		t.Fatalf("跳过校验后应能连通: %v", err)
	}
	resp.Body.Close()
}

// TestFetchInsecure_DeniedWithoutPermission 是 #401 的回归锚点：
// 没有 net:insecure-tls 权限时，X-Fetch-Insecure 必须被忽略，请求因证书校验失败。
// 这正是插件报"全部候选网关均不可用"的那一刻。
func TestFetchInsecure_DeniedWithoutPermission(t *testing.T) {
	srv, upstreamSaw := newSelfSignedGateway(t)

	raw := doHTTPRequest(srv.URL+"/access_code_verify", http.MethodGet,
		`{"X-Fetch-Insecure":"1","x-access-code":"YWJj"}`, "", false)

	if !strings.Contains(raw, `"error"`) {
		t.Fatalf("无权限时应因证书校验失败，却成功了: %s", raw)
	}
	if !strings.Contains(raw, "certificate") {
		t.Errorf("错误应来自证书校验，实际: %s", raw)
	}
	if got := upstreamSaw(); got != "" {
		t.Errorf("请求根本不该到达上游，但上游收到了 X-Fetch-Insecure=%q", got)
	}
}

// TestFetchInsecure_AllowedWithPermission 是修复的正向断言：
// 有权限时自签网关可连通，并且能读到 204 上的 Set-Cookie——
// 即 #401 描述的 verifyAccessCode 链路真正跑通。
func TestFetchInsecure_AllowedWithPermission(t *testing.T) {
	srv, _ := newSelfSignedGateway(t)

	raw := doHTTPRequest(srv.URL+"/access_code_verify", http.MethodGet,
		`{"X-Fetch-Insecure":"1","x-access-code":"YWJj"}`, "", true)

	var got struct {
		Status      int                 `json:"status"`
		HeadersList map[string][]string `json:"headersList"`
		Error       string              `json:"error"`
	}
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal %q: %v", raw, err)
	}
	if got.Error != "" {
		t.Fatalf("有权限时应连通，却报错: %s", got.Error)
	}
	if got.Status != http.StatusNoContent {
		t.Errorf("status = %d, want 204", got.Status)
	}
	// 与 fetch_cookie_test.go 的修复合并起来才是完整链路：连得上 + 读得到 cookie。
	cookies := got.HeadersList["Set-Cookie"]
	if len(cookies) != 1 || !strings.HasPrefix(cookies[0], "os-access-code=abc123") {
		t.Errorf("Set-Cookie = %#v, want 单条 os-access-code=abc123", cookies)
	}
}

// TestFetchInsecure_ControlHeaderNotForwarded 锁住插件那句"零副作用"里真正错的部分：
// X-Fetch-Insecure 是内部控制头，无论权限如何都不能转发给上游。
// 修复前它会被 req.Header.Set 原样带上去。
func TestFetchInsecure_ControlHeaderNotForwarded(t *testing.T) {
	srv, upstreamSaw := newSelfSignedGateway(t)

	raw := doHTTPRequest(srv.URL+"/access_code_verify", http.MethodGet,
		`{"X-Fetch-Insecure":"1","X-Custom":"kept"}`, "", true)
	if strings.Contains(raw, `"error"`) {
		t.Fatalf("有权限时应连通: %s", raw)
	}
	if got := upstreamSaw(); got != "" {
		t.Errorf("上游收到了内部控制头 X-Fetch-Insecure=%q，应被剥离", got)
	}
	// 顺带确认剥离逻辑没有误伤普通自定义头
	if !strings.Contains(raw, "204") {
		t.Errorf("响应异常: %s", raw)
	}
}

// TestFetchInsecure_NoHeaderStillVerifies 确认这个口子只在显式带头时打开：
// 有权限但没带 X-Fetch-Insecure 的请求仍须完整校验证书。
func TestFetchInsecure_NoHeaderStillVerifies(t *testing.T) {
	srv, _ := newSelfSignedGateway(t)

	raw := doHTTPRequest(srv.URL+"/access_code_verify", http.MethodGet, `{}`, "", true)
	if !strings.Contains(raw, "certificate") {
		t.Errorf("未带 X-Fetch-Insecure 时应仍校验证书，实际: %s", raw)
	}
}

// TestFetchInsecure_PlainHTTPUnaffected 确认非 TLS 目标不受影响
// （插件的 fetchMaybe 给每个请求都注入了该头，含局域网 http 网关）。
func TestFetchInsecure_PlainHTTPUnaffected(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("X-Fetch-Insecure")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	for _, allow := range []bool{false, true} {
		raw := doHTTPRequest(srv.URL, http.MethodGet, `{"X-Fetch-Insecure":"1"}`, "", allow)
		if strings.Contains(raw, `"error"`) {
			t.Errorf("allowInsecure=%v: 明文 HTTP 应始终连通，实际: %s", allow, raw)
		}
		if seen != "" {
			t.Errorf("allowInsecure=%v: 上游收到了 X-Fetch-Insecure=%q", allow, seen)
		}
	}
}

// TestFetchInsecure_NoRedirectCombination 覆盖两个控制头叠加的分支：
// 插件的 verifyAccessCode 同时带 X-Fetch-Insecure 与 X-Fetch-No-Redirect
// （后者用于手动跟随 302 以收集中间跳的 Set-Cookie），走的是
// insecureNoRedirectHTTPClient 这条 case。
func TestFetchInsecure_NoRedirectCombination(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/access_code_verify" {
			w.Header().Add("Set-Cookie", "hop=1; Path=/")
			w.Header().Set("Location", srv.URL+"/final")
			w.WriteHeader(http.StatusFound)
			return
		}
		w.Header().Add("Set-Cookie", "os-access-code=final; Path=/")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	raw := doHTTPRequest(srv.URL+"/access_code_verify", http.MethodGet,
		`{"X-Fetch-Insecure":"1","X-Fetch-No-Redirect":"1"}`, "", true)

	var got struct {
		Status      int                 `json:"status"`
		HeadersList map[string][]string `json:"headersList"`
		Error       string              `json:"error"`
	}
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal %q: %v", raw, err)
	}
	if got.Error != "" {
		t.Fatalf("应连通: %s", got.Error)
	}
	// 没跟随重定向 → 拿到 302 本身，且中间跳的 Set-Cookie 可读
	if got.Status != http.StatusFound {
		t.Errorf("status = %d, want 302（不应自动跟随）", got.Status)
	}
	if c := got.HeadersList["Set-Cookie"]; len(c) != 1 || !strings.HasPrefix(c[0], "hop=1") {
		t.Errorf("应读到 302 上的 Set-Cookie，实际 %#v", c)
	}
}

// TestFetchInsecure_EndToEndViaJSEnv 走完整的 JS 层：
// SetAllowInsecureTLS 是权限在运行时层的唯一投影，这里验证它真的作用到
// env 内的 fetch()，而不只是 doHTTPRequest 的参数好使。
func TestFetchInsecure_EndToEndViaJSEnv(t *testing.T) {
	srv, _ := newSelfSignedGateway(t)

	probe := fmt.Sprintf(`
		fetch(%q, { headers: { 'X-Fetch-Insecure': '1' } }).then(function(r) {
			__out = JSON.stringify({ status: r.status, cookies: r.headers.getSetCookie() });
		}).catch(function(e) { __out = JSON.stringify({ err: String(e) }); });
	`, srv.URL+"/access_code_verify")

	for _, tc := range []struct {
		name  string
		allow bool
	}{
		{"denied", false},
		{"allowed", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewJSEnvManager()
			t.Cleanup(manager.SignalShutdown)

			envID := "test-insecure-" + tc.name
			if err := manager.CreateEnv(envID, polyfillJS, 1); err != nil {
				t.Fatalf("CreateEnv: %v", err)
			}
			t.Cleanup(func() { manager.DestroyEnv(envID) })

			if err := manager.SetAllowInsecureTLS(envID, tc.allow); err != nil {
				t.Fatalf("SetAllowInsecureTLS: %v", err)
			}

			out := runProbeInEnv(t, manager, envID, probe)

			var got struct {
				Status  int      `json:"status"`
				Cookies []string `json:"cookies"`
				Err     string   `json:"err"`
			}
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				t.Fatalf("unmarshal %q: %v", out, err)
			}

			if !tc.allow {
				if got.Err == "" {
					t.Fatalf("无权限时 fetch 应 reject，实际 status=%d", got.Status)
				}
				if !strings.Contains(got.Err, "certificate") {
					t.Errorf("应是证书错误，实际: %s", got.Err)
				}
				return
			}
			if got.Err != "" {
				t.Fatalf("有权限时 fetch 应成功: %s", got.Err)
			}
			if got.Status != http.StatusNoContent {
				t.Errorf("status = %d, want 204", got.Status)
			}
			if len(got.Cookies) != 1 || !strings.HasPrefix(got.Cookies[0], "os-access-code=abc123") {
				t.Errorf("getSetCookie() = %#v, want 单条 os-access-code", got.Cookies)
			}
		})
	}
}

// runProbeInEnv 在调用方已建好并配置好的 env 里跑探针代码，轮询 __out。
// 与 fetch_cookie_test.go 的 runFetchProbe 同范式，但不自建 env——
// 本文件需要在 CreateEnv 之后、跑代码之前插入 SetAllowInsecureTLS。
func runProbeInEnv(t *testing.T, manager *JSEnvManager, envID, probeJS string) string {
	t.Helper()
	code := "var __out = '';\n" + probeJS
	if _, err := manager.ExecuteJS(context.Background(), envID, code, 5000); err != nil {
		t.Fatalf("ExecuteJS: %v", err)
	}
	for range 40 {
		res, _ := manager.ExecuteJS(context.Background(), envID, "__out", 1000)
		if res != nil && res.Result != "" && res.Result != "undefined" {
			return res.Result
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("probe 未在时限内产出结果")
	return ""
}
