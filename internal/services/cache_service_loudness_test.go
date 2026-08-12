package services

import (
	"testing"

	"songloft/internal/database/testutil"
	"songloft/internal/models"
)

// newLoudnessCacheService 构造带真实 config 的 CacheService，用于响度配置读写测试。
func newLoudnessCacheService(t *testing.T) *CacheService {
	t.Helper()
	mdb := testutil.OpenMemoryDB(t)
	return NewCacheService(t.TempDir(), NewConfigService(mdb.ConfigRepository()))
}

// TestNormalizeLoudness_NilConfigService 测试里 &CacheService{} 直接构造（configService 为 nil）
// 时回落默认 -16，不 panic。seek_stream_test 等就是这么构造的，必须保持可用。
func TestNormalizeLoudness_NilConfigService(t *testing.T) {
	cs := &CacheService{}
	if v := cs.NormalizeLoudness(); v != defaultNormalizeLoudness {
		t.Errorf("nil configService: got %v want %v", v, defaultNormalizeLoudness)
	}
	// nil 接收者也安全
	var nilCs *CacheService
	if v := nilCs.NormalizeLoudness(); v != defaultNormalizeLoudness {
		t.Errorf("nil receiver: got %v want %v", v, defaultNormalizeLoudness)
	}
}

// TestNormalizeLoudness_DefaultEqualsConst 默认 -16 时滤镜串与缓存标记逐字等于旧常量，
// 保证升级后已均衡歌曲行为不变、seek_stream_test 的默认断言继续通过。
func TestNormalizeLoudness_DefaultEqualsConst(t *testing.T) {
	cs := newLoudnessCacheService(t)
	if got := cs.normalizeLoudnessFilter(); got != loudnormFilter {
		t.Errorf("默认滤镜串: got %q want %q (const)", got, loudnormFilter)
	}
	if got := cs.normalizeLoudnessTag(); got != "norm-16" {
		t.Errorf("默认缓存标记: got %q want %q", got, "norm-16")
	}
}

// TestNormalizeLoudness_CustomValue 设置 -14 后滤镜串/缓存标记/NormalizeLoudness 都反映新值。
func TestNormalizeLoudness_CustomValue(t *testing.T) {
	cs := newLoudnessCacheService(t)
	if err := cs.configService.Set(VolumeNormalizeLoudnessKey, "-14"); err != nil {
		t.Fatalf("set loudness: %v", err)
	}
	if v := cs.NormalizeLoudness(); v != -14 {
		t.Errorf("NormalizeLoudness: got %v want -14", v)
	}
	if got := cs.normalizeLoudnessFilter(); got != "loudnorm=I=-14:LRA=11:TP=-1.5" {
		t.Errorf("filter: got %q want loudnorm=I=-14:LRA=11:TP=-1.5", got)
	}
	if got := cs.normalizeLoudnessTag(); got != "norm-14" {
		t.Errorf("tag: got %q want norm-14", got)
	}
}

// TestNormalizeLoudness_FractionalValue 小数响度也按 %g 精确反映。
func TestNormalizeLoudness_FractionalValue(t *testing.T) {
	cs := newLoudnessCacheService(t)
	if err := cs.configService.Set(VolumeNormalizeLoudnessKey, "-14.5"); err != nil {
		t.Fatalf("set loudness: %v", err)
	}
	if got := cs.normalizeLoudnessFilter(); got != "loudnorm=I=-14.5:LRA=11:TP=-1.5" {
		t.Errorf("filter: got %q want loudnorm=I=-14.5:LRA=11:TP=-1.5", got)
	}
	if got := cs.normalizeLoudnessTag(); got != "norm-14.5" {
		t.Errorf("tag: got %q want norm-14.5", got)
	}
}

// TestNormalizeLoudness_InvalidFallsBack 非法字符串回落默认，不 panic、不崩转码。
func TestNormalizeLoudness_InvalidFallsBack(t *testing.T) {
	cs := newLoudnessCacheService(t)
	if err := cs.configService.Set(VolumeNormalizeLoudnessKey, "not-a-number"); err != nil {
		t.Fatalf("set loudness: %v", err)
	}
	if v := cs.NormalizeLoudness(); v != defaultNormalizeLoudness {
		t.Errorf("invalid config: got %v want %v", v, defaultNormalizeLoudness)
	}
}

// TestNormalizeLoudness_OutOfRangeClamps 越界值夹到边界（坏配置不崩，仍给一个合法响度）。
func TestNormalizeLoudness_OutOfRangeClamps(t *testing.T) {
	cs := newLoudnessCacheService(t)
	for _, c := range []struct {
		in   string
		want float64
	}{
		{"-100", minNormalizeLoudness},
		{"0", maxNormalizeLoudness}, // 0 > -5，夹到上限 -5
		{"100", maxNormalizeLoudness},
	} {
		if err := cs.configService.Set(VolumeNormalizeLoudnessKey, c.in); err != nil {
			t.Fatalf("set %s: %v", c.in, err)
		}
		if v := cs.NormalizeLoudness(); v != c.want {
			t.Errorf("clamp(%s): got %v want %v", c.in, v, c.want)
		}
	}
}

// TestValidateNormalizeLoudness 校验函数对边界内放行、越界拒绝。
func TestValidateNormalizeLoudness(t *testing.T) {
	for _, v := range []float64{-40, -23, -16, -14, -5, -14.5} {
		if err := ValidateNormalizeLoudness(v); err != nil {
			t.Errorf("valid %g should pass, got %v", v, err)
		}
	}
	for _, v := range []float64{-40.1, -50, 0, 5, -4.9} {
		if err := ValidateNormalizeLoudness(v); err == nil {
			t.Errorf("invalid %g should fail, got nil", v)
		}
	}
}

// TestTranscodedFileName_NormalizeTag 含响度值的缓存键区分不同响度产物，
// 默认 -16 时文件名含 "norm-16."。
func TestTranscodedFileName_NormalizeTag(t *testing.T) {
	cs := newLoudnessCacheService(t)
	local := &models.Song{ID: 42, Type: "local"}

	// 默认 -16
	if name := cs.transcodedFileName(local, "mp3", 0, -1, true); name != "42.tc.norm-16.mp3" {
		t.Errorf("default normalize name: got %q want %q", name, "42.tc.norm-16.mp3")
	}

	// -14 后文件名变 norm-14，与 -16 产物不共用缓存
	if err := cs.configService.Set(VolumeNormalizeLoudnessKey, "-14"); err != nil {
		t.Fatalf("set loudness: %v", err)
	}
	if name := cs.transcodedFileName(local, "mp3", 0, -1, true); name != "42.tc.norm-14.mp3" {
		t.Errorf("-14 name: got %q want %q", name, "42.tc.norm-14.mp3")
	}
	// 不均衡的产物不受响度影响
	if name := cs.transcodedFileName(local, "mp3", 0, -1, false); name != "42.tc.mp3" {
		t.Errorf("non-normalize name: got %q want %q", name, "42.tc.mp3")
	}
}
