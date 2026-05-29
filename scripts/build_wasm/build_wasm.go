package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "build", "-o", "cdd-go.wasm", "./cmd/cdd-go")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build WASM: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("WASM built successfully")
}
