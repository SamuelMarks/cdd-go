// Package arazzo provides strongly-typed structures, parsers, validators, and emitters for the Arazzo Specification v1.1.0.
package arazzo

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ghodss/yaml"
)

// extractExtensions filters a raw JSON object map for specification extensions starting with "x-".
// Parameter raw is the top-level key-value map of the JSON object, and it returns the extension map or nil if none exist.
func extractExtensions(raw map[string]json.RawMessage) map[string]json.RawMessage {
	var ext map[string]json.RawMessage
	for k, v := range raw {
		if strings.HasPrefix(k, "x-") {
			if ext == nil {
				ext = make(map[string]json.RawMessage)
			}
			ext[k] = v
		}
	}
	return ext
}

// populateDocumentExtensions walks the raw JSON representation of an Arazzo document to capture "^x-" extensions at all levels.
// Parameter data is the JSON byte slice, and doc is the target Arazzo struct to populate.
func populateDocumentExtensions(data []byte, doc *Arazzo) {
	var rawRoot map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawRoot); err != nil {
		return
	}
	doc.Extensions = extractExtensions(rawRoot)

	if rawInfoBytes, ok := rawRoot["info"]; ok {
		var rawInfo map[string]json.RawMessage
		if json.Unmarshal(rawInfoBytes, &rawInfo) == nil {
			doc.Info.Extensions = extractExtensions(rawInfo)
		}
	}

	if rawSDBytes, ok := rawRoot["sourceDescriptions"]; ok {
		var rawSDs []map[string]json.RawMessage
		if json.Unmarshal(rawSDBytes, &rawSDs) == nil {
			for i := 0; i < len(rawSDs) && i < len(doc.SourceDescriptions); i++ {
				doc.SourceDescriptions[i].Extensions = extractExtensions(rawSDs[i])
			}
		}
	}

	if rawCompBytes, ok := rawRoot["components"]; ok && doc.Components != nil {
		var rawComp map[string]json.RawMessage
		if json.Unmarshal(rawCompBytes, &rawComp) == nil {
			doc.Components.Extensions = extractExtensions(rawComp)
		}
	}

	if rawWfBytes, ok := rawRoot["workflows"]; ok {
		var rawWfs []map[string]json.RawMessage
		if json.Unmarshal(rawWfBytes, &rawWfs) == nil {
			for i := 0; i < len(rawWfs) && i < len(doc.Workflows); i++ {
				doc.Workflows[i].Extensions = extractExtensions(rawWfs[i])
				if rawStepsBytes, ok := rawWfs[i]["steps"]; ok {
					var rawSteps []map[string]json.RawMessage
					if json.Unmarshal(rawStepsBytes, &rawSteps) == nil {
						for j := 0; j < len(rawSteps) && j < len(doc.Workflows[i].Steps); j++ {
							doc.Workflows[i].Steps[j].Extensions = extractExtensions(rawSteps[j])
						}
					}
				}
			}
		}
	}
}

// Parse reads an Arazzo 1.1.0 specification from the provided io.Reader and unmarshals it into an Arazzo struct.
// Parameter r is the input stream (supporting JSON or YAML), and it returns the parsed Arazzo document and any error.
func Parse(r io.Reader) (*Arazzo, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseBytes(data)
}

// ParseBytes unmarshals a byte slice containing JSON or YAML into an Arazzo 1.1.0 struct.
// Parameter data is the raw specification payload, and it returns the parsed Arazzo document and any error.
func ParseBytes(data []byte) (*Arazzo, error) {
	var doc Arazzo
	errJSON := json.Unmarshal(data, &doc)
	if errJSON == nil && doc.Arazzo != "" {
		populateDocumentExtensions(data, &doc)
		return &doc, nil
	}

	jsonBytes, errYAMLConv := yaml.YAMLToJSON(data)
	if errYAMLConv == nil {
		var yamlDoc Arazzo
		if errUnmarshal := json.Unmarshal(jsonBytes, &yamlDoc); errUnmarshal == nil && yamlDoc.Arazzo != "" {
			populateDocumentExtensions(jsonBytes, &yamlDoc)
			return &yamlDoc, nil
		}
	}

	if errJSON != nil {
		return nil, errJSON
	}
	return nil, fmt.Errorf("missing required 'arazzo' version field in specification")
}

// Validate checks an Arazzo 1.1.0 document for structural conformance, unique identifiers, reference integrity, and DAG acyclicity.
// Parameter doc is the Arazzo specification to validate, and it returns an error if any constraint is violated.
func Validate(doc *Arazzo) error {
	if doc == nil {
		return fmt.Errorf("arazzo document is nil")
	}
	if !strings.HasPrefix(doc.Arazzo, "1.0.") && !strings.HasPrefix(doc.Arazzo, "1.1.") {
		return fmt.Errorf("unsupported arazzo version %q: must be 1.0.x or 1.1.x", doc.Arazzo)
	}
	if strings.TrimSpace(doc.Info.Title) == "" {
		return fmt.Errorf("arazzo info.title is required")
	}
	if strings.TrimSpace(doc.Info.Version) == "" {
		return fmt.Errorf("arazzo info.version is required")
	}
	if len(doc.SourceDescriptions) == 0 {
		return fmt.Errorf("arazzo sourceDescriptions must contain at least one entry")
	}
	sdNames := make(map[string]bool, len(doc.SourceDescriptions))
	for _, sd := range doc.SourceDescriptions {
		if strings.TrimSpace(sd.Name) == "" || strings.TrimSpace(sd.URL) == "" {
			return fmt.Errorf("sourceDescription name and url are required")
		}
		if sd.Type != "" && sd.Type != SourceTypeOpenAPI && sd.Type != SourceTypeAsyncAPI && sd.Type != SourceTypeArazzo {
			return fmt.Errorf("invalid sourceDescription type %q for %q", sd.Type, sd.Name)
		}
		if sdNames[sd.Name] {
			return fmt.Errorf("duplicate sourceDescription name %q", sd.Name)
		}
		sdNames[sd.Name] = true
	}

	if len(doc.Workflows) == 0 {
		return fmt.Errorf("arazzo workflows must contain at least one workflow")
	}

	wfMap := make(map[string]Workflow, len(doc.Workflows))
	for _, wf := range doc.Workflows {
		if strings.TrimSpace(wf.WorkflowID) == "" {
			return fmt.Errorf("workflowId is required")
		}
		if _, exists := wfMap[wf.WorkflowID]; exists {
			return fmt.Errorf("duplicate workflowId %q", wf.WorkflowID)
		}
		wfMap[wf.WorkflowID] = wf
	}

	// Validate workflow-level DAG dependencies and check for cycles
	for _, wf := range doc.Workflows {
		for _, dep := range wf.DependsOn {
			if _, exists := wfMap[dep]; !exists {
				return fmt.Errorf("workflow %q dependsOn unknown workflowId %q", wf.WorkflowID, dep)
			}
		}
		for _, param := range wf.Parameters {
			if err := validateParameterOrReusable(param, doc.Components); err != nil {
				return fmt.Errorf("workflow %q parameter error: %w", wf.WorkflowID, err)
			}
		}
		for _, sa := range wf.SuccessActions {
			if err := validateSuccessActionOrReusable(sa, doc.Components); err != nil {
				return fmt.Errorf("workflow %q successAction error: %w", wf.WorkflowID, err)
			}
		}
		for _, fa := range wf.FailureActions {
			if err := validateFailureActionOrReusable(fa, doc.Components); err != nil {
				return fmt.Errorf("workflow %q failureAction error: %w", wf.WorkflowID, err)
			}
		}
		if err := validateWorkflowSteps(wf, doc.Components); err != nil {
			return err
		}
	}

	return checkWorkflowCycles(wfMap)
}

// validateWorkflowSteps validates the steps of a single Workflow, including step targets, references, and step-level DAG acyclicity.
// Parameter wf is the Workflow to inspect, and comps is the document's Components registry.
func validateWorkflowSteps(wf Workflow, comps *Components) error {
	if len(wf.Steps) == 0 {
		return fmt.Errorf("workflow %q must contain at least one step", wf.WorkflowID)
	}
	stepMap := make(map[string]Step, len(wf.Steps))
	for _, step := range wf.Steps {
		if strings.TrimSpace(step.StepID) == "" {
			return fmt.Errorf("workflow %q contains a step with an empty stepId", wf.WorkflowID)
		}
		if _, exists := stepMap[step.StepID]; exists {
			return fmt.Errorf("workflow %q contains duplicate stepId %q", wf.WorkflowID, step.StepID)
		}
		targetCount := 0
		if step.OperationID != "" {
			targetCount++
		}
		if step.OperationPath != "" {
			targetCount++
		}
		if step.ChannelPath != "" {
			targetCount++
		}
		if step.WorkflowID != "" {
			targetCount++
		}
		if targetCount != 1 {
			return fmt.Errorf("step %q in workflow %q must specify exactly one of operationId, operationPath, channelPath, or workflowId (got %d)", step.StepID, wf.WorkflowID, targetCount)
		}
		for _, p := range step.Parameters {
			if err := validateParameterOrReusable(p, comps); err != nil {
				return fmt.Errorf("step %q parameter error: %w", step.StepID, err)
			}
		}
		for _, sa := range step.OnSuccess {
			if err := validateSuccessActionOrReusable(sa, comps); err != nil {
				return fmt.Errorf("step %q onSuccess error: %w", step.StepID, err)
			}
		}
		for _, fa := range step.OnFailure {
			if err := validateFailureActionOrReusable(fa, comps); err != nil {
				return fmt.Errorf("step %q onFailure error: %w", step.StepID, err)
			}
		}
		stepMap[step.StepID] = step
	}

	for _, step := range wf.Steps {
		for _, dep := range step.DependsOn {
			if _, exists := stepMap[dep]; !exists {
				return fmt.Errorf("step %q in workflow %q dependsOn unknown stepId %q", step.StepID, wf.WorkflowID, dep)
			}
		}
	}
	return checkStepCycles(wf.WorkflowID, stepMap)
}

// validateParameterOrReusable verifies that an inline Parameter or a $components.parameters reference is valid.
// Parameter pr is the ParameterOrReusable union, and comps is the Components registry.
func validateParameterOrReusable(pr ParameterOrReusable, comps *Components) error {
	if pr.Reusable != nil {
		const prefix = "$components.parameters."
		if !strings.HasPrefix(pr.Reusable.Reference, prefix) {
			return fmt.Errorf("invalid parameter reference %q: must start with %s", pr.Reusable.Reference, prefix)
		}
		key := strings.TrimPrefix(pr.Reusable.Reference, prefix)
		if comps == nil || comps.Parameters == nil {
			return fmt.Errorf("parameter reference %q cannot be resolved: components.parameters is empty", pr.Reusable.Reference)
		}
		if _, ok := comps.Parameters[key]; !ok {
			return fmt.Errorf("parameter reference %q not found in components.parameters", pr.Reusable.Reference)
		}
		return nil
	}
	if pr.Parameter == nil || strings.TrimSpace(pr.Parameter.Name) == "" {
		return fmt.Errorf("parameter name is required")
	}
	return nil
}

// validateSuccessActionOrReusable verifies that an inline SuccessAction or a $components.successActions reference is valid.
// Parameter sr is the SuccessActionOrReusable union, and comps is the Components registry.
func validateSuccessActionOrReusable(sr SuccessActionOrReusable, comps *Components) error {
	if sr.Reusable != nil {
		const prefix = "$components.successActions."
		if !strings.HasPrefix(sr.Reusable.Reference, prefix) {
			return fmt.Errorf("invalid successAction reference %q: must start with %s", sr.Reusable.Reference, prefix)
		}
		key := strings.TrimPrefix(sr.Reusable.Reference, prefix)
		if comps == nil || comps.SuccessActions == nil {
			return fmt.Errorf("successAction reference %q cannot be resolved: components.successActions is empty", sr.Reusable.Reference)
		}
		if _, ok := comps.SuccessActions[key]; !ok {
			return fmt.Errorf("successAction reference %q not found in components.successActions", sr.Reusable.Reference)
		}
		return nil
	}
	if sr.Action == nil || strings.TrimSpace(sr.Action.Name) == "" {
		return fmt.Errorf("successAction name is required")
	}
	if sr.Action.Type != SuccessTypeEnd && sr.Action.Type != SuccessTypeGoto {
		return fmt.Errorf("invalid successAction type %q", sr.Action.Type)
	}
	return nil
}

// validateFailureActionOrReusable verifies that an inline FailureAction or a $components.failureActions reference is valid.
// Parameter fr is the FailureActionOrReusable union, and comps is the Components registry.
func validateFailureActionOrReusable(fr FailureActionOrReusable, comps *Components) error {
	if fr.Reusable != nil {
		const prefix = "$components.failureActions."
		if !strings.HasPrefix(fr.Reusable.Reference, prefix) {
			return fmt.Errorf("invalid failureAction reference %q: must start with %s", fr.Reusable.Reference, prefix)
		}
		key := strings.TrimPrefix(fr.Reusable.Reference, prefix)
		if comps == nil || comps.FailureActions == nil {
			return fmt.Errorf("failureAction reference %q cannot be resolved: components.failureActions is empty", fr.Reusable.Reference)
		}
		if _, ok := comps.FailureActions[key]; !ok {
			return fmt.Errorf("failureAction reference %q not found in components.failureActions", fr.Reusable.Reference)
		}
		return nil
	}
	if fr.Action == nil || strings.TrimSpace(fr.Action.Name) == "" {
		return fmt.Errorf("failureAction name is required")
	}
	if fr.Action.Type != FailureTypeEnd && fr.Action.Type != FailureTypeRetry && fr.Action.Type != FailureTypeGoto {
		return fmt.Errorf("invalid failureAction type %q", fr.Action.Type)
	}
	return nil
}

// checkWorkflowCycles performs a depth-first cycle detection pass over workflow DependsOn edges.
// Parameter wfMap maps workflowId to Workflow, and it returns an error if a cycle is detected.
func checkWorkflowCycles(wfMap map[string]Workflow) error {
	state := make(map[string]int, len(wfMap))
	var visit func(id string) error
	visit = func(id string) error {
		state[id] = 1
		for _, dep := range wfMap[id].DependsOn {
			if state[dep] == 1 {
				return fmt.Errorf("cyclic workflow dependency detected involving %q and %q", id, dep)
			}
			if state[dep] == 0 {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		state[id] = 2
		return nil
	}
	for id := range wfMap {
		if state[id] == 0 {
			if err := visit(id); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkStepCycles performs a depth-first cycle detection pass over step DependsOn edges within a workflow.
// Parameter wfID is the enclosing workflow identifier, stepMap maps stepId to Step, and it returns an error if a cycle is detected.
func checkStepCycles(wfID string, stepMap map[string]Step) error {
	state := make(map[string]int, len(stepMap))
	var visit func(id string) error
	visit = func(id string) error {
		state[id] = 1
		for _, dep := range stepMap[id].DependsOn {
			if state[dep] == 1 {
				return fmt.Errorf("cyclic step dependency in workflow %q involving %q and %q", wfID, id, dep)
			}
			if state[dep] == 0 {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		state[id] = 2
		return nil
	}
	for id := range stepMap {
		if state[id] == 0 {
			if err := visit(id); err != nil {
				return err
			}
		}
	}
	return nil
}
