package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/samuel/cdd-go/src/openapi"
)

// Code represents the structured code example
type Code struct {
	Imports      *string `json:"imports,omitempty"`
	WrapperStart *string `json:"wrapper_start,omitempty"`
	Snippet      string  `json:"snippet"`
	WrapperEnd   *string `json:"wrapper_end,omitempty"`
}

// Operation struct for the JSON output
type Operation struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	OperationId string `json:"operationId"`
	Code        Code   `json:"code"`
}

// DocsJSONOutput is the root of the JSON array elements
type DocsJSONOutput struct {
	Language   string      `json:"language"`
	Operations []Operation `json:"operations"`
}

func runToDocsJSON(args []string) error {
	fs := flag.NewFlagSet("to_docs_json", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var in string
	var out string
	var noImports bool
	var noWrapping bool

	fs.StringVar(&in, "i", envOrDefault("CDD_GO_INPUT", ""), "Input file path")
	fs.StringVar(&in, "input", envOrDefault("CDD_GO_INPUT", ""), "Input file path")
	fs.StringVar(&out, "o", envOrDefault("CDD_GO_OUTPUT", ""), "Output file path")
	fs.BoolVar(&noImports, "no-imports", envOrDefaultBool("CDD_GO_NO_IMPORTS", false), "Omit imports")
	fs.BoolVar(&noWrapping, "no-wrapping", envOrDefaultBool("CDD_GO_NO_WRAPPING", false), "Omit wrapping")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if in == "" {
		return fmt.Errorf("input file is required")
	}

	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()

	oa, err := openapi.Parse(f)
	if err != nil {
		return err
	}

	var operations []Operation
	for path, item := range oa.Paths {
		methods := map[string]*openapi.Operation{
			"GET":     item.Get,
			"PUT":     item.Put,
			"POST":    item.Post,
			"DELETE":  item.Delete,
			"OPTIONS": item.Options,
			"HEAD":    item.Head,
			"PATCH":   item.Patch,
			"TRACE":   item.Trace,
		}

		for method, op := range methods {
			if op == nil {
				continue
			}

			opID := op.OperationID

			snippet := fmt.Sprintf(`        client := client.NewAPIClient(client.NewConfiguration())
        resp, r, err := client.DefaultApi.%s(context.Background()).Execute()
        if err != nil {
                fmt.Fprintf(os.Stderr, "Error: %%v\n", err)
                os.Exit(1)
        }
        fmt.Fprintf(os.Stdout, "Response: %%v\n", resp)`, opID)

			code := Code{
				Snippet: snippet,
			}

			if !noImports {
				imports := `import (
        "context"
        "fmt"
        "os"

        "github.com/your/client"
)`
				code.Imports = &imports
			}
			if !noWrapping {
				wrapperStart := "func main() {"
				wrapperEnd := "}"
				code.WrapperStart = &wrapperStart
				code.WrapperEnd = &wrapperEnd
			}

			operations = append(operations, Operation{
				Method:      method,
				Path:        path,
				OperationId: opID,
				Code:        code,
			})
		}
	}

	if operations == nil {
		operations = []Operation{}
	}

	result := []DocsJSONOutput{
		{
			Language:   "go",
			Operations: operations,
		},
	}

	outTarget := os.Stdout
	if out != "" {
		fileTarget, err := os.Create(out)
		if err != nil {
			return err
		}
		defer fileTarget.Close()
		outTarget = fileTarget
	}

	encoder := json.NewEncoder(outTarget)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return err
	}

	return nil
}
