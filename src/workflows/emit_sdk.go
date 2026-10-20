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
	"github.com/SamuelMarks/cdd-go/src/classes"
	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// inferStepHTTPMethod determines the HTTP method for an Arazzo step from its OperationPath or OperationID.
// Parameter step is the Arazzo Step, and it returns the uppercase HTTP verb.
func inferStepHTTPMethod(step arazzo.Step) string {
	if step.OperationPath != "" {
		lowerPath := strings.ToLower(step.OperationPath)
		for _, v := range []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"} {
			if strings.HasSuffix(lowerPath, "/"+v) {
				return strings.ToUpper(v)
			}
		}
	}
	lowerOp := strings.ToLower(step.OperationID)
	switch {
	case strings.HasPrefix(lowerOp, "get") || strings.HasPrefix(lowerOp, "list") || strings.HasPrefix(lowerOp, "fetch") || strings.HasPrefix(lowerOp, "find"):
		if step.RequestBody == nil {
			return "GET"
		}
	case strings.HasPrefix(lowerOp, "delete") || strings.HasPrefix(lowerOp, "remove"):
		return "DELETE"
	case strings.HasPrefix(lowerOp, "put") || strings.HasPrefix(lowerOp, "replace"):
		return "PUT"
	case strings.HasPrefix(lowerOp, "patch") || strings.HasPrefix(lowerOp, "update"):
		return "PATCH"
	}
	return "POST"
}

// EmitSDKWorkflows generates a Go AST File (`package workflows`) implementing the Arazzo 1.1.0 Client SDK workflow orchestrator.
// Parameter doc is the validated Arazzo 1.1.0 specification, and it returns the constructed *dst.File and any generation error.
func EmitSDKWorkflows(doc *arazzo.Arazzo) (*dst.File, error) {
	if doc == nil {
		return nil, fmt.Errorf("arazzo document is nil")
	}
	if err := arazzo.Validate(doc); err != nil {
		return nil, err
	}

	needsStrings := false
	needsTime := false
	for _, wf := range doc.Workflows {
		for _, step := range wf.Steps {
			if step.RequestBody != nil && len(step.RequestBody.Payload) > 0 {
				needsStrings = true
			}
			if step.Timeout != nil && *step.Timeout > 0 {
				needsTime = true
			}
		}
	}

	importSpecs := []dst.Spec{
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"context"`}},
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"fmt"`}},
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
	}
	if needsStrings {
		importSpecs = append(importSpecs, &dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"strings"`}})
	}
	if needsTime {
		importSpecs = append(importSpecs, &dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"time"`}})
	}

	decls := []dst.Decl{
		&dst.GenDecl{
			Tok:   token.IMPORT,
			Specs: importSpecs,
		},
	}

	// Emit Arazzo metadata variables as strongly-typed AST constants/variables
	infoSpecs := []dst.Spec{
		&dst.ValueSpec{
			Names:  []*dst.Ident{dst.NewIdent("ArazzoVersion")},
			Values: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(doc.Arazzo)}},
			Decs:   dst.ValueSpecDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// ArazzoVersion holds the semantic version of the Arazzo specification."}}},
		},
		&dst.ValueSpec{
			Names:  []*dst.Ident{dst.NewIdent("ArazzoTitle")},
			Values: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(doc.Info.Title)}},
			Decs:   dst.ValueSpecDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// ArazzoTitle holds the title of the Arazzo specification."}}},
		},
		&dst.ValueSpec{
			Names:  []*dst.Ident{dst.NewIdent("ArazzoInfoVersion")},
			Values: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(doc.Info.Version)}},
			Decs:   dst.ValueSpecDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// ArazzoInfoVersion holds the document version of the Arazzo specification."}}},
		},
	}
	if doc.Info.Summary != "" {
		infoSpecs = append(infoSpecs, &dst.ValueSpec{
			Names:  []*dst.Ident{dst.NewIdent("ArazzoSummary")},
			Values: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(doc.Info.Summary)}},
			Decs:   dst.ValueSpecDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// ArazzoSummary holds the short summary of the Arazzo specification."}}},
		})
	}
	if doc.Info.Description != "" {
		infoSpecs = append(infoSpecs, &dst.ValueSpec{
			Names:  []*dst.Ident{dst.NewIdent("ArazzoDescription")},
			Values: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(doc.Info.Description)}},
			Decs:   dst.ValueSpecDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// ArazzoDescription holds the description of the Arazzo specification."}}},
		})
	}

	infoDecl := &dst.GenDecl{
		Tok:   token.VAR,
		Specs: infoSpecs,
	}
	infoDecl.Decs.Start.Append("// ArazzoVersion and companion variables store top-level Arazzo document metadata.")
	if doc.Self != "" {
		infoDecl.Decs.Start.Append(fmt.Sprintf("// SelfURI: %s", doc.Self))
	}
	if len(doc.Extensions) > 0 {
		extB, _ := json.Marshal(doc.Extensions)
		infoDecl.Decs.Start.Append(fmt.Sprintf("// RootExtensionsJSON: %s", string(extB)))
	}
	if len(doc.Info.Extensions) > 0 {
		infoExtB, _ := json.Marshal(doc.Info.Extensions)
		infoDecl.Decs.Start.Append(fmt.Sprintf("// InfoExtensionsJSON: %s", string(infoExtB)))
	}
	decls = append(decls, infoDecl)

	// Emit SourceDescriptionsMetadata variable
	sdElts := make([]dst.Expr, 0, len(doc.SourceDescriptions))
	for _, sd := range doc.SourceDescriptions {
		sdElts = append(sdElts, &dst.KeyValueExpr{
			Key:   &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(sd.Name)},
			Value: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(sd.URL)},
		})
	}
	sdDecl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("SourceDescriptionsMetadata")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Type: &dst.MapType{Key: dst.NewIdent("string"), Value: dst.NewIdent("string")},
						Elts: sdElts,
					},
				},
			},
		},
	}
	sdDecl.Decs.Start.Append("// SourceDescriptionsMetadata maps sourceDescription names to their base URLs.")
	for _, sd := range doc.SourceDescriptions {
		if sd.Type != "" {
			sdDecl.Decs.Start.Append(fmt.Sprintf("// SourceType: %s=%s", sd.Name, sd.Type))
		}
	}
	decls = append(decls, sdDecl)

	// Emit ComponentsMetadata variable if present
	if doc.Components != nil {
		compJSON, _ := json.Marshal(doc.Components)
		compDecl := &dst.GenDecl{
			Tok: token.VAR,
			Specs: []dst.Spec{
				&dst.ValueSpec{
					Names:  []*dst.Ident{dst.NewIdent("ComponentsMetadataJSON")},
					Values: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(string(compJSON))}},
				},
			},
		}
		compDecl.Decs.Start.Append("// ComponentsMetadataJSON holds the serialized reusable Arazzo components.")
		if len(doc.Components.Extensions) > 0 {
			cExtB, _ := json.Marshal(doc.Components.Extensions)
			compDecl.Decs.Start.Append(fmt.Sprintf("// ComponentsExtensionsJSON: %s", string(cExtB)))
		}
		decls = append(decls, compDecl)
	}

	// Emit WorkflowClient struct and constructor
	clientStruct := &dst.GenDecl{
		Tok: token.TYPE,
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent("WorkflowClient"),
				Type: &dst.StructType{
					Fields: &dst.FieldList{
						List: []*dst.Field{
							{
								Names: []*dst.Ident{dst.NewIdent("BaseURL")},
								Type:  dst.NewIdent("string"),
								Decs:  dst.FieldDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// BaseURL is the default endpoint base URL for workflow steps."}}},
							},
							{
								Names: []*dst.Ident{dst.NewIdent("HTTPClient")},
								Type:  &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("Client")}},
								Decs:  dst.FieldDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// HTTPClient is the underlying HTTP client used to dispatch workflow steps."}}},
							},
							{
								Names: []*dst.Ident{dst.NewIdent("SourceURLs")},
								Type:  &dst.MapType{Key: dst.NewIdent("string"), Value: dst.NewIdent("string")},
								Decs:  dst.FieldDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// SourceURLs maps sourceDescription names to resolved runtime URLs."}}},
							},
						},
					},
				},
			},
		},
	}
	clientStruct.Decs.Start.Append("// WorkflowClient orchestrates multi-step Arazzo 1.1.0 workflows across linked APIs.")
	decls = append(decls, clientStruct)

	newClientFunc := &dst.FuncDecl{
		Name: dst.NewIdent("NewWorkflowClient"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("baseURL")}, Type: dst.NewIdent("string")}},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{{Type: &dst.StarExpr{X: dst.NewIdent("WorkflowClient")}}},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.UnaryExpr{
							Op: token.AND,
							X: &dst.CompositeLit{
								Type: dst.NewIdent("WorkflowClient"),
								Elts: []dst.Expr{
									&dst.KeyValueExpr{Key: dst.NewIdent("BaseURL"), Value: dst.NewIdent("baseURL")},
									&dst.KeyValueExpr{
										Key: dst.NewIdent("HTTPClient"),
										Value: &dst.UnaryExpr{
											Op: token.AND,
											X:  &dst.CompositeLit{Type: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("Client")}},
										},
									},
									&dst.KeyValueExpr{Key: dst.NewIdent("SourceURLs"), Value: dst.NewIdent("SourceDescriptionsMetadata")},
								},
							},
						},
					},
				},
			},
		},
	}
	newClientFunc.Decs.Start.Append("// NewWorkflowClient constructs a WorkflowClient initialized with the given baseURL. Parameter baseURL is the target server URL.")
	decls = append(decls, newClientFunc)

	// Emit per-workflow Input struct, Output struct, and Execute<WorkflowPascal> method
	for _, wf := range doc.Workflows {
		wfPascal := naming.ToPascalCase(wf.WorkflowID)
		inputStructName := wfPascal + "Input"
		outputStructName := wfPascal + "Output"

		inputSchema := wf.Inputs
		if inputSchema == nil {
			inputSchema = &openapi.Schema{Type: "object"}
		}
		inTS, err := classes.EmitTypeWithOptions(inputStructName, inputSchema, classes.EmitOptions{IncludeGormTags: false, OmitEmpty: true, UsePascalCaseFieldNames: true})
		if err != nil {
			return nil, err
		}
		inDecl := &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{inTS}}
		inDecl.Decs.Start.Append(fmt.Sprintf("// %s holds the typed input parameters for workflow %s.", inputStructName, wf.WorkflowID))
		decls = append(decls, inDecl)

		outFields := &dst.FieldList{}
		outKeys := make([]string, 0, len(wf.Outputs))
		for k := range wf.Outputs {
			outKeys = append(outKeys, k)
		}
		sort.Strings(outKeys)
		for _, k := range outKeys {
			ov := wf.Outputs[k]
			tagStr := fmt.Sprintf("`json:%q`", k)
			if ov.Selector != nil {
				tagStr = fmt.Sprintf("`json:%q arazzo_context:%q`", k, ov.Selector.Context)
			}
			field := &dst.Field{
				Names: []*dst.Ident{dst.NewIdent(naming.ToPascalCase(k))},
				Type:  dst.NewIdent("string"),
				Tag:   &dst.BasicLit{Kind: token.STRING, Value: tagStr},
			}
			field.Decs.Start.Append(fmt.Sprintf("// %s holds the extracted %s workflow output.", naming.ToPascalCase(k), k))
			outFields.List = append(outFields.List, field)
		}
		outDecl := &dst.GenDecl{
			Tok: token.TYPE,
			Specs: []dst.Spec{
				&dst.TypeSpec{
					Name: dst.NewIdent(outputStructName),
					Type: &dst.StructType{Fields: outFields},
				},
			},
		}
		outDecl.Decs.Start.Append(fmt.Sprintf("// %s holds the typed output values for workflow %s.", outputStructName, wf.WorkflowID))
		decls = append(decls, outDecl)

		execMethod := emitWorkflowExecMethod(wf, wfPascal, inputStructName, outputStructName)
		decls = append(decls, execMethod)
	}

	f := &dst.File{
		Name:  dst.NewIdent("workflows"),
		Decls: decls,
	}
	f.Decs.Start.Append("// Package workflows provides strongly-typed Arazzo 1.1.0 Client SDK workflow orchestration.")
	return f, nil
}

// emitWorkflowExecMethod builds the Execute<WorkflowPascal> method AST for a single Arazzo Workflow.
// Parameter wf is the Workflow definition, wfPascal is its PascalCase name, inputStructName is the Input struct name, and outputStructName is the Output struct name.
func emitWorkflowExecMethod(wf arazzo.Workflow, wfPascal, inputStructName, outputStructName string) *dst.FuncDecl {
	bodyStmts := []dst.Stmt{
		&dst.AssignStmt{
			Lhs: []dst.Expr{dst.NewIdent("stepOutputs")},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.CallExpr{
					Fun: dst.NewIdent("make"),
					Args: []dst.Expr{
						&dst.MapType{
							Key:   dst.NewIdent("string"),
							Value: &dst.MapType{Key: dst.NewIdent("string"), Value: dst.NewIdent("string")},
						},
					},
				},
			},
		},
	}

	// Emit prerequisite DependsOn calls
	for _, dep := range wf.DependsOn {
		depPascal := naming.ToPascalCase(dep)
		depCall := &dst.IfStmt{
			Init: &dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("_"), dst.NewIdent("err")},
				Tok: token.DEFINE,
				Rhs: []dst.Expr{
					&dst.CallExpr{
						Fun: &dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent("Execute" + depPascal)},
						Args: []dst.Expr{
							dst.NewIdent("ctx"),
							&dst.CompositeLit{Type: dst.NewIdent(depPascal + "Input")},
						},
					},
				},
			},
			Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
			Body: &dst.BlockStmt{
				List: []dst.Stmt{
					&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("nil"), dst.NewIdent("err")}},
				},
			},
		}
		depCall.Decs.Start.Append(fmt.Sprintf("// PrerequisiteWorkflow: %s", dep))
		bodyStmts = append(bodyStmts, depCall)
	}

	// Emit each step block with structured AST statements and HTTP dispatch
	for _, step := range wf.Steps {
		stepBlock := &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.AssignStmt{
					Lhs: []dst.Expr{
						&dst.IndexExpr{X: dst.NewIdent("stepOutputs"), Index: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(step.StepID)}},
					},
					Tok: token.ASSIGN,
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun:  dst.NewIdent("make"),
							Args: []dst.Expr{&dst.MapType{Key: dst.NewIdent("string"), Value: dst.NewIdent("string")}},
						},
					},
				},
			},
		}

		ctxIdent := "ctx"
		if step.Timeout != nil && *step.Timeout > 0 {
			ctxIdent = "stepCtx"
			stepBlock.List = append(stepBlock.List,
				&dst.AssignStmt{
					Lhs: []dst.Expr{dst.NewIdent("stepCtx"), dst.NewIdent("cancel")},
					Tok: token.DEFINE,
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("WithTimeout")},
							Args: []dst.Expr{
								dst.NewIdent("ctx"),
								&dst.BinaryExpr{
									X:  &dst.BasicLit{Kind: token.INT, Value: strconv.FormatInt(*step.Timeout, 10)},
									Op: token.MUL,
									Y:  &dst.SelectorExpr{X: dst.NewIdent("time"), Sel: dst.NewIdent("Millisecond")},
								},
							},
						},
					},
				},
				&dst.DeferStmt{
					Call: &dst.CallExpr{Fun: dst.NewIdent("cancel")},
				},
			)
		}

		targetEndpoint := step.OperationID
		targetKind := "operationId"
		if targetEndpoint == "" && step.OperationPath != "" {
			targetEndpoint = step.OperationPath
			targetKind = "operationPath"
		}
		if targetEndpoint == "" && step.ChannelPath != "" {
			targetEndpoint = step.ChannelPath
			targetKind = "channelPath"
		}
		if targetEndpoint == "" && step.WorkflowID != "" {
			targetEndpoint = step.WorkflowID
			targetKind = "workflowId"
		}

		httpVerb := inferStepHTTPMethod(step)
		bodyReaderExpr := dst.Expr(dst.NewIdent("nil"))
		if step.RequestBody != nil && len(step.RequestBody.Payload) > 0 {
			bodyReaderExpr = &dst.CallExpr{
				Fun:  &dst.SelectorExpr{X: dst.NewIdent("strings"), Sel: dst.NewIdent("NewReader")},
				Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(string(step.RequestBody.Payload))}},
			}
		}

		reqStmts := []dst.Stmt{
			&dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("req"), dst.NewIdent("err")},
				Tok: token.DEFINE,
				Rhs: []dst.Expr{
					&dst.CallExpr{
						Fun: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("NewRequestWithContext")},
						Args: []dst.Expr{
							dst.NewIdent(ctxIdent),
							&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(httpVerb)},
							&dst.BinaryExpr{
								X:  &dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent("BaseURL")},
								Op: token.ADD,
								Y:  &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("/" + strings.TrimPrefix(targetEndpoint, "/"))},
							},
							bodyReaderExpr,
						},
					},
				},
			},
			&dst.IfStmt{
				Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
				Body: &dst.BlockStmt{
					List: []dst.Stmt{
						&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("nil"), dst.NewIdent("err")}},
					},
				},
			},
		}
		if step.RequestBody != nil && step.RequestBody.ContentType != "" {
			reqStmts = append(reqStmts, &dst.ExprStmt{
				X: &dst.CallExpr{
					Fun: &dst.SelectorExpr{
						X:   &dst.SelectorExpr{X: dst.NewIdent("req"), Sel: dst.NewIdent("Header")},
						Sel: dst.NewIdent("Set"),
					},
					Args: []dst.Expr{
						&dst.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
						&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(step.RequestBody.ContentType)},
					},
				},
			})
		}

		reqStmts = append(reqStmts,
			&dst.IfStmt{
				Cond: &dst.BinaryExpr{
					X:  &dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent("HTTPClient")},
					Op: token.NEQ,
					Y:  dst.NewIdent("nil"),
				},
				Body: &dst.BlockStmt{
					List: []dst.Stmt{
						&dst.AssignStmt{
							Lhs: []dst.Expr{dst.NewIdent("resp"), dst.NewIdent("errDo")},
							Tok: token.DEFINE,
							Rhs: []dst.Expr{
								&dst.CallExpr{
									Fun:  &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent("HTTPClient")}, Sel: dst.NewIdent("Do")},
									Args: []dst.Expr{dst.NewIdent("req")},
								},
							},
						},
						&dst.IfStmt{
							Cond: &dst.BinaryExpr{X: dst.NewIdent("errDo"), Op: token.NEQ, Y: dst.NewIdent("nil")},
							Body: &dst.BlockStmt{
								List: []dst.Stmt{
									&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("nil"), dst.NewIdent("errDo")}},
								},
							},
						},
						&dst.IfStmt{
							Cond: &dst.BinaryExpr{X: dst.NewIdent("resp"), Op: token.NEQ, Y: dst.NewIdent("nil")},
							Body: &dst.BlockStmt{
								List: []dst.Stmt{
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
												&dst.ReturnStmt{
													Results: []dst.Expr{
														dst.NewIdent("nil"),
														&dst.CallExpr{
															Fun: &dst.SelectorExpr{X: dst.NewIdent("fmt"), Sel: dst.NewIdent("Errorf")},
															Args: []dst.Expr{
																&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("step " + step.StepID + " failed with status %d")},
																&dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
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
				},
			},
		)

		stepBlock.List = append(stepBlock.List, reqStmts...)

		stepOutKeys := make([]string, 0, len(step.Outputs))
		for k := range step.Outputs {
			stepOutKeys = append(stepOutKeys, k)
		}
		sort.Strings(stepOutKeys)
		for _, k := range stepOutKeys {
			ov := step.Outputs[k]
			valStr := ov.Expression
			if ov.Selector != nil {
				valStr = ov.Selector.Context + "#" + ov.Selector.Selector
			}
			stepBlock.List = append(stepBlock.List, &dst.AssignStmt{
				Lhs: []dst.Expr{
					&dst.IndexExpr{
						X:     &dst.IndexExpr{X: dst.NewIdent("stepOutputs"), Index: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(step.StepID)}},
						Index: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(k)},
					},
				},
				Tok: token.ASSIGN,
				Rhs: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(valStr)}},
			})
		}

		stepBlock.Decs.Start.Append(fmt.Sprintf("// StepID: %s (%s=%s)", step.StepID, targetKind, targetEndpoint))
		bodyStmts = append(bodyStmts, stepBlock)
	}

	// Build final output composite literal
	outElts := []dst.Expr{}
	outKeys := make([]string, 0, len(wf.Outputs))
	for k := range wf.Outputs {
		outKeys = append(outKeys, k)
	}
	sort.Strings(outKeys)
	for _, k := range outKeys {
		ov := wf.Outputs[k]
		exprVal := ov.Expression
		if ov.Selector != nil {
			exprVal = ov.Selector.Selector
		}
		outElts = append(outElts, &dst.KeyValueExpr{
			Key:   dst.NewIdent(naming.ToPascalCase(k)),
			Value: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(exprVal)},
		})
	}

	bodyStmts = append(bodyStmts,
		&dst.IfStmt{
			Cond: &dst.BinaryExpr{
				X:  &dst.CallExpr{Fun: dst.NewIdent("len"), Args: []dst.Expr{dst.NewIdent("stepOutputs")}},
				Op: token.LSS,
				Y:  &dst.BasicLit{Kind: token.INT, Value: "0"},
			},
			Body: &dst.BlockStmt{
				List: []dst.Stmt{
					&dst.ReturnStmt{
						Results: []dst.Expr{
							dst.NewIdent("nil"),
							&dst.CallExpr{
								Fun:  &dst.SelectorExpr{X: dst.NewIdent("fmt"), Sel: dst.NewIdent("Errorf")},
								Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"invalid input: %+v"`}, dst.NewIdent("input")},
							},
						},
					},
				},
			},
		},
		&dst.ReturnStmt{
			Results: []dst.Expr{
				&dst.UnaryExpr{
					Op: token.AND,
					X: &dst.CompositeLit{
						Type: dst.NewIdent(outputStructName),
						Elts: outElts,
					},
				},
				dst.NewIdent("nil"),
			},
		},
	)

	fn := &dst.FuncDecl{
		Recv: &dst.FieldList{
			List: []*dst.Field{
				{Names: []*dst.Ident{dst.NewIdent("w")}, Type: &dst.StarExpr{X: dst.NewIdent("WorkflowClient")}},
			},
		},
		Name: dst.NewIdent("Execute" + wfPascal),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
					{Names: []*dst.Ident{dst.NewIdent("input")}, Type: dst.NewIdent(inputStructName)},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: &dst.StarExpr{X: dst.NewIdent(outputStructName)}},
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Body: &dst.BlockStmt{List: bodyStmts},
	}

	fn.Decs.Start.Append(fmt.Sprintf("// Execute%s executes the %s Arazzo workflow. Parameter ctx is the execution context and input holds workflow inputs.", wfPascal, wf.WorkflowID))
	fn.Decs.Start.Append(fmt.Sprintf("// WorkflowID: %s", wf.WorkflowID))
	if wf.Summary != "" {
		fn.Decs.Start.Append(fmt.Sprintf("// Summary: %s", wf.Summary))
	}
	if wf.Description != "" {
		fn.Decs.Start.Append(fmt.Sprintf("// Description: %s", wf.Description))
	}
	if len(wf.DependsOn) > 0 {
		fn.Decs.Start.Append(fmt.Sprintf("// DependsOn: %s", strings.Join(wf.DependsOn, ",")))
	}
	if len(wf.Parameters) > 0 {
		pB, _ := json.Marshal(wf.Parameters)
		fn.Decs.Start.Append(fmt.Sprintf("// WorkflowParametersJSON: %s", string(pB)))
	}
	if len(wf.SuccessActions) > 0 {
		saB, _ := json.Marshal(wf.SuccessActions)
		fn.Decs.Start.Append(fmt.Sprintf("// WorkflowSuccessActionsJSON: %s", string(saB)))
	}
	if len(wf.FailureActions) > 0 {
		faB, _ := json.Marshal(wf.FailureActions)
		fn.Decs.Start.Append(fmt.Sprintf("// WorkflowFailureActionsJSON: %s", string(faB)))
	}
	if len(wf.Extensions) > 0 {
		extB, _ := json.Marshal(wf.Extensions)
		fn.Decs.Start.Append(fmt.Sprintf("// WorkflowExtensionsJSON: %s", string(extB)))
	}

	return fn
}
