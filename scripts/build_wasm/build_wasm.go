package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if err := os.MkdirAll("bin", 0755); err != nil {
		fmt.Println("Failed to create bin dir:", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", "bin/cdd-go.wasm", "./cmd/cdd-go")
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("WASM build failed:", err)
		os.Exit(1)
	}
}
