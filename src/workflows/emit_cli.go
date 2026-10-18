// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
	"encoding/json"
	"fmt"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/dave/dst"
)

// EmitCLIWorkflows generates a Go AST File (`package main`) containing Cobra commands that wrap and reuse the generated SDK `workflows.WorkflowClient`.
// Parameter doc is the validated Arazzo 1.1.0 specification, and it returns the constructed *dst.File and any generation error.
func EmitCLIWorkflows(doc *arazzo.Arazzo) (*dst.File, error) {
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
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"context"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"fmt"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"os"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/workflows"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/spf13/cobra"`}},
			},
		},
	}

	shortStr := doc.Info.Summary
	if shortStr == "" {
		shortStr = doc.Info.Title
	}
	longStr := doc.Info.Description
	if longStr == "" {
		longStr = doc.Info.Title
	}

	rootCmdDecl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("WorkflowCmd")},
				Values: []dst.Expr{
					&dst.UnaryExpr{
						Op: token.AND,
						X: &dst.CompositeLit{
							Type: &dst.SelectorExpr{X: dst.NewIdent("cobra"), Sel: dst.NewIdent("Command")},
							Elts: []dst.Expr{
								&dst.KeyValueExpr{Key: dst.NewIdent("Use"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"workflow"`}},
								&dst.KeyValueExpr{Key: dst.NewIdent("Short"), Value: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(shortStr)}},
								&dst.KeyValueExpr{Key: dst.NewIdent("Long"), Value: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(longStr)}},
								&dst.KeyValueExpr{Key: dst.NewIdent("Version"), Value: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(doc.Info.Version)}},
							},
						},
					},
				},
			},
		},
	}
	rootCmdDecl.Decs.Start.Append("// WorkflowCmd is the root Cobra command for executing Arazzo 1.1.0 workflows via the SDK.")
	rootCmdDecl.Decs.Start.Append(fmt.Sprintf("// ArazzoTitle: %s", doc.Info.Title))
	for _, wf := range doc.Workflows {
		wfPascal := naming.ToPascalCase(wf.WorkflowID)
		decls = append(decls, emitCLIWorkflowSubcommand(wf, wfPascal))
	}

	decls = append([]dst.Decl{decls[0], rootCmdDecl}, decls[1:]...)

	f := &dst.File{
		Name:  dst.NewIdent("main"),
		Decls: decls,
	}
	f.Decs.Start.Append("// Package main provides the Cobra CLI workflow runner reusing the generated SDK workflows package.")
	return f, nil
}

// emitCLIWorkflowSubcommand builds the `<WorkflowPascal>WorkflowCmd` Cobra subcommand AST declaration for a single Workflow.
// Parameter wf is the Arazzo Workflow, and wfPascal is its PascalCase identifier.
func emitCLIWorkflowSubcommand(wf arazzo.Workflow, wfPascal string) *dst.GenDecl {
	useName := naming.ToSnakeCase(wf.WorkflowID)
	shortStr := wf.Summary
	if shortStr == "" {
		shortStr = "Execute workflow " + wf.WorkflowID
	}

	elts := []dst.Expr{
		&dst.KeyValueExpr{Key: dst.NewIdent("Use"), Value: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(useName)}},
		&dst.KeyValueExpr{Key: dst.NewIdent("Short"), Value: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(shortStr)}},
	}
	if wf.Description != "" {
		elts = append(elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("Long"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(wf.Description)},
		})
	}

	// RunE body instantiates workflows.NewWorkflowClient(os.Getenv("BASE_URL")) and calls client.Execute<WfPascal>(context.Background(), workflows.<WfPascal>Input{})
	runE := &dst.KeyValueExpr{
		Key: dst.NewIdent("RunE"),
		Value: &dst.FuncLit{
			Type: &dst.FuncType{
				Params: &dst.FieldList{
					List: []*dst.Field{
						{Names: []*dst.Ident{dst.NewIdent("cmd")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("cobra"), Sel: dst.NewIdent("Command")}}},
						{Names: []*dst.Ident{dst.NewIdent("args")}, Type: &dst.ArrayType{Elt: dst.NewIdent("string")}},
					},
				},
				Results: &dst.FieldList{
					List: []*dst.Field{{Type: dst.NewIdent("error")}},
				},
			},
			Body: &dst.BlockStmt{
				List: []dst.Stmt{
					&dst.AssignStmt{
						Lhs: []dst.Expr{dst.NewIdent("client")},
						Tok: token.DEFINE,
						Rhs: []dst.Expr{
							&dst.CallExpr{
								Fun: &dst.SelectorExpr{X: dst.NewIdent("workflows"), Sel: dst.NewIdent("NewWorkflowClient")},
								Args: []dst.Expr{
									&dst.CallExpr{
										Fun:  &dst.SelectorExpr{X: dst.NewIdent("os"), Sel: dst.NewIdent("Getenv")},
										Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"BASE_URL"`}},
									},
								},
							},
						},
					},
					&dst.AssignStmt{
						Lhs: []dst.Expr{dst.NewIdent("out"), dst.NewIdent("err")},
						Tok: token.DEFINE,
						Rhs: []dst.Expr{
							&dst.CallExpr{
								Fun: &dst.SelectorExpr{X: dst.NewIdent("client"), Sel: dst.NewIdent("Execute" + wfPascal)},
								Args: []dst.Expr{
									&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Background")}},
									&dst.CompositeLit{Type: &dst.SelectorExpr{X: dst.NewIdent("workflows"), Sel: dst.NewIdent(wfPascal + "Input")}},
								},
							},
						},
					},
					&dst.IfStmt{
						Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
						Body: &dst.BlockStmt{
							List: []dst.Stmt{
								&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("err")}},
							},
						},
					},
					&dst.ExprStmt{
						X: &dst.CallExpr{
							Fun: &dst.SelectorExpr{X: dst.NewIdent("fmt"), Sel: dst.NewIdent("Fprintf")},
							Args: []dst.Expr{
								&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("cmd"), Sel: dst.NewIdent("OutOrStdout")}},
								&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("%+v")},
								dst.NewIdent("out"),
							},
						},
					},
					&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("nil")}},
				},
			},
		},
	}
	elts = append(elts, runE)

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent(wfPascal + "WorkflowCmd")},
				Values: []dst.Expr{
					&dst.UnaryExpr{
						Op: token.AND,
						X: &dst.CompositeLit{
							Type: &dst.SelectorExpr{X: dst.NewIdent("cobra"), Sel: dst.NewIdent("Command")},
							Elts: elts,
						},
					},
				},
			},
		},
	}

	decl.Decs.Start.Append(fmt.Sprintf("// %sWorkflowCmd executes the %s workflow via the SDK.", wfPascal, wf.WorkflowID))
	decl.Decs.Start.Append(fmt.Sprintf("// WorkflowID: %s", wf.WorkflowID))
	if len(wf.DependsOn) > 0 {
		decl.Decs.Start.Append(fmt.Sprintf("// DependsOn: %s", strings.Join(wf.DependsOn, ",")))
	}
	if wf.Inputs != nil && len(wf.Inputs.Properties) > 0 {
		propKeys := make([]string, 0, len(wf.Inputs.Properties))
		for k := range wf.Inputs.Properties {
			propKeys = append(propKeys, k)
		}
		sort.Strings(propKeys)
		for _, k := range propKeys {
			prop := wf.Inputs.Properties[k]
			decl.Decs.Start.Append(fmt.Sprintf("// FlagInput: %s:%s", k, prop.Type))
		}
	}
	wfJSON, _ := json.Marshal(wf)
	decl.Decs.Start.Append(fmt.Sprintf("// WorkflowJSON: %s", string(wfJSON)))

	return decl
}
