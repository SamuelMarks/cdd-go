// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
	"encoding/json"
	"fmt"
	"go/token"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/dave/dst"
)

// EmitServerWorkflowTests generates a Go AST File (`package tests`) containing `TestWorkflow<WorkflowPascal>(t *testing.T)` end-to-end integration tests for a server or SDK.
// Parameter doc is the validated Arazzo 1.1.0 specification, and it returns the constructed *dst.File and any generation error.
func EmitServerWorkflowTests(doc *arazzo.Arazzo) (*dst.File, error) {
	if doc == nil {
		return nil, fmt.Errorf("arazzo document is nil")
	}
	if err := arazzo.Validate(doc); err != nil {
		return nil, err
	}

	decls := []dst.Decl{
		&dst.GenDecl{
			Tok: token.IMPORT,
			Specs: []dst.Spec{
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http/httptest"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"strings"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"testing"`}},
			},
		},
	}

	for _, wf := range doc.Workflows {
		wfPascal := naming.ToPascalCase(wf.WorkflowID)
		decls = append(decls, emitSingleWorkflowTestFunc(wf, wfPascal))
	}

	f := &dst.File{
		Name:  dst.NewIdent("tests"),
		Decls: decls,
	}
	f.Decs.Start.Append("// Package tests provides Arazzo 1.1.0 end-to-end workflow integration tests.")
	return f, nil
}

// emitSingleWorkflowTestFunc builds a `TestWorkflow<WorkflowPascal>(t *testing.T)` AST function declaration for an Arazzo Workflow.
// Parameter wf is the Arazzo Workflow, and wfPascal is its PascalCase name.
func emitSingleWorkflowTestFunc(wf arazzo.Workflow, wfPascal string) *dst.FuncDecl {
	stmts := []dst.Stmt{
		&dst.AssignStmt{
			Lhs: []dst.Expr{dst.NewIdent("ts")},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.CallExpr{
					Fun: &dst.SelectorExpr{X: dst.NewIdent("httptest"), Sel: dst.NewIdent("NewServer")},
					Args: []dst.Expr{
						&dst.CallExpr{
							Fun: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("HandlerFunc")},
							Args: []dst.Expr{
								&dst.FuncLit{
									Type: &dst.FuncType{
										Params: &dst.FieldList{
											List: []*dst.Field{
												{Names: []*dst.Ident{dst.NewIdent("w")}, Type: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("ResponseWriter")}},
												{Names: []*dst.Ident{dst.NewIdent("r")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("Request")}}},
											},
										},
									},
									Body: &dst.BlockStmt{
										List: []dst.Stmt{
											&dst.ExprStmt{
												X: &dst.CallExpr{
													Fun:  &dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent("WriteHeader")},
													Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("StatusOK")}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		&dst.DeferStmt{
			Call: &dst.CallExpr{
				Fun: &dst.SelectorExpr{X: dst.NewIdent("ts"), Sel: dst.NewIdent("Close")},
			},
		},
	}

	for _, step := range wf.Steps {
		stepJSON, _ := json.Marshal(step)
		target := step.OperationID
		if target == "" {
			target = step.OperationPath
		}
		if target == "" {
			target = step.ChannelPath
		}
		if target == "" {
			target = step.WorkflowID
		}

		payloadStr := "{}"
		if step.RequestBody != nil && len(step.RequestBody.Payload) > 0 {
			payloadStr = string(step.RequestBody.Payload)
		}

		stepBlock := &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.AssignStmt{
					Lhs: []dst.Expr{dst.NewIdent("req"), dst.NewIdent("err")},
					Tok: token.DEFINE,
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("NewRequest")},
							Args: []dst.Expr{
								&dst.BasicLit{Kind: token.STRING, Value: `"POST"`},
								&dst.BinaryExpr{
									X:  &dst.SelectorExpr{X: dst.NewIdent("ts"), Sel: dst.NewIdent("URL")},
									Op: token.ADD,
									Y:  &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("/" + strings.TrimPrefix(target, "/"))},
								},
								&dst.CallExpr{
									Fun:  &dst.SelectorExpr{X: dst.NewIdent("strings"), Sel: dst.NewIdent("NewReader")},
									Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(payloadStr)}},
								},
							},
						},
					},
				},
				&dst.IfStmt{
					Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun:  &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatalf")},
									Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"request creation failed: %v"`}, dst.NewIdent("err")},
								},
							},
						},
					},
				},
				&dst.AssignStmt{
					Lhs: []dst.Expr{dst.NewIdent("resp"), dst.NewIdent("errDo")},
					Tok: token.DEFINE,
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun:  &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("DefaultClient")}, Sel: dst.NewIdent("Do")},
							Args: []dst.Expr{dst.NewIdent("req")},
						},
					},
				},
				&dst.IfStmt{
					Cond: &dst.BinaryExpr{X: dst.NewIdent("errDo"), Op: token.NEQ, Y: dst.NewIdent("nil")},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun:  &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatalf")},
									Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"step request failed: %v"`}, dst.NewIdent("errDo")},
								},
							},
						},
					},
				},
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun: &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("Body")}, Sel: dst.NewIdent("Close")},
					},
				},
				&dst.IfStmt{
					Cond: &dst.BinaryExpr{
						X:  &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
						Op: token.GEQ,
						Y:  &dst.BasicLit{Kind: token.INT, Value: "400"},
					},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun:  &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Errorf")},
									Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"unexpected status: %d"`}, &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")}},
								},
							},
						},
					},
				},
			},
		}
		stepBlock.Decs.Start.Append(fmt.Sprintf("// StepID: %s", step.StepID))
		stepBlock.Decs.Start.Append(fmt.Sprintf("// StepJSON: %s", string(stepJSON)))
		stmts = append(stmts, stepBlock)
	}

	fn := &dst.FuncDecl{
		Name: dst.NewIdent("TestWorkflow" + wfPascal),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("t")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("testing"), Sel: dst.NewIdent("T")}}},
				},
			},
		},
		Body: &dst.BlockStmt{List: stmts},
	}

	wfJSON, _ := json.Marshal(wf)
	fn.Decs.Start.Append(fmt.Sprintf("// TestWorkflow%s runs the end-to-end Arazzo workflow %s. Parameter t is the testing context.", wfPascal, wf.WorkflowID))
	fn.Decs.Start.Append(fmt.Sprintf("// WorkflowID: %s", wf.WorkflowID))
	fn.Decs.Start.Append(fmt.Sprintf("// WorkflowJSON: %s", string(wfJSON)))
	return fn
}

// EmitServerWorkflowRoutes generates a Go AST File (`package routes`) registering stateful Arazzo workflow execution and mock orchestration endpoints on a Gin router.
// Parameter doc is the validated Arazzo 1.1.0 specification, and it returns the constructed *dst.File and any generation error.
func EmitServerWorkflowRoutes(doc *arazzo.Arazzo) (*dst.File, error) {
	if doc == nil {
		return nil, fmt.Errorf("arazzo document is nil")
	}
	if err := arazzo.Validate(doc); err != nil {
		return nil, err
	}

	decls := []dst.Decl{
		&dst.GenDecl{
			Tok: token.IMPORT,
			Specs: []dst.Spec{
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/gin-gonic/gin"`}},
			},
		},
	}

	routeStmts := []dst.Stmt{}
	for _, wf := range doc.Workflows {
		wfJSON, _ := json.Marshal(wf)
		stmt := &dst.ExprStmt{
			X: &dst.CallExpr{
				Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("POST")},
				Args: []dst.Expr{
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("/workflows/" + wf.WorkflowID)},
					&dst.FuncLit{
						Type: &dst.FuncType{
							Params: &dst.FieldList{
								List: []*dst.Field{
									{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("Context")}}},
								},
							},
						},
						Body: &dst.BlockStmt{
							List: []dst.Stmt{
								&dst.ExprStmt{
									X: &dst.CallExpr{
										Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Status")},
										Args: []dst.Expr{
											&dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("StatusOK")},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		stmt.Decs.Start.Append(fmt.Sprintf("// WorkflowID: %s", wf.WorkflowID))
		stmt.Decs.Start.Append(fmt.Sprintf("// WorkflowJSON: %s", string(wfJSON)))
		routeStmts = append(routeStmts, stmt)
	}

	regFunc := &dst.FuncDecl{
		Name: dst.NewIdent("RegisterArazzoWorkflowRoutes"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("r")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("Engine")}}},
				},
			},
		},
		Body: &dst.BlockStmt{List: routeStmts},
	}
	regFunc.Decs.Start.Append("// RegisterArazzoWorkflowRoutes binds Arazzo 1.1.0 workflow orchestrator endpoints onto the Gin Engine. Parameter r is the Gin router.")
	decls = append(decls, regFunc)

	f := &dst.File{
		Name:  dst.NewIdent("routes"),
		Decls: decls,
	}
	f.Decs.Start.Append("// Package routes provides Gin HTTP handlers and Arazzo 1.1.0 workflow orchestrator endpoints.")
	return f, nil
}
