package main

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestBuildWasm(t *testing.T) {
	oldExecCommand := execCommand
	oldExecLookPath := execLookPath
	oldOsExit := osExit
	oldStdout := osStdout
	oldStderr := osStderr
	defer func() {
		execCommand = oldExecCommand
		execLookPath = oldExecLookPath
		osExit = oldOsExit
		osStdout = oldStdout
		osStderr = oldStderr
	}()

	osStdout = nil
	osStderr = nil

	// Test success path with wasm-opt available
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/wasm-opt", nil
	}
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo")
	}
	main()

	// Test success path without wasm-opt available
	execLookPath = func(file string) (string, error) {
		return "", errors.New("not found")
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
