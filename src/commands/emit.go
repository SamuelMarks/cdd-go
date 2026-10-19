// Package commands provides parsing and emitting of Cobra CLI commands to and from OpenAPI Operations.
package commands

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Operation object into a Cobra command AST representation.
func Emit(path, method string, op *openapi.Operation) *dst.GenDecl {
	if op == nil {
		return nil
	}

	opID := op.OperationID
	if opID == "" {
		cleanPath := strings.ReplaceAll(path, "/", "_")
		cleanPath = strings.ReplaceAll(cleanPath, "{", "")
		cleanPath = strings.ReplaceAll(cleanPath, "}", "")
		opID = strings.ToLower(method) + naming.ToPascalCase(cleanPath)
	}

	useName := naming.ToSnakeCase(opID)
	cl := &dst.CompositeLit{
		Type: &dst.SelectorExpr{
			X:   dst.NewIdent("cobra"),
			Sel: dst.NewIdent("Command"),
		},
		Elts: []dst.Expr{
			&dst.KeyValueExpr{
				Key:   dst.NewIdent("Use"),
				Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", useName)},
			},
		},
	}

	if op.Summary != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("Short"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", op.Summary)},
		})
	}

	if op.Description != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("Long"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", op.Description)},
		})
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent(naming.ToPascalCase(opID) + "Cmd")},
				Values: []dst.Expr{
					&dst.UnaryExpr{
						Op: token.AND,
						X:  cl,
					},
				},
			},
		},
		Decs: dst.GenDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.NewLine,
				After:  dst.NewLine,
			},
		},
	}

	decl.Decs.Start.Append(fmt.Sprintf("// Method: %s", strings.ToUpper(method)))
	decl.Decs.Start.Append(fmt.Sprintf("// Path: %s", path))

	return decl
}

// EmitMcpCmd generates the MCP CLI subcommand with stdio transport bindings.
func EmitMcpCmd() *dst.GenDecl {
	cl := &dst.CompositeLit{
		Type: &dst.SelectorExpr{
			X:   dst.NewIdent("cobra"),
			Sel: dst.NewIdent("Command"),
		},
		Elts: []dst.Expr{
			&dst.KeyValueExpr{
				Key:   dst.NewIdent("Use"),
				Value: &dst.BasicLit{Kind: token.STRING, Value: `"mcp"`},
			},
			&dst.KeyValueExpr{
				Key:   dst.NewIdent("Short"),
				Value: &dst.BasicLit{Kind: token.STRING, Value: `"Start the Model Context Protocol (MCP) server over stdio"`},
			},
			&dst.KeyValueExpr{
				Key: dst.NewIdent("Run"),
				Value: &dst.FuncLit{
					Type: &dst.FuncType{
						Params: &dst.FieldList{
							List: []*dst.Field{
								{
									Names: []*dst.Ident{dst.NewIdent("cmd")},
									Type: &dst.StarExpr{
										X: &dst.SelectorExpr{
											X:   dst.NewIdent("cobra"),
											Sel: dst.NewIdent("Command"),
										},
									},
								},
								{
									Names: []*dst.Ident{dst.NewIdent("args")},
									Type: &dst.ArrayType{
										Elt: dst.NewIdent("string"),
									},
								},
							},
						},
					},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun:  &dst.SelectorExpr{X: dst.NewIdent("cmd"), Sel: dst.NewIdent("SetIn")},
									Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("os"), Sel: dst.NewIdent("Stdin")}},
								},
							},
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun:  &dst.SelectorExpr{X: dst.NewIdent("cmd"), Sel: dst.NewIdent("SetOut")},
									Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("os"), Sel: dst.NewIdent("Stdout")}},
								},
							},
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun: &dst.SelectorExpr{X: dst.NewIdent("cmd"), Sel: dst.NewIdent("Println")},
									Args: []dst.Expr{
										&dst.BasicLit{Kind: token.STRING, Value: `"MCP server started on stdio"`},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("McpCmd")},
				Values: []dst.Expr{
					&dst.UnaryExpr{
						Op: token.AND,
						X:  cl,
					},
				},
			},
		},
		Decs: dst.GenDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.NewLine,
				After:  dst.NewLine,
			},
		},
	}

	decl.Decs.Start.Append("// McpCmd represents the MCP CLI subcommand and stdio transport bindings.")

	return decl
}
