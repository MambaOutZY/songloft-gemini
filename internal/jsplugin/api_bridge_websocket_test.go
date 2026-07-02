package jsplugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

const webSocketEchoJSCode = `
globalThis.onInit = async function() {};
globalThis.onDeinit = async function() {};
globalThis.onHTTPRequest = async function(req) {
    return { statusCode: 200, headers: {}, body: "ok" };
};
globalThis.onWebSocket = async function(req, socket) {
    socket.onMessage(async function(event) {
        if (event.isBinary) {
            await socket.send(event.data);
            return;
        }
        await socket.send("echo:" + event.data);
    });
};
`

func loadWebSocketTestPlugin(t *testing.T, entryPath string, permissions []string) (*Manager, http.Handler) {
	t.Helper()

	pluginsDir, dataDir, repo, _ := setupTestEnv(t)
	ctx := context.Background()

	manifest := testManifest(entryPath)
	manifest.Permissions = permissions
	zipData := createTestPluginZip(t, manifest, webSocketEchoJSCode)

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

func TestInboundWebSocketEcho(t *testing.T) {
	_, router := loadWebSocketTestPlugin(t, "ws-echo", []string{PermWebSocket})
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/jsplugin/ws-echo/api/inbound?x=1"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			t.Fatalf("dial websocket: status=%d err=%v", resp.StatusCode, err)
		}
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
		t.Fatalf("write text: %v", err)
	}
	messageType, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read text echo: %v", err)
	}
	if messageType != websocket.TextMessage || string(payload) != "echo:ping" {
		t.Fatalf("text echo = type %d payload %q, want text echo:ping", messageType, string(payload))
	}

	binaryPayload := []byte{0x01, 0x02, 0x7f}
	if err := conn.WriteMessage(websocket.BinaryMessage, binaryPayload); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	messageType, payload, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read binary echo: %v", err)
	}
	if messageType != websocket.BinaryMessage || string(payload) != string(binaryPayload) {
		t.Fatalf("binary echo = type %d payload %v, want %v", messageType, payload, binaryPayload)
	}
}

func TestInboundWebSocketRequiresPermission(t *testing.T) {
	_, router := loadWebSocketTestPlugin(t, "ws-no-perm", []string{})
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/jsplugin/ws-no-perm/api/inbound"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected websocket dial to fail without websocket permission")
	}
	if resp == nil {
		t.Fatalf("expected HTTP response on failed dial, got nil: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}
