package services

import (
	"testing"
	"time"

	"songloft/internal/database/testutil"
)

func TestAutoScanner_GetConfig_Default(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	configService := NewConfigService(mdb.ConfigRepository())
	as := NewAutoScanner(nil, configService)

	cfg := as.GetConfig()
	if cfg.Enabled {
		t.Error("default enabled should be false")
	}
	if cfg.IntervalSeconds != 3600 {
		t.Errorf("default interval: got %d want 3600", cfg.IntervalSeconds)
	}
}

func TestAutoScanner_GetConfig_Persisted(t *testing.T) {
	mdb := testutil.OpenMemoryDB(t)
	configService := NewConfigService(mdb.ConfigRepository())
	if err := configService.SetJSON("auto_scan", AutoScanConfig{Enabled: true, IntervalSeconds: 600}); err != nil {
		t.Fatal(err)
	}
	as := NewAutoScanner(nil, configService)

	cfg := as.GetConfig()
	if !cfg.Enabled {
		t.Error("expected enabled=true")
	}
	if cfg.IntervalSeconds != 600 {
		t.Errorf("interval: got %d want 600", cfg.IntervalSeconds)
	}
}

func TestAutoScanner_ApplyConfig_StartStop(t *testing.T) {
	as := NewAutoScanner(nil, nil)

	as.ApplyConfig(AutoScanConfig{Enabled: true, IntervalSeconds: 60})

	// done channel 存在说明 goroutine 在运行
	as.mu.Lock()
	if as.done == nil {
		t.Error("auto scanner should be running after ApplyConfig(enabled=true)")
	}
	as.mu.Unlock()

	as.Stop()

	as.mu.Lock()
	if as.done != nil {
		t.Error("auto scanner should be stopped after Stop()")
	}
	as.mu.Unlock()
}

func TestAutoScanner_ApplyConfig_Disabled(t *testing.T) {
	as := NewAutoScanner(nil, nil)

	as.ApplyConfig(AutoScanConfig{Enabled: false, IntervalSeconds: 3600})

	as.mu.Lock()
	if as.done != nil {
		t.Error("auto scanner should not be running when disabled")
	}
	as.mu.Unlock()
}

func TestAutoScanner_ApplyConfig_Reconfigure(t *testing.T) {
	as := NewAutoScanner(nil, nil)

	as.ApplyConfig(AutoScanConfig{Enabled: true, IntervalSeconds: 60})
	as.mu.Lock()
	firstDone := as.done
	as.mu.Unlock()

	as.ApplyConfig(AutoScanConfig{Enabled: true, IntervalSeconds: 120})
	as.mu.Lock()
	secondDone := as.done
	as.mu.Unlock()

	// 旧 goroutine 应该已结束（done channel 关闭）
	select {
	case <-firstDone:
	case <-time.After(time.Second):
		t.Error("old goroutine should have stopped on reconfigure")
	}

	if secondDone == firstDone {
		t.Error("reconfigure should create a new done channel")
	}

	as.Stop()
}

func TestAutoScanner_StopIdempotent(t *testing.T) {
	as := NewAutoScanner(nil, nil)
	// Stop on never-started scanner should not panic
	as.Stop()
	as.Stop()
}
