package main

import (
	"os/exec"
	"testing"
)

func TestTestServers(t *testing.T) {
	oldExecCommand := execCommand
	oldRunCmd := runCmd
	oldOsChdir := osChdir
	oldOsExit := osExit
	defer func() {
		execCommand = oldExecCommand
		runCmd = oldRunCmd
		osChdir = oldOsChdir
		osExit = oldOsExit
	}()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo")
	}
	if err := oldRunCmd("echo"); err != nil {
		t.Fatalf("expected nil error from echo command, got %v", err)
	}

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	if err := oldRunCmd("false"); err == nil {
		t.Fatalf("expected non-nil error from false command")
	}

	exitCode := 0
	osExit = func(code int) {
		exitCode = code
	}

	var chdirs []string
	var cmds []string
	runCmd = func(name string, args ...string) error {
		cmds = append(cmds, name)
		return nil
	}
	osChdir = func(dir string) error {
		chdirs = append(chdirs, dir)
		return nil
	}

	main()
	if exitCode != 0 {
		t.Errorf("expected exitCode 0 on success, got %d", exitCode)
	}
	if len(chdirs) != 4 {
		t.Errorf("expected 4 chdir calls, got %d (%v)", len(chdirs), chdirs)
	}
	if len(cmds) != 4 {
		t.Errorf("expected 4 go commands executed, got %d", len(cmds))
	}

	// Test command failure path
	exitCode = 0
	runCmd = func(name string, args ...string) error {
		return exec.Command("false").Run()
	}
	runMain()
	if exitCode != 1 {
		t.Errorf("expected exitCode 1 when server build/test fails, got %d", exitCode)
	}

	// Test chdir failure path
	exitCode = 0
	osChdir = func(dir string) error {
		return exec.Command("false").Run()
	}
	runMain()
	if exitCode != 1 {
		t.Errorf("expected exitCode 1 when chdir fails, got %d", exitCode)
	}
}
