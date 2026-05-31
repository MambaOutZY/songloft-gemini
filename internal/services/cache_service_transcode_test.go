package services

import (
	"os"
	"path/filepath"
	"testing"

	"songloft/internal/models"
)

func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"mp3", "mp3"},
		{"MP3", "mp3"},
		{"mpeg", "mp3"},
		{".mp3", "mp3"},
		{"m4a", "m4a"},
		{"aac", "m4a"},
		{"mp4", "m4a"},
		{"ogg", "ogg"},
		{"vorbis", "ogg"},
		{"flac", "flac"},
		{"wav", "wav"},
		{"wave", "wav"},
		{"wma", "wma"},
		{"asf", "wma"},
		{"ape", "ape"},
		{"unknown", "unknown"},
		{"", ""},
	}
	for _, tt := range tests {
		got := NormalizeFormat(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeFormat(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNeedsTranscode(t *testing.T) {
	tests := []struct {
		src    string
		target string
		want   bool
	}{
		{"mp3", "", false},
		{"mp3", "mp3", false},
		{"mpeg", "mp3", false},
		{"MP3", "mp3", false},
		{"wma", "mp3", true},
		{"ape", "flac", true},
		{"flac", "mp3", true},
		{"m4a", "aac", false},
		{"", "mp3", false}, // 未知源格式不转码，避免误判
	}
	for _, tt := range tests {
		got := NeedsTranscode(tt.src, tt.target)
		if got != tt.want {
			t.Errorf("NeedsTranscode(%q, %q) = %v, want %v", tt.src, tt.target, got, tt.want)
		}
	}
}

func TestFfmpegArgs(t *testing.T) {
	tests := []struct {
		format  string
		encoder string
		wantErr bool
	}{
		{"mp3", "libmp3lame", false},
		{"ogg", "libvorbis", false},
		{"m4a", "aac", false},
		{"flac", "flac", false},
		{"wav", "pcm_s16le", false},
		{"xyz", "", true},
	}
	for _, tt := range tests {
		enc, _, _, err := ffmpegArgs(tt.format)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ffmpegArgs(%q) should error", tt.format)
			}
			continue
		}
		if err != nil {
			t.Errorf("ffmpegArgs(%q) unexpected error: %v", tt.format, err)
			continue
		}
		if enc != tt.encoder {
			t.Errorf("ffmpegArgs(%q) encoder = %q, want %q", tt.format, enc, tt.encoder)
		}
	}
}

func TestTranscodedFileName(t *testing.T) {
	cs := &CacheService{cacheDir: "/tmp/test"}

	// 本地歌曲（无 cacheKey）
	local := &models.Song{ID: 42, Type: "local"}
	name := cs.transcodedFileName(local, "mp3")
	if name != "42.tc.mp3" {
		t.Errorf("transcodedFileName(local) = %q, want %q", name, "42.tc.mp3")
	}

	// 插件来源歌曲（有 cacheKey）
	remote := &models.Song{
		ID:              123,
		Type:            "remote",
		PluginEntryPath: "my-source",
		DedupKey:        "platform:12345",
	}
	name = cs.transcodedFileName(remote, "ogg")
	expected := "123.my-source_platform_12345.tc.ogg"
	if name != expected {
		t.Errorf("transcodedFileName(remote) = %q, want %q", name, expected)
	}
}

func TestFindTranscodedFile(t *testing.T) {
	tmpDir := t.TempDir()
	cs := &CacheService{cacheDir: tmpDir}

	song := &models.Song{ID: 100, Type: "local", Format: "wma"}

	// 不存在时应 miss
	if _, ok := cs.FindTranscodedFile(song, "mp3"); ok {
		t.Error("FindTranscodedFile should miss when file does not exist")
	}

	// 创建转码文件
	dir, _ := cs.getCachePath(song.ID, "")
	os.MkdirAll(dir, 0755)
	name := cs.transcodedFileName(song, "mp3")
	path := filepath.Join(dir, name)
	os.WriteFile(path, []byte("fake mp3"), 0644)

	// 现在应该命中
	found, ok := cs.FindTranscodedFile(song, "mp3")
	if !ok {
		t.Fatal("FindTranscodedFile should hit after creating file")
	}
	if found != path {
		t.Errorf("FindTranscodedFile path = %q, want %q", found, path)
	}

	// 不同格式应 miss
	if _, ok := cs.FindTranscodedFile(song, "ogg"); ok {
		t.Error("FindTranscodedFile should miss for different format")
	}
}
