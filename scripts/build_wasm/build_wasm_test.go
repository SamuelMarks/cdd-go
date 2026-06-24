package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func TestBuildWasm(t *testing.T) {
	// Backup and restore
	oldExecCommand := execCommand
	oldOsExit := osExit
	oldStdout := osStdout
	oldStderr := osStderr
	defer func() {
		execCommand = oldExecCommand
		osExit = oldOsExit
		osStdout = oldStdout
		osStderr = oldStderr
	}()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	osStdout = nil
	osStderr = nil

	// Test success path
	execCommand = func(name string, arg ...string) *exec.Cmd {
		// Just run a simple command that succeeds, like "true" or "echo"
		return exec.Command("echo")
	}
	main()

	// Test error path
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
	}
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("false")
	}
	main()
	if !exitCalled {
		t.Error("expected osExit to be called on failure")
	}
}
