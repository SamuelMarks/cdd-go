// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
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
	wfPascal := strings.TrimPrefix(fn.Name.Name, "TestWorkflow")
	wf := arazzo.Workflow{
		WorkflowID: strings.ToLower(wfPascal[:1]) + wfPascal[1:],
	}
	for _, c := range fn.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "WorkflowID:") {
			wf.WorkflowID = strings.TrimSpace(strings.TrimPrefix(line, "WorkflowID:"))
		} else if strings.HasPrefix(line, "DependsOn:") {
			csv := strings.TrimSpace(strings.TrimPrefix(line, "DependsOn:"))
			if csv != "" {
				wf.DependsOn = strings.Split(csv, ",")
			}
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
		if strings.HasPrefix(line, "WorkflowID:") {
			wf.WorkflowID = strings.TrimSpace(strings.TrimPrefix(line, "WorkflowID:"))
		} else if strings.HasPrefix(line, "Step:") {
			rest := strings.TrimSpace(strings.TrimPrefix(line, "Step:"))
			step := arazzo.Step{}
			if openIdx := strings.Index(rest, "("); openIdx != -1 && strings.HasSuffix(rest, ")") {
				step.StepID = strings.TrimSpace(rest[:openIdx])
				meta := rest[openIdx+1 : len(rest)-1]
				if kv := strings.SplitN(meta, "=", 2); len(kv) == 2 {
					switch kv[0] {
					case "operationPath":
						step.OperationPath = kv[1]
					case "channelPath":
						step.ChannelPath = kv[1]
					case "workflowId":
						step.WorkflowID = kv[1]
					default:
						step.OperationID = kv[1]
					}
				}
			}
			if step.StepID != "" {
				wf.Steps = append(wf.Steps, step)
			}
		}
	}
	if wf.WorkflowID == "" {
		if call, ok := es.X.(*dst.CallExpr); ok && len(call.Args) > 0 {
			if lit, ok := call.Args[0].(*dst.BasicLit); ok {
				rawPath := strings.Trim(lit.Value, `"`)
				if strings.HasPrefix(rawPath, "/workflows/") {
					wf.WorkflowID = strings.TrimPrefix(rawPath, "/workflows/")
				}
			}
		}
	}
	if wf.WorkflowID != "" {
		if len(wf.Steps) == 0 {
			wf.Steps = []arazzo.Step{{StepID: "routeStep", OperationID: wf.WorkflowID}}
		}
		return wf, true
	}
	return arazzo.Workflow{}, false
}
