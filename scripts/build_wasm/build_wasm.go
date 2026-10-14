// Package main compiles the cdd-go CLI binary to a WebAssembly (GOOS=js GOARCH=wasm) target.
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

// osStdout is an injectable writer wrapping os.Stdout.
var osStdout = os.Stdout

// osStderr is an injectable writer wrapping os.Stderr.
var osStderr = os.Stderr

// main executes go build with GOOS=js and GOARCH=wasm for ./cmd/cdd-go.
func main() {
	cmd := execCommand("go", "build", "-o", "cdd-go.wasm", "./cmd/cdd-go")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = osStdout
	cmd.Stderr = osStderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(osStderr, "Failed to build WASM:", err)
		osExit(1)
	}
	fmt.Println("WASM built successfully")
}
