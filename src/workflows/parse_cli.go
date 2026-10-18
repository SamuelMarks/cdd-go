// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
	"encoding/json"
	"fmt"
	"go/token"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// ParseCLIWorkflows extracts Arazzo 1.1.0 workflows and Info metadata from a Cobra CLI workflow runner AST file (`sdk_workflow_cli.go`).
// Parameter file is the parsed dst.File of the CLI workflow file, and it returns the reconstructed Workflows, Info, and any error.
func ParseCLIWorkflows(file *dst.File) ([]arazzo.Workflow, *arazzo.Info, error) {
	if file == nil {
		return nil, nil, fmt.Errorf("dst.File is nil")
	}

	info := &arazzo.Info{
		Title:   "CLI Workflows",
		Version: "1.0.0",
	}
	var workflows []arazzo.Workflow

	for _, decl := range file.Decls {
		gd, ok := decl.(*dst.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*dst.ValueSpec)
			if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
				continue
			}
			varName := vs.Names[0].Name
			if varName == "WorkflowCmd" {
				parseCLIRootCommand(gd, vs, info)
			} else if strings.HasSuffix(varName, "WorkflowCmd") {
				wf := parseCLISubcommand(gd, vs, varName)
				workflows = append(workflows, wf)
			}
		}
	}

	return workflows, info, nil
}

// parseCLIRootCommand extracts arazzo.Info fields from the root WorkflowCmd Cobra command declaration.
// Parameter gd is the enclosing GenDecl, vs is the ValueSpec, and info is the target arazzo.Info struct.
func parseCLIRootCommand(gd *dst.GenDecl, vs *dst.ValueSpec, info *arazzo.Info) {
	for _, c := range gd.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "ArazzoTitle:") {
			info.Title = strings.TrimSpace(strings.TrimPrefix(line, "ArazzoTitle:"))
		}
	}
	unary, ok := vs.Values[0].(*dst.UnaryExpr)
	if !ok {
		return
	}
	cl, ok := unary.X.(*dst.CompositeLit)
	if !ok {
		return
	}
	for _, elt := range cl.Elts {
		kv, ok := elt.(*dst.KeyValueExpr)
		if !ok {
			continue
		}
		kIdent, kOk := kv.Key.(*dst.Ident)
		vLit, vOk := kv.Value.(*dst.BasicLit)
		if !kOk || !vOk || vLit.Kind != token.STRING {
			continue
		}
		val, _ := strconv.Unquote(vLit.Value)
		switch kIdent.Name {
		case "Short":
			info.Summary = val
		case "Long":
			info.Description = val
		case "Version":
			info.Version = val
		}
	}
}

// parseCLISubcommand reconstructs an arazzo.Workflow from a `<WorkflowPascal>WorkflowCmd` Cobra subcommand declaration.
// Parameter gd is the enclosing GenDecl, vs is the ValueSpec, and varName is the variable identifier.
func parseCLISubcommand(gd *dst.GenDecl, vs *dst.ValueSpec, varName string) arazzo.Workflow {
	for _, c := range gd.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "WorkflowJSON:") {
			raw := strings.TrimSpace(strings.TrimPrefix(line, "WorkflowJSON:"))
			var wf arazzo.Workflow
			if json.Unmarshal([]byte(raw), &wf) == nil && wf.WorkflowID != "" {
				return wf
			}
		}
	}

	wfPascal := strings.TrimSuffix(varName, "WorkflowCmd")
	wf := arazzo.Workflow{
		WorkflowID: strings.ToLower(wfPascal[:1]) + wfPascal[1:],
	}

	for _, c := range gd.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "WorkflowID:") {
			wf.WorkflowID = strings.TrimSpace(strings.TrimPrefix(line, "WorkflowID:"))
		} else if strings.HasPrefix(line, "DependsOn:") {
			csv := strings.TrimSpace(strings.TrimPrefix(line, "DependsOn:"))
			if csv != "" {
				wf.DependsOn = strings.Split(csv, ",")
			}
		} else if strings.HasPrefix(line, "FlagInput:") {
			raw := strings.TrimSpace(strings.TrimPrefix(line, "FlagInput:"))
			parts := strings.SplitN(raw, ":", 2)
			if len(parts) == 2 {
				if wf.Inputs == nil {
					wf.Inputs = &openapi.Schema{Type: "object", Properties: make(map[string]openapi.Schema)}
				}
				wf.Inputs.Properties[parts[0]] = openapi.Schema{Type: parts[1]}
			}
		}
	}

	if unary, ok := vs.Values[0].(*dst.UnaryExpr); ok {
		if cl, ok := unary.X.(*dst.CompositeLit); ok {
			for _, elt := range cl.Elts {
				if kv, ok := elt.(*dst.KeyValueExpr); ok {
					kIdent, kOk := kv.Key.(*dst.Ident)
					vLit, vOk := kv.Value.(*dst.BasicLit)
					if kOk && vOk && vLit.Kind == token.STRING {
						val, _ := strconv.Unquote(vLit.Value)
						switch kIdent.Name {
						case "Short":
							wf.Summary = val
						case "Long":
							wf.Description = val
						}
					}
				}
			}
		}
	}

	if len(wf.Steps) == 0 {
		wf.Steps = []arazzo.Step{
			{StepID: "executeStep", OperationID: wf.WorkflowID},
		}
	}

	return wf
}
