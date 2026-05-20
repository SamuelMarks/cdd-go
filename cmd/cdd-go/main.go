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
		fmt.Fprintf(stderr, "error: %v\n", err)
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

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("expected 'from_openapi', 'to_openapi', 'to_docs_json', or 'server_json_rpc' subcommands")
	}

	subcommand := args[0]
	var in, out string

	switch subcommand {
	case "-h", "--help", "help":
		fmt.Println("cdd-go is a Code-Driven Development tool for Go.")
		fmt.Println("\nUsage:")
		fmt.Println("  cdd-go [subcommand] [flags]")
		fmt.Println("\nSubcommands:")
		fmt.Println("  from_openapi     Generate code from OpenAPI spec")
		fmt.Println("  to_openapi       Generate OpenAPI spec from code")
		fmt.Println("  to_docs_json     Generate documentation JSON from OpenAPI spec")
		fmt.Println("  server_json_rpc  Run a JSON-RPC server exposing the CLI")
		fmt.Println("\nFlags:")
		fmt.Println("  -h, --help       Show this help message")
		fmt.Println("  -v, --version    Show version information")
		return nil
	case "-v", "--version", "version":
		fmt.Println("cdd-go version 0.0.1")
		return nil
	case "server_json_rpc":
		return runServerJSONRPC(args[1:])
	case "from_openapi":
		if len(args) < 2 {
			return fmt.Errorf("expected 'to_sdk', 'to_sdk_cli', or 'to_server' subcommands for from_openapi")
		}
		subsubcommand := args[1]
		fs := flag.NewFlagSet("from_openapi "+subsubcommand, flag.ContinueOnError)
		fs.SetOutput(stderr)

		fs.StringVar(&in, "i", envOrDefault("CDD_GO_INPUT", ""), "Input file path")
		var inputDir string
		fs.StringVar(&inputDir, "input-dir", envOrDefault("CDD_GO_INPUT_DIR", ""), "Input directory path")
		fs.StringVar(&out, "o", envOrDefault("CDD_GO_OUTPUT", ""), "Output directory path")

		var noGithubActions, noInstallablePackage, tests bool
		fs.BoolVar(&noGithubActions, "no-github-actions", envOrDefaultBool("CDD_GO_NO_GITHUB_ACTIONS", false), "Do not generate GitHub Actions")
		fs.BoolVar(&noInstallablePackage, "no-installable-package", envOrDefaultBool("CDD_GO_NO_INSTALLABLE_PACKAGE", false), "Do not generate installable package scaffolding")
		fs.BoolVar(&tests, "create-composable-tests", envOrDefaultBool("CDD_GO_CREATE_COMPOSABLE_TESTS", false), "Create composable tests & mocks")
		fs.BoolVar(&tests, "tests", envOrDefaultBool("CDD_GO_CREATE_COMPOSABLE_TESTS", false), "Alias for create-composable-tests")

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
		return cdd.RunFromOpenAPI(subsubcommand, inputTarget, out, noGithubActions, noInstallablePackage, tests)
	case "to_openapi":
		fs := flag.NewFlagSet("to_openapi", flag.ContinueOnError)
		fs.SetOutput(stderr)
		fs.StringVar(&in, "i", envOrDefault("CDD_GO_INPUT", ""), "Input file or directory path")
		fs.StringVar(&out, "o", envOrDefault("CDD_GO_OUTPUT", "openapi.json"), "Output file path")

		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		return cdd.RunToOpenAPI(in, out)
	case "to_docs_json":
		return runToDocsJSON(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}
