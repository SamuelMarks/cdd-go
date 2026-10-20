package cdd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/arazzo"
)

func TestFromArazzoAndToArazzo(t *testing.T) {
	if err := FromArazzo("to_sdk", "", "", false, false, false); err == nil {
		t.Error("expected error on empty input")
	}
	if err := FromArazzo("to_sdk", "nonexistent_arazzo.json", "", false, false, false); err == nil {
		t.Error("expected error on missing input")
	}

	tmpDir := t.TempDir()
	invalidSyntaxFile := filepath.Join(tmpDir, "invalid.json")
	_ = os.WriteFile(invalidSyntaxFile, []byte(`{invalid`), 0644)
	if err := FromArazzo("to_sdk", invalidSyntaxFile, tmpDir, false, false, false); err == nil {
		t.Error("expected parse error")
	}

	invalidSemanticFile := filepath.Join(tmpDir, "invalid_sem.json")
	_ = os.WriteFile(invalidSemanticFile, []byte(`{"arazzo": "1.1.0", "info": {"title": "", "version": "1"}}`), 0644)
	if err := FromArazzo("to_sdk", invalidSemanticFile, tmpDir, false, false, false); err == nil {
		t.Error("expected validation error")
	}

	// Write a local OpenAPI spec and an Arazzo spec referencing it
	oaSpecPath := filepath.Join(tmpDir, "openapi.json")
	_ = os.WriteFile(oaSpecPath, []byte(`{"openapi": "3.2.0", "info": {"title": "Pet", "version": "1.0.0"}, "paths": {"/pets": {"get": {"operationId": "listPets"}}}}`), 0644)

	arazzoSpecPath := filepath.Join(tmpDir, "arazzo.json")
	arazzoJSON := `{
  "arazzo": "1.1.0",
  "info": {"title": "Pet Workflows", "version": "1.0.0"},
  "sourceDescriptions": [{"name": "pet", "url": "openapi.json", "type": "openapi"}],
  "workflows": [
    {
      "workflowId": "adoptPet",
      "steps": [
        {"stepId": "listStep", "operationId": "listPets"}
      ]
    }
  ]
}`
	_ = os.WriteFile(arazzoSpecPath, []byte(arazzoJSON), 0644)

	// 1. to_sdk with tests
	sdkDir := filepath.Join(tmpDir, "out_sdk")
	if err := FromArazzo("to_sdk", arazzoSpecPath, sdkDir, false, false, true); err != nil {
		t.Fatalf("FromArazzo to_sdk failed: %v", err)
	}

	// 2. to_sdk_cli with tests
	cliDir := filepath.Join(tmpDir, "out_cli")
	if err := FromArazzo("to_sdk_cli", arazzoSpecPath, cliDir, true, true, true); err != nil {
		t.Fatalf("FromArazzo to_sdk_cli failed: %v", err)
	}

	// 3. to_server
	srvDir := filepath.Join(tmpDir, "out_srv")
	if err := FromArazzo("to_server", arazzoSpecPath, srvDir, true, true, true); err != nil {
		t.Fatalf("FromArazzo to_server failed: %v", err)
	}

	// 4. unknown subsubcommand
	if err := FromArazzo("unknown_target", arazzoSpecPath, tmpDir, true, true, false); err == nil {
		t.Error("expected error on unknown subcommand")
	}

	// 5. ToArazzo from directory (to directory and to .json / .yaml files)
	if err := ToArazzo("", ""); err == nil {
		t.Error("expected error on empty input")
	}
	if err := ToArazzo("nonexistent_dir_xyz", filepath.Join(tmpDir, "out.json")); err == nil {
		t.Error("expected error on missing input")
	}

	outDirTarget := filepath.Join(tmpDir, "regen_dir")
	if err := ToArazzo(cliDir, outDirTarget); err != nil {
		t.Fatalf("ToArazzo to dir failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDirTarget, "arazzo.json")); err != nil {
		t.Errorf("expected arazzo.json in output dir: %v", err)
	}

	yamlTarget := filepath.Join(tmpDir, "regen.yaml")
	if err := ToArazzo(srvDir, yamlTarget); err != nil {
		t.Fatalf("ToArazzo to yaml failed: %v", err)
	}

	// Test ToArazzo from single CLI file and single Server test file
	singleCLI := filepath.Join(cliDir, "sdk_workflow_cli.go")
	if err := ToArazzo(singleCLI, filepath.Join(tmpDir, "from_cli.json")); err != nil {
		t.Fatalf("ToArazzo from single CLI file failed: %v", err)
	}

	singleSrvTest := filepath.Join(srvDir, "tests", "workflows_test.go")
	if err := ToArazzo(singleSrvTest, filepath.Join(tmpDir, "from_srv.json")); err != nil {
		t.Fatalf("ToArazzo from single server test file failed: %v", err)
	}

	// Test ToArazzo from an empty Go file (triggers defaultWorkflow fallback)
	emptyGoFile := filepath.Join(tmpDir, "empty.go")
	_ = os.WriteFile(emptyGoFile, []byte(`package main`), 0644)
	if err := ToArazzo(emptyGoFile, filepath.Join(tmpDir, "from_empty.json")); err != nil {
		t.Fatalf("ToArazzo from empty file failed: %v", err)
	}

	// Test ToArazzo parse error on corrupt Go file
	corruptGoFile := filepath.Join(tmpDir, "corrupt.go")
	_ = os.WriteFile(corruptGoFile, []byte(`package main; func broken(`), 0644)
	if err := ToArazzo(corruptGoFile, filepath.Join(tmpDir, "corrupt_out.json")); err == nil {
		t.Error("expected parse error on corrupt Go file")
	}

	// Test Sync with truth == "arazzo" and truth == "code" with arazzo.json output
	if err := Sync(arazzoSpecPath, filepath.Join(tmpDir, "sync_sdk"), "arazzo"); err != nil {
		t.Fatalf("Sync truth=arazzo failed: %v", err)
	}
	if err := Sync(sdkDir, filepath.Join(tmpDir, "sync_arazzo.json"), "code"); err != nil {
		t.Fatalf("Sync truth=code to arazzo.json failed: %v", err)
	}

	// Test ExecuteGeneratorRouter for Arazzo tools
	if msg, err := ExecuteGeneratorRouter("cdd_generate_arazzo_sdk", MCPToolParams{Input: arazzoSpecPath, Output: filepath.Join(tmpDir, "router_sdk")}); err != nil || !strings.Contains(msg, "Arazzo") {
		t.Errorf("unexpected ExecuteGeneratorRouter result: %q, %v", msg, err)
	}
	if msg, err := ExecuteGeneratorRouter("cdd_sync_arazzo", MCPToolParams{Input: sdkDir, Output: filepath.Join(tmpDir, "router_arazzo.json")}); err != nil || !strings.Contains(msg, "Arazzo") {
		t.Errorf("unexpected ExecuteGeneratorRouter result: %q, %v", msg, err)
	}

	// Test error branches in GenerateArazzoSDK, GenerateArazzoCLI, GenerateArazzoServerRoutes, GenerateArazzoTests
	blockingFile := filepath.Join(tmpDir, "blocker")
	_ = os.WriteFile(blockingFile, []byte("x"), 0644)
	validDoc, _ := arazzo.ParseBytes([]byte(arazzoJSON))

	if err := GenerateArazzoSDK(validDoc, blockingFile); err == nil {
		t.Error("expected MkdirAll error")
	}
	if err := GenerateArazzoCLI(validDoc, blockingFile); err == nil {
		t.Error("expected MkdirAll error")
	}
	if err := GenerateArazzoServerRoutes(validDoc, blockingFile); err == nil {
		t.Error("expected MkdirAll error")
	}
	if err := GenerateArazzoTests(validDoc, blockingFile); err == nil {
		t.Error("expected MkdirAll error")
	}

	if err := GenerateArazzoSDK(nil, tmpDir); err == nil {
		t.Error("expected error on nil doc")
	}
	if err := GenerateArazzoCLI(nil, tmpDir); err == nil {
		t.Error("expected error on nil doc")
	}
	if err := GenerateArazzoServerRoutes(nil, tmpDir); err == nil {
		t.Error("expected error on nil doc")
	}
	if err := GenerateArazzoTests(nil, tmpDir); err == nil {
		t.Error("expected error on nil doc")
	}

	// Test FromArazzo when output directory is a file (triggers sub-generator errors)
	if err := FromArazzo("to_sdk", arazzoSpecPath, blockingFile, true, true, false); err == nil {
		t.Error("expected error when outDir is a file for to_sdk")
	}
	if err := FromArazzo("to_sdk_cli", arazzoSpecPath, blockingFile, true, true, false); err == nil {
		t.Error("expected error when outDir is a file for to_sdk_cli")
	}
	if err := FromArazzo("to_server", arazzoSpecPath, blockingFile, true, true, false); err == nil {
		t.Error("expected error when outDir is a file for to_server")
	}

	// Test FromArazzo when tests directory is blocked by a file
	dirWithTestsBlocker := filepath.Join(tmpDir, "tests_blocked")
	_ = os.MkdirAll(dirWithTestsBlocker, 0755)
	_ = os.WriteFile(filepath.Join(dirWithTestsBlocker, "tests"), []byte("x"), 0644)
	if err := FromArazzo("to_sdk", arazzoSpecPath, dirWithTestsBlocker, true, true, true); err == nil {
		t.Error("expected error when tests dir is blocked for to_sdk")
	}
	if err := FromArazzo("to_sdk_cli", arazzoSpecPath, dirWithTestsBlocker, true, true, true); err == nil {
		t.Error("expected error when tests dir is blocked for to_sdk_cli")
	}
	if err := FromArazzo("to_server", arazzoSpecPath, dirWithTestsBlocker, true, true, true); err == nil {
		t.Error("expected error when tests dir is blocked for to_server")
	}

	// Test FromArazzo when routes directory is blocked by a file
	dirWithRoutesBlocker := filepath.Join(tmpDir, "routes_blocked")
	_ = os.MkdirAll(dirWithRoutesBlocker, 0755)
	_ = os.WriteFile(filepath.Join(dirWithRoutesBlocker, "routes"), []byte("x"), 0644)
	if err := FromArazzo("to_server", arazzoSpecPath, dirWithRoutesBlocker, true, true, true); err == nil {
		t.Error("expected error when routes dir is blocked for to_server")
	}

	// Test FromArazzo when sdk_workflow_cli.go is a directory
	dirWithCLIBlocker := filepath.Join(tmpDir, "cli_blocked")
	_ = os.MkdirAll(filepath.Join(dirWithCLIBlocker, "sdk_workflow_cli.go"), 0755)
	if err := FromArazzo("to_sdk_cli", arazzoSpecPath, dirWithCLIBlocker, true, true, false); err == nil {
		t.Error("expected error when sdk_workflow_cli.go is a directory")
	}

	wfCode, err := os.ReadFile(filepath.Join(sdkDir, "workflows", "workflows.go"))
	if err != nil {
		t.Fatalf("failed to read generated workflows.go: %v", err)
	}
	if strings.Contains(string(wfCode), "WorkflowJSON:") || strings.Contains(string(wfCode), "StepJSON:") {
		t.Errorf("did not expect comment-stuffed WorkflowJSON or StepJSON in generated workflows.go")
	}
}
