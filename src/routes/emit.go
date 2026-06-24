package routes

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitHandlerInterface generates a Go interface for an OpenAPI PathItem.
func EmitHandlerInterface(path string, pathItem *openapi.PathItem) (*dst.GenDecl, error) {
	if pathItem == nil {
		return nil, fmt.Errorf("PathItem is nil")
	}

	interfaceName := "Handler" + toCamelCase(path)
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

	// Native MCP SSE Endpoint Generation Support
	iface.Methods.List = append(iface.Methods.List, &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("HandleMcpSse")},
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("c")},
						Type: &dst.StarExpr{
							X: &dst.SelectorExpr{
								X:   dst.NewIdent("gin"),
								Sel: dst.NewIdent("Context"),
							},
						},
					},
				},
			},
			Results: &dst.FieldList{},
		},
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// HandleMcpSse exposes an SSE endpoint for MCP server integration."},
			},
		},
	})
	iface.Methods.List = append(iface.Methods.List, &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("HandleMcpMessage")},
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("c")},
						Type: &dst.StarExpr{
							X: &dst.SelectorExpr{
								X:   dst.NewIdent("gin"),
								Sel: dst.NewIdent("Context"),
							},
						},
					},
				},
			},
			Results: &dst.FieldList{},
		},
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// HandleMcpMessage handles incoming MCP messages (e.g., tool calls) over HTTP."},
			},
		},
	})
	iface.Methods.List = append(iface.Methods.List, &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("WithMcpAuth")},
		Type: &dst.FuncType{
			Params: &dst.FieldList{},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{
						Type: &dst.SelectorExpr{
							X:   dst.NewIdent("gin"),
							Sel: dst.NewIdent("HandlerFunc"),
						},
					},
				},
			},
		},
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// WithMcpAuth provides HTTP Request/Auth Bridging to secure MCP endpoints."},
			},
		},
	})
	iface.Methods.List = append(iface.Methods.List, &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("MapMcpToolToRoute")},
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("toolName")}, Type: dst.NewIdent("string")},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{
						Type: &dst.SelectorExpr{
							X:   dst.NewIdent("gin"),
							Sel: dst.NewIdent("HandlerFunc"),
						},
					},
				},
			},
		},
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// MapMcpToolToRoute provides a Dynamic API-to-Tool Proxy for resolving incoming tool calls to backend route handlers."},
			},
		},
	})

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(interfaceName),
		Type: iface,
	}

	if pathItem.Summary != "" {
		ts.Decs.Start.Append(fmt.Sprintf("// %s", pathItem.Summary))
	}

	return &dst.GenDecl{
		Tok:   token.TYPE,
		Specs: []dst.Spec{ts},
	}, nil
}

func emitMethodSignature(method string, op *openapi.Operation) *dst.Field {
	name := method
	if op.OperationID != "" {
		name = strings.ToUpper(op.OperationID[:1]) + op.OperationID[1:]
	}

	fType := &dst.FuncType{
		Params: &dst.FieldList{
			List: []*dst.Field{
				{
					Names: []*dst.Ident{dst.NewIdent("c")},
					Type: &dst.StarExpr{
						X: &dst.SelectorExpr{
							X:   dst.NewIdent("gin"),
							Sel: dst.NewIdent("Context"),
						},
					},
				},
			},
		},
		Results: &dst.FieldList{},
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

// EmitHandlerStruct generates a set of declarations representing the handler struct and its methods for a given API path.
func EmitHandlerStruct(path string, pathItem *openapi.PathItem) ([]dst.Decl, error) {
	if pathItem == nil {
		return nil, fmt.Errorf("PathItem is nil")
	}

	camelName := toCamelCase(path)
	interfaceName := "Handler" + camelName
	structName := "handler" + camelName

	decls := []dst.Decl{}

	structDecl := &dst.GenDecl{
		Tok: token.TYPE,
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent(structName),
				Type: &dst.StructType{
					Fields: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("daos")}, Type: &dst.SelectorExpr{X: dst.NewIdent("daos"), Sel: dst.NewIdent("DAOFactory")}},
						},
					},
				},
			},
		},
	}
	decls = append(decls, structDecl)

	factoryFunc := &dst.FuncDecl{
		Name: dst.NewIdent("New" + interfaceName),
		Type: &dst.FuncType{
			Params:  &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("daos")}, Type: &dst.SelectorExpr{X: dst.NewIdent("daos"), Sel: dst.NewIdent("DAOFactory")}}}},
			Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent(interfaceName)}}},
		},
		Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{Type: dst.NewIdent(structName), Elts: []dst.Expr{&dst.KeyValueExpr{Key: dst.NewIdent("daos"), Value: dst.NewIdent("daos")}}}}}}}},
	}
	decls = append(decls, factoryFunc)

	addMethod := func(name string) {
		methodFunc := &dst.FuncDecl{
			Recv: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("h")}, Type: &dst.StarExpr{X: dst.NewIdent(structName)}}}},
			Name: dst.NewIdent(name),
			Type: &dst.FuncType{
				Params:  &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("Context")}}}}},
				Results: &dst.FieldList{},
			},
			Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("JSON")}, Args: []dst.Expr{dst.NewIdent("501"), &dst.CompositeLit{Type: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("H")}, Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"error"`}, Value: &dst.BasicLit{Kind: token.STRING, Value: `"Not implemented"`}}}}}}}}},
		}
		decls = append(decls, methodFunc)
	}

	if pathItem.Get != nil {
		name := "Get"
		if pathItem.Get.OperationID != "" {
			name = strings.ToUpper(pathItem.Get.OperationID[:1]) + pathItem.Get.OperationID[1:]
		}
		addMethod(name)
	}
	if pathItem.Post != nil {
		name := "Post"
		if pathItem.Post.OperationID != "" {
			name = strings.ToUpper(pathItem.Post.OperationID[:1]) + pathItem.Post.OperationID[1:]
		}
		addMethod(name)
	}
	if pathItem.Put != nil {
		name := "Put"
		if pathItem.Put.OperationID != "" {
			name = strings.ToUpper(pathItem.Put.OperationID[:1]) + pathItem.Put.OperationID[1:]
		}
		addMethod(name)
	}
	if pathItem.Delete != nil {
		name := "Delete"
		if pathItem.Delete.OperationID != "" {
			name = strings.ToUpper(pathItem.Delete.OperationID[:1]) + pathItem.Delete.OperationID[1:]
		}
		addMethod(name)
	}
	if pathItem.Patch != nil {
		name := "Patch"
		if pathItem.Patch.OperationID != "" {
			name = strings.ToUpper(pathItem.Patch.OperationID[:1]) + pathItem.Patch.OperationID[1:]
		}
		addMethod(name)
	}
	if pathItem.Options != nil {
		name := "Options"
		if pathItem.Options.OperationID != "" {
			name = strings.ToUpper(pathItem.Options.OperationID[:1]) + pathItem.Options.OperationID[1:]
		}
		addMethod(name)
	}
	if pathItem.Head != nil {
		name := "Head"
		if pathItem.Head.OperationID != "" {
			name = strings.ToUpper(pathItem.Head.OperationID[:1]) + pathItem.Head.OperationID[1:]
		}
		addMethod(name)
	}
	if pathItem.Trace != nil {
		name := "Trace"
		if pathItem.Trace.OperationID != "" {
			name = strings.ToUpper(pathItem.Trace.OperationID[:1]) + pathItem.Trace.OperationID[1:]
		}
		addMethod(name)
	}

	addMethod("HandleMcpSse")
	addMethod("HandleMcpMessage")

	withMcpAuthFunc := &dst.FuncDecl{
		Recv: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("h")}, Type: &dst.StarExpr{X: dst.NewIdent(structName)}}}},
		Name: dst.NewIdent("WithMcpAuth"),
		Type: &dst.FuncType{
			Params:  &dst.FieldList{},
			Results: &dst.FieldList{List: []*dst.Field{{Type: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("HandlerFunc")}}}},
		},
		Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{&dst.FuncLit{Type: &dst.FuncType{Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("Context")}}}}}}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Next")}, Args: []dst.Expr{}}}}}}}}}},
	}
	decls = append(decls, withMcpAuthFunc)

	mapMcpToolFunc := &dst.FuncDecl{
		Recv: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("h")}, Type: &dst.StarExpr{X: dst.NewIdent(structName)}}}},
		Name: dst.NewIdent("MapMcpToolToRoute"),
		Type: &dst.FuncType{
			Params:  &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("toolName")}, Type: dst.NewIdent("string")}}},
			Results: &dst.FieldList{List: []*dst.Field{{Type: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("HandlerFunc")}}}},
		},
		Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{&dst.FuncLit{Type: &dst.FuncType{Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("Context")}}}}}}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Next")}, Args: []dst.Expr{}}}}}}}}}},
	}
	decls = append(decls, mapMcpToolFunc)

	registerFunc := &dst.FuncDecl{
		Recv: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("h")}, Type: &dst.StarExpr{X: dst.NewIdent(structName)}}}},
		Name: dst.NewIdent("Register"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("r")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("Engine")}}}}},
		},
		Body: &dst.BlockStmt{List: []dst.Stmt{}},
	}
	decls = append(decls, registerFunc)

	return decls, nil
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "/")
	var res string
	for _, p := range parts {
		p = strings.ReplaceAll(p, "{", "")
		p = strings.ReplaceAll(p, "}", "")
		if p != "" {
			res += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	if res == "" {
		res = "Root"
	}
	return res
}
