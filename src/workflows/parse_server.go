// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/dave/dst"
)

// ParseServerWorkflows extracts Arazzo 1.1.0 Workflow definitions from a server E2E test file (`tests/workflows_test.go`) or workflow route file (`routes/workflows_routes.go`).
// Parameter file is the parsed dst.File, and it returns the slice of reconstructed Workflows and any error.
func ParseServerWorkflows(file *dst.File) ([]arazzo.Workflow, error) {
	if file == nil {
		return nil, fmt.Errorf("dst.File is nil")
	}

	var workflows []arazzo.Workflow

	for _, decl := range file.Decls {
		fn, ok := decl.(*dst.FuncDecl)
		if !ok {
			continue
		}
		if strings.HasPrefix(fn.Name.Name, "TestWorkflow") {
			wf := parseWorkflowTestFunc(fn)
			workflows = append(workflows, wf)
		} else if fn.Name.Name == "RegisterArazzoWorkflowRoutes" && fn.Body != nil {
			for _, stmt := range fn.Body.List {
				if wf, ok := parseWorkflowRouteStmt(stmt); ok {
					workflows = append(workflows, wf)
				}
			}
		}
	}

	return workflows, nil
}

// parseWorkflowTestFunc extracts an arazzo.Workflow from a `TestWorkflow<Name>` AST function declaration.
// Parameter fn is the FuncDecl AST node, and it returns the reconstructed arazzo.Workflow.
func parseWorkflowTestFunc(fn *dst.FuncDecl) arazzo.Workflow {
	for _, c := range fn.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "WorkflowJSON:") {
			raw := strings.TrimSpace(strings.TrimPrefix(line, "WorkflowJSON:"))
			var wf arazzo.Workflow
			if json.Unmarshal([]byte(raw), &wf) == nil && wf.WorkflowID != "" {
				return wf
			}
		}
	}

	wfPascal := strings.TrimPrefix(fn.Name.Name, "TestWorkflow")
	wf := arazzo.Workflow{
		WorkflowID: strings.ToLower(wfPascal[:1]) + wfPascal[1:],
	}
	for _, c := range fn.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "WorkflowID:") {
			wf.WorkflowID = strings.TrimSpace(strings.TrimPrefix(line, "WorkflowID:"))
		}
	}

	if fn.Body != nil {
		for _, stmt := range fn.Body.List {
			if block, ok := stmt.(*dst.BlockStmt); ok {
				step := parseSDKStepBlock(block)
				if step.StepID != "" {
					wf.Steps = append(wf.Steps, step)
				}
			}
		}
	}

	if len(wf.Steps) == 0 {
		wf.Steps = []arazzo.Step{
			{StepID: "testStep", OperationID: wf.WorkflowID},
		}
	}
	return wf
}

// parseWorkflowRouteStmt extracts an arazzo.Workflow from a route registration statement inside RegisterArazzoWorkflowRoutes.
// Parameter stmt is the AST Stmt, and it returns the parsed Workflow and true if found.
func parseWorkflowRouteStmt(stmt dst.Stmt) (arazzo.Workflow, bool) {
	es, ok := stmt.(*dst.ExprStmt)
	if !ok {
		return arazzo.Workflow{}, false
	}
	var wf arazzo.Workflow
	for _, c := range es.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "WorkflowJSON:") {
			raw := strings.TrimSpace(strings.TrimPrefix(line, "WorkflowJSON:"))
			if json.Unmarshal([]byte(raw), &wf) == nil && wf.WorkflowID != "" {
				return wf, true
			}
		} else if strings.HasPrefix(line, "WorkflowID:") {
			wf.WorkflowID = strings.TrimSpace(strings.TrimPrefix(line, "WorkflowID:"))
		}
	}
	if wf.WorkflowID != "" {
		wf.Steps = []arazzo.Step{{StepID: "routeStep", OperationID: wf.WorkflowID}}
		return wf, true
	}
	return arazzo.Workflow{}, false
}
