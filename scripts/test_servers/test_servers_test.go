package main

import (
	"os/exec"
	"testing"
)

func TestTestServers(t *testing.T) {
	oldExecCommand := execCommand
	oldRunCmd := runCmd
	oldOsChdir := osChdir
	defer func() {
		execCommand = oldExecCommand
		runCmd = oldRunCmd
		osChdir = oldOsChdir
	}()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo")
	}

	oldRunCmd("echo")

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	oldRunCmd("false")

	runCmd = func(name string, args ...string) error {
		return nil
	}
	osChdir = func(dir string) error {
		return nil
	}

	main()

	// Test failure path
	runCmd = func(name string, args ...string) error {
		return exec.Command("false").Run()
	}
	runMain()

	// Test chdir failure
	osChdir = func(dir string) error {
		return exec.Command("false").Run()
	}
	runMain()
}
