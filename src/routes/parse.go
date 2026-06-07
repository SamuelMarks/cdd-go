package routes

import (
	"fmt"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// ParseHandlerInterface parses a dst.InterfaceType into an OpenAPI PathItem.
func ParseHandlerInterface(ts *dst.TypeSpec) (*openapi.PathItem, error) {
	if ts == nil {
		return nil, fmt.Errorf("TypeSpec is nil")
	}

	iface, ok := ts.Type.(*dst.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("TypeSpec is not an interface")
	}

	pathItem := &openapi.PathItem{}
	var mcpExt map[string]interface{}

	if len(ts.Decs.Start) > 0 {
		desc := ""
		for _, doc := range ts.Decs.Start {
			desc += strings.TrimSpace(strings.TrimPrefix(doc, "//")) + " "
		}
		pathItem.Summary = strings.TrimSpace(desc)
	}

	for _, field := range iface.Methods.List {
		if len(field.Names) == 0 {
			continue // Embedded interface
		}
		methodName := field.Names[0].Name

		// Track MCP server handler methods
		if methodName == "HandleMcpSse" || methodName == "HandleMcpMessage" || methodName == "WithMcpAuth" || methodName == "MapMcpToolToRoute" {
			if mcpExt == nil {
				mcpExt = make(map[string]interface{})
			}
			if methodName == "HandleMcpSse" {
				mcpExt["sse"] = true
			} else if methodName == "HandleMcpMessage" {
				mcpExt["message"] = true
			} else if methodName == "WithMcpAuth" {
				mcpExt["auth"] = true
			} else if methodName == "MapMcpToolToRoute" {
				mcpExt["proxy"] = true
			}
			continue
		}

		op := &openapi.Operation{
			OperationID: methodName, Responses: make(openapi.Responses, 0),
		}

		if len(field.Decs.Start) > 0 {
			desc := ""
			for _, doc := range field.Decs.Start {
				desc += strings.TrimSpace(strings.TrimPrefix(doc, "//")) + " "
			}
			op.Summary = strings.TrimSpace(desc)
		}

		// Simple heuristic to map method name to HTTP verb
		nameLower := strings.ToLower(methodName)
		if strings.HasPrefix(nameLower, "get") {
			pathItem.Get = op
		} else if strings.HasPrefix(nameLower, "post") {
			pathItem.Post = op
		} else if strings.HasPrefix(nameLower, "put") {
			pathItem.Put = op
		} else if strings.HasPrefix(nameLower, "delete") {
			pathItem.Delete = op
		} else if strings.HasPrefix(nameLower, "patch") {
			pathItem.Patch = op
		} else if strings.HasPrefix(nameLower, "options") {
			pathItem.Options = op
		} else if strings.HasPrefix(nameLower, "head") {
			pathItem.Head = op
		} else if strings.HasPrefix(nameLower, "trace") {
			pathItem.Trace = op
		}
	}

	if mcpExt != nil {
		pathItem.Extensions = make(map[string]interface{})
		pathItem.Extensions["x-mcp-server"] = mcpExt
	}

	return pathItem, nil
}
