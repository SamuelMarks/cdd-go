package cdd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateIdP(t *testing.T) {
	outDir := t.TempDir()

	if err := GenerateIdP(outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "idp", "idp_handlers.go")); os.IsNotExist(err) {
		t.Errorf("expected idp_handlers.go to exist")
	}
}

func TestGenerateIdP_Errors(t *testing.T) {
	if err := GenerateIdP("/dev/null/invalid"); err == nil {
		t.Errorf("expected error for invalid directory")
	}

	outDir := t.TempDir()
	os.MkdirAll(filepath.Join(outDir, "idp"), 0555) // read-only
	if err := GenerateIdP(outDir); err == nil {
		t.Errorf("expected error for read-only directory")
	}
	os.Chmod(filepath.Join(outDir, "idp"), 0755) // restore
}
