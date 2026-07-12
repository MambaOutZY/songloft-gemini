package jsplugin

import (
	"encoding/json"
	"net"
	"testing"
)

func newTestTCPHandler() *BridgeHandler {
	h := &BridgeHandler{
		service: &JSService{
			plugin: &JSPlugin{EntryPath: "test-tcp-plugin"},
		},
	}
	// HasActiveTCPSockets 通过 service.bridgeHandler 反查 socket，补上反向引用
	h.service.bridgeHandler = h
	return h
}

// startEchoServer 在 127.0.0.1 上起一个 TCP echo server，返回其端口和关闭函数。
func startEchoServer(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						return
					}
					_, _ = c.Write(buf[:n])
				}
			}(conn)
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port, func() { _ = ln.Close() }
}

func TestNetTCPConnectSendClose(t *testing.T) {
	port, stop := startEchoServer(t)
	defer stop()

	h := newTestTCPHandler()

	// connect
	connReq, _ := json.Marshal(map[string]any{"host": "127.0.0.1", "port": port, "timeout": 2000})
	result, err := h.netTCPConnect(string(connReq))
	if err != nil {
		t.Fatalf("netTCPConnect failed: %v", err)
	}
	var resp struct {
		SocketID   string `json:"socketId"`
		RemoteAddr string `json:"remoteAddr"`
	}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("unmarshal connect response: %v", err)
	}
	if resp.SocketID == "" {
		t.Fatal("expected non-empty socketId")
	}

	// 连接期间应有活跃 TCP socket（用于阻止空闲休眠）
	if !h.service.HasActiveTCPSockets() {
		t.Error("expected active TCP socket while connected")
	}

	// send
	sendReq, _ := json.Marshal(map[string]string{"socketId": resp.SocketID, "data": "PING\n"})
	if _, err := h.netTCPSend(string(sendReq)); err != nil {
		t.Fatalf("netTCPSend failed: %v", err)
	}

	// close
	closeReq, _ := json.Marshal(map[string]string{"socketId": resp.SocketID})
	if _, err := h.netTCPClose(string(closeReq)); err != nil {
		t.Fatalf("netTCPClose failed: %v", err)
	}

	// 关闭后应无活跃 TCP socket
	if h.service.HasActiveTCPSockets() {
		t.Error("expected no active TCP socket after close")
	}
}

func TestNetTCPConnectRejectsPublicAddress(t *testing.T) {
	h := newTestTCPHandler()
	connReq, _ := json.Marshal(map[string]any{"host": "8.8.8.8", "port": 53})
	if _, err := h.netTCPConnect(string(connReq)); err == nil {
		t.Fatal("expected public address to be rejected")
	}
}

func TestNetTCPSocketQuota(t *testing.T) {
	port, stop := startEchoServer(t)
	defer stop()

	h := newTestTCPHandler()
	connReq, _ := json.Marshal(map[string]any{"host": "127.0.0.1", "port": port})

	var ids []string
	for i := range maxSocketsPerPlugin {
		result, err := h.netTCPConnect(string(connReq))
		if err != nil {
			t.Fatalf("connect %d failed: %v", i, err)
		}
		var r struct {
			SocketID string `json:"socketId"`
		}
		_ = json.Unmarshal([]byte(result), &r)
		ids = append(ids, r.SocketID)
	}
	// 第 9 个应超配额被拒
	if _, err := h.netTCPConnect(string(connReq)); err == nil {
		t.Fatalf("expected quota error after %d sockets", maxSocketsPerPlugin)
	}

	// 清理
	for _, id := range ids {
		closeReq, _ := json.Marshal(map[string]string{"socketId": id})
		_, _ = h.netTCPClose(string(closeReq))
	}
}

func TestIsPrivateHostAllowed(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"127.0.0.1", true},
		{"192.168.1.1", true},
		{"10.0.0.5", true},
		{"172.16.0.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
	}
	for _, c := range cases {
		if got := isPrivateHostAllowed(c.host); got != c.want {
			t.Errorf("isPrivateHostAllowed(%q) = %v, want %v", c.host, got, c.want)
		}
	}
}
