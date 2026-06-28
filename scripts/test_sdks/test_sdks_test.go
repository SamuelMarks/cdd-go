package main

import (
	"os/exec"
	"testing"
	"time"
)

func TestTestSdks(t *testing.T) {
	// Backup
	oldExecCommand := execCommand
	oldRunCmd := runCmd
	oldOsChdir := osChdir
	oldTimeSleep := timeSleep
	oldWaitTimeout := waitTimeout
	oldWaitTimeoutOAS3 := waitTimeoutOAS3
	oldOsExit := osExit
	defer func() {
		execCommand = oldExecCommand
		runCmd = oldRunCmd
		osChdir = oldOsChdir
		timeSleep = oldTimeSleep
		waitTimeout = oldWaitTimeout
		waitTimeoutOAS3 = oldWaitTimeoutOAS3
		osExit = oldOsExit
	}()

	oldRunCmd("echo")
	oldRunCmd("false")

	timeSleep = func(d time.Duration) {}
	waitTimeout = 1 * time.Millisecond
	waitTimeoutOAS3 = 1 * time.Millisecond
	osExit = func(code int) {}

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo") // always succeeds
	}
	runCmd = func(name string, args ...string) error {
		return nil
	}
	osChdir = func(dir string) error {
		return nil
	}

	runCmd("echo")
	runCmd("false")

	main()

	// Try failure path where JVM fails to start and falls back to prism
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "docker" && len(args) > 0 && args[0] == "run" && (args[len(args)-1] == "swaggerapi/petstore" || args[len(args)-1] == "openapitools/openapi-petstore") {
			return exec.Command("false")
		}
		if name == "curl" {
			return exec.Command("true")
		}
		return exec.Command("echo")
	}
	runMain()

	// Test failure on curl but success on docker
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "curl" {
			return exec.Command("false")
		}
		return exec.Command("echo")
	}
	runCmd = func(name string, args ...string) error {
		return exec.Command("false").Run()
	}
	runMain()

	// Test npx fallback to docker
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "docker" && len(args) > 0 && args[0] == "run" && (args[len(args)-1] == "swaggerapi/petstore" || args[len(args)-1] == "openapitools/openapi-petstore") {
			return exec.Command("false")
		}
		if name == "npx" {
			return exec.Command("this_command_does_not_exist_12345")
		}
		if name == "curl" {
			return exec.Command("true")
		}
		return exec.Command("echo")
	}
	runCmd = func(name string, args ...string) error {
		return nil
	}
	runMain()

}

func TestStartPrismCoverage(t *testing.T) {
	oldExecCommand := execCommand
	oldTimeSleep := timeSleep
	defer func() {
		execCommand = oldExecCommand
		timeSleep = oldTimeSleep
	}()

	timeSleep = func(d time.Duration) {}
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo") // always succeeds
	}

	// 1. absolute path that exists
	startPrism("/tmp", "name1", "8080")

	// 2. absolute path that does not exist
	startPrism("/tmp/does_not_exist_123", "name2", "8080")

	// 3. relative path that does not exist, and parent does not exist
	startPrism("does_not_exist_123", "name3", "8080")

	// 4. npx fallback
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "npx" {
			return exec.Command("this_command_does_not_exist_12345")
		}
		return exec.Command("echo")
	}
	startPrism("/tmp", "name4", "8080")
}

func TestMissingBranches(t *testing.T) {
	oldExecCommand := execCommand
	oldTimeSleep := timeSleep
	oldOsChdir := osChdir
	oldWaitTimeout := waitTimeout
	oldWaitTimeoutOAS3 := waitTimeoutOAS3
	oldOsExit := osExit
	defer func() {
		execCommand = oldExecCommand
		timeSleep = oldTimeSleep
		osChdir = oldOsChdir
		waitTimeout = oldWaitTimeout
		waitTimeoutOAS3 = oldWaitTimeoutOAS3
		osExit = oldOsExit
	}()

	timeSleep = func(d time.Duration) {}
	osChdir = func(dir string) error { return nil }
	waitTimeout = 1 * time.Millisecond
	waitTimeoutOAS3 = 1 * time.Millisecond
	osExit = func(code int) {}

	// Cover startPrism lines
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo")
	}
	startPrism("test_sdks", "name5", "8080")

	// Cover runMain failure to start JVM
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "curl" {
			return exec.Command("false")
		}
		if name == "docker" && len(args) > 0 && args[0] == "run" {
			return exec.Command("false")
		}
		return exec.Command("echo")
	}
	runMain()
}
