// Package main provides the CLI entrypoint and subcommand dispatcher for cdd-go.
package main

import (
	"flag"
	"io"
	"os"

	"github.com/SamuelMarks/cdd-go/cdd"
)

// runToDocsJSON parses CLI flags for the to_docs_json subcommand and writes the JSON output to stdout or a file.
// Parameter args is the slice of CLI arguments, and it returns any execution error.
func runToDocsJSON(args []string) error {
	return runToDocsJSONWriter(args, os.Stdout)
}

// runToDocsJSONWriter parses CLI flags for the to_docs_json subcommand and writes the JSON output to defaultOut when -o is not specified.
// Parameter args is the slice of CLI arguments, defaultOut is the fallback writer, and it returns any execution error.
func runToDocsJSONWriter(args []string, defaultOut io.Writer) error {
	fs := flag.NewFlagSet("to_docs_json", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var in, out string
	var noImports, noWrapping bool

	fs.StringVar(&in, "i", envOrDefault("CDD_INPUT", ""), "Path or URL to the OpenAPI specification.")
	fs.StringVar(&in, "input", envOrDefault("CDD_INPUT", ""), "Path or URL to the OpenAPI specification.")
	fs.StringVar(&out, "o", envOrDefault("CDD_OUTPUT", ""), "Output file or directory path.")
	fs.StringVar(&out, "output", envOrDefault("CDD_OUTPUT", ""), "Output file or directory path.")
	fs.BoolVar(&noImports, "no-imports", envOrDefaultBool("CDD_NO_IMPORTS", false), "Omit the imports field.")
	fs.BoolVar(&noWrapping, "no-wrapping", envOrDefaultBool("CDD_NO_WRAPPING", false), "Omit the wrapper fields.")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if out != "" {
		return cdd.ToDocsJSON(in, out, noImports, noWrapping)
	}
	return cdd.ToDocsJSONWriter(in, defaultOut, noImports, noWrapping)
}
