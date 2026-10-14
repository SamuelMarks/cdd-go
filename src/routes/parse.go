// Package routes provides parsing and emitting of Gin HTTP handler interfaces and structs to and from OpenAPI PathItems.
package routes

import (
	"fmt"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// ParseHandlerInterface parses a dst.InterfaceType into an OpenAPI PathItem.
// Parameter ts is the AST type specification for the handler interface, and it returns the parsed PathItem and error.
func ParseHandlerInterface(ts *dst.TypeSpec) (*openapi.PathItem, error) {
	if ts == nil {
		return nil, fmt.Errorf("TypeSpec is nil")
	}

	iface, ok := ts.Type.(*dst.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("TypeSpec is not an InterfaceType")
	}

	pathItem := &openapi.PathItem{}

	if len(ts.Decs.Start) > 0 {
		desc := ""
		for _, doc := range ts.Decs.Start {
			line := strings.TrimSpace(strings.TrimPrefix(doc, "//"))
			if strings.HasPrefix(line, "Path:") {
				continue
			}
			desc += line + " "
		}
		pathItem.Summary = strings.TrimSpace(desc)
	}

	mcpExt := make(map[string]bool)

	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue
		}
		methodName := method.Names[0].Name
		if methodName == "HandleMcpSse" {
			mcpExt["sse"] = true
			continue
		}
		if methodName == "HandleMcpMessage" {
			mcpExt["message"] = true
			continue
		}
		if methodName == "WithMcpAuth" {
			mcpExt["auth"] = true
			continue
		}
		if methodName == "MapMcpToolToRoute" {
			mcpExt["proxy"] = true
			continue
		}

		op := &openapi.Operation{
			OperationID: strings.ToLower(methodName[:1]) + methodName[1:],
		}

		var httpVerb string
		if len(method.Decs.Start) > 0 {
			desc := ""
			for _, doc := range method.Decs.Start {
				line := strings.TrimSpace(strings.TrimPrefix(doc, "//"))
				if strings.HasPrefix(line, "METHOD:") {
					httpVerb = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "METHOD:")))
					continue
				}
				desc += line + " "
			}
			op.Summary = strings.TrimSpace(desc)
		}

		nameLower := strings.ToLower(methodName)
		if httpVerb == "" {
			for _, verb := range []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"} {
				if strings.HasPrefix(nameLower, verb) {
					httpVerb = verb
					break
				}
			}
		}

		switch httpVerb {
		case "get":
			pathItem.Get = op
		case "post":
			pathItem.Post = op
		case "put":
			pathItem.Put = op
		case "delete":
			pathItem.Delete = op
		case "patch":
			pathItem.Patch = op
		case "options":
			pathItem.Options = op
		case "head":
			pathItem.Head = op
		case "trace":
			pathItem.Trace = op
		}
	}

	if len(mcpExt) > 0 {
		pathItem.Extensions = make(map[string]map[string]bool)
		pathItem.Extensions["x-mcp-server"] = mcpExt
	}

	return pathItem, nil
}
