package cdd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

// DocsJSONOutput is the root structure for the documentation JSON output.
type DocsJSONOutput struct {
	Language   string          `json:"language"`
	Operations []DocsOperation `json:"operations"`
}

// DocsOperation represents a single API operation documented in the JSON output.
type DocsOperation struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	OperationId string   `json:"operationId"`
	Code        DocsCode `json:"code"`
}

// DocsCode contains the code snippet and metadata for a specific operation.
type DocsCode struct {
	Imports      *string `json:"imports,omitempty"`
	WrapperStart *string `json:"wrapperStart,omitempty"`
	WrapperEnd   *string `json:"wrapperEnd,omitempty"`
	Snippet      string  `json:"snippet"`
}

// GenerateDocsJson generates JSON documentation with code snippets for an OpenAPI specification.
func GenerateDocsJson(in string, out string, noImports bool, noWrapping bool) error {
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
