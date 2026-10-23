// Package clients provides parsing and emitting of Go client SDK interfaces to and from OpenAPI PathItems.
package clients

import (
	"fmt"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// ParseClientInterface parses a dst.InterfaceType into an OpenAPI PathItem.
// Parameter ts is the AST type specification for the client interface, and it returns the parsed PathItem and error.
func ParseClientInterface(ts *dst.TypeSpec) (*openapi.PathItem, error) {
	if ts == nil {
		return nil, fmt.Errorf("TypeSpec is nil")
	}

	iface, ok := ts.Type.(*dst.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("TypeSpec is not an InterfaceType")
	}

	pathItem := &openapi.PathItem{}
	pathItem.Summary, _ = docstrings.ParseSummaryAndTag(ts.Decs.Start, "Path:")

	mcpExt := make(map[string]bool)

	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue
		}
		methodName := method.Names[0].Name
		if methodName == "GetTools" {
			mcpExt["tools"] = true
			continue
		}
		if methodName == "GetResources" {
			mcpExt["resources"] = true
			continue
		}
		if methodName == "ExecuteTool" {
			mcpExt["execute"] = true
			continue
		}
		if methodName == "AuthLogin" {
			continue
		}

		summary, rawVerb := docstrings.ParseSummaryAndTag(method.Decs.Start, "METHOD:")
		op := &openapi.Operation{
			OperationID: strings.ToLower(methodName[:1]) + methodName[1:],
			Summary:     summary,
		}
		httpVerb := strings.ToLower(rawVerb)

		nameLower := strings.ToLower(methodName)
		if httpVerb == "" {
			for _, verb := range []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"} {
				if strings.HasPrefix(nameLower, verb) {
					httpVerb = verb
					break
				}
			}
		}
		if httpVerb == "" {
			httpVerb = "post"
		}

		pathItem.SetOperation(httpVerb, op)
	}

	if len(mcpExt) > 0 {
		pathItem.Extensions = make(map[string]map[string]bool)
		pathItem.Extensions["x-mcp"] = mcpExt
	}

	return pathItem, nil
}
