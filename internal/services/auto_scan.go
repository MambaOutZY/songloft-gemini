package services

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	autoScanMinInterval = 60
	autoScanMaxInterval = 86400
)

type AutoScanConfig struct {
	Enabled         bool `json:"enabled"`
	IntervalSeconds int  `json:"interval_seconds"`
}

type AutoScanner struct {
	songService   *SongService
	configService *ConfigService

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func NewAutoScanner(songService *SongService, configService *ConfigService) *AutoScanner {
	return &AutoScanner{
		songService:   songService,
		configService: configService,
	}
}

func (a *AutoScanner) GetConfig() AutoScanConfig {
	cfg := AutoScanConfig{
		Enabled:         false,
		IntervalSeconds: 3600,
	}
	if a.configService != nil {
		_ = a.configService.GetJSON("auto_scan", &cfg)
	}
	return cfg
}

func (a *AutoScanner) ApplyConfig(cfg AutoScanConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stopLocked()

	if !cfg.Enabled {
		return
	}

	interval := cfg.IntervalSeconds
	if interval < autoScanMinInterval {
		slog.Warn("自动扫描间隔过小，已钳位到最小值", "requested", interval, "min", autoScanMinInterval)
		interval = autoScanMinInterval
	}
	if interval > autoScanMaxInterval {
		slog.Warn("自动扫描间隔过大，已钳位到最大值", "requested", interval, "max", autoScanMaxInterval)
		interval = autoScanMaxInterval
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.done = make(chan struct{})

	go a.run(ctx, time.Duration(interval)*time.Second)
}

func (a *AutoScanner) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stopLocked()
}

func (a *AutoScanner) stopLocked() {
	if a.cancel != nil {
		a.cancel()
		<-a.done
		a.cancel = nil
		a.done = nil
	}
}

func (a *AutoScanner) run(ctx context.Context, interval time.Duration) {
	defer close(a.done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	slog.Info("自动扫描已启动", "interval", interval)
	for {
		select {
		case <-ctx.Done():
			slog.Info("自动扫描已停止")
			return
		case <-ticker.C:
			if err := a.songService.ScanAndImportAsync(false, nil); err != nil {
				slog.Debug("自动扫描跳过：已有扫描在进行", "error", err)
			} else {
				slog.Info("自动扫描已触发")
			}
		}
	}
}
