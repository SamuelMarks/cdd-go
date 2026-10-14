// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
package cdd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

// DocsJSONOutput is the root structure for the documentation JSON output.
type DocsJSONOutput struct {
	// Language specifies the target programming language ("go").
	Language string `json:"language"`
	// Operations lists the documented API operations and their snippets.
	Operations []DocsOperation `json:"operations"`
}

// DocsOperation represents a single API operation documented in the JSON output.
type DocsOperation struct {
	// Method is the uppercase HTTP verb (e.g., "GET", "POST").
	Method string `json:"method"`
	// Path is the OpenAPI URL path template.
	Path string `json:"path"`
	// OperationId is the unique identifier for the operation.
	OperationId string `json:"operationId"`
	// Code contains the generated code snippet and optional wrapper/imports.
	Code DocsCode `json:"code"`
}

// DocsCode contains the code snippet and metadata for a specific operation.
type DocsCode struct {
	// Imports holds the Go import block when imports are enabled.
	Imports *string `json:"imports,omitempty"`
	// WrapperStart holds the opening function wrapper when wrapping is enabled.
	WrapperStart *string `json:"wrapperStart,omitempty"`
	// WrapperEnd holds the closing function wrapper when wrapping is enabled.
	WrapperEnd *string `json:"wrapperEnd,omitempty"`
	// Snippet holds the main body of the generated example call.
	Snippet string `json:"snippet"`
}

// ToDocsJSON generates JSON documentation with code snippets for an OpenAPI specification and writes it to out or os.Stdout.
// Parameter in is the input OpenAPI file path, out is the optional output file path, noImports omits imports, and noWrapping omits main() wrapping.
func ToDocsJSON(in string, out string, noImports bool, noWrapping bool) error {
	var outTarget io.Writer = os.Stdout
	if out != "" {
		fileTarget, err := os.Create(out)
		if err != nil {
			return err
		}
		defer fileTarget.Close()
		outTarget = fileTarget
	}
	return ToDocsJSONWriter(in, outTarget, noImports, noWrapping)
}

// ToDocsJSONWriter generates JSON documentation with code snippets for an OpenAPI specification and writes it to w.
// Parameter in is the input OpenAPI file path, w is the destination io.Writer, noImports omits imports, and noWrapping omits main() wrapping.
func ToDocsJSONWriter(in string, w io.Writer, noImports bool, noWrapping bool) error {
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

	paths := make([]string, 0, len(oa.Paths))
	for path := range oa.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	methodOrder := []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}

	for _, path := range paths {
		item := oa.Paths[path]
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

		for _, method := range methodOrder {
			op := methods[method]
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
	fmt.Fprintf(os.Stderr, "Error: %%v
", err)
	os.Exit(1)
}
fmt.Fprintf(os.Stdout, "Response: %%v
", resp)`, opID)

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

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
