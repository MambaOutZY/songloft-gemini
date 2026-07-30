package services

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hanxi/tag"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"songloft/internal/database"
	"songloft/internal/models"
)

// FindSidecarLyricFile locates a .lrc file beside the given audio file.
// Returns the matched path and its mod time, or ("", zero) if none found.
//
// Candidate order:
//   - <base>.lrc / .LRC / .Lrc
//   - <full filename>.lrc / .LRC / .Lrc  (e.g. "a.mp3.lrc")
//
// Zero-byte files and directories are treated as not found.
func FindSidecarLyricFile(audioPath string) (string, time.Time) {
	dir := filepath.Dir(audioPath)
	base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	full := filepath.Base(audioPath)

	candidates := []string{
		filepath.Join(dir, base+".lrc"),
		filepath.Join(dir, base+".LRC"),
		filepath.Join(dir, base+".Lrc"),
		filepath.Join(dir, full+".lrc"),
		filepath.Join(dir, full+".LRC"),
		filepath.Join(dir, full+".Lrc"),
	}

	for _, p := range candidates {
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		if st.IsDir() || st.Size() == 0 {
			continue
		}
		return p, st.ModTime()
	}
	return "", time.Time{}
}

// ReadSidecarLyric reads and decodes the .lrc file for the given audio path.
// Handles UTF-16 LE/BE (via BOM detection) and GBK (via tag.FixEncoding).
// Returns "" if no sidecar is found, the file is empty, or only whitespace.
func ReadSidecarLyric(audioPath string) string {
	lrcPath, _ := FindSidecarLyricFile(audioPath)
	if lrcPath == "" {
		return ""
	}

	raw, err := os.ReadFile(lrcPath)
	if err != nil {
		slog.Debug("ReadSidecarLyric: read failed", "path", lrcPath, "error", err)
		return ""
	}

	content := decodeBytes(raw)
	if strings.TrimSpace(content) == "" {
		return ""
	}
	return content
}

// decodeBytes detects UTF-16 BOM and decodes accordingly; otherwise uses tag.FixEncoding.
// UTF-8 BOM (EF BB BF) is stripped if present.
func decodeBytes(raw []byte) string {
	if len(raw) >= 3 && raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF {
		raw = raw[3:]
	}
	if len(raw) >= 2 {
		// UTF-16 LE BOM: FF FE
		if raw[0] == 0xFF && raw[1] == 0xFE {
			decoder := unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder()
			decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(raw), decoder))
			if err == nil {
				return string(decoded)
			}
		}
		// UTF-16 BE BOM: FE FF
		if raw[0] == 0xFE && raw[1] == 0xFF {
			decoder := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM).NewDecoder()
			decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(raw), decoder))
			if err == nil {
				return string(decoded)
			}
		}
	}
	return tag.FixEncoding(raw)
}

// SidecarLyricForSong returns the sidecar .lrc content for the song, or "" if not applicable.
// Excluded: non-local, empty FilePath, CUE split tracks, manual lyric source.
func SidecarLyricForSong(song *models.Song) string {
	if song.Type != models.TypeLocal {
		return ""
	}
	if song.FilePath == "" {
		return ""
	}
	if song.CueSourcePath != "" {
		return ""
	}
	if song.LyricSource == models.LyricSourceManual {
		return ""
	}
	return ReadSidecarLyric(song.FilePath)
}

// needsSidecarLyricImport determines whether a song needs re-extraction for sidecar lyric.
// Three-level short circuit:
//  1. LyricSource is file or manual → already imported or user-adjusted, skip
//  2. Song's directory not in lyricDirs → no lrc in that dir, skip
//  3. Actually stat the sidecar file for this specific audio
func needsSidecarLyricImport(info database.LocalPathInfo, filePath string, lyricDirs map[string]struct{}) bool {
	if info.LyricSource == models.LyricSourceFile || info.LyricSource == models.LyricSourceManual {
		return false
	}
	dir := filepath.Dir(filePath)
	if _, ok := lyricDirs[dir]; !ok {
		return false
	}
	lrcPath, _ := FindSidecarLyricFile(filePath)
	return lrcPath != ""
}

// shouldApplyScanLyric decides whether scan-extracted lyric should be written to the song.
// - manual: never overwrite
// - non-empty newLyric: always apply
// - empty newLyric: only apply if song has no existing lyric at all (don't wipe existing)
func shouldApplyScanLyric(song *models.Song, newLyric string) bool {
	if song.LyricSource == models.LyricSourceManual {
		return false
	}
	if newLyric != "" {
		return true
	}
	return song.Lyric == "" && song.LyricRemoteURL == ""
}
