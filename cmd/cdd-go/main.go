// Package main provides the CLI entrypoint and subcommand dispatcher for cdd-go.
package main

import (
	"github.com/SamuelMarks/cdd-go/cdd"

	"flag"
	"fmt"
	"os"
)

// osExit is an injectable function variable wrapping os.Exit.
var osExit = os.Exit

// osGetwd is an injectable function variable wrapping os.Getwd.
var osGetwd = os.Getwd

// stderr is an injectable writer wrapping os.Stderr.
var stderr = os.Stderr

// main is the CLI entrypoint for cdd-go.
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(stderr, err.Error())
		osExit(1)
	}
}

// envOrDefault returns the value of environment variable key or def if unset.
// Parameter key is the environment variable name, and def is the fallback string.
func envOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

// envOrDefaultBool returns the boolean value of environment variable key or def if unset.
// Parameter key is the environment variable name, and def is the fallback boolean.
func envOrDefaultBool(key string, def bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1"
	}
	return def
}

// envOrDefaultInt returns the integer value of environment variable key or def if unset or invalid.
// Parameter key is the environment variable name, and def is the fallback integer.
func envOrDefaultInt(key string, def int) int {
	if val := os.Getenv(key); val != "" {
		var i int
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i
		}
	}
	return def
}

// run parses CLI arguments and executes the requested subcommand.
// Parameter args is the slice of CLI arguments excluding the binary name, and it returns any execution error.
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
		fmt.Println("  from_arazzo     Generate code from an Arazzo 1.1.0 specification.")
		fmt.Println("  to_arazzo       Generate an Arazzo 1.1.0 specification from source code.")
		fmt.Println("  sync            Synchronize an OpenAPI or Arazzo specification with source code.")
		fmt.Println("  to_docs_json    Generate JSON documentation with code snippets for an OpenAPI specification.")
		fmt.Println("  serve_json_rpc  Expose CLI interface as a JSON-RPC server.")
		fmt.Println("  mcp             Run the generator as an MCP server over stdio.")
		fmt.Println("\nFlags:")
		fmt.Println("  -h, --help       Show this help message")
		fmt.Println("  -v, --version    Show version information")
		return nil
	case "-v", "--version", "version":
		fmt.Println("0.0.4")
		return nil
	case "mcp":
		if len(args) > 1 && (args[1] == "-h" || args[1] == "--help" || args[1] == "help") {
			fmt.Println("Usage of mcp:")
			fmt.Println("  Run the Model Context Protocol (MCP) server over standard I/O (stdio).")
			return nil
		}
		return cdd.ServeMCPStdio(args[1:])
	case "serve_json_rpc":
		return runServeJSONRPC(args[1:])
	case "from_openapi":
		if len(args) < 2 {
			return fmt.Errorf("expected 'to_sdk', 'to_sdk_cli', or 'to_server' subcommands for from_openapi")
		}
		subsubcommand := args[1]
		if subsubcommand == "-h" || subsubcommand == "--help" || subsubcommand == "help" {
			fmt.Println("Usage of from_openapi:")
			fmt.Println("  cdd-go from_openapi [to_sdk | to_sdk_cli | to_server] [flags]")
			return nil
		}
		flagArgs := args[2:]
		if len(subsubcommand) > 0 && subsubcommand[0] == '-' {
			subsubcommand = "to_sdk"
			flagArgs = args[1:]
		}
		fs := flag.NewFlagSet("from_openapi "+subsubcommand, flag.ContinueOnError)
		fs.SetOutput(stderr)

		fs.StringVar(&in, "i", envOrDefault("CDD_INPUT", ""), "Path or URL to the OpenAPI specification.")
		fs.StringVar(&in, "input", envOrDefault("CDD_INPUT", ""), "Path or URL to the OpenAPI specification.")
		var inputDir string
		fs.StringVar(&inputDir, "d", envOrDefault("CDD_INPUT_DIR", ""), "Directory containing OpenAPI specifications.")
		fs.StringVar(&inputDir, "input-dir", envOrDefault("CDD_INPUT_DIR", ""), "Directory containing OpenAPI specifications.")
		fs.StringVar(&out, "o", envOrDefault("CDD_OUTPUT", ""), "Output file or directory path.")
		fs.StringVar(&out, "output", envOrDefault("CDD_OUTPUT", ""), "Output file or directory path.")

		var noGithubActions, noInstallablePackage, tests, mcp bool
		fs.BoolVar(&noGithubActions, "no-github-actions", envOrDefaultBool("CDD_NO_GITHUB_ACTIONS", false), "Do not generate GitHub Actions scaffolding.")
		fs.BoolVar(&noInstallablePackage, "no-installable-package", envOrDefaultBool("CDD_NO_INSTALLABLE_PACKAGE", false), "Do not generate installable package scaffolding.")
		fs.BoolVar(&tests, "tests", envOrDefaultBool("CDD_TESTS", false), "Generate integration tests and mocks.")
		fs.BoolVar(&mcp, "m", envOrDefaultBool("CDD_MCP", false), "Generate Model Context Protocol (MCP) server and adapter.")
		fs.BoolVar(&mcp, "mcp", envOrDefaultBool("CDD_MCP", false), "Generate Model Context Protocol (MCP) server and adapter.")

		if err := fs.Parse(flagArgs); err != nil {
			if err == flag.ErrHelp {
				return nil
			}
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
		return cdd.FromOpenAPI(subsubcommand, inputTarget, out, noGithubActions, noInstallablePackage, tests, mcp)
	case "to_openapi":
		fs := flag.NewFlagSet("to_openapi", flag.ContinueOnError)
		fs.SetOutput(stderr)
		fs.StringVar(&in, "i", envOrDefault("CDD_INPUT", ""), "Path to source code directory or file.")
		fs.StringVar(&in, "input", envOrDefault("CDD_INPUT", ""), "Path to source code directory or file.")
		fs.StringVar(&out, "o", envOrDefault("CDD_OUTPUT", "openapi.json"), "Output file or directory path.")
		fs.StringVar(&out, "output", envOrDefault("CDD_OUTPUT", "openapi.json"), "Output file or directory path.")
		if err := fs.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return nil
			}
			return err
		}

		return cdd.ToOpenAPI(in, out)
	case "from_arazzo":
		if len(args) < 2 {
			return fmt.Errorf("expected 'to_sdk', 'to_sdk_cli', or 'to_server' subcommands for from_arazzo")
		}
		subsubcommand := args[1]
		if subsubcommand == "-h" || subsubcommand == "--help" || subsubcommand == "help" {
			fmt.Println("Usage of from_arazzo:")
			fmt.Println("  cdd-go from_arazzo [to_sdk | to_sdk_cli | to_server] [flags]")
			return nil
		}
		fs := flag.NewFlagSet("from_arazzo "+subsubcommand, flag.ContinueOnError)
		fs.SetOutput(stderr)

		fs.StringVar(&in, "i", envOrDefault("CDD_INPUT", ""), "Path to the Arazzo 1.1.0 specification.")
		fs.StringVar(&in, "input", envOrDefault("CDD_INPUT", ""), "Path to the Arazzo 1.1.0 specification.")
		var inputDir string
		fs.StringVar(&inputDir, "d", envOrDefault("CDD_INPUT_DIR", ""), "Directory containing Arazzo specifications.")
		fs.StringVar(&inputDir, "input-dir", envOrDefault("CDD_INPUT_DIR", ""), "Directory containing Arazzo specifications.")
		fs.StringVar(&out, "o", envOrDefault("CDD_OUTPUT", ""), "Output directory path.")
		fs.StringVar(&out, "output", envOrDefault("CDD_OUTPUT", ""), "Output directory path.")

		var noGithubActions, noInstallablePackage, tests bool
		fs.BoolVar(&noGithubActions, "no-github-actions", envOrDefaultBool("CDD_NO_GITHUB_ACTIONS", false), "Do not generate GitHub Actions scaffolding.")
		fs.BoolVar(&noInstallablePackage, "no-installable-package", envOrDefaultBool("CDD_NO_INSTALLABLE_PACKAGE", false), "Do not generate installable package scaffolding.")
		fs.BoolVar(&tests, "tests", envOrDefaultBool("CDD_TESTS", false), "Generate workflow E2E tests.")

		if err := fs.Parse(args[2:]); err != nil {
			if err == flag.ErrHelp {
				return nil
			}
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
		return cdd.FromArazzo(subsubcommand, inputTarget, out, noGithubActions, noInstallablePackage, tests)
	case "to_arazzo":
		fs := flag.NewFlagSet("to_arazzo", flag.ContinueOnError)
		fs.SetOutput(stderr)
		fs.StringVar(&in, "i", envOrDefault("CDD_INPUT", ""), "Path to source code directory or file.")
		fs.StringVar(&in, "input", envOrDefault("CDD_INPUT", ""), "Path to source code directory or file.")
		fs.StringVar(&out, "o", envOrDefault("CDD_OUTPUT", "arazzo.json"), "Output file or directory path.")
		fs.StringVar(&out, "output", envOrDefault("CDD_OUTPUT", "arazzo.json"), "Output file or directory path.")
		if err := fs.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return nil
			}
			return err
		}

		return cdd.ToArazzo(in, out)
	case "sync":
		fs := flag.NewFlagSet("sync", flag.ContinueOnError)
		fs.SetOutput(stderr)
		fs.StringVar(&in, "i", envOrDefault("CDD_INPUT", ""), "Input path (source dir or openapi spec).")
		fs.StringVar(&in, "input", envOrDefault("CDD_INPUT", ""), "Input path (source dir or openapi spec).")
		fs.StringVar(&out, "o", envOrDefault("CDD_OUTPUT", "openapi.json"), "Output file or directory path.")
		fs.StringVar(&out, "output", envOrDefault("CDD_OUTPUT", "openapi.json"), "Output file or directory path.")
		var truth string
		fs.StringVar(&truth, "t", envOrDefault("CDD_TRUTH", "code"), "Source of truth: 'code' (or 'go'), 'openapi' (or 'spec', 'swagger'), 'arazzo'.")
		fs.StringVar(&truth, "truth", envOrDefault("CDD_TRUTH", "code"), "Source of truth: 'code' (or 'go'), 'openapi' (or 'spec', 'swagger'), 'arazzo'.")
		if err := fs.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return nil
			}
			return err
		}

		return cdd.Sync(in, out, truth)
	case "to_docs_json":
		return runToDocsJSON(args[1:])
	default:
		return fmt.Errorf("Error: Unknown or incomplete command: %s", subcommand)
	}
}
