// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
	"encoding/json"
	"fmt"
	"go/token"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/SamuelMarks/cdd-go/src/classes"
	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// ParseSDKWorkflows extracts an Arazzo 1.1.0 document from a Go AST File (`package workflows`).
// Parameter file is the parsed dst.File of the SDK workflows package, and it returns the reconstructed *arazzo.Arazzo specification and any error.
func ParseSDKWorkflows(file *dst.File) (*arazzo.Arazzo, error) {
	if file == nil {
		return nil, fmt.Errorf("dst.File is nil")
	}

	doc := &arazzo.Arazzo{
		Arazzo: "1.1.0",
		Info: arazzo.Info{
			Title:   "Extracted Workflows",
			Version: "1.0.0",
		},
	}

	inputSchemas := make(map[string]*openapi.Schema)
	outputMaps := make(map[string]map[string]arazzo.OutputValue)

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *dst.GenDecl:
			parseSDKGenDecl(d, doc, inputSchemas, outputMaps)
		case *dst.FuncDecl:
			if strings.HasPrefix(d.Name.Name, "Execute") && d.Recv != nil {
				wf := parseSDKExecuteMethod(d, inputSchemas, outputMaps)
				doc.Workflows = append(doc.Workflows, wf)
			}
		}
	}

	if len(doc.SourceDescriptions) == 0 {
		doc.SourceDescriptions = []arazzo.SourceDescription{
			{Name: "default", URL: "./openapi.json", Type: arazzo.SourceTypeOpenAPI},
		}
	}

	return doc, nil
}

// parseSDKGenDecl inspects a top-level GenDecl in the SDK workflows file for metadata variables and Input/Output structs.
// Parameter d is the GenDecl, doc is the target Arazzo document, inputSchemas stores parsed input schemas, and outputMaps stores parsed workflow outputs.
func parseSDKGenDecl(d *dst.GenDecl, doc *arazzo.Arazzo, inputSchemas map[string]*openapi.Schema, outputMaps map[string]map[string]arazzo.OutputValue) {
	if d.Tok == token.VAR {
		for _, spec := range d.Specs {
			vs, ok := spec.(*dst.ValueSpec)
			if !ok || len(vs.Names) == 0 {
				continue
			}
			varName := vs.Names[0].Name
			switch varName {
			case "ArazzoVersion":
				if len(vs.Values) > 0 {
					if bl, ok := vs.Values[0].(*dst.BasicLit); ok && bl.Kind == token.STRING {
						if unq, err := strconv.Unquote(bl.Value); err == nil && unq != "" {
							doc.Arazzo = unq
						}
					}
				}
				for _, c := range d.Decs.Start {
					line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
					if strings.HasPrefix(line, "SelfURI:") {
						doc.Self = strings.TrimSpace(strings.TrimPrefix(line, "SelfURI:"))
					} else if strings.HasPrefix(line, "ArazzoInfoJSON:") {
						raw := strings.TrimSpace(strings.TrimPrefix(line, "ArazzoInfoJSON:"))
						_ = json.Unmarshal([]byte(raw), &doc.Info)
					} else if strings.HasPrefix(line, "RootExtensionsJSON:") {
						raw := strings.TrimSpace(strings.TrimPrefix(line, "RootExtensionsJSON:"))
						_ = json.Unmarshal([]byte(raw), &doc.Extensions)
					} else if strings.HasPrefix(line, "InfoExtensionsJSON:") {
						raw := strings.TrimSpace(strings.TrimPrefix(line, "InfoExtensionsJSON:"))
						_ = json.Unmarshal([]byte(raw), &doc.Info.Extensions)
					}
				}
			case "SourceDescriptionsMetadata":
				parsedFromDoc := false
				for _, c := range d.Decs.Start {
					line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
					if strings.HasPrefix(line, "SourceDescriptionsJSON:") {
						raw := strings.TrimSpace(strings.TrimPrefix(line, "SourceDescriptionsJSON:"))
						if json.Unmarshal([]byte(raw), &doc.SourceDescriptions) == nil {
							parsedFromDoc = true
						}
					}
				}
				if !parsedFromDoc && len(vs.Values) > 0 {
					if cl, ok := vs.Values[0].(*dst.CompositeLit); ok {
						for _, elt := range cl.Elts {
							if kv, ok := elt.(*dst.KeyValueExpr); ok {
								kLit, kOk := kv.Key.(*dst.BasicLit)
								vLit, vOk := kv.Value.(*dst.BasicLit)
								if kOk && vOk {
									kStr, _ := strconv.Unquote(kLit.Value)
									vStr, _ := strconv.Unquote(vLit.Value)
									doc.SourceDescriptions = append(doc.SourceDescriptions, arazzo.SourceDescription{
										Name: kStr,
										URL:  vStr,
										Type: arazzo.SourceTypeOpenAPI,
									})
								}
							}
						}
					}
				}
			case "ComponentsMetadataJSON":
				if len(vs.Values) > 0 {
					if bl, ok := vs.Values[0].(*dst.BasicLit); ok && bl.Kind == token.STRING {
						if unq, err := strconv.Unquote(bl.Value); err == nil {
							var comp arazzo.Components
							if json.Unmarshal([]byte(unq), &comp) == nil {
								doc.Components = &comp
							}
						}
					}
				}
				for _, c := range d.Decs.Start {
					line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
					if strings.HasPrefix(line, "ComponentsExtensionsJSON:") && doc.Components != nil {
						raw := strings.TrimSpace(strings.TrimPrefix(line, "ComponentsExtensionsJSON:"))
						_ = json.Unmarshal([]byte(raw), &doc.Components.Extensions)
					}
				}
			}
		}
	} else if d.Tok == token.TYPE {
		for _, spec := range d.Specs {
			ts, ok := spec.(*dst.TypeSpec)
			if !ok {
				continue
			}
			tName := ts.Name.Name
			if strings.HasSuffix(tName, "Input") {
				wfPascal := strings.TrimSuffix(tName, "Input")
				var schemaFromDoc *openapi.Schema
				for _, c := range d.Decs.Start {
					line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
					if strings.HasPrefix(line, "InputSchemaJSON:") {
						raw := strings.TrimSpace(strings.TrimPrefix(line, "InputSchemaJSON:"))
						var s openapi.Schema
						if json.Unmarshal([]byte(raw), &s) == nil {
							schemaFromDoc = &s
						}
					}
				}
				if schemaFromDoc != nil {
					inputSchemas[wfPascal] = schemaFromDoc
				} else if parsed, err := classes.ParseType(ts); err == nil && len(parsed.Properties) > 0 {
					parsed.Description = ""
					inputSchemas[wfPascal] = parsed
				}
			} else if strings.HasSuffix(tName, "Output") {
				wfPascal := strings.TrimSuffix(tName, "Output")
				for _, c := range d.Decs.Start {
					line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
					if strings.HasPrefix(line, "WorkflowOutputsJSON:") {
						raw := strings.TrimSpace(strings.TrimPrefix(line, "WorkflowOutputsJSON:"))
						var outs map[string]arazzo.OutputValue
						if json.Unmarshal([]byte(raw), &outs) == nil {
							outputMaps[wfPascal] = outs
						}
					}
				}
			}
		}
	}
}

// parseSDKExecuteMethod reconstructs an arazzo.Workflow from an Execute<WorkflowPascal> method declaration.
// Parameter fn is the FuncDecl AST node, inputSchemas maps workflow PascalCase names to Input schemas, and outputMaps maps workflow PascalCase names to OutputValue maps.
func parseSDKExecuteMethod(fn *dst.FuncDecl, inputSchemas map[string]*openapi.Schema, outputMaps map[string]map[string]arazzo.OutputValue) arazzo.Workflow {
	wfPascal := strings.TrimPrefix(fn.Name.Name, "Execute")
	wfID := strings.ToLower(wfPascal[:1]) + wfPascal[1:]

	wf := arazzo.Workflow{
		WorkflowID: wfID,
		Inputs:     inputSchemas[wfPascal],
		Outputs:    outputMaps[wfPascal],
	}

	for _, c := range fn.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		switch {
		case strings.HasPrefix(line, "WorkflowID:"):
			wf.WorkflowID = strings.TrimSpace(strings.TrimPrefix(line, "WorkflowID:"))
		case strings.HasPrefix(line, "Summary:"):
			wf.Summary = strings.TrimSpace(strings.TrimPrefix(line, "Summary:"))
		case strings.HasPrefix(line, "Description:"):
			wf.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
		case strings.HasPrefix(line, "DependsOn:"):
			csv := strings.TrimSpace(strings.TrimPrefix(line, "DependsOn:"))
			if csv != "" {
				wf.DependsOn = strings.Split(csv, ",")
			}
		case strings.HasPrefix(line, "WorkflowParametersJSON:"):
			raw := strings.TrimSpace(strings.TrimPrefix(line, "WorkflowParametersJSON:"))
			_ = json.Unmarshal([]byte(raw), &wf.Parameters)
		case strings.HasPrefix(line, "WorkflowSuccessActionsJSON:"):
			raw := strings.TrimSpace(strings.TrimPrefix(line, "WorkflowSuccessActionsJSON:"))
			_ = json.Unmarshal([]byte(raw), &wf.SuccessActions)
		case strings.HasPrefix(line, "WorkflowFailureActionsJSON:"):
			raw := strings.TrimSpace(strings.TrimPrefix(line, "WorkflowFailureActionsJSON:"))
			_ = json.Unmarshal([]byte(raw), &wf.FailureActions)
		case strings.HasPrefix(line, "WorkflowExtensionsJSON:"):
			raw := strings.TrimSpace(strings.TrimPrefix(line, "WorkflowExtensionsJSON:"))
			_ = json.Unmarshal([]byte(raw), &wf.Extensions)
		}
	}

	if fn.Body != nil {
		for _, stmt := range fn.Body.List {
			if ifStmt, ok := stmt.(*dst.IfStmt); ok && len(wf.DependsOn) == 0 {
				if assign, ok := ifStmt.Init.(*dst.AssignStmt); ok && len(assign.Rhs) > 0 {
					if call, ok := assign.Rhs[0].(*dst.CallExpr); ok {
						if sel, ok := call.Fun.(*dst.SelectorExpr); ok && strings.HasPrefix(sel.Sel.Name, "Execute") {
							depPascal := strings.TrimPrefix(sel.Sel.Name, "Execute")
							if depPascal != "" {
								wf.DependsOn = append(wf.DependsOn, strings.ToLower(depPascal[:1])+depPascal[1:])
							}
						}
					}
				}
			} else if block, ok := stmt.(*dst.BlockStmt); ok {
				step := parseSDKStepBlock(block)
				if step.StepID != "" {
					wf.Steps = append(wf.Steps, step)
				}
			} else if retStmt, ok := stmt.(*dst.ReturnStmt); ok && len(wf.Outputs) == 0 && len(retStmt.Results) > 0 {
				if unary, ok := retStmt.Results[0].(*dst.UnaryExpr); ok {
					if cl, ok := unary.X.(*dst.CompositeLit); ok && len(cl.Elts) > 0 {
						wf.Outputs = make(map[string]arazzo.OutputValue)
						for _, elt := range cl.Elts {
							if kv, ok := elt.(*dst.KeyValueExpr); ok {
								kIdent, kOk := kv.Key.(*dst.Ident)
								vLit, vOk := kv.Value.(*dst.BasicLit)
								if kOk && vOk {
									vStr, _ := strconv.Unquote(vLit.Value)
									wf.Outputs[naming.ToSnakeCase(kIdent.Name)] = arazzo.OutputValue{Expression: vStr}
								}
							}
						}
					}
				}
			}
		}
	}

	if len(wf.Steps) == 0 {
		wf.Steps = []arazzo.Step{
			{StepID: "defaultStep", OperationID: wf.WorkflowID + "Op"},
		}
	}

	return wf
}

// parseSDKStepBlock reconstructs an arazzo.Step from a step BlockStmt inside an Execute<WorkflowPascal> method.
// Parameter block is the step AST BlockStmt, and it returns the parsed arazzo.Step.
func parseSDKStepBlock(block *dst.BlockStmt) arazzo.Step {
	var step arazzo.Step
	for _, c := range block.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "StepJSON:") {
			raw := strings.TrimSpace(strings.TrimPrefix(line, "StepJSON:"))
			if json.Unmarshal([]byte(raw), &step) == nil && step.StepID != "" {
				return step
			}
		} else if strings.HasPrefix(line, "StepID:") {
			step.StepID = strings.TrimSpace(strings.TrimPrefix(line, "StepID:"))
		}
	}

	// Inspect AST statements in block for stepOutputs["<stepID>"] and http.NewRequestWithContext URL
	for _, s := range block.List {
		assign, ok := s.(*dst.AssignStmt)
		if !ok || len(assign.Lhs) == 0 || len(assign.Rhs) == 0 {
			continue
		}
		if idx, ok := assign.Lhs[0].(*dst.IndexExpr); ok {
			if ident, ok := idx.X.(*dst.Ident); ok && ident.Name == "stepOutputs" && step.StepID == "" {
				if lit, ok := idx.Index.(*dst.BasicLit); ok {
					step.StepID, _ = strconv.Unquote(lit.Value)
				}
			}
		}
		if call, ok := assign.Rhs[0].(*dst.CallExpr); ok && len(call.Args) >= 3 {
			if bin, ok := call.Args[2].(*dst.BinaryExpr); ok {
				if lit, ok := bin.Y.(*dst.BasicLit); ok {
					if pathStr, err := strconv.Unquote(lit.Value); err == nil {
						step.OperationID = strings.TrimPrefix(pathStr, "/")
					}
				}
			}
		}
	}

	if step.StepID != "" && step.OperationID == "" && step.OperationPath == "" && step.ChannelPath == "" && step.WorkflowID == "" {
		step.OperationID = step.StepID
	}
	return step
}
