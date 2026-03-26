package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/samuel/cdd-go/src/openapi"
)

type DocsJSONOutput struct {
	Language   string          `json:"language"`
	Operations []DocsOperation `json:"operations"`
}

type DocsOperation struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	OperationId string   `json:"operationId"`
	Code        DocsCode `json:"code"`
}

type DocsCode struct {
	Imports      *string `json:"imports,omitempty"`
	WrapperStart *string `json:"wrapperStart,omitempty"`
	WrapperEnd   *string `json:"wrapperEnd,omitempty"`
	Snippet      string  `json:"snippet"`
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

	var operations []DocsOperation

	for path, item := range oa.Paths {
		methods := map[string]*openapi.Operation{
			"get":     item.Get,
			"put":     item.Put,
			"post":    item.Post,
			"delete":  item.Delete,
			"options": item.Options,
			"head":    item.Head,
			"patch":   item.Patch,
			"trace":   item.Trace,
		}

		for method, op := range methods {
			if op == nil {
				continue
			}

			opID := op.OperationID
			if opID == "" {
				opID = "request"
			}

			var imports *string
			if !noImports {
				i := "import (\n\t\"context\"\n\t\"fmt\"\n\t\"os\"\n\t\"github.com/your/client\"\n)"
				imports = &i
			}

			var wrapperStart *string
			var wrapperEnd *string
			if !noWrapping {
				ws := "func main() {"
				wrapperStart = &ws
				we := "}"
				wrapperEnd = &we
			}

			snippet := fmt.Sprintf(`client := client.NewAPIClient(client.NewConfiguration())
resp, r, err := client.DefaultApi.%s(context.Background()).Execute()
if err != nil {
	fmt.Fprintf(os.Stderr, "Error: %%v\n", err)
	os.Exit(1)
}
fmt.Fprintf(os.Stdout, "Response: %%v\n", resp)`, opID)

			operations = append(operations, DocsOperation{
				Method:      strings.ToUpper(method),
				Path:        path,
				OperationId: opID,
				Code: DocsCode{
					Imports:      imports,
					WrapperStart: wrapperStart,
					WrapperEnd:   wrapperEnd,
					Snippet:      snippet,
				},
			})
		}
	}

	if operations == nil {
		operations = []DocsOperation{}
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
