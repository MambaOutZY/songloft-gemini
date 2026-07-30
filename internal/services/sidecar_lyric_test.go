package services

import (
	"os"
	"path/filepath"
	"testing"

	"songloft/internal/database"
	"songloft/internal/models"
)

func TestFindSidecarLyricFile(t *testing.T) {
	dir := t.TempDir()

	// Create audio file placeholder
	audioPath := filepath.Join(dir, "song.mp3")
	os.WriteFile(audioPath, []byte("fake"), 0644)

	tests := []struct {
		name      string
		setup     func()
		audioPath string
		wantFile  string
	}{
		{
			name:      "lowercase .lrc",
			setup:     func() { os.WriteFile(filepath.Join(dir, "song.lrc"), []byte("[00:01]hi"), 0644) },
			audioPath: audioPath,
			wantFile:  filepath.Join(dir, "song.lrc"),
		},
		{
			name: "uppercase .LRC",
			setup: func() {
				os.Remove(filepath.Join(dir, "song.lrc"))
				os.WriteFile(filepath.Join(dir, "song.LRC"), []byte("[00:01]hi"), 0644)
			},
			audioPath: audioPath,
			wantFile:  filepath.Join(dir, "song.LRC"),
		},
		{
			name: "mixed case .Lrc",
			setup: func() {
				os.Remove(filepath.Join(dir, "song.LRC"))
				os.WriteFile(filepath.Join(dir, "song.Lrc"), []byte("[00:01]hi"), 0644)
			},
			audioPath: audioPath,
			wantFile:  filepath.Join(dir, "song.Lrc"),
		},
		{
			name: "full filename variant song.mp3.lrc",
			setup: func() {
				os.Remove(filepath.Join(dir, "song.Lrc"))
				os.WriteFile(filepath.Join(dir, "song.mp3.lrc"), []byte("[00:01]hi"), 0644)
			},
			audioPath: audioPath,
			wantFile:  filepath.Join(dir, "song.mp3.lrc"),
		},
		{
			name: "full filename variant song.mp3.LRC",
			setup: func() {
				os.Remove(filepath.Join(dir, "song.mp3.lrc"))
				os.WriteFile(filepath.Join(dir, "song.mp3.LRC"), []byte("[00:01]hi"), 0644)
			},
			audioPath: audioPath,
			wantFile:  filepath.Join(dir, "song.mp3.LRC"),
		},
		{
			name: "zero-byte lrc is ignored",
			setup: func() {
				os.Remove(filepath.Join(dir, "song.mp3.LRC"))
				os.WriteFile(filepath.Join(dir, "song.lrc"), []byte{}, 0644)
			},
			audioPath: audioPath,
			wantFile:  "",
		},
		{
			name: "directory named song.lrc is ignored",
			setup: func() {
				os.Remove(filepath.Join(dir, "song.lrc"))
				os.MkdirAll(filepath.Join(dir, "song.lrc"), 0755)
			},
			audioPath: audioPath,
			wantFile:  "",
		},
		{
			name:      "no lrc at all",
			setup:     func() { os.RemoveAll(filepath.Join(dir, "song.lrc")) },
			audioPath: audioPath,
			wantFile:  "",
		},
		{
			name:      "empty audio path",
			setup:     func() {},
			audioPath: "",
			wantFile:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, _ := FindSidecarLyricFile(tt.audioPath)
			if got != tt.wantFile {
				t.Errorf("FindSidecarLyricFile() = %q, want %q", got, tt.wantFile)
			}
		})
	}
}

func TestFindSidecarLyricFile_BasePreferredOverFull(t *testing.T) {
	dir := t.TempDir()
	audioPath := filepath.Join(dir, "a.mp3")
	os.WriteFile(audioPath, []byte("fake"), 0644)

	// Both base.lrc and full.lrc exist — base wins
	os.WriteFile(filepath.Join(dir, "a.lrc"), []byte("[00:01]base"), 0644)
	os.WriteFile(filepath.Join(dir, "a.mp3.lrc"), []byte("[00:01]full"), 0644)

	got, _ := FindSidecarLyricFile(audioPath)
	if got != filepath.Join(dir, "a.lrc") {
		t.Errorf("expected base variant to win, got %q", got)
	}
}

func TestReadSidecarLyric(t *testing.T) {
	dir := t.TempDir()
	audioPath := filepath.Join(dir, "song.flac")
	os.WriteFile(audioPath, []byte("fake"), 0644)

	t.Run("utf8 normal", func(t *testing.T) {
		os.WriteFile(filepath.Join(dir, "song.lrc"), []byte("[00:01.00]Hello World"), 0644)
		got := ReadSidecarLyric(audioPath)
		if got != "[00:01.00]Hello World" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("whitespace-only returns empty", func(t *testing.T) {
		os.WriteFile(filepath.Join(dir, "song.lrc"), []byte("   \n\t  \n"), 0644)
		got := ReadSidecarLyric(audioPath)
		if got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("UTF-16 LE with BOM", func(t *testing.T) {
		// BOM FF FE + "Hi" in UTF-16LE
		content := []byte{0xFF, 0xFE, 'H', 0x00, 'i', 0x00}
		os.WriteFile(filepath.Join(dir, "song.lrc"), content, 0644)
		got := ReadSidecarLyric(audioPath)
		if got != "Hi" {
			t.Errorf("got %q, want %q", got, "Hi")
		}
	})

	t.Run("UTF-16 BE with BOM", func(t *testing.T) {
		// BOM FE FF + "Hi" in UTF-16BE
		content := []byte{0xFE, 0xFF, 0x00, 'H', 0x00, 'i'}
		os.WriteFile(filepath.Join(dir, "song.lrc"), content, 0644)
		got := ReadSidecarLyric(audioPath)
		if got != "Hi" {
			t.Errorf("got %q, want %q", got, "Hi")
		}
	})

	t.Run("UTF-8 BOM stripped", func(t *testing.T) {
		// UTF-8 BOM EF BB BF + "[00:01]Hi"
		content := append([]byte{0xEF, 0xBB, 0xBF}, []byte("[00:01]Hi")...)
		os.WriteFile(filepath.Join(dir, "song.lrc"), content, 0644)
		got := ReadSidecarLyric(audioPath)
		if got != "[00:01]Hi" {
			t.Errorf("got %q, want %q", got, "[00:01]Hi")
		}
	})

	t.Run("no sidecar returns empty", func(t *testing.T) {
		os.Remove(filepath.Join(dir, "song.lrc"))
		got := ReadSidecarLyric(audioPath)
		if got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})
}

func TestSidecarLyricForSong(t *testing.T) {
	dir := t.TempDir()
	audioPath := filepath.Join(dir, "test.mp3")
	os.WriteFile(audioPath, []byte("fake"), 0644)
	os.WriteFile(filepath.Join(dir, "test.lrc"), []byte("[00:01]lyrics"), 0644)

	tests := []struct {
		name string
		song *models.Song
		want string
	}{
		{
			name: "normal local song",
			song: &models.Song{Type: models.TypeLocal, FilePath: audioPath, LyricSource: models.LyricSourceScraped},
			want: "[00:01]lyrics",
		},
		{
			name: "remote song excluded",
			song: &models.Song{Type: models.TypeRemote, FilePath: audioPath},
			want: "",
		},
		{
			name: "CUE track excluded",
			song: &models.Song{Type: models.TypeLocal, FilePath: audioPath, CueSourcePath: "/some.cue"},
			want: "",
		},
		{
			name: "manual source excluded",
			song: &models.Song{Type: models.TypeLocal, FilePath: audioPath, LyricSource: models.LyricSourceManual},
			want: "",
		},
		{
			name: "empty file path excluded",
			song: &models.Song{Type: models.TypeLocal, FilePath: ""},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SidecarLyricForSong(tt.song)
			if got != tt.want {
				t.Errorf("SidecarLyricForSong() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNeedsSidecarLyricImport(t *testing.T) {
	dir := t.TempDir()
	audioPath := filepath.Join(dir, "x.mp3")
	os.WriteFile(audioPath, []byte("fake"), 0644)
	os.WriteFile(filepath.Join(dir, "x.lrc"), []byte("[00:01]test"), 0644)

	lyricDirs := map[string]struct{}{dir: {}}

	tests := []struct {
		name      string
		info      database.LocalPathInfo
		filePath  string
		lyricDirs map[string]struct{}
		want      bool
	}{
		{
			name:      "file source skips without stat",
			info:      database.LocalPathInfo{LyricSource: models.LyricSourceFile},
			filePath:  "/nonexistent/dir/x.mp3",
			lyricDirs: lyricDirs,
			want:      false,
		},
		{
			name:      "manual source skips without stat",
			info:      database.LocalPathInfo{LyricSource: models.LyricSourceManual},
			filePath:  "/nonexistent/dir/x.mp3",
			lyricDirs: lyricDirs,
			want:      false,
		},
		{
			name:      "dir not in lyricDirs",
			info:      database.LocalPathInfo{LyricSource: models.LyricSourceScraped},
			filePath:  filepath.Join("/other/dir", "x.mp3"),
			lyricDirs: lyricDirs,
			want:      false,
		},
		{
			name:      "nil lyricDirs",
			info:      database.LocalPathInfo{LyricSource: models.LyricSourceScraped},
			filePath:  audioPath,
			lyricDirs: nil,
			want:      false,
		},
		{
			name:      "scraped + dir has lrc + file exists",
			info:      database.LocalPathInfo{LyricSource: models.LyricSourceScraped},
			filePath:  audioPath,
			lyricDirs: lyricDirs,
			want:      true,
		},
		{
			name:      "scraped + dir has lrc but no matching lrc file",
			info:      database.LocalPathInfo{LyricSource: models.LyricSourceScraped},
			filePath:  filepath.Join(dir, "other.mp3"),
			lyricDirs: lyricDirs,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := needsSidecarLyricImport(tt.info, tt.filePath, tt.lyricDirs)
			if got != tt.want {
				t.Errorf("needsSidecarLyricImport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldApplyScanLyric(t *testing.T) {
	tests := []struct {
		name     string
		song     *models.Song
		newLyric string
		want     bool
	}{
		{
			name:     "manual never overwritten",
			song:     &models.Song{LyricSource: models.LyricSourceManual, Lyric: "old"},
			newLyric: "new",
			want:     false,
		},
		{
			name:     "non-empty lyric always applies",
			song:     &models.Song{LyricSource: models.LyricSourceScraped, Lyric: "old"},
			newLyric: "new",
			want:     true,
		},
		{
			name:     "empty lyric does not wipe existing lyric",
			song:     &models.Song{LyricSource: models.LyricSourceScraped, Lyric: "existing"},
			newLyric: "",
			want:     false,
		},
		{
			name:     "empty lyric does not wipe existing remote url",
			song:     &models.Song{LyricSource: models.LyricSourceScraped, LyricRemoteURL: "http://x"},
			newLyric: "",
			want:     false,
		},
		{
			name:     "empty lyric applies when song has no lyric at all",
			song:     &models.Song{LyricSource: "", Lyric: "", LyricRemoteURL: ""},
			newLyric: "",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldApplyScanLyric(tt.song, tt.newLyric)
			if got != tt.want {
				t.Errorf("shouldApplyScanLyric() = %v, want %v", got, tt.want)
			}
		})
	}
}
