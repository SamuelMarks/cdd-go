// Package workflows provides AST emitting and parsing of Arazzo 1.1.0 workflows across Client SDKs, CLIs, and Server E2E test suites.
package workflows

import (
	"encoding/json"
	"fmt"
	"go/token"
	"reflect"
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
	outputContexts := make(map[string]map[string]string)

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *dst.GenDecl:
			parseSDKGenDecl(d, doc, inputSchemas, outputContexts)
		case *dst.FuncDecl:
			if strings.HasPrefix(d.Name.Name, "Execute") && d.Recv != nil {
				wf := parseSDKExecuteMethod(d, inputSchemas, outputContexts)
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
// Parameter d is the GenDecl, doc is the target Arazzo document, inputSchemas stores parsed input schemas, and outputContexts stores parsed selector contexts.
func parseSDKGenDecl(d *dst.GenDecl, doc *arazzo.Arazzo, inputSchemas map[string]*openapi.Schema, outputContexts map[string]map[string]string) {
	if d.Tok == token.VAR {
		for _, c := range d.Decs.Start {
			line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
			if strings.HasPrefix(line, "SelfURI:") {
				doc.Self = strings.TrimSpace(strings.TrimPrefix(line, "SelfURI:"))
			} else if strings.HasPrefix(line, "RootExtensionsJSON:") {
				raw := strings.TrimSpace(strings.TrimPrefix(line, "RootExtensionsJSON:"))
				_ = json.Unmarshal([]byte(raw), &doc.Extensions)
			} else if strings.HasPrefix(line, "InfoExtensionsJSON:") {
				raw := strings.TrimSpace(strings.TrimPrefix(line, "InfoExtensionsJSON:"))
				_ = json.Unmarshal([]byte(raw), &doc.Info.Extensions)
			}
		}

		sourceTypes := make(map[string]arazzo.SourceDescriptionType)
		for _, c := range d.Decs.Start {
			line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
			if strings.HasPrefix(line, "SourceType:") {
				kv := strings.TrimSpace(strings.TrimPrefix(line, "SourceType:"))
				if parts := strings.SplitN(kv, "=", 2); len(parts) == 2 {
					sourceTypes[parts[0]] = arazzo.SourceDescriptionType(parts[1])
				}
			}
		}

		for _, spec := range d.Specs {
			vs, ok := spec.(*dst.ValueSpec)
			if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
				continue
			}
			varName := vs.Names[0].Name
			switch varName {
			case "ArazzoVersion":
				if bl, ok := vs.Values[0].(*dst.BasicLit); ok && bl.Kind == token.STRING {
					if unq, err := strconv.Unquote(bl.Value); err == nil && unq != "" {
						doc.Arazzo = unq
					}
				}
			case "ArazzoTitle":
				if bl, ok := vs.Values[0].(*dst.BasicLit); ok && bl.Kind == token.STRING {
					doc.Info.Title, _ = strconv.Unquote(bl.Value)
				}
			case "ArazzoInfoVersion":
				if bl, ok := vs.Values[0].(*dst.BasicLit); ok && bl.Kind == token.STRING {
					doc.Info.Version, _ = strconv.Unquote(bl.Value)
				}
			case "ArazzoSummary":
				if bl, ok := vs.Values[0].(*dst.BasicLit); ok && bl.Kind == token.STRING {
					doc.Info.Summary, _ = strconv.Unquote(bl.Value)
				}
			case "ArazzoDescription":
				if bl, ok := vs.Values[0].(*dst.BasicLit); ok && bl.Kind == token.STRING {
					doc.Info.Description, _ = strconv.Unquote(bl.Value)
				}
			case "SourceDescriptionsMetadata":
				if cl, ok := vs.Values[0].(*dst.CompositeLit); ok {
					for _, elt := range cl.Elts {
						if kv, ok := elt.(*dst.KeyValueExpr); ok {
							kLit, kOk := kv.Key.(*dst.BasicLit)
							vLit, vOk := kv.Value.(*dst.BasicLit)
							if kOk && vOk {
								kStr, _ := strconv.Unquote(kLit.Value)
								vStr, _ := strconv.Unquote(vLit.Value)
								sdType := sourceTypes[kStr]
								if sdType == "" {
									sdType = arazzo.SourceTypeOpenAPI
								}
								doc.SourceDescriptions = append(doc.SourceDescriptions, arazzo.SourceDescription{
									Name: kStr,
									URL:  vStr,
									Type: sdType,
								})
							}
						}
					}
				}
			case "ComponentsMetadataJSON":
				if bl, ok := vs.Values[0].(*dst.BasicLit); ok && bl.Kind == token.STRING {
					if unq, err := strconv.Unquote(bl.Value); err == nil {
						var comp arazzo.Components
						if json.Unmarshal([]byte(unq), &comp) == nil {
							doc.Components = &comp
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
				if parsed, err := classes.ParseType(ts); err == nil && len(parsed.Properties) > 0 {
					parsed.Description = ""
					for k, prop := range parsed.Properties {
						prop.Description = ""
						parsed.Properties[k] = prop
					}
					inputSchemas[wfPascal] = parsed
				}
			} else if strings.HasSuffix(tName, "Output") {
				wfPascal := strings.TrimSuffix(tName, "Output")
				if st, ok := ts.Type.(*dst.StructType); ok && st.Fields != nil {
					ctxMap := make(map[string]string)
					for _, f := range st.Fields.List {
						if f.Tag != nil && len(f.Names) > 0 {
							rawTag, err := strconv.Unquote(f.Tag.Value)
							if err == nil {
								stTag := reflect.StructTag(rawTag)
								if ctxVal := stTag.Get("arazzo_context"); ctxVal != "" {
									jsonKey := strings.Split(stTag.Get("json"), ",")[0]
									ctxMap[jsonKey] = ctxVal
								}
							}
						}
					}
					if len(ctxMap) > 0 {
						outputContexts[wfPascal] = ctxMap
					}
				}
			}
		}
	}
}

// parseSDKExecuteMethod reconstructs an arazzo.Workflow from an Execute<WorkflowPascal> method declaration.
// Parameter fn is the FuncDecl AST node, inputSchemas maps workflow PascalCase names to Input schemas, and outputContexts maps workflow PascalCase names to selector contexts.
func parseSDKExecuteMethod(fn *dst.FuncDecl, inputSchemas map[string]*openapi.Schema, outputContexts map[string]map[string]string) arazzo.Workflow {
	wfPascal := strings.TrimPrefix(fn.Name.Name, "Execute")
	wfID := strings.ToLower(wfPascal[:1]) + wfPascal[1:]

	wf := arazzo.Workflow{
		WorkflowID: wfID,
		Inputs:     inputSchemas[wfPascal],
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
						ctxMap := outputContexts[wfPascal]
						for _, elt := range cl.Elts {
							if kv, ok := elt.(*dst.KeyValueExpr); ok {
								kIdent, kOk := kv.Key.(*dst.Ident)
								vLit, vOk := kv.Value.(*dst.BasicLit)
								if kOk && vOk {
									vStr, _ := strconv.Unquote(vLit.Value)
									outKey := naming.ToSnakeCase(kIdent.Name)
									if ctxMap != nil && ctxMap[outKey] != "" {
										wf.Outputs[outKey] = arazzo.OutputValue{
											Selector: &arazzo.Selector{
												Context:  ctxMap[outKey],
												Selector: vStr,
											},
										}
									} else {
										wf.Outputs[outKey] = arazzo.OutputValue{Expression: vStr}
									}
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

// parseSDKStepBlock reconstructs an arazzo.Step from a step BlockStmt inside an Execute<WorkflowPascal> method or E2E test function.
// Parameter block is the step AST BlockStmt, and it returns the parsed arazzo.Step.
func parseSDKStepBlock(block *dst.BlockStmt) arazzo.Step {
	var step arazzo.Step
	targetKind := "operationId"
	var headerTarget string

	for _, c := range block.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "StepID:") {
			rest := strings.TrimSpace(strings.TrimPrefix(line, "StepID:"))
			if openIdx := strings.Index(rest, "("); openIdx != -1 && strings.HasSuffix(rest, ")") {
				step.StepID = strings.TrimSpace(rest[:openIdx])
				meta := rest[openIdx+1 : len(rest)-1]
				if kv := strings.SplitN(meta, "=", 2); len(kv) == 2 {
					targetKind = kv[0]
					headerTarget = kv[1]
				}
			} else {
				step.StepID = rest
			}
		}
	}

	// Inspect AST statements in block for stepOutputs, context.WithTimeout, http.NewRequestWithContext, Content-Type header, and step outputs
	for _, s := range block.List {
		switch st := s.(type) {
		case *dst.AssignStmt:
			// Check stepOutputs["<stepID>"] initialization or stepOutputs["<stepID>"]["<outKey>"] = "<val>"
			if idx, ok := st.Lhs[0].(*dst.IndexExpr); ok {
				if ident, ok := idx.X.(*dst.Ident); ok && ident.Name == "stepOutputs" && step.StepID == "" {
					if lit, ok := idx.Index.(*dst.BasicLit); ok {
						step.StepID, _ = strconv.Unquote(lit.Value)
					}
				} else if innerIdx, ok := idx.X.(*dst.IndexExpr); ok {
					if ident, ok := innerIdx.X.(*dst.Ident); ok && ident.Name == "stepOutputs" {
						outKeyLit, kOk := idx.Index.(*dst.BasicLit)
						outValLit, vOk := st.Rhs[0].(*dst.BasicLit)
						if kOk && vOk {
							outKey, _ := strconv.Unquote(outKeyLit.Value)
							outVal, _ := strconv.Unquote(outValLit.Value)
							if step.Outputs == nil {
								step.Outputs = make(map[string]arazzo.OutputValue)
							}
							if strings.Contains(outVal, "#$.") {
								parts := strings.SplitN(outVal, "#", 2)
								step.Outputs[outKey] = arazzo.OutputValue{
									Selector: &arazzo.Selector{Context: parts[0], Selector: parts[1]},
								}
							} else {
								step.Outputs[outKey] = arazzo.OutputValue{Expression: outVal}
							}
						}
					}
				}
			}

			// Check context.WithTimeout(ctx, <ms> * time.Millisecond) or http.NewRequestWithContext / http.NewRequest
			if call, ok := st.Rhs[0].(*dst.CallExpr); ok {
				if sel, ok := call.Fun.(*dst.SelectorExpr); ok && sel.Sel.Name == "WithTimeout" && len(call.Args) >= 2 {
					if bin, ok := call.Args[1].(*dst.BinaryExpr); ok {
						if lit, ok := bin.X.(*dst.BasicLit); ok && lit.Kind == token.INT {
							if ms, err := strconv.ParseInt(lit.Value, 10, 64); err == nil {
								step.Timeout = &ms
							}
						}
					}
				}
				if len(call.Args) >= 3 {
					urlArgIdx := 2
					bodyArgIdx := 3
					if len(call.Args) == 3 {
						// http.NewRequest(method, url, body) or NewRequestWithContext(ctx, method, url)
						if _, isBin := call.Args[1].(*dst.BinaryExpr); isBin {
							urlArgIdx = 1
							bodyArgIdx = 2
						}
					}
					if bin, ok := call.Args[urlArgIdx].(*dst.BinaryExpr); ok {
						if lit, ok := bin.Y.(*dst.BasicLit); ok {
							if pathStr, err := strconv.Unquote(lit.Value); err == nil && headerTarget == "" {
								headerTarget = strings.TrimPrefix(pathStr, "/")
							}
						}
					}
					if bodyArgIdx < len(call.Args) {
						if bodyCall, ok := call.Args[bodyArgIdx].(*dst.CallExpr); ok && len(bodyCall.Args) == 1 {
							if payloadLit, ok := bodyCall.Args[0].(*dst.BasicLit); ok {
								if payloadStr, err := strconv.Unquote(payloadLit.Value); err == nil && payloadStr != "" && payloadStr != "{}" {
									step.RequestBody = &arazzo.RequestBody{
										Payload: json.RawMessage(payloadStr),
									}
								}
							}
						}
					}
				}
			}
		case *dst.ExprStmt:
			if call, ok := st.X.(*dst.CallExpr); ok && len(call.Args) == 2 {
				if sel, ok := call.Fun.(*dst.SelectorExpr); ok && sel.Sel.Name == "Set" {
					kLit, kOk := call.Args[0].(*dst.BasicLit)
					vLit, vOk := call.Args[1].(*dst.BasicLit)
					if kOk && vOk && kLit.Value == `"Content-Type"` {
						if step.RequestBody != nil {
							step.RequestBody.ContentType, _ = strconv.Unquote(vLit.Value)
						}
					}
				}
			}
		}
	}

	switch targetKind {
	case "operationPath":
		step.OperationPath = headerTarget
	case "channelPath":
		step.ChannelPath = headerTarget
	case "workflowId":
		step.WorkflowID = headerTarget
	default:
		step.OperationID = headerTarget
	}

	if step.StepID != "" && step.OperationID == "" && step.OperationPath == "" && step.ChannelPath == "" && step.WorkflowID == "" {
		step.OperationID = step.StepID
	}
	return step
}
