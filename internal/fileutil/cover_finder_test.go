package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindExternalCover_PriorityOrder(t *testing.T) {
	tmpDir := t.TempDir()

	// cover.png should be found first (highest priority)
	coverPath := filepath.Join(tmpDir, "cover.png")
	writeFile(t, coverPath)

	// Also create folder.jpg (lower priority)
	writeFile(t, filepath.Join(tmpDir, "folder.jpg"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("expected cover.png to be found")
	}
	if filepath.Base(result) != "cover.png" {
		t.Errorf("expected cover.png, got %s", filepath.Base(result))
	}
}

func TestFindExternalCover_CoverJpg(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, filepath.Join(tmpDir, "cover.jpg"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != "cover.jpg" {
		t.Errorf("expected cover.jpg, got %s", filepath.Base(result))
	}
}

func TestFindExternalCover_FolderJpg(t *testing.T) {
	tmpDir := t.TempDir()

	// Only folder.jpg, no cover.*
	writeFile(t, filepath.Join(tmpDir, "folder.jpg"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != "folder.jpg" {
		t.Errorf("expected folder.jpg, got %s", filepath.Base(result))
	}
}

func TestFindExternalCover_DotFolder(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, filepath.Join(tmpDir, ".folder.png"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != ".folder.png" {
		t.Errorf("expected .folder.png, got %s", filepath.Base(result))
	}
}

func TestFindExternalCover_AlbumArtSmall(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, filepath.Join(tmpDir, "AlbumArtSmall.jpg"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != "AlbumArtSmall.jpg" {
		t.Errorf("expected AlbumArtSmall.jpg, got %s", filepath.Base(result))
	}
}

func TestFindExternalCover_CaseInsensitiveExt(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, filepath.Join(tmpDir, "cover.JPG"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// On case-insensitive filesystems (macOS), os.Stat("cover.jpg") matches
	// "cover.JPG", so the returned path may use the lowercase extension.
	// The important thing is that the file is found.
	base := strings.ToLower(filepath.Base(result))
	if base != "cover.jpg" && base != "cover.jpeg" {
		t.Errorf("expected cover.jpg or cover.JPG, got %s", filepath.Base(result))
	}
}

func TestFindExternalCover_SingleImageFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// No conventional name, but a single image in the directory
	writeFile(t, filepath.Join(tmpDir, "random-artwork.png"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != "random-artwork.png" {
		t.Errorf("expected random-artwork.png, got %s", filepath.Base(result))
	}
}

func TestFindExternalCover_MultipleImagesNoFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// Multiple images, none with conventional names → no fallback
	writeFile(t, filepath.Join(tmpDir, "a.png"))
	writeFile(t, filepath.Join(tmpDir, "b.jpg"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result for multiple images, got %s", result)
	}
}

func TestFindExternalCover_NonImageIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	// Only a single image, but also a non-image file
	writeFile(t, filepath.Join(tmpDir, "notes.txt"))
	writeFile(t, filepath.Join(tmpDir, "my-cover.png"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != "my-cover.png" {
		t.Errorf("expected my-cover.png, got %s", filepath.Base(result))
	}
}

func TestFindExternalCover_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result, got %s", result)
	}
}

func TestFindExternalCover_NonexistentDir(t *testing.T) {
	result, err := FindExternalCover("/nonexistent/path/12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result, got %s", result)
	}
}

func TestFindExternalCover_WebpFormat(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, filepath.Join(tmpDir, "cover.webp"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != "cover.webp" {
		t.Errorf("expected cover.webp, got %s", filepath.Base(result))
	}
}

func TestFindExternalCover_PngPriorityOverJpg(t *testing.T) {
	tmpDir := t.TempDir()

	// Both cover.png and cover.jpg exist → png takes priority
	writeFile(t, filepath.Join(tmpDir, "cover.jpg"))
	writeFile(t, filepath.Join(tmpDir, "cover.png"))

	result, err := FindExternalCover(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(result) != "cover.png" {
		t.Errorf("expected cover.png (higher priority), got %s", filepath.Base(result))
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("fake-image-data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}
