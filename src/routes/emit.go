// Package routes provides parsing and emitting of Gin HTTP handler interfaces and structs to and from OpenAPI PathItems.
package routes

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitHandlerInterface generates a Go interface for an OpenAPI PathItem.
// Parameter path is the URL path, pathItem is the OpenAPI PathItem, and it returns the AST declaration and any error.
func EmitHandlerInterface(path string, pathItem *openapi.PathItem) (*dst.GenDecl, error) {
	return EmitHandlerInterfaceWithOptions(path, pathItem, true)
}

// EmitHandlerInterfaceWithOptions generates a Go interface for an OpenAPI PathItem with configurable MCP method emission.
// Parameter path is the URL path, pathItem is the OpenAPI PathItem, includeMCP enables MCP methods, and it returns the AST declaration and any error.
func EmitHandlerInterfaceWithOptions(path string, pathItem *openapi.PathItem, includeMCP bool) (*dst.GenDecl, error) {
	if pathItem == nil {
		return nil, fmt.Errorf("PathItem is nil")
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("invalid path %q: must start with '/'", path)
	}

	interfaceName := "Handler" + naming.ToCamelCasePath(path)
	iface := &dst.InterfaceType{
		Methods: &dst.FieldList{},
	}

	for _, mo := range openapi.PathItemOperations(pathItem) {
		verb := strings.ToUpper(mo.Method[:1]) + mo.Method[1:]
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature(verb, mo.Operation))
	}

	if includeMCP {
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

// emitMethodSignature builds a dst.Field method signature for a route handler interface method.
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

// openAPIToGinPath converts OpenAPI curly-brace path parameters (e.g., /users/{id}) into Gin colon parameters (/users/:id).
// Parameter path is the OpenAPI path template, and it returns the Gin-compatible route path.
func openAPIToGinPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") && len(p) > 2 {
			parts[i] = ":" + p[1:len(p)-1]
		}
	}
	return strings.Join(parts, "/")
}

// EmitRouteBundle validates path and pathItem once and returns both the Handler interface declaration and the handler struct/method declarations.
// Parameter path is the URL path, pathItem is the OpenAPI PathItem, and it returns the interface declaration, struct declarations, and any error.
func EmitRouteBundle(path string, pathItem *openapi.PathItem) (*dst.GenDecl, []dst.Decl, error) {
	ifaceDecl, err := EmitHandlerInterface(path, pathItem)
	if err != nil {
		return nil, nil, err
	}
	structDecls, _ := EmitHandlerStruct(path, pathItem)
	return ifaceDecl, structDecls, nil
}

// EmitHandlerStruct generates a set of declarations representing the handler struct and its methods for a given API path.
// Parameter path is the URL path, pathItem is the OpenAPI PathItem, and it returns the AST declarations and any error.
func EmitHandlerStruct(path string, pathItem *openapi.PathItem) ([]dst.Decl, error) {
	if pathItem == nil {
		return nil, fmt.Errorf("PathItem is nil")
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("invalid path %q: must start with '/'", path)
	}

	camelName := naming.ToCamelCasePath(path)
	interfaceName := "Handler" + camelName
	structName := "handler" + camelName

	decls := []dst.Decl{}

	daosField := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("daos")},
		Type:  &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("daos"), Sel: dst.NewIdent("DAOFactory")}},
	}
	daosField.Decs.Start.Append("// daos provides data access objects for the route handler.")

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(structName),
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{daosField},
			},
		},
	}
	ts.Decs.Start.Append("// " + structName + " implements " + interfaceName + ".")

	structDecl := &dst.GenDecl{
		Tok:   token.TYPE,
		Specs: []dst.Spec{ts},
	}
	decls = append(decls, structDecl)

	factoryFunc := &dst.FuncDecl{
		Name: dst.NewIdent("New" + interfaceName),
		Type: &dst.FuncType{
			Params:  &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("daos")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("daos"), Sel: dst.NewIdent("DAOFactory")}}}}},
			Results: &dst.FieldList{List: []*dst.Field{{Type: &dst.StarExpr{X: dst.NewIdent(structName)}}}},
		},
		Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{Type: dst.NewIdent(structName), Elts: []dst.Expr{&dst.KeyValueExpr{Key: dst.NewIdent("daos"), Value: dst.NewIdent("daos")}}}}}}}},
	}
	factoryFunc.Decs.Start.Append("// New" + interfaceName + " creates a new " + structName + " instance. Parameter daos is the DAO factory.")
	decls = append(decls, factoryFunc)

	ginPath := openAPIToGinPath(path)
	registerStmts := []dst.Stmt{}

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
		methodFunc.Decs.Start.Append("// " + name + " handles the HTTP route operation. Parameter c is the Gin context.")
		decls = append(decls, methodFunc)
	}

	addRouteRegistration := func(ginVerb string, handlerMethod string) {
		registerStmts = append(registerStmts, &dst.ExprStmt{
			X: &dst.CallExpr{
				Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent(ginVerb)},
				Args: []dst.Expr{
					&dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", ginPath)},
					&dst.SelectorExpr{X: dst.NewIdent("h"), Sel: dst.NewIdent(handlerMethod)},
				},
			},
		})
	}

	for _, mo := range openapi.PathItemOperations(pathItem) {
		name := strings.ToUpper(mo.Method[:1]) + mo.Method[1:]
		if mo.Operation.OperationID != "" {
			name = strings.ToUpper(mo.Operation.OperationID[:1]) + mo.Operation.OperationID[1:]
		}
		addMethod(name)
		upperVerb := strings.ToUpper(mo.Method)
		if upperVerb == "TRACE" {
			registerStmts = append(registerStmts, &dst.ExprStmt{
				X: &dst.CallExpr{
					Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("Handle")},
					Args: []dst.Expr{
						&dst.BasicLit{Kind: token.STRING, Value: `"TRACE"`},
						&dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", ginPath)},
						&dst.SelectorExpr{X: dst.NewIdent("h"), Sel: dst.NewIdent(name)},
					},
				},
			})
		} else {
			addRouteRegistration(upperVerb, name)
		}
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
	withMcpAuthFunc.Decs.Start.Append("// WithMcpAuth returns a Gin middleware bridging HTTP auth to MCP.")
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
	mapMcpToolFunc.Decs.Start.Append("// MapMcpToolToRoute resolves an MCP tool name to a Gin HandlerFunc. Parameter toolName is the MCP tool identifier.")
	decls = append(decls, mapMcpToolFunc)

	registerFunc := &dst.FuncDecl{
		Recv: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("h")}, Type: &dst.StarExpr{X: dst.NewIdent(structName)}}}},
		Name: dst.NewIdent("Register"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("r")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("Engine")}}}}},
		},
		Body: &dst.BlockStmt{List: registerStmts},
	}
	registerFunc.Decs.Start.Append("// Register binds all route handlers for " + path + " onto the Gin engine. Parameter r is the Gin engine.")
	decls = append(decls, registerFunc)

	return decls, nil
}
