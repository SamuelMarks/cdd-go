package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateCoverage(t *testing.T) {
	dir := t.TempDir()

	// File doc (1/1) + Type A (1/1) + Field X (1/1) + Type B (0/1) = 3/4 = 75%
	os.WriteFile(filepath.Join(dir, "a.go"), []byte(`// Package a doc
package a
// Doc
type A struct {
// X doc
X int
}
type B struct {
int
}
`), 0644)

	cov, err := calculateCoverage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cov != 75.0 {
		t.Errorf("expected 75.0%% coverage, got %f", cov)
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
	// b.go: File doc (0/1) + Func C (1/1) + Var D (0/1) + Const E (0/1) = 1/4
	// Total with a.go (3/4): 4/8 = 50%
	os.WriteFile(filepath.Join(dir, "b.go"), []byte(`package a
// C
func C() {}
var D int
const E = 1
`), 0644)
	cov, err = calculateCoverage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cov != 50.0 {
		t.Errorf("expected 50.0%% coverage, got %f", cov)
	}

	// Hit ValueSpec doc: File doc (1/1) + Var F (1/1) = 2/2 -> Total 6/10 = 60%
	os.WriteFile(filepath.Join(dir, "c.go"), []byte(`// Package a
package a
// F
var F int
`), 0644)
	cov, err = calculateCoverage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cov != 60.0 {
		t.Errorf("expected 60.0%% coverage, got %f", cov)
	}

	// Hit InterfaceType methods: File doc (1/1) + Type Iface (1/1) + Method DocM (1/1) + Method UndocM (0/1) + embedded A = 3/4 -> Total 9/14
	os.WriteFile(filepath.Join(dir, "d.go"), []byte(`// Package a
package a
// Iface doc
type Iface interface {
	// DocM doc
	DocM()
	UndocM()
	A
}
`), 0644)
	cov, err = calculateCoverage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cov < 64.0 || cov > 65.0 {
		t.Errorf("expected ~64.3%% coverage (9/14), got %f", cov)
	}
}

func TestMainCoverage(t *testing.T) {
	dir := t.TempDir()

	// Test success
	os.WriteFile(filepath.Join(dir, "a.go"), []byte(`// Package a
package a
// Doc
type A struct{}
`), 0644)
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

	main()
}

func TestRunMainEmpty(t *testing.T) {
	dir := t.TempDir()
	runMain(dir)
}

func TestMainError(t *testing.T) {
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	dir := t.TempDir()
	os.Chdir(dir)
	os.MkdirAll("src", 0755)
	os.WriteFile("src/bad.go", []byte("package bad; var var"), 0644)

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
	os.WriteFile("src/a.go", []byte(`// Package a
package a
// A is doc
type A int
`), 0644)
	os.MkdirAll("cmd", 0755)
	os.MkdirAll("cdd", 0755)

	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	exitCode := 0
	osExit = func(code int) { exitCode = code }

	main()
	if exitCode != 0 {
		t.Errorf("expected exitCode 0 for 100%% coverage, got %d", exitCode)
	}

	// Test <100% coverage triggers osExit(1)
	os.WriteFile("src/undoc.go", []byte("package a\ntype Undoc int\n"), 0644)
	main()
	if exitCode != 1 {
		t.Errorf("expected exitCode 1 when doc coverage < 100%%, got %d", exitCode)
	}
}
