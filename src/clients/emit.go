// Package clients provides parsing and emitting of Go client SDK interfaces to and from OpenAPI PathItems.
package clients

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/functions"
	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitClientInterface generates a Go interface for an OpenAPI PathItem to be used as a client.
// Parameter path is the URL path, pathItem is the OpenAPI PathItem, and it returns the AST declaration and any error.
func EmitClientInterface(path string, pathItem *openapi.PathItem) (*dst.GenDecl, error) {
	return EmitClientInterfaceWithOptions(path, pathItem, true)
}

// EmitClientInterfaceWithOptions generates a Go interface for an OpenAPI PathItem with configurable MCP adapter emission.
// Parameter path is the URL path, pathItem is the OpenAPI PathItem, includeMCP enables MCP methods, and it returns the AST declaration and any error.
func EmitClientInterfaceWithOptions(path string, pathItem *openapi.PathItem, includeMCP bool) (*dst.GenDecl, error) {
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

	for _, mo := range openapi.PathItemOperations(pathItem) {
		verb := strings.ToUpper(mo.Method[:1]) + mo.Method[1:]
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature(verb, mo.Operation))
	}

	if includeMCP {
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
		iface.Methods.List = append(iface.Methods.List, &dst.Field{
			Names: []*dst.Ident{dst.NewIdent("AuthLogin")},
			Type: &dst.FuncType{
				Params: &dst.FieldList{
					List: []*dst.Field{
						{Names: []*dst.Ident{dst.NewIdent("username")}, Type: dst.NewIdent("string")},
						{Names: []*dst.Ident{dst.NewIdent("password")}, Type: dst.NewIdent("string")},
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
					Start: dst.Decorations{"// AuthLogin authenticates against the integrated IdP and stores the Bearer session token."},
				},
			},
		})
	}

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

// EmitClientOperationFuncs emits strongly-typed standalone client operation function declarations for a PathItem using src/functions.
// Parameter path is the URL path, pathItem is the OpenAPI PathItem, and it returns the slice of emitted function declarations.
func EmitClientOperationFuncs(path string, pathItem *openapi.PathItem) []dst.Decl {
	if pathItem == nil {
		return nil
	}
	var decls []dst.Decl
	for _, mo := range openapi.PathItemOperations(pathItem) {
		opCopy := *mo.Operation
		if opCopy.OperationID == "" {
			opCopy.OperationID = mo.Method + naming.ToCamelCasePath(path)
		}
		if fn, err := functions.EmitOperation(&opCopy); err == nil && fn != nil {
			fn.Decs.Start.Prepend(fmt.Sprintf("// METHOD: %s", strings.ToUpper(mo.Method)), fmt.Sprintf("// Path: %s", path))
			decls = append(decls, fn)
		}
	}
	return decls
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
