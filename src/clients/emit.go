// Package clients provides parsing and emitting of Go client SDK interfaces to and from OpenAPI PathItems.
package clients

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitClientInterface generates a Go interface for an OpenAPI PathItem to be used as a client.
func EmitClientInterface(path string, pathItem *openapi.PathItem) (*dst.GenDecl, error) {
	if pathItem == nil {
		return nil, fmt.Errorf("PathItem is nil")
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("invalid path %q: must start with '/'", path)
	}

	interfaceName := "Client" + naming.ToCamelCasePath(path)
	iface := &dst.InterfaceType{
		Methods: &dst.FieldList{},
	}

	if pathItem.Get != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Get", pathItem.Get))
	}
	if pathItem.Post != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Post", pathItem.Post))
	}
	if pathItem.Put != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Put", pathItem.Put))
	}
	if pathItem.Delete != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Delete", pathItem.Delete))
	}
	if pathItem.Patch != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Patch", pathItem.Patch))
	}
	if pathItem.Options != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Options", pathItem.Options))
	}
	if pathItem.Head != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Head", pathItem.Head))
	}
	if pathItem.Trace != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Trace", pathItem.Trace))
	}

	// Native MCP Adapters
	iface.Methods.List = append(iface.Methods.List, &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("GetTools")},
		Type: &dst.FuncType{
			Params: &dst.FieldList{},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: &dst.ArrayType{Elt: dst.NewIdent("string")}},
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// GetTools returns native MCP tools for this client interface."},
			},
		},
	})
	iface.Methods.List = append(iface.Methods.List, &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("GetResources")},
		Type: &dst.FuncType{
			Params: &dst.FieldList{},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: &dst.ArrayType{Elt: dst.NewIdent("string")}},
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// GetResources returns native MCP resources for this client interface."},
			},
		},
	})
	iface.Methods.List = append(iface.Methods.List, &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("ExecuteTool")},
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("toolName")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("args")}, Type: &dst.MapType{Key: dst.NewIdent("string"), Value: dst.NewIdent("string")}},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: dst.NewIdent("string")},
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// ExecuteTool routes an LLM tool execution request natively to the corresponding client method."},
			},
		},
	})

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(interfaceName),
		Type: iface,
	}

	decl := &dst.GenDecl{
		Tok:   token.TYPE,
		Specs: []dst.Spec{ts},
	}

	pathComment := fmt.Sprintf("// Path: %s", path)
	ts.Decs.Start.Append(pathComment)
	decl.Decs.Start.Append(pathComment)
	if pathItem.Summary != "" {
		summaryComment := fmt.Sprintf("// %s", pathItem.Summary)
		ts.Decs.Start.Append(summaryComment)
		decl.Decs.Start.Append(summaryComment)
	}

	return decl, nil
}

// emitMethodSignature builds a dst.Field method signature for a client interface method.
// Parameter method is the HTTP method name, and op is the OpenAPI Operation.
func emitMethodSignature(method string, op *openapi.Operation) *dst.Field {
	name := method
	if op.OperationID != "" {
		name = strings.ToUpper(op.OperationID[:1]) + op.OperationID[1:]
	}

	fType := &dst.FuncType{
		Params: &dst.FieldList{
			List: []*dst.Field{
				{
					Names: []*dst.Ident{dst.NewIdent("req")},
					Type: &dst.StarExpr{
						X: &dst.SelectorExpr{
							X:   dst.NewIdent("http"),
							Sel: dst.NewIdent("Request"),
						},
					},
				},
			},
		},
		Results: &dst.FieldList{
			List: []*dst.Field{
				{
					Type: &dst.StarExpr{
						X: &dst.SelectorExpr{
							X:   dst.NewIdent("http"),
							Sel: dst.NewIdent("Response"),
						},
					},
				},
				{
					Type: dst.NewIdent("error"),
				},
			},
		},
	}

	f := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent(name)},
		Type:  fType,
	}

	f.Decs.Start.Append(fmt.Sprintf("// METHOD: %s", method))
	if op.Summary != "" {
		f.Decs.Start.Append(fmt.Sprintf("// %s", op.Summary))
	}

	return f
}
