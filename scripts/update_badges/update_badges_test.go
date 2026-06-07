package main

import (
	"os"
	"strings"
	"testing"
)

func TestGetColor(t *testing.T) {
	if getColor(95) != "brightgreen" {
		t.Error("expected brightgreen")
	}
	if getColor(85) != "green" {
		t.Error("expected green")
	}
	if getColor(75) != "yellowgreen" {
		t.Error("expected yellowgreen")
	}
	if getColor(65) != "yellow" {
		t.Error("expected yellow")
	}
	if getColor(55) != "orange" {
		t.Error("expected orange")
	}
	if getColor(45) != "red" {
		t.Error("expected red")
	}
}

func TestParseCoverage(t *testing.T) {
	if parseCoverage("total: (statements) 95.5%", `total:\s+\(statements\)\s+([0-9.]+)%`) != 95.5 {
		t.Error("expected 95.5")
	}
	if parseCoverage("total: (statements) fail%", `total:\s+\(statements\)\s+([0-9.]+)%`) != 0.0 {
		t.Error("expected 0.0")
	}
	if parseCoverage("100.0%", `([0-9.]+)%`) != 100.0 {
		t.Error("expected 100.0")
	}
}

func TestFormatCoverage(t *testing.T) {
	if formatCoverage(95.5) != "95.5" {
		t.Error("expected 95.5")
	}
	if formatCoverage(100.0) != "100" {
		t.Error("expected 100")
	}
}

func TestMainFunc(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	main()

	readmeContent := `[![Test Coverage](https://img.shields.io/badge/test_coverage-0%25-red.svg)](#)
[![Doc Coverage](https://img.shields.io/badge/doc_coverage-0%25-red.svg)](#)`
	os.WriteFile("README.md", []byte(readmeContent), 0644)

	main()

	content, _ := os.ReadFile("README.md")
	if !strings.Contains(string(content), "test_coverage") {
		t.Error("README.md was mangled")
	}

	// hit read error
	os.Chmod("README.md", 0000)
	main()
	os.Chmod("README.md", 0644)

	// Make an invalid PATH to force exec to fail
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")

	// Test getDocCov and getTestCov individually to ensure their error branches are hit
	getTestCov()
	getDocCov()

	main()
	os.Setenv("PATH", oldPath)
}
func TestGetDocCovErr(t *testing.T) {
	// mock path again to trigger doc failure
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	res := getDocCov()
	if res != 0.0 {
		t.Errorf("Expected 0.0 got %v", res)
	}
	os.Setenv("PATH", oldPath)
}

func TestGetDocCovFakeFail(t *testing.T) {
	// Create a fake doc_cover.go that fails
	os.MkdirAll("scripts/doc_cover", 0755)
	os.WriteFile("scripts/doc_cover/doc_cover.go", []byte("package main\nfunc main() { panic(\"fail\") }"), 0644)
	defer os.RemoveAll("scripts")

	if getDocCov() != 0.0 {
		t.Error("Expected 0.0")
	}
}

func TestGetDocCovSuccess(t *testing.T) {
	os.MkdirAll("scripts/doc_cover", 0755)
	os.WriteFile("scripts/doc_cover/doc_cover.go", []byte("package main\nimport \"fmt\"\nfunc main() { fmt.Print(\"100.0%\") }"), 0644)
	defer os.RemoveAll("scripts")

	if getDocCov() != 100.0 {
		t.Error("Expected 100.0")
	}
}
