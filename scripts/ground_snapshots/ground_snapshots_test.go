package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGroundSnapshots(t *testing.T) {
	// 1. Default run with non-existent directory falls back to canonical OpenAPI & Arazzo snapshots
	if err := runGrounding([]string{"-snapshots-dir", "nonexistent_snapshots_dir"}); err != nil {
		t.Fatalf("unexpected error on canonical fallback: %v", err)
	}

	// 2. Invalid flag
	if err := runGrounding([]string{"-invalid-flag"}); err == nil {
		t.Error("expected error on invalid flag")
	}

	// 3. Test with ML_FRAMEWORK_SNAPSHOTS_DIR env var and populated directory
	snapDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(snapDir, "spec.json"), []byte(canonicalOpenAPISnapshot), 0644)
	_ = os.WriteFile(filepath.Join(snapDir, "workflow.yaml"), []byte(canonicalArazzoSnapshot), 0644)
	_ = os.WriteFile(filepath.Join(snapDir, "unrelated.json"), []byte(`{"hello": "world"}`), 0644)

	t.Setenv("ML_FRAMEWORK_SNAPSHOTS_DIR", snapDir)
	if err := runGrounding([]string{}); err != nil {
		t.Fatalf("unexpected error on populated snapDir: %v", err)
	}

	// 4. Test verifySnapshotFile error paths
	if err := verifySnapshotFile("nonexistent_file.json", snapDir); err == nil {
		t.Error("expected error on missing file")
	}

	// Invalid Arazzo file (has workflows but missing info.title so Validate fails in FromArazzo)
	badDir := t.TempDir()
	badAzPath := filepath.Join(badDir, "bad_arazzo.json")
	_ = os.WriteFile(badAzPath, []byte(`{"arazzo": "1.1.0", "info": {"title": "", "version": "1"}, "workflows": [{"workflowId": "w", "steps": [{"stepId": "s", "operationId": "o"}]}]}`), 0644)
	if err := runGrounding([]string{"-snapshots-dir", badDir}); err == nil {
		t.Error("expected error on invalid Arazzo snapshot")
	}

	// Test FromArazzo to_sdk_cli failure by blocking az_cli
	validAzPath := filepath.Join(snapDir, "valid_az.json")
	_ = os.WriteFile(validAzPath, []byte(canonicalArazzoSnapshot), 0644)
	workBlockCLI := t.TempDir()
	_ = os.WriteFile(filepath.Join(workBlockCLI, "az_cli"), []byte("x"), 0644)
	if err := verifySnapshotFile(validAzPath, workBlockCLI); err == nil {
		t.Error("expected error when az_cli is blocked")
	}

	// Test FromArazzo to_server failure by blocking az_srv
	workBlockSrv := t.TempDir()
	_ = os.WriteFile(filepath.Join(workBlockSrv, "az_srv"), []byte("x"), 0644)
	if err := verifySnapshotFile(validAzPath, workBlockSrv); err == nil {
		t.Error("expected error when az_srv is blocked")
	}

	// Test ToArazzo failure by blocking regen_arazzo.json as a directory
	workBlockRegen := t.TempDir()
	_ = os.MkdirAll(filepath.Join(workBlockRegen, "regen_arazzo.json"), 0755)
	if err := verifySnapshotFile(validAzPath, workBlockRegen); err == nil {
		t.Error("expected error when regen_arazzo.json is a directory")
	}

	// Test FromOpenAPI failure by blocking oa_sdk
	validOAPath := filepath.Join(snapDir, "spec.json")
	workBlockOA := t.TempDir()
	_ = os.MkdirAll(filepath.Join(workBlockOA, "oa_sdk", "models", "trainingconfig.go"), 0755)
	if err := verifySnapshotFile(validOAPath, workBlockOA); err == nil {
		t.Error("expected error when oa_sdk models file is blocked")
	}

	// Test ToOpenAPI failure by blocking regen_openapi.json as a directory
	workBlockOARegen := t.TempDir()
	_ = os.MkdirAll(filepath.Join(workBlockOARegen, "regen_openapi.json", "openapi.json"), 0755)
	if err := verifySnapshotFile(validOAPath, workBlockOARegen); err == nil {
		t.Error("expected error when regen_openapi.json is a directory")
	}

	// 5. Test main() success and failure
	exitCode := 0
	oldExit := osExit
	osExit = func(c int) { exitCode = c }
	defer func() { osExit = oldExit }()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ground_snapshots", "-snapshots-dir", snapDir}
	main()
	if exitCode != 0 {
		t.Errorf("expected exitCode 0, got %d", exitCode)
	}

	os.Args = []string{"ground_snapshots", "-invalid-flag"}
	main()
	if exitCode != 1 {
		t.Errorf("expected exitCode 1, got %d", exitCode)
	}
}
