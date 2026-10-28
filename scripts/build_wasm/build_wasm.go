// Package main compiles and optimizes the cdd-go CLI binary to a WebAssembly (GOOS=js GOARCH=wasm) target.
package main

import (
	"fmt"
	"os"
	"os/exec"
)

// osExit is an injectable function variable wrapping os.Exit.
var osExit = os.Exit

// execCommand is an injectable function variable wrapping exec.Command.
var execCommand = exec.Command

// execLookPath is an injectable function variable wrapping exec.LookPath.
var execLookPath = exec.LookPath

// osStdout is an injectable writer wrapping os.Stdout.
var osStdout = os.Stdout

// osStderr is an injectable writer wrapping os.Stderr.
var osStderr = os.Stderr

// main executes go build with size-reducing flags (-trimpath, -ldflags="-s -w", GOEXPERIMENT=nojsonv2)
// and optional wasm-opt post-processing for ./cmd/cdd-go.
func main() {
	cmd := execCommand("go", "build", "-trimpath", "-ldflags=-s -w", "-o", "cdd-go.wasm", "./cmd/cdd-go")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm", "GOEXPERIMENT=nojsonv2")
	cmd.Stdout = osStdout
	cmd.Stderr = osStderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(osStderr, "Failed to build WASM:", err)
		osExit(1)
	}
	if wasmOptPath, err := execLookPath("wasm-opt"); err == nil && wasmOptPath != "" {
		optCmd := execCommand(
			wasmOptPath,
			"-Oz",
			"--enable-bulk-memory-opt",
			"--enable-nontrapping-float-to-int",
			"--enable-sign-ext",
			"--enable-mutable-globals",
			"--strip-debug",
			"--strip-dwarf",
			"--strip-producers",
			"cdd-go.wasm",
			"-o",
			"cdd-go.wasm",
		)
		optCmd.Stdout = osStdout
		optCmd.Stderr = osStderr
		_ = optCmd.Run()
	}
	fmt.Println("WASM built successfully")
}
