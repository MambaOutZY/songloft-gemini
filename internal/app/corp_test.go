package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Cross-Origin-Resource-Policy 是 Lynx Web 端的一个跨仓库契约：那边的宿主页必须开
// COEP: require-corp（web-core 依赖 SharedArrayBuffer，不开就白屏），于是这边每个被
// <img>/<audio>/<video> 以 no-cors 模式取用的响应都必须声明可被嵌入，否则浏览器阻断。
//
// 它值得闸门而不只是注释，是因为两种坏法都不出声：
//   - 头没了     → 封面/插件图标/音频流全挂，但 DevTools 里状态码仍是 200 (OK)
//   - 中间件没挂 → 函数还在、单元测试还绿，线上却一个响应也没有这个头
//
// 所以下面一条验行为，一条验它真的挂在 setupBaseRouter 的链上。

func TestCORPMiddlewareSetsHeader(t *testing.T) {
	const body = "payload"
	handler := corpMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/songs/1/cover", nil))

	// Result().Header 而不是 Header()：见下一个用例的说明。
	if got := rec.Result().Header.Get("Cross-Origin-Resource-Policy"); got != "cross-origin" {
		t.Errorf("Cross-Origin-Resource-Policy = %q, want %q", got, "cross-origin")
	}
	// 中间件必须是透明的：状态码与响应体都不能被它改动。
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != body {
		t.Errorf("body = %q, want %q", rec.Body.String(), body)
	}
}

// 头必须在 handler 写响应之前就设好。放在 next.ServeHTTP 之后设是个很容易犯又完全
// 静默的错误：一旦 handler 已经 WriteHeader，真实的 ResponseWriter 早把响应头发出去了，
// 后续对 Header() 的修改直接丢弃。
//
// 断言必须读 Result().Header，不能读 Header()。ResponseRecorder.Header() 返回的是那张
// 还能改的 map，所以「先 WriteHeader 再 Set」在它眼里照样能读到——本用例第一版就是这么
// 写的，反向注入那个错误时它是绿的，即一个只会证明自己的假闸门。Result() 用的是
// WriteHeader 那一刻 clone 的 snapHeader，与真实连接上的行为一致。
func TestCORPHeaderSurvivesHandlerWritingFirst(t *testing.T) {
	handler := corpMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPartialContent) // 模拟 Range 请求的音频流
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/songs/1/play", nil))

	if got := rec.Result().Header.Get("Cross-Origin-Resource-Policy"); got != "cross-origin" {
		t.Errorf("206 响应缺少 CORP 头（音频流会被 COEP 阻断）: got %q", got)
	}
}

// corpMiddleware 存在但没被 Use，等于完全没有这个功能，且没有任何运行时信号。
// 断言源码里真的挂上了——构造一个完整的 App 才能端到端验证这条链，代价远高于
// 它能多抓到的东西。
func TestCORPMiddlewareIsRegistered(t *testing.T) {
	src, err := os.ReadFile("routers.go")
	if err != nil {
		t.Fatalf("读取 routers.go 失败: %v", err)
	}
	if !strings.Contains(string(src), "a.router.Use(corpMiddleware())") {
		t.Error("setupBaseRouter 没有 a.router.Use(corpMiddleware())：" +
			"CORP 头不会出现在任何响应上，Web 端跨源封面/音频会被 COEP 静默阻断")
	}
}
