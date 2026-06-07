package main

import (
	"github.com/SamuelMarks/cdd-go/cdd"

	"flag"
	"fmt"
	"os"
)

var osExit = os.Exit
var osGetwd = os.Getwd
var stderr = os.Stderr

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(stderr, err.Error())
		osExit(1)
	}
}

func envOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func envOrDefaultBool(key string, def bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1"
	}
	return def
}

func envOrDefaultInt(key string, def int) int {
	if val := os.Getenv(key); val != "" {
		var i int
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i
		}
	}
	return def
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Error: Unknown or incomplete command: ")
	}

	subcommand := args[0]
	var in, out string

	switch subcommand {
	case "-h", "--help", "help":
		fmt.Println("cdd-go is a Code-Driven Development tool for Go.")
		fmt.Println("\nUsage:")
		fmt.Println("  cdd-go [subcommand] [flags]")
		fmt.Println("\nSubcommands:")
		fmt.Println("  from_openapi    Generate code from an OpenAPI specification.")
		fmt.Println("  to_openapi      Generate an OpenAPI specification from source code.")
		fmt.Println("  to_docs_json    Generate JSON documentation with code snippets for an OpenAPI specification.")
		fmt.Println("  serve_json_rpc  Expose CLI interface as a JSON-RPC server.")
		fmt.Println("\nFlags:")
		fmt.Println("  -h, --help       Show this help message")
		fmt.Println("  -v, --version    Show version information")
		return nil
	case "-v", "--version", "version":
		fmt.Println("0.0.2")
		return nil
	case "mcp":
		return runServeMCPStdio(args[1:])
	case "serve_json_rpc":
		return runServeJSONRPC(args[1:])
	case "from_openapi":
		if len(args) < 2 {
			return fmt.Errorf("expected 'to_sdk', 'to_sdk_cli', or 'to_server' subcommands for from_openapi")
		}
		subsubcommand := args[1]
		fs := flag.NewFlagSet("from_openapi "+subsubcommand, flag.ContinueOnError)
		fs.SetOutput(stderr)

		fs.StringVar(&in, "i", envOrDefault("CDD_INPUT", ""), "Path or URL to the OpenAPI specification.")
		fs.StringVar(&in, "input", envOrDefault("CDD_INPUT", ""), "Path or URL to the OpenAPI specification.")
		var inputDir string
		fs.StringVar(&inputDir, "input-dir", envOrDefault("CDD_INPUT_DIR", ""), "Directory containing OpenAPI specifications.")
		fs.StringVar(&out, "o", envOrDefault("CDD_OUTPUT", ""), "Output file or directory path.")
		fs.StringVar(&out, "output", envOrDefault("CDD_OUTPUT", ""), "Output file or directory path.")

		var noGithubActions, noInstallablePackage, tests bool
		fs.BoolVar(&noGithubActions, "no-github-actions", envOrDefaultBool("CDD_NO_GITHUB_ACTIONS", false), "Do not generate GitHub Actions scaffolding.")
		fs.BoolVar(&noInstallablePackage, "no-installable-package", envOrDefaultBool("CDD_NO_INSTALLABLE_PACKAGE", false), "Do not generate installable package scaffolding.")
		fs.BoolVar(&tests, "tests", envOrDefaultBool("CDD_TESTS", false), "Generate integration tests and mocks.")

		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if out == "" {
			pwd, err := osGetwd()
			if err != nil {
				pwd = "."
			}
			out = pwd
		}
		inputTarget := in
		if inputTarget == "" {
			inputTarget = inputDir
		}
		return cdd.GenerateFromOpenApi(subsubcommand, inputTarget, out, noGithubActions, noInstallablePackage, tests)
	case "to_openapi":
		fs := flag.NewFlagSet("to_openapi", flag.ContinueOnError)
		fs.SetOutput(stderr)
		fs.StringVar(&in, "i", envOrDefault("CDD_INPUT", ""), "Path to source code directory or file.")
		fs.StringVar(&in, "input", envOrDefault("CDD_INPUT", ""), "Path to source code directory or file.")
		fs.StringVar(&out, "o", envOrDefault("CDD_OUTPUT", "openapi.json"), "Output file or directory path.")
		fs.StringVar(&out, "output", envOrDefault("CDD_OUTPUT", "openapi.json"), "Output file or directory path.")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		return cdd.GenerateToOpenApi(in, out)
	case "to_docs_json":
		return runToDocsJSON(args[1:])
	default:
		return fmt.Errorf("Error: Unknown or incomplete command: %s", subcommand)
	}
}
