// Package cdd provides the core orchestration for OpenAPI, Arazzo 1.1.0, and Go AST code generation.
package cdd

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/SamuelMarks/cdd-go/src/workflows"
	"github.com/dave/dst/decorator"
)

// FromArazzo generates Go Client SDKs, CLIs (reusing the SDK), or Server E2E workflow suites from an Arazzo 1.1.0 specification.
// Parameter subsubcommand selects the target ("to_sdk", "to_sdk_cli", "to_server"), in is the input Arazzo spec path, outDir is the destination directory, noGithubActions omits CI workflow files, noInstallablePackage omits go.mod, and tests enables E2E test suite generation.
func FromArazzo(subsubcommand, in, outDir string, noGithubActions, noInstallablePackage, tests bool) error {
	if in == "" {
		return fmt.Errorf("input file or directory is required")
	}

	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()

	doc, err := arazzo.Parse(f)
	if err != nil {
		return err
	}
	if err := arazzo.Validate(doc); err != nil {
		return err
	}

	_ = os.MkdirAll(outDir, 0755)

	if !noInstallablePackage {
		GenerateGoMod(outDir)
	}
	if !noGithubActions {
		GenerateGithubActions(outDir)
	}

	// If any sourceDescription points to a local OpenAPI file relative to the Arazzo file, generate its models/clients too
	specDir := filepath.Dir(in)
	for _, sd := range doc.SourceDescriptions {
		if sd.Type == "" || sd.Type == arazzo.SourceTypeOpenAPI {
			localPath := sd.URL
			if !filepath.IsAbs(localPath) {
				localPath = filepath.Join(specDir, sd.URL)
			}
			if info, statErr := os.Stat(localPath); statErr == nil && !info.IsDir() {
				_ = FromOpenAPI(subsubcommand, localPath, outDir, true, true, tests, false)
			}
		}
	}

	switch subsubcommand {
	case "to_sdk":
		if err := GenerateArazzoSDK(doc, outDir); err != nil {
			return err
		}
		if tests {
			if err := GenerateArazzoTests(doc, outDir); err != nil {
				return err
			}
		}
	case "to_sdk_cli":
		if err := GenerateArazzoSDK(doc, outDir); err != nil {
			return err
		}
		if err := GenerateArazzoCLI(doc, outDir); err != nil {
			return err
		}
		if tests {
			if err := GenerateArazzoTests(doc, outDir); err != nil {
				return err
			}
		}
	case "to_server":
		if err := GenerateArazzoSDK(doc, outDir); err != nil {
			return err
		}
		if err := GenerateArazzoServerRoutes(doc, outDir); err != nil {
			return err
		}
		if err := GenerateArazzoTests(doc, outDir); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown subcommand for from_arazzo: %s", subsubcommand)
	}

	return nil
}

// GenerateArazzoSDK emits the `workflows/workflows.go` SDK workflow orchestrator package from an Arazzo 1.1.0 specification.
// Parameter doc is the parsed Arazzo document, outDir is the destination directory, and it returns any generation error.
func GenerateArazzoSDK(doc *arazzo.Arazzo, outDir string) error {
	wfDir := filepath.Join(outDir, "workflows")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		return err
	}
	file, err := workflows.EmitSDKWorkflows(doc)
	if err != nil {
		return err
	}
	return WriteDstFile(filepath.Join(wfDir, "workflows.go"), file)
}

// GenerateArazzoCLI emits the `sdk_workflow_cli.go` Cobra CLI workflow runner (reusing the SDK `workflows` package) from an Arazzo 1.1.0 specification.
// Parameter doc is the parsed Arazzo document, outDir is the destination directory, and it returns any generation error.
func GenerateArazzoCLI(doc *arazzo.Arazzo, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	file, err := workflows.EmitCLIWorkflows(doc)
	if err != nil {
		return err
	}
	return WriteDstFile(filepath.Join(outDir, "sdk_workflow_cli.go"), file)
}

// GenerateArazzoServerRoutes emits the `routes/workflows_routes.go` Gin workflow orchestrator routes from an Arazzo 1.1.0 specification.
// Parameter doc is the parsed Arazzo document, outDir is the destination directory, and it returns any generation error.
func GenerateArazzoServerRoutes(doc *arazzo.Arazzo, outDir string) error {
	routesDir := filepath.Join(outDir, "routes")
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		return err
	}
	file, err := workflows.EmitServerWorkflowRoutes(doc)
	if err != nil {
		return err
	}
	return WriteDstFile(filepath.Join(routesDir, "workflows_routes.go"), file)
}

// GenerateArazzoTests emits the `tests/workflows_test.go` end-to-end workflow test suite from an Arazzo 1.1.0 specification.
// Parameter doc is the parsed Arazzo document, outDir is the destination directory, and it returns any generation error.
func GenerateArazzoTests(doc *arazzo.Arazzo, outDir string) error {
	testsDir := filepath.Join(outDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		return err
	}
	file, err := workflows.EmitServerWorkflowTests(doc)
	if err != nil {
		return err
	}
	return WriteDstFile(filepath.Join(testsDir, "workflows_test.go"), file)
}

// ToArazzo extracts an Arazzo 1.1.0 specification from Go source code (SDK workflows, CLI commands, or Server E2E workflow tests).
// Parameter in is the input Go file or directory path, outPath is the output Arazzo file or directory path, and it returns any extraction error.
func ToArazzo(in, outPath string) error {
	if in == "" {
		return fmt.Errorf("input file or directory is required")
	}

	if filepath.Ext(outPath) == "" {
		_ = os.MkdirAll(outPath, 0755)
		outPath = filepath.Join(outPath, "arazzo.json")
	} else {
		_ = os.MkdirAll(filepath.Dir(outPath), 0755)
	}

	return GenerateArazzo(in, outPath)
}

// GenerateArazzo walks the input Go file or directory, parses all workflow AST constructs, and writes the resulting Arazzo 1.1.0 document to outPath.
// Parameter inputPath is the Go source path, outPath is the output file path (.json or .yaml/.yml), and it returns any error.
func GenerateArazzo(inputPath, outPath string) error {
	var goFiles []string
	info, err := os.Stat(inputPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		_ = filepath.Walk(inputPath, func(p string, fi os.FileInfo, walkErr error) error {
			if walkErr == nil && !fi.IsDir() && strings.HasSuffix(p, ".go") {
				goFiles = append(goFiles, p)
			}
			return nil
		})
	} else if strings.HasSuffix(inputPath, ".go") {
		goFiles = append(goFiles, inputPath)
	}

	var resultDoc *arazzo.Arazzo
	var extraWorkflows []arazzo.Workflow
	var cliInfo *arazzo.Info

	fset := token.NewFileSet()
	for _, fpath := range goFiles {
		file, parseErr := decorator.ParseFile(fset, fpath, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("failed to parse %s: %w", fpath, parseErr)
		}

		base := filepath.Base(fpath)
		switch {
		case base == "workflows.go" || (file.Name != nil && file.Name.Name == "workflows"):
			sdkDoc, _ := workflows.ParseSDKWorkflows(file)
			if len(sdkDoc.Workflows) > 0 {
				resultDoc = sdkDoc
			}
		case base == "sdk_workflow_cli.go" || strings.HasSuffix(base, "cli.go"):
			wfs, parsedInfo, _ := workflows.ParseCLIWorkflows(file)
			if len(wfs) > 0 {
				extraWorkflows = append(extraWorkflows, wfs...)
				cliInfo = parsedInfo
			}
		default:
			wfs, _ := workflows.ParseServerWorkflows(file)
			if len(wfs) > 0 {
				extraWorkflows = append(extraWorkflows, wfs...)
			}
		}
	}

	if resultDoc == nil {
		resultDoc = &arazzo.Arazzo{
			Arazzo: "1.1.0",
			Info: arazzo.Info{
				Title:   "Generated Arazzo Workflows",
				Version: "1.0.0",
			},
			SourceDescriptions: []arazzo.SourceDescription{
				{Name: "default", URL: "./openapi.json", Type: arazzo.SourceTypeOpenAPI},
			},
		}
		if cliInfo != nil && cliInfo.Title != "" {
			resultDoc.Info = *cliInfo
		}
	}

	seenWf := make(map[string]bool, len(resultDoc.Workflows))
	for _, wf := range resultDoc.Workflows {
		seenWf[wf.WorkflowID] = true
	}
	for _, wf := range extraWorkflows {
		if !seenWf[wf.WorkflowID] {
			resultDoc.Workflows = append(resultDoc.Workflows, wf)
			seenWf[wf.WorkflowID] = true
		}
	}

	if len(resultDoc.Workflows) == 0 {
		resultDoc.Workflows = []arazzo.Workflow{
			{
				WorkflowID: "defaultWorkflow",
				Steps: []arazzo.Step{
					{StepID: "defaultStep", OperationID: "defaultOp"},
				},
			},
		}
	}

	var buf bytes.Buffer
	ext := strings.ToLower(filepath.Ext(outPath))
	if ext == ".yaml" || ext == ".yml" {
		_ = arazzo.EmitYAML(&buf, resultDoc)
	} else {
		_ = arazzo.Emit(&buf, resultDoc)
	}

	return os.WriteFile(outPath, buf.Bytes(), 0644)
}
