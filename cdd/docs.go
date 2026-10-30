// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
package cdd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/SamuelMarks/cdd-go/src/naming"
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

// docsDefaultStdout is an injectable writer for ToDocsJSON when out is empty.
var docsDefaultStdout io.Writer = os.Stdout

// ToDocsJSON generates JSON documentation with code snippets for an OpenAPI specification and writes it to out or docsDefaultStdout.
// Parameter in is the input OpenAPI file path, out is the optional output file or directory path, noImports omits imports, and noWrapping omits main() wrapping.
func ToDocsJSON(in string, out string, noImports bool, noWrapping bool) error {
	var outTarget io.Writer = docsDefaultStdout
	if out != "" {
		if stat, err := os.Stat(out); err == nil && stat.IsDir() {
			out = filepath.Join(out, "docs.json")
		}
		fileTarget, err := os.Create(out)
		if err != nil {
			return err
		}
		defer fileTarget.Close()
		outTarget = fileTarget
	}
	return ToDocsJSONWriter(in, outTarget, noImports, noWrapping)
}

// buildDocsOperationParamLines builds sample Go code lines for operation parameters and request bodies.
// Parameter itemParams is the slice of PathItem-level parameters, and op is the OpenAPI Operation.
func buildDocsOperationParamLines(itemParams []openapi.Parameter, op *openapi.Operation) string {
	var lines []string
	allParams := make([]openapi.Parameter, 0, len(itemParams)+len(op.Parameters))
	allParams = append(allParams, itemParams...)
	allParams = append(allParams, op.Parameters...)
	for _, p := range allParams {
		paramName := p.Name
		if paramName == "" && p.Ref != "" {
			parts := strings.Split(p.Ref, "/")
			paramName = parts[len(parts)-1]
		}
		if paramName == "" {
			continue
		}
		fieldName := naming.ToPascalCase(paramName)
		sampleVal := `"example"`
		if p.Schema != nil {
			switch p.Schema.Type {
			case "integer":
				sampleVal = "1"
			case "number":
				sampleVal = "1.0"
			case "boolean":
				sampleVal = "true"
			}
		}
		lines = append(lines, fmt.Sprintf("// Param %s (%s): %s = %s", fieldName, p.In, paramName, sampleVal))
	}
	if op.RequestBody != nil {
		bodyType := "map[string]any{}"
		if mt, ok := op.RequestBody.Content["application/json"]; ok && mt.Schema != nil && mt.Schema.Ref != "" {
			parts := strings.Split(mt.Schema.Ref, "/")
			bodyType = "models." + naming.ToPascalCase(parts[len(parts)-1]) + "{}"
		}
		lines = append(lines, fmt.Sprintf("body := %s\n_ = body", bodyType))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// buildDocsOperationsForPathItem builds DocsOperation entries for all HTTP operations in a PathItem.
// Parameter path is the OpenAPI path or webhook name, item is the PathItem, noImports omits imports, and noWrapping omits main() wrapping.
func buildDocsOperationsForPathItem(path string, item *openapi.PathItem, noImports bool, noWrapping bool) []DocsOperation {
	var ops []DocsOperation
	for _, mo := range openapi.PathItemOperations(item) {
		method := mo.Method
		op := mo.Operation

		opID := op.OperationID
		if opID == "" {
			opID = "request"
		}
		clientIface := "Client" + naming.ToCamelCasePath(path)
		methodName := naming.ToPascalCase(opID)

		var imports *string
		if !noImports {
			i := "import (\n\t\"context\"\n\t\"fmt\"\n\t\"net/http\"\n\t\"os\"\n\t\"generated_sdk/client\"\n)"
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

		paramLines := buildDocsOperationParamLines(item.Parameters, op)
		snippet := fmt.Sprintf("var sdk client.%s\n%sreq, _ := http.NewRequestWithContext(context.Background(), %q, %q, nil)\n// OperationID: %s\nresp, err := sdk.%s(req)\nif err != nil {\n\tfmt.Fprintf(os.Stderr, \"Error: %%v\\n\", err)\n\tos.Exit(1)\n}\nfmt.Fprintf(os.Stdout, \"Response: %%v\\n\", resp)", clientIface, paramLines, strings.ToUpper(method), path, opID, methodName)

		ops = append(ops, DocsOperation{
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
	return ops
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

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	var operations []DocsOperation

	if azDoc, errAz := arazzo.ParseBytes(data); errAz == nil && len(azDoc.Workflows) > 0 {
		for _, wf := range azDoc.Workflows {
			wfPascal := naming.ToPascalCase(wf.WorkflowID)
			var imports *string
			if !noImports {
				i := "import (\n\t\"context\"\n\t\"fmt\"\n\t\"os\"\n\t\"generated_sdk/workflows\"\n)"
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
			snippet := fmt.Sprintf("client := workflows.NewWorkflowClient(os.Getenv(\"BASE_URL\"))\nout, err := client.Execute%s(context.Background(), workflows.%sInput{})\nif err != nil {\n\tfmt.Fprintf(os.Stderr, \"Error: %%v\\n\", err)\n\tos.Exit(1)\n}\nfmt.Fprintf(os.Stdout, \"Workflow Output: %%+v\\n\", out)", wfPascal, wfPascal)
			operations = append(operations, DocsOperation{
				Method:      "WORKFLOW",
				Path:        "/workflows/" + wf.WorkflowID,
				OperationId: wf.WorkflowID,
				Code: DocsCode{
					Imports:      imports,
					WrapperStart: wrapperStart,
					WrapperEnd:   wrapperEnd,
					Snippet:      snippet,
				},
			})
		}
		result := []DocsJSONOutput{{Language: "go", Operations: operations}}
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	oa, err := openapi.Parse(bytes.NewReader(data))
	if err != nil {
		return err
	}

	paths := make([]string, 0, len(oa.Paths))
	for path := range oa.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		item := oa.Paths[path]
		operations = append(operations, buildDocsOperationsForPathItem(path, &item, noImports, noWrapping)...)
	}

	webhooks := make([]string, 0, len(oa.Webhooks))
	for wh := range oa.Webhooks {
		webhooks = append(webhooks, wh)
	}
	sort.Strings(webhooks)

	for _, wh := range webhooks {
		item := oa.Webhooks[wh]
		operations = append(operations, buildDocsOperationsForPathItem(wh, &item, noImports, noWrapping)...)
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
