package main

import (
	"flag"
	"os"

	"github.com/SamuelMarks/cdd-go/cdd"
)

func runToDocsJSON(args []string) error {
	fs := flag.NewFlagSet("to_docs_json", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var in string
	var out string
	var noImports bool
	var noWrapping bool

	fs.StringVar(&in, "i", envOrDefault("CDD_INPUT", ""), "Path or URL to the OpenAPI specification.")
	fs.StringVar(&in, "input", envOrDefault("CDD_INPUT", ""), "Path or URL to the OpenAPI specification.")
	fs.StringVar(&out, "o", envOrDefault("CDD_OUTPUT", ""), "Output file or directory path.")
	fs.StringVar(&out, "output", envOrDefault("CDD_OUTPUT", ""), "Output file or directory path.")
	fs.BoolVar(&noImports, "no-imports", envOrDefaultBool("CDD_NO_IMPORTS", false), "Omit the imports field.")
	fs.BoolVar(&noWrapping, "no-wrapping", envOrDefaultBool("CDD_NO_WRAPPING", false), "Omit the wrapper fields.")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return cdd.ToDocsJSON(in, out, noImports, noWrapping)
}
