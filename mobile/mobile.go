// Package mobile 提供 gomobile 可调用的入口函数，用于将 Go 后端嵌入移动客户端。
// 通过 gomobile bind 编译为 Android .aar / iOS .xcframework 后，
// Flutter 经 MethodChannel 调用这些函数在本机启动 HTTP 服务器。
package mobile

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"songloft/internal/app"
	"songloft/internal/config"
)

var (
	mu       sync.Mutex
	instance *mobileServer
)

type mobileServer struct {
	app      *app.App
	server   *http.Server
	listener net.Listener
	port     int
}

// Start 启动内嵌后端。
//   - dataDir: 应用沙盒数据目录（DB/封面/缓存/插件都放这里）
//   - musicDir: 音乐扫描目录
//   - port: 监听端口（0 = 自动分配）
//
// 返回实际监听端口号。
func Start(dataDir, musicDir string, port int) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return instance.port, nil
	}

	if os.Getenv("GOMEMLIMIT") == "" {
		debug.SetMemoryLimit(512 * 1024 * 1024)
	}
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(50)
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return -1, fmt.Errorf("创建数据目录失败: %w", err)
	}

	dbPath := filepath.Join(dataDir, "songloft.db")
	portStr := "0"
	if port > 0 {
		portStr = fmt.Sprintf("%d", port)
	}

	cfg := config.NewAppConfig(portStr, dbPath, "", "", "")
	cfg.MusicDir = musicDir

	var emptyFS embed.FS
	a := app.NewApp(cfg, emptyFS)
	if err := a.Init(); err != nil {
		return -1, fmt.Errorf("后端初始化失败: %w", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		a.Close()
		return -1, fmt.Errorf("监听端口失败: %w", err)
	}

	actualPort := ln.Addr().(*net.TCPAddr).Port
	srv := &http.Server{Handler: a.BuildHandler()}

	instance = &mobileServer{
		app:      a,
		server:   srv,
		listener: ln,
		port:     actualPort,
	}

	go func() {
		slog.Info("移动端后端已启动", "port", actualPort)
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP 服务异常退出", "error", err)
		}
	}()

	return actualPort, nil
}

// Stop 优雅停止后端
func Stop() {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	instance.server.Shutdown(ctx)
	instance.app.Close()
	instance = nil
	slog.Info("移动端后端已停止")
}

// IsRunning 检查后端是否在运行
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return instance != nil
}

// GetPort 获取当前监听端口，未运行时返回 0
func GetPort() int {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		return 0
	}
	return instance.port
}
