package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateCoverage(t *testing.T) {
	dir := t.TempDir()

	// Create dummy go file with coverage 50%
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n// Doc\ntype A struct{}\ntype B struct{}\n"), 0644)

	cov, err := calculateCoverage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cov != 50.0 {
		t.Errorf("expected 50.0%% coverage, got %f", cov)
	}

	// Empty dir
	emptyDir := filepath.Join(dir, "empty")
	os.MkdirAll(emptyDir, 0755)
	cov, err = calculateCoverage(emptyDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cov != 100.0 {
		t.Errorf("expected 100.0%% coverage for empty dir, got %f", cov)
	}

	// Error reading directory
	_, err = calculateCoverage("/non/existent/dir")
	if err == nil {
		t.Errorf("expected error reading non-existent dir")
	}

	// Parse error
	os.WriteFile(filepath.Join(dir, "bad.go"), []byte("package a; var var var"), 0644)
	_, err = calculateCoverage(dir)
	if err == nil {
		t.Errorf("expected parse error")
	}

	// Add more stuff
	os.Remove(filepath.Join(dir, "bad.go"))
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package a\n// C\nfunc C() {}\nvar D int\nconst E = 1\n"), 0644)
	cov, err = calculateCoverage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cov != 40.0 { // 5 total, 2 with docs
		t.Errorf("expected 40.0%% coverage, got %f", cov)
	}

	// Hit ValueSpec doc
	os.WriteFile(filepath.Join(dir, "c.go"), []byte("package a\n// F\nvar F int\n"), 0644)
	cov, err = calculateCoverage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cov != 50.0 { // 6 total, 3 with docs
		t.Errorf("expected 50.0%% coverage, got %f", cov)
	}
}

func TestMainCoverage(t *testing.T) {
	dir := t.TempDir()

	// Test success
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n// Doc\ntype A struct{}\n"), 0644)
	runMain(dir)

	// Test error
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		if code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	}

	runMain("/non/existent/dir")
	if !exitCalled {
		t.Errorf("expected os.Exit to be called")
	}
}

func TestMainDirectly(t *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	osExit = func(code int) {}

	// main normally runs with src, cmd, cdd. Since we are inside scripts, those might not exist or might exist.
	// But it covers the code path.
	main()
}

func TestRunMainEmpty(t *testing.T) {
	dir := t.TempDir()
	runMain(dir)
}

func TestMainError(t *testing.T) {
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Go to a place where "src" exists but is unreadable or just a file
	dir := t.TempDir()
	os.Chdir(dir)
	os.WriteFile("src", []byte("file, not dir"), 0644)

	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	exitCalled := false
	osExit = func(code int) { exitCalled = true }

	main()
	if !exitCalled {
		t.Errorf("expected osExit for error in main")
	}
}

func TestMainEmptyDirs(t *testing.T) {
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	dir := t.TempDir()
	os.Chdir(dir)
	os.MkdirAll("src", 0755)
	os.MkdirAll("cmd", 0755)
	os.MkdirAll("cdd", 0755)

	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	osExit = func(code int) {}

	main()
}

func TestMainWithFiles(t *testing.T) {
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	dir := t.TempDir()
	os.Chdir(dir)
	os.MkdirAll("src", 0755)
	os.WriteFile("src/a.go", []byte("package a\n// A is doc\ntype A int\n"), 0644)
	os.MkdirAll("cmd", 0755)
	os.MkdirAll("cdd", 0755)

	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	osExit = func(code int) {}

	main()
}
