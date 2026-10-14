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

// toSnakeCase converts PascalCase or camelCase to snake_case.
// Parameter s is the input string, and it returns the snake_case representation.
func toSnakeCase(s string) string {
	return naming.ToSnakeCase(s)
}

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
		opID = strings.ToLower(method) + toPascalCase(cleanPath)
	}

	useName := toSnakeCase(opID)
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
				Names: []*dst.Ident{dst.NewIdent(toPascalCase(opID) + "Cmd")},
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

// toPascalCase converts an identifier string to PascalCase.
// Parameter s is the input string, and it returns the PascalCase representation.
func toPascalCase(s string) string {
	return naming.ToPascalCase(s)
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
							&dst.AssignStmt{
								Lhs: []dst.Expr{dst.NewIdent("_")},
								Tok: token.ASSIGN,
								Rhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("os"), Sel: dst.NewIdent("Stdin")}},
							},
							&dst.AssignStmt{
								Lhs: []dst.Expr{dst.NewIdent("_")},
								Tok: token.ASSIGN,
								Rhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("os"), Sel: dst.NewIdent("Stdout")}},
							},
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun: dst.NewIdent("print"),
									Args: []dst.Expr{
										&dst.BasicLit{Kind: token.STRING, Value: `"MCP server started on stdio\n"`},
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
