// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
	"go/token"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/dave/dst"
)

func TestEmitAndParseCLIWorkflows(t *testing.T) {
	if _, err := EmitCLIWorkflows(nil); err == nil {
		t.Error("expected error on nil doc")
	}
	if _, err := EmitCLIWorkflows(&arazzo.Arazzo{Arazzo: "bad"}); err == nil {
		t.Error("expected validation error")
	}

	doc := sampleFullArazzoDoc()
	doc.Info.Summary = ""
	doc.Info.Description = ""
	doc.Workflows[1].Summary = ""

	file, err := EmitCLIWorkflows(doc)
	if err != nil {
		t.Fatalf("EmitCLIWorkflows failed: %v", err)
	}

	wfs, info, err := ParseCLIWorkflows(file)
	if err != nil {
		t.Fatalf("ParseCLIWorkflows failed: %v", err)
	}
	if info.Title != doc.Info.Title {
		t.Errorf("expected title %q, got %q", doc.Info.Title, info.Title)
	}
	if len(wfs) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(wfs))
	}

	if _, _, err := ParseCLIWorkflows(nil); err == nil {
		t.Error("expected error on nil file")
	}

	// Test manual Cobra CLI AST without WorkflowJSON annotation
	manualCLI := &dst.File{
		Name: dst.NewIdent("main"),
		Decls: []dst.Decl{
			&dst.FuncDecl{Name: dst.NewIdent("ignoreFunc")},
			&dst.GenDecl{
				Tok: token.VAR,
				Specs: []dst.Spec{
					&dst.TypeSpec{Name: dst.NewIdent("IgnoreType")},
					&dst.ValueSpec{
						Names:  []*dst.Ident{dst.NewIdent("WorkflowCmd")},
						Values: []dst.Expr{dst.NewIdent("notUnary")},
					},
					&dst.ValueSpec{
						Names: []*dst.Ident{dst.NewIdent("WorkflowCmd")},
						Values: []dst.Expr{
							&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("notComposite")},
						},
					},
					&dst.ValueSpec{
						Names: []*dst.Ident{dst.NewIdent("WorkflowCmd")},
						Values: []dst.Expr{
							&dst.UnaryExpr{
								Op: token.AND,
								X: &dst.CompositeLit{
									Elts: []dst.Expr{
										dst.NewIdent("notKV"),
										&dst.KeyValueExpr{Key: dst.NewIdent("Short"), Value: dst.NewIdent("notLit")},
										&dst.KeyValueExpr{Key: dst.NewIdent("Short"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"S"`}},
										&dst.KeyValueExpr{Key: dst.NewIdent("Long"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"L"`}},
										&dst.KeyValueExpr{Key: dst.NewIdent("Version"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"2.0.0"`}},
									},
								},
							},
						},
					},
				},
			},
			&dst.GenDecl{
				Tok: token.VAR,
				Decs: dst.GenDeclDecorations{
					NodeDecs: dst.NodeDecs{
						Start: dst.Decorations{
							"// WorkflowID: customWf",
							"// DependsOn: dep1,dep2",
							"// FlagInput: userId:string",
						},
					},
				},
				Specs: []dst.Spec{
					&dst.ValueSpec{
						Names: []*dst.Ident{dst.NewIdent("CustomWorkflowCmd")},
						Values: []dst.Expr{
							&dst.UnaryExpr{
								Op: token.AND,
								X: &dst.CompositeLit{
									Elts: []dst.Expr{
										&dst.KeyValueExpr{Key: dst.NewIdent("Short"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"Custom Short"`}},
										&dst.KeyValueExpr{Key: dst.NewIdent("Long"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"Custom Long"`}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	manualWfs, manualInfo, err := ParseCLIWorkflows(manualCLI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manualInfo.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %q", manualInfo.Version)
	}
	if len(manualWfs) != 1 || manualWfs[0].WorkflowID != "customWf" {
		t.Errorf("unexpected manual workflows: %+v", manualWfs)
	}
}
