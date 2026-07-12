package jsplugin

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

const (
	tcpReadBufferSize     = 65535
	tcpDialTimeoutDefault = 10 * time.Second
	tcpDialTimeoutMax     = 60 * time.Second
)

// managedTCPSocket 跟踪一条由插件发起的出站 TCP 连接。
type managedTCPSocket struct {
	id      string
	conn    net.Conn
	cancel  context.CancelFunc
	done    chan struct{}
	sendMu  sync.Mutex
	closeMu sync.Once
}

func (h *BridgeHandler) handleTCP(action, data string) (string, error) {
	switch action {
	case "net.tcpConnect":
		return h.netTCPConnect(data)
	case "net.tcpSend":
		return h.netTCPSend(data)
	case "net.tcpClose":
		return h.netTCPClose(data)
	default:
		return "", fmt.Errorf("unknown tcp action: %s", action)
	}
}

func (h *BridgeHandler) netTCPConnect(data string) (string, error) {
	var params struct {
		Host    string `json:"host"`
		Port    int    `json:"port"`
		Timeout int    `json:"timeout"` // 毫秒，0 表示默认
	}
	if err := json.Unmarshal([]byte(data), &params); err != nil {
		return "", fmt.Errorf("netTCPConnect: %w", err)
	}
	if params.Host == "" || params.Port <= 0 || params.Port > 65535 {
		return "", fmt.Errorf("netTCPConnect: invalid host/port %q:%d", params.Host, params.Port)
	}

	// 安全约束：仅允许连接私有 / 回环地址，避免成为开放代理（SSRF）。
	if !isPrivateHostAllowed(params.Host) {
		return "", fmt.Errorf("netTCPConnect: only private/loopback addresses are allowed, refused %q", params.Host)
	}

	// 配额：复用 UDP 的每插件 socket 上限（TCP 独立计数）。
	count := 0
	h.tcpSockets.Range(func(_, _ any) bool {
		count++
		return count < maxSocketsPerPlugin
	})
	if count >= maxSocketsPerPlugin {
		return "", fmt.Errorf("netTCPConnect: max %d sockets per plugin", maxSocketsPerPlugin)
	}

	timeout := tcpDialTimeoutDefault
	if params.Timeout > 0 {
		timeout = min(time.Duration(params.Timeout)*time.Millisecond, tcpDialTimeoutMax)
	}

	addr := net.JoinHostPort(params.Host, fmt.Sprintf("%d", params.Port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return "", fmt.Errorf("netTCPConnect: dial %q: %w", addr, err)
	}

	socketID := fmt.Sprintf("tcp-%d", h.socketIDSeq.Add(1))
	ctx, cancel := context.WithCancel(context.Background())

	sock := &managedTCPSocket{
		id:     socketID,
		conn:   conn,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	h.tcpSockets.Store(socketID, sock)

	go h.tcpReadLoop(ctx, sock)

	slog.Info("jsplugin: TCP socket connected",
		"plugin", h.service.plugin.EntryPath,
		"socketId", socketID,
		"remoteAddr", conn.RemoteAddr().String())

	result, _ := json.Marshal(map[string]string{
		"socketId":   socketID,
		"localAddr":  conn.LocalAddr().String(),
		"remoteAddr": conn.RemoteAddr().String(),
	})
	return string(result), nil
}

func (h *BridgeHandler) tcpReadLoop(ctx context.Context, sock *managedTCPSocket) {
	defer close(sock.done)

	entryPath := h.service.plugin.EntryPath
	reader := bufio.NewReader(sock.conn)
	buf := make([]byte, tcpReadBufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := reader.Read(buf)
		if n > 0 {
			// base64 编码原始字节：TCP 是字节流，一次 Read 可能在多字节 UTF-8
			// 字符中间截断，直接转 string 会被 json.Marshal 替换为 U+FFFD 而永久
			// 损坏。编码字节可安全穿过 JSON/JS 字符串，插件 atob 后跨 chunk 拼接。
			event := &TcpDataEvent{
				SocketID: sock.id,
				Data:     base64.StdEncoding.EncodeToString(buf[:n]),
			}
			eventJSON, mErr := json.Marshal(event)
			if mErr != nil {
				slog.Debug("jsplugin: TCP event marshal failed",
					"plugin", entryPath, "socketId", sock.id, "error", mErr)
			} else if pErr := h.postHostEvent("tcp_data", sock.id, string(eventJSON)); pErr != nil {
				slog.Debug("jsplugin: TCP host event push failed",
					"plugin", entryPath, "socketId", sock.id, "error", pErr)
			}
		}
		if err != nil {
			// 主动 close（ctx 取消）时静默退出，不推送 close 事件。
			if ctx.Err() != nil {
				return
			}
			// 对端关闭或读错误：通知 JS 并清理。
			h.emitTCPClose(sock)
			return
		}
	}
}

// emitTCPClose 推送 tcp_close 事件并从跟踪表移除，保证只触发一次。
func (h *BridgeHandler) emitTCPClose(sock *managedTCPSocket) {
	sock.closeMu.Do(func() {
		h.tcpSockets.Delete(sock.id)
		event := &TcpCloseEvent{SocketID: sock.id}
		eventJSON, err := json.Marshal(event)
		if err != nil {
			return
		}
		if err := h.postHostEvent("tcp_close", sock.id, string(eventJSON)); err != nil {
			slog.Debug("jsplugin: TCP close host event push failed",
				"plugin", h.service.plugin.EntryPath, "socketId", sock.id, "error", err)
		}
	})
}

func (h *BridgeHandler) netTCPSend(data string) (string, error) {
	var params struct {
		SocketID string `json:"socketId"`
		Data     string `json:"data"`
	}
	if err := json.Unmarshal([]byte(data), &params); err != nil {
		return "", fmt.Errorf("netTCPSend: %w", err)
	}

	val, ok := h.tcpSockets.Load(params.SocketID)
	if !ok {
		return "", fmt.Errorf("netTCPSend: socket %q not found", params.SocketID)
	}
	sock := val.(*managedTCPSocket)

	// 与 UDP 对齐：JS 侧 btoa 编码后传入；非 base64 时容错当原始字符串。
	payload, decErr := base64.StdEncoding.DecodeString(params.Data)
	if decErr != nil {
		payload = []byte(params.Data)
	}

	sock.sendMu.Lock()
	_, err := sock.conn.Write(payload)
	sock.sendMu.Unlock()
	if err != nil {
		return "", fmt.Errorf("netTCPSend: %w", err)
	}
	return "", nil
}

func (h *BridgeHandler) netTCPClose(data string) (string, error) {
	var params struct {
		SocketID string `json:"socketId"`
	}
	if err := json.Unmarshal([]byte(data), &params); err != nil {
		return "", fmt.Errorf("netTCPClose: %w", err)
	}

	val, ok := h.tcpSockets.LoadAndDelete(params.SocketID)
	if !ok {
		return "", nil
	}
	sock := val.(*managedTCPSocket)
	// 标记已关闭，避免 readLoop 再推送 tcp_close。
	sock.closeMu.Do(func() {})

	sock.cancel()
	_ = sock.conn.Close()
	<-sock.done

	slog.Info("jsplugin: TCP socket closed",
		"plugin", h.service.plugin.EntryPath,
		"socketId", params.SocketID)
	return "", nil
}

func (h *BridgeHandler) cleanupTCPSockets() {
	h.tcpSockets.Range(func(key, value any) bool {
		sock := value.(*managedTCPSocket)
		sock.closeMu.Do(func() {})
		sock.cancel()
		_ = sock.conn.Close()
		<-sock.done
		h.tcpSockets.Delete(key)
		slog.Info("jsplugin: closed TCP socket on cleanup",
			"plugin", h.service.plugin.EntryPath,
			"socketId", key)
		return true
	})
}

// isPrivateHostAllowed 解析 host 并要求其所有地址均为私有 / 回环 / 链路本地地址。
// 仅当全部解析结果都是内网地址时才放行，防止插件把 TCP 能力用作公网 SSRF 代理。
func isPrivateHostAllowed(host string) bool {
	// host 直接是 IP 字面量时无需 DNS 解析。
	if ip := net.ParseIP(host); ip != nil {
		return isPrivateIP(ip)
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, ip := range ips {
		if !isPrivateIP(ip) {
			return false
		}
	}
	return true
}

// isPrivateIP 判断 IP 是否为内网 / 保留地址（与 services.whitelist 中的判定保持一致）。
func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}
