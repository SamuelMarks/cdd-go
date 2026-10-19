// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
	"go/token"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/dave/dst"
)

func TestEmitAndParseServerWorkflows(t *testing.T) {
	if _, err := EmitServerWorkflowTests(nil); err == nil {
		t.Error("expected error on nil doc")
	}
	if _, err := EmitServerWorkflowTests(&arazzo.Arazzo{Arazzo: "bad"}); err == nil {
		t.Error("expected validation error")
	}
	if _, err := EmitServerWorkflowRoutes(nil); err == nil {
		t.Error("expected error on nil doc")
	}
	if _, err := EmitServerWorkflowRoutes(&arazzo.Arazzo{Arazzo: "bad"}); err == nil {
		t.Error("expected validation error")
	}

	doc := sampleFullArazzoDoc()
	testFile, err := EmitServerWorkflowTests(doc)
	if err != nil {
		t.Fatalf("EmitServerWorkflowTests failed: %v", err)
	}

	wfsFromTests, err := ParseServerWorkflows(testFile)
	if err != nil {
		t.Fatalf("ParseServerWorkflows from tests failed: %v", err)
	}
	if len(wfsFromTests) != 2 {
		t.Fatalf("expected 2 workflows from tests, got %d", len(wfsFromTests))
	}

	routesFile, err := EmitServerWorkflowRoutes(doc)
	if err != nil {
		t.Fatalf("EmitServerWorkflowRoutes failed: %v", err)
	}

	wfsFromRoutes, err := ParseServerWorkflows(routesFile)
	if err != nil {
		t.Fatalf("ParseServerWorkflows from routes failed: %v", err)
	}
	if len(wfsFromRoutes) != 2 {
		t.Fatalf("expected 2 workflows from routes, got %d", len(wfsFromRoutes))
	}

	if _, err := ParseServerWorkflows(nil); err == nil {
		t.Error("expected error on nil file")
	}

	// Test fallback branches in ParseServerWorkflows when WorkflowJSON is absent
	manualServerFile := &dst.File{
		Name: dst.NewIdent("tests"),
		Decls: []dst.Decl{
			&dst.GenDecl{Tok: token.VAR},
			&dst.FuncDecl{
				Name: dst.NewIdent("TestWorkflowManual"),
				Decs: dst.FuncDeclDecorations{
					NodeDecs: dst.NodeDecs{
						Start: dst.Decorations{"// WorkflowID: manualWf"},
					},
				},
				Body: &dst.BlockStmt{
					List: []dst.Stmt{
						&dst.BlockStmt{
							Decs: dst.BlockStmtDecorations{
								NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// StepID: stepOne"}},
							},
						},
					},
				},
			},
			&dst.FuncDecl{
				Name: dst.NewIdent("TestWorkflowEmpty"),
			},
			&dst.FuncDecl{
				Name: dst.NewIdent("RegisterArazzoWorkflowRoutes"),
				Body: &dst.BlockStmt{
					List: []dst.Stmt{
						&dst.ReturnStmt{},
						&dst.ExprStmt{
							X: dst.NewIdent("noop"),
						},
						&dst.ExprStmt{
							X: dst.NewIdent("routeCall"),
							Decs: dst.ExprStmtDecorations{
								NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// WorkflowID: routeOnlyWf"}},
							},
						},
						&dst.ExprStmt{
							X: &dst.CallExpr{
								Fun:  &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("POST")},
								Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"/workflows/astRouteWf"`}},
							},
						},
					},
				},
			},
		},
	}

	manualWfs, err := ParseServerWorkflows(manualServerFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manualWfs) != 4 {
		t.Fatalf("expected 4 manual workflows, got %d", len(manualWfs))
	}
}
