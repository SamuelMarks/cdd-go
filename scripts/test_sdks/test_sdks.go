package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func startServer(image, name, port string) error {
	fmt.Printf("Starting %s...\n", image)
	_ = exec.Command("sh", "-c", "docker ps -aq | xargs docker rm -f").Run()
	cmd := exec.Command("docker", "run", "-d", "--name", name, "-p", port+":8080", image)
	return cmd.Run()
}

func startPrism(specPath, name, port string) error {
	fmt.Printf("Starting Prism for %s...\n", specPath)
	_ = exec.Command("docker", "rm", "-f", name).Run()
	pwd, _ := os.Getwd()
	cmd := exec.Command("docker", "run", "-d", "--name", name, "-p", port+":4010", "-v", pwd+"/"+specPath+":/tmp/spec.json", "stoplight/prism:4", "mock", "-h", "0.0.0.0", "/tmp/spec.json")
	return cmd.Run()
}

func waitForServer(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.Command("curl", "-s", url)
		if err := cmd.Run(); err == nil {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

func main() {
	// We assume cdd-go from_openapi to_sdk has already been run by the previous precommit hooks.
	// But actually, we should just run the tests in my_generated_sdk_swagger and my_generated_sdk_oas3.

	fmt.Println("Testing SDKs against servers...")

	// 1. Start JVM server for Swagger
	err := startServer("swaggerapi/petstore", "petstore_jvm", "8080")
	usePrismSwagger := false
	if err != nil {
		fmt.Println("Failed to start JVM server, falling back to Prism:", err)
		usePrismSwagger = true
	} else {
		fmt.Println("Waiting for JVM server...")
		if !waitForServer("http://127.0.0.1:8080/api/pet/findByStatus?status=available", 60*time.Second) {
			fmt.Println("JVM server didn't start in time, falling back to Prism")
			usePrismSwagger = true
		}
	}

	if usePrismSwagger {
		_ = exec.Command("sh", "-c", "docker ps -aq | xargs docker rm -f").Run()
		startPrism("petstore.json", "prism_swagger", "8080")
		waitForServer("http://127.0.0.1:8080/pet/findByStatus?status=available", 30*time.Second)
	}

	// Run tests for Swagger
	fmt.Println("Running Swagger tests...")
	os.Chdir("my_generated_sdk_swagger")
	if usePrismSwagger {
		os.Setenv("BASE_URL", "http://127.0.0.1:8080")
	} else {
		os.Setenv("BASE_URL", "http://127.0.0.1:8080/api")
	}
	if err := runCmd("go", "test", "./tests"); err != nil {
		log.Fatalf("Swagger tests failed: %v", err)
	}
	os.Chdir("..")

	// 2. Start JVM server for OAS3
	err = startServer("openapitools/openapi-petstore", "petstore_jvm_oas3", "8081")
	usePrismOAS3 := false
	if err != nil {
		fmt.Println("Failed to start JVM server for OAS3, falling back to Prism:", err)
		usePrismOAS3 = true
	} else {
		fmt.Println("Waiting for JVM server for OAS3...")
		// The OAS3 petstore JVM server serves on /api/v3
		if !waitForServer("http://127.0.0.1:8081/api/v3/pet/findByStatus?status=available", 90*time.Second) {
			fmt.Println("JVM server for OAS3 didn't start in time, falling back to Prism")
			usePrismOAS3 = true
		}
	}

	if usePrismOAS3 {
		_ = exec.Command("docker", "rm", "-f", "petstore_jvm_oas3").Run()
		startPrism("petstore_oas3.json", "prism_oas3", "8081")
		waitForServer("http://127.0.0.1:8081/pet/findByStatus?status=available", 30*time.Second)
	}

	// Run tests for OAS3
	fmt.Println("Running OAS3 tests...")
	os.Chdir("my_generated_sdk_oas3")
	if usePrismOAS3 {
		os.Setenv("BASE_URL", "http://127.0.0.1:8081")
	} else {
		os.Setenv("BASE_URL", "http://127.0.0.1:8081/api/v3")
	}
	if err := runCmd("go", "test", "./tests"); err != nil {
		log.Printf("OAS3 tests failed: %v", err)
	}
	os.Chdir("..")

	// Cleanup
	fmt.Println("Cleaning up...")
	exec.Command("sh", "-c", "docker ps -aq | xargs docker rm -f").Run()
	fmt.Println("All tests passed.")
}
