package cdd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateMiddlewares(t *testing.T) {
	outDir := t.TempDir()

	if err := GenerateMiddlewares(outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "middlewares", "middlewares.go")); os.IsNotExist(err) {
		t.Errorf("expected middlewares.go to exist")
	}
}

func TestGenerateMiddlewares_Errors(t *testing.T) {
	if err := GenerateMiddlewares("/dev/null/invalid"); err == nil {
		t.Errorf("expected error for invalid directory")
	}

	outDir := t.TempDir()
	os.MkdirAll(filepath.Join(outDir, "middlewares"), 0555) // read-only
	if err := GenerateMiddlewares(outDir); err == nil {
		t.Errorf("expected error for read-only directory")
	}
	os.Chmod(filepath.Join(outDir, "middlewares"), 0755) // restore
}
