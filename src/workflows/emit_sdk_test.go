// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
	"encoding/json"
	"go/token"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

func sampleFullArazzoDoc() *arazzo.Arazzo {
	timeout := int64(2500)
	return &arazzo.Arazzo{
		Arazzo: "1.1.0",
		Self:   "https://example.com/arazzo.json",
		Info: arazzo.Info{
			Title:       "Sample Suite",
			Summary:     "Suite Summary",
			Description: "Suite Description",
			Version:     "1.2.3",
			Extensions:  map[string]json.RawMessage{"x-info": json.RawMessage(`"info"`)}},
		SourceDescriptions: []arazzo.SourceDescription{
			{Name: "api", URL: "./openapi.json", Type: arazzo.SourceTypeOpenAPI},
		},
		Components: &arazzo.Components{
			Parameters: map[string]arazzo.Parameter{
				"trace": {Name: "X-Trace", In: arazzo.ParamInHeader, Value: arazzo.NewStringValue("$inputs.trace")},
			},
			SuccessActions: map[string]arazzo.SuccessAction{
				"endOk": {Name: "endOk", Type: arazzo.SuccessTypeEnd},
			},
			FailureActions: map[string]arazzo.FailureAction{
				"endFail": {Name: "endFail", Type: arazzo.FailureTypeEnd},
			},
			Extensions: map[string]json.RawMessage{"x-comp": json.RawMessage(`true`)},
		},
		Workflows: []arazzo.Workflow{
			{
				WorkflowID:  "authFlow",
				Summary:     "Auth summary",
				Description: "Auth description",
				Inputs: &openapi.Schema{
					Type: "object",
					Properties: map[string]openapi.Schema{
						"username": {Type: "string"},
					},
				},
				Parameters: []arazzo.ParameterOrReusable{
					{Reusable: &arazzo.ReusableObject{Reference: "$components.parameters.trace"}},
				},
				SuccessActions: []arazzo.SuccessActionOrReusable{
					{Reusable: &arazzo.ReusableObject{Reference: "$components.successActions.endOk"}},
				},
				FailureActions: []arazzo.FailureActionOrReusable{
					{Reusable: &arazzo.ReusableObject{Reference: "$components.failureActions.endFail"}},
				},
				Outputs: map[string]arazzo.OutputValue{
					"token": {Expression: "$steps.login.outputs.jwt"},
					"sel":   {Selector: &arazzo.Selector{Context: "$response.body", Selector: "$.id"}},
				},
				Steps: []arazzo.Step{
					{
						StepID:      "login",
						OperationID: "loginOp",
						Timeout:     &timeout,
						RequestBody: &arazzo.RequestBody{
							ContentType: "application/json",
							Payload:     json.RawMessage(`{"u":"a"}`),
						},
						Outputs: map[string]arazzo.OutputValue{
							"jwt":    {Expression: "$response.body#/jwt"},
							"selOut": {Selector: &arazzo.Selector{Context: "$response.body", Selector: "$.jwt"}},
						},
					},
					{
						StepID:        "byPath",
						OperationPath: "{$sourceDescriptions.api.url}#/paths/~1me/get",
					},
					{
						StepID:      "byChannel",
						ChannelPath: "{$sourceDescriptions.api.url}#/channels/events",
					},
					{
						StepID:      "getStep",
						OperationID: "getProfile",
					},
					{
						StepID:      "deleteStep",
						OperationID: "deleteProfile",
					},
					{
						StepID:      "putStep",
						OperationID: "putProfile",
					},
					{
						StepID:      "patchStep",
						OperationID: "patchProfile",
					},
				},
				Extensions: map[string]json.RawMessage{"x-wf": json.RawMessage(`1`)},
			},
			{
				WorkflowID: "subCaller",
				DependsOn:  []string{"authFlow"},
				Steps: []arazzo.Step{
					{
						StepID:     "callAuth",
						WorkflowID: "authFlow",
					},
				},
			},
		},
		Extensions: map[string]json.RawMessage{"x-root": json.RawMessage(`"r"`)},
	}
}

func TestEmitAndParseSDKWorkflows(t *testing.T) {
	if _, err := EmitSDKWorkflows(nil); err == nil {
		t.Error("expected error on nil doc")
	}

	invalidDoc := &arazzo.Arazzo{Arazzo: "bad"}
	if _, err := EmitSDKWorkflows(invalidDoc); err == nil {
		t.Error("expected validation error")
	}

	badInputSchemaDoc := sampleFullArazzoDoc()
	badInputSchemaDoc.Workflows[0].Inputs = &openapi.Schema{Type: "unsupported-type"}
	if _, err := EmitSDKWorkflows(badInputSchemaDoc); err == nil {
		t.Error("expected schema emit error")
	}

	doc := sampleFullArazzoDoc()
	file, err := EmitSDKWorkflows(doc)
	if err != nil {
		t.Fatalf("EmitSDKWorkflows failed: %v", err)
	}

	parsed, err := ParseSDKWorkflows(file)
	if err != nil {
		t.Fatalf("ParseSDKWorkflows failed: %v", err)
	}

	if err := arazzo.Validate(parsed); err != nil {
		t.Fatalf("parsed doc failed validation: %v", err)
	}
	if parsed.Self != doc.Self {
		t.Errorf("expected self %q, got %q", doc.Self, parsed.Self)
	}
	if len(parsed.Workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(parsed.Workflows))
	}
	if parsed.Workflows[0].WorkflowID != "authFlow" {
		t.Errorf("expected authFlow, got %q", parsed.Workflows[0].WorkflowID)
	}

	if _, err := ParseSDKWorkflows(nil); err == nil {
		t.Error("expected error on nil file")
	}

	// Test fallback branches in ParseSDKWorkflows when comments are absent
	manualFile := &dst.File{
		Name: dst.NewIdent("workflows"),
		Decls: []dst.Decl{
			&dst.GenDecl{
				Tok: token.VAR,
				Specs: []dst.Spec{
					&dst.ValueSpec{},
					&dst.ValueSpec{
						Names: []*dst.Ident{dst.NewIdent("SourceDescriptionsMetadata")},
						Values: []dst.Expr{
							&dst.CompositeLit{
								Elts: []dst.Expr{
									&dst.KeyValueExpr{
										Key:   &dst.BasicLit{Kind: token.STRING, Value: `"manual"`},
										Value: &dst.BasicLit{Kind: token.STRING, Value: `"./manual.json"`},
									},
								},
							},
						},
					},
				},
			},
			&dst.GenDecl{
				Tok: token.TYPE,
				Specs: []dst.Spec{
					&dst.ValueSpec{},
					&dst.TypeSpec{
						Name: dst.NewIdent("ManualInput"),
						Type: &dst.StructType{
							Fields: &dst.FieldList{
								List: []*dst.Field{
									{
										Names: []*dst.Ident{dst.NewIdent("UserID")},
										Type:  dst.NewIdent("string"),
										Tag:   &dst.BasicLit{Kind: token.STRING, Value: "`" + `json:"userId"` + "`"},
									},
								},
							},
						},
					},
				},
			},
			&dst.FuncDecl{
				Recv: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("WorkflowClient")}}},
				Name: dst.NewIdent("ExecuteManual"),
				Type: &dst.FuncType{},
				Body: &dst.BlockStmt{
					List: []dst.Stmt{
						&dst.IfStmt{
							Init: &dst.AssignStmt{
								Lhs: []dst.Expr{dst.NewIdent("_"), dst.NewIdent("err")},
								Tok: token.DEFINE,
								Rhs: []dst.Expr{
									&dst.CallExpr{
										Fun: &dst.SelectorExpr{X: dst.NewIdent("w"), Sel: dst.NewIdent("ExecuteAuthPrereq")},
									},
								},
							},
							Cond: dst.NewIdent("true"),
							Body: &dst.BlockStmt{},
						},
						&dst.BlockStmt{
							List: []dst.Stmt{
								&dst.ExprStmt{X: dst.NewIdent("noop")},
								&dst.AssignStmt{
									Lhs: []dst.Expr{
										&dst.IndexExpr{
											X:     dst.NewIdent("stepOutputs"),
											Index: &dst.BasicLit{Kind: token.STRING, Value: `"astStep"`},
										},
									},
									Tok: token.ASSIGN,
									Rhs: []dst.Expr{dst.NewIdent("nil")},
								},
								&dst.AssignStmt{
									Lhs: []dst.Expr{dst.NewIdent("req"), dst.NewIdent("err")},
									Tok: token.DEFINE,
									Rhs: []dst.Expr{
										&dst.CallExpr{
											Fun: dst.NewIdent("NewRequestWithContext"),
											Args: []dst.Expr{
												dst.NewIdent("ctx"),
												&dst.BasicLit{Kind: token.STRING, Value: `"POST"`},
												&dst.BinaryExpr{
													X:  dst.NewIdent("baseURL"),
													Op: token.ADD,
													Y:  &dst.BasicLit{Kind: token.STRING, Value: `"/astOp"`},
												},
											},
										},
									},
								},
							},
						},
						&dst.BlockStmt{
							Decs: dst.BlockStmtDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// StepID: manualStep"}}},
							List: []dst.Stmt{
								&dst.AssignStmt{
									Lhs: []dst.Expr{dst.NewIdent("req"), dst.NewIdent("err")},
									Tok: token.DEFINE,
									Rhs: []dst.Expr{
										&dst.CallExpr{
											Fun: dst.NewIdent("NewRequest"),
											Args: []dst.Expr{
												&dst.BasicLit{Kind: token.STRING, Value: `"GET"`},
												&dst.BinaryExpr{
													X:  dst.NewIdent("baseURL"),
													Op: token.ADD,
													Y:  &dst.BasicLit{Kind: token.STRING, Value: `"/threeArgOp"`},
												},
												dst.NewIdent("nil"),
											},
										},
									},
								},
							},
						},
						&dst.BlockStmt{
							Decs: dst.BlockStmtDecorations{NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// StepID: fallbackStep"}}},
						},
						&dst.ReturnStmt{
							Results: []dst.Expr{
								&dst.UnaryExpr{
									Op: token.AND,
									X: &dst.CompositeLit{
										Elts: []dst.Expr{
											&dst.KeyValueExpr{
												Key:   dst.NewIdent("TokenVal"),
												Value: &dst.BasicLit{Kind: token.STRING, Value: `"$response.body#/token"`},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			&dst.FuncDecl{
				Recv: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("WorkflowClient")}}},
				Name: dst.NewIdent("ExecuteEmptyBody"),
				Type: &dst.FuncType{},
			},
		},
	}

	manualParsed, err := ParseSDKWorkflows(manualFile)
	if err != nil {
		t.Fatalf("unexpected error on manualFile: %v", err)
	}
	if len(manualParsed.Workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(manualParsed.Workflows))
	}

	emptyFile := &dst.File{Name: dst.NewIdent("workflows")}
	emptyParsed, err := ParseSDKWorkflows(emptyFile)
	if err != nil || len(emptyParsed.SourceDescriptions) != 1 {
		t.Errorf("expected default sourceDescription on empty file, got %v, %v", emptyParsed, err)
	}
}
