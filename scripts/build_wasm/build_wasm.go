package main

import (
	"fmt"
	"os"
	"os/exec"
)

var osExit = os.Exit
var execCommand = exec.Command
var osStdout = os.Stdout
var osStderr = os.Stderr

func main() {
	cmd := execCommand("go", "build", "-o", "cdd-go.wasm", "./cmd/cdd-go")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = osStdout
	cmd.Stderr = osStderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(osStderr, "Failed to build WASM: %v\n", err)
		osExit(1)
	}
	fmt.Println("WASM built successfully")
}
