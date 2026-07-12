package jsplugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
)

const udpDuringHTTPRequestJSCode = `
globalThis.onInit = async function() {};
globalThis.onDeinit = async function() {};
globalThis.onHTTPRequest = async function(req) {
    var received = [];
    var bind = await songloft.net.udpBind({ address: "127.0.0.1:0" });
    songloft.net.onData(bind.socketId, function(event) {
        received.push(atob(event.data));
    });
    await songloft.net.udpSend(bind.socketId, "hello", bind.localAddr);
    await new Promise(function(resolve) { setTimeout(resolve, 300); });
    await songloft.net.udpClose(bind.socketId);
    return {
        statusCode: 200,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ received: received })
    };
};
`

func loadNetIntegrationTestPlugin(t *testing.T, entryPath string) (*Manager, http.Handler) {
	return loadNetIntegrationTestPluginWithCode(t, entryPath, udpDuringHTTPRequestJSCode)
}

func loadNetIntegrationTestPluginWithCode(t *testing.T, entryPath, jsCode string) (*Manager, http.Handler) {
	t.Helper()

	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	ctx := context.Background()

	manifest := testManifest(entryPath)
	manifest.Permissions = []string{PermNet}
	zipData := createTestPluginZip(t, manifest, jsCode)

	zipFileName := entryPath + ".jsplugin.zip"
	if err := os.WriteFile(filepath.Join(pluginsDir, zipFileName), zipData, 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	plugin := &JSPlugin{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		EntryPath:   manifest.EntryPath,
		Main:        manifest.Main,
		Permissions: manifest.Permissions,
		Status:      JSPluginStatusActive,
		FilePath:    zipFileName,
	}
	if err := repo.Create(ctx, plugin); err != nil {
		t.Fatalf("create plugin record: %v", err)
	}

	manager := NewManager(repo, pluginsDir, dataDir, "", nil, nil)
	t.Cleanup(func() { manager.Close() })

	if err := manager.LoadPlugin(ctx, plugin); err != nil {
		t.Skipf("LoadPlugin failed (may need QuickJS runtime): %v", err)
	}

	router := chi.NewRouter()
	manager.RegisterAPIRoutes(router)
	return manager, router
}

func TestUDPHostEventDeliveredDuringHTTPRequest(t *testing.T) {
	_, router := loadNetIntegrationTestPlugin(t, "udp-during-http")
	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/jsplugin/udp-during-http/api/scan")
	if err != nil {
		t.Fatalf("GET plugin route: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Received []string `json:"received"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal body %s: %v", string(body), err)
	}
	if len(payload.Received) != 1 || payload.Received[0] != "hello" {
		t.Fatalf("received = %#v, want [hello]", payload.Received)
	}
}

// tcpEchoJSCodeTemplate 建立 TCP 连接、send 含多字节 UTF-8 的文本、通过 onData
// 收回 echo 数据并解码。%d 由测试填入 echo server 端口。验证完整链路：
// JS tcpConnect → bridge → Go dial/readLoop → postHostEvent("tcp_data") →
// __dispatchHostEvent → sock.onData，且 base64 通道对多字节 UTF-8 二进制安全。
// 文本 → 字节串用 unescape(encodeURIComponent)，字节串 → 文本用 decodeURIComponent(escape)。
const tcpEchoJSCodeTemplate = `
globalThis.onInit = async function() {};
globalThis.onDeinit = async function() {};
globalThis.onHTTPRequest = async function(req) {
    var text = "状态:播放中♪ end";
    var received = "";
    var sock = await songloft.net.tcpConnect("127.0.0.1", %d, { timeout: 2000 });
    sock.onData(function(data) { received += atob(data); }); // 累积原始字节串
    await sock.send(unescape(encodeURIComponent(text)));      // UTF-8 文本 → 字节串
    await new Promise(function(resolve) { setTimeout(resolve, 300); });
    await sock.close();
    return {
        statusCode: 200,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text: decodeURIComponent(escape(received)) })
    };
};
`

func TestTCPHostEventDeliveredDuringHTTPRequest(t *testing.T) {
	port, stop := startEchoServer(t)
	defer stop()

	jsCode := fmt.Sprintf(tcpEchoJSCodeTemplate, port)
	_, router := loadNetIntegrationTestPluginWithCode(t, "tcp-during-http", jsCode)
	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/jsplugin/tcp-during-http/api/scan")
	if err != nil {
		t.Fatalf("GET plugin route: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal body %s: %v", string(body), err)
	}
	// 多字节 UTF-8 经 TCP base64 往返应完好还原（raw string 方案会损坏为 U+FFFD）
	if want := "状态:播放中♪ end"; payload.Text != want {
		t.Fatalf("text = %q, want %q", payload.Text, want)
	}
}
