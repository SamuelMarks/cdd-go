// Package main builds and runs unit tests for generated Swagger 2.0 and OpenAPI 3.0 server projects.
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// execCommand is an injectable function variable wrapping exec.Command.
var execCommand = exec.Command

// runCmd executes an external command and pipes its output to stdout and stderr.
var runCmd = func(name string, args ...string) error {
	cmd := execCommand(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// osChdir is an injectable function variable wrapping os.Chdir.
var osChdir = os.Chdir

// runMain executes go test and go build inside the generated Swagger and OpenAPI 3.0 server directories.
func runMain() {
	fmt.Println("Testing Generated Servers...")

	// Swagger
	fmt.Println("Running Swagger server tests...")
	if err := osChdir("my_generated_server_swagger"); err != nil {
		log.Printf("Failed to chdir into my_generated_server_swagger: %v", err)
	} else {
		// Run go test in the generated directory
		if err := runCmd("go", "test", "./..."); err != nil {
			log.Printf("Swagger server tests failed: %v", err)
		}
		// Build the server entrypoint
		if err := runCmd("go", "build", "./cmd/server/..."); err != nil {
			log.Printf("Swagger server build failed: %v", err)
		}
		osChdir("..")
	}

	// OAS3
	fmt.Println("Running OAS3 server tests...")
	if err := osChdir("my_generated_server_oas3"); err != nil {
		log.Printf("Failed to chdir into my_generated_server_oas3: %v", err)
	} else {
		// Run go test in the generated directory
		if err := runCmd("go", "test", "./..."); err != nil {
			log.Printf("OAS3 server tests failed: %v", err)
		}
		// Build the server entrypoint
		if err := runCmd("go", "build", "./cmd/server/..."); err != nil {
			log.Printf("OAS3 server build failed: %v", err)
		}
		osChdir("..")
	}

	fmt.Println("Server tests complete.")
}

// main is the entrypoint for the generated server test runner script.
func main() {
	runMain()
}
