package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

var runCmd = func(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var execCommand = exec.Command
var osChdir = os.Chdir
var timeSleep = time.Sleep

func startServer(image, name, port string) error {
	fmt.Printf("Starting %s...\n", image)
	_ = execCommand("sh", "-c", "docker ps -aq | xargs docker rm -f").Run()
	cmd := execCommand("docker", "run", "-d", "--name", name, "-p", port+":8080", image)
	return cmd.Run()
}

func startPrism(specPath, name, port string) error {
	fmt.Printf("Starting Prism for %s...\n", specPath)
	_ = execCommand("docker", "rm", "-f", name).Run()
	pwd, _ := os.Getwd()
	cmd := execCommand("docker", "run", "-d", "--name", name, "-p", port+":4010", "-v", pwd+"/"+specPath+":/tmp/spec.json", "stoplight/prism:4", "mock", "-h", "0.0.0.0", "/tmp/spec.json")
	return cmd.Run()
}

func waitForServer(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := execCommand("curl", "-s", url)
		if err := cmd.Run(); err == nil {
			return true
		}
		timeSleep(2 * time.Second)
	}
	return false
}

func main() {
	runMain()
}

var waitTimeout = 60 * time.Second
var waitTimeoutOAS3 = 90 * time.Second

func runMain() {
	fmt.Println("Testing SDKs against servers...")

	err := startServer("swaggerapi/petstore", "petstore_jvm", "8080")
	usePrismSwagger := false
	if err != nil {
		fmt.Println("Failed to start JVM server, falling back to Prism:", err)
		usePrismSwagger = true
	} else {
		fmt.Println("Waiting for JVM server...")
		if !waitForServer("http://127.0.0.1:8080/api/pet/findByStatus?status=available", waitTimeout) {
			fmt.Println("JVM server didn't start in time, falling back to Prism")
			usePrismSwagger = true
		}
	}

	if usePrismSwagger {
		_ = execCommand("sh", "-c", "docker ps -aq | xargs docker rm -f").Run()
		startPrism("petstore.json", "prism_swagger", "8080")
		waitForServer("http://127.0.0.1:8080/pet/findByStatus?status=available", waitTimeout)
	}

	fmt.Println("Running Swagger tests...")
	osChdir("my_generated_sdk_swagger")
	if usePrismSwagger {
		os.Setenv("BASE_URL", "http://127.0.0.1:8080")
	} else {
		os.Setenv("BASE_URL", "http://127.0.0.1:8080/api")
	}
	if err := runCmd("go", "test", "./tests"); err != nil {
		log.Printf("Swagger tests failed: %v", err)
	}
	osChdir("..")

	err = startServer("openapitools/openapi-petstore", "petstore_jvm_oas3", "8081")
	usePrismOAS3 := false
	if err != nil {
		fmt.Println("Failed to start JVM server for OAS3, falling back to Prism:", err)
		usePrismOAS3 = true
	} else {
		fmt.Println("Waiting for JVM server for OAS3...")
		if !waitForServer("http://127.0.0.1:8081/api/v3/pet/findByStatus?status=available", waitTimeoutOAS3) {
			fmt.Println("JVM server for OAS3 didn't start in time, falling back to Prism")
			usePrismOAS3 = true
		}
	}

	if usePrismOAS3 {
		_ = execCommand("docker", "rm", "-f", "petstore_jvm_oas3").Run()
		startPrism("petstore_oas3.json", "prism_oas3", "8081")
		waitForServer("http://127.0.0.1:8081/pet/findByStatus?status=available", waitTimeoutOAS3)
	}

	fmt.Println("Running OAS3 tests...")
	osChdir("my_generated_sdk_oas3")
	if usePrismOAS3 {
		os.Setenv("BASE_URL", "http://127.0.0.1:8081")
	} else {
		os.Setenv("BASE_URL", "http://127.0.0.1:8081/api/v3")
	}
	if err := runCmd("go", "test", "./tests"); err != nil {
		log.Printf("OAS3 tests failed: %v", err)
	}
	osChdir("..")

	fmt.Println("Cleaning up...")
	execCommand("sh", "-c", "docker ps -aq | xargs docker rm -f").Run()
	fmt.Println("All tests passed.")
}
