package cdd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

func TestGenerateDAOs(t *testing.T) {
	outDir := t.TempDir()
	oa := &openapi.OpenAPI{
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{
				"User": {Type: "object"},
			},
		},
	}

	if err := GenerateDAOs(oa, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFiles := []string{
		"user_dao.go", "user_stub.go", "user_gorm.go", "factory.go", "factory_impl.go",
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(outDir, "daos", f)); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}

	if err := GenerateDAOs(&openapi.OpenAPI{}, outDir); err != nil {
		t.Errorf("unexpected error with empty openapi: %v", err)
	}
}

func TestGenerateDAOs_Errors(t *testing.T) {
	// Test MkdirAll error
	oa := &openapi.OpenAPI{Components: &openapi.Components{Schemas: map[string]openapi.Schema{"User": {Type: "object"}}}}
	if err := GenerateDAOs(oa, "/dev/null/invalid"); err == nil {
		t.Errorf("expected error for invalid directory")
	}

	outDir := t.TempDir()
	daosDir := filepath.Join(outDir, "daos")
	os.MkdirAll(daosDir, 0755)

	// Block writing user_dao.go
	os.MkdirAll(filepath.Join(daosDir, "user_dao.go"), 0755)
	if err := GenerateDAOs(oa, outDir); err == nil {
		t.Errorf("expected error for write user_dao.go")
	}
	os.RemoveAll(filepath.Join(daosDir, "user_dao.go"))

	// Block writing user_stub.go
	os.MkdirAll(filepath.Join(daosDir, "user_stub.go"), 0755)
	if err := GenerateDAOs(oa, outDir); err == nil {
		t.Errorf("expected error for write user_stub.go")
	}
	os.RemoveAll(filepath.Join(daosDir, "user_stub.go"))

	// Block writing user_gorm.go
	os.MkdirAll(filepath.Join(daosDir, "user_gorm.go"), 0755)
	if err := GenerateDAOs(oa, outDir); err == nil {
		t.Errorf("expected error for write user_gorm.go")
	}
	os.RemoveAll(filepath.Join(daosDir, "user_gorm.go"))

	// Block writing factory.go
	os.MkdirAll(filepath.Join(daosDir, "factory.go"), 0755)
	if err := GenerateDAOs(oa, outDir); err == nil {
		t.Errorf("expected error for write factory.go")
	}
	os.RemoveAll(filepath.Join(daosDir, "factory.go"))

	// Block writing factory_impl.go
	os.MkdirAll(filepath.Join(daosDir, "factory_impl.go"), 0755)
	if err := GenerateDAOs(oa, outDir); err == nil {
		t.Errorf("expected error for write factory_impl.go")
	}
	os.RemoveAll(filepath.Join(daosDir, "factory_impl.go"))

	// Block writing daos_test.go
	os.MkdirAll(filepath.Join(daosDir, "daos_test.go"), 0755)
	if err := GenerateDAOs(oa, outDir); err == nil {
		t.Errorf("expected error for write daos_test.go")
	}
	os.RemoveAll(filepath.Join(daosDir, "daos_test.go"))

	// Test Emit error paths (simulated via missing required schema components or unknown type in emit helpers if they handled it, but here we can just pass an empty map element to trigger schema == nil error where applicable)
	oaErr := &openapi.OpenAPI{Components: &openapi.Components{Schemas: map[string]openapi.Schema{"ErrorSchema": {Type: "unknown-error-emit"}}}}
	if err := GenerateDAOs(oaErr, outDir); err == nil {
		t.Errorf("expected error for unknown-error")
	}

	oaErrStub := &openapi.OpenAPI{Components: &openapi.Components{Schemas: map[string]openapi.Schema{"ErrorSchema": {Type: "unknown-error-stub-emit"}}}}
	if err := GenerateDAOs(oaErrStub, outDir); err == nil {
		t.Errorf("expected error for unknown-error-stub")
	}

	oaErrConcrete := &openapi.OpenAPI{Components: &openapi.Components{Schemas: map[string]openapi.Schema{"ErrorSchema": {Type: "unknown-error-concrete-emit"}}}}
	if err := GenerateDAOs(oaErrConcrete, outDir); err == nil {
		t.Errorf("expected error for unknown-error-concrete")
	}
}

func TestGenerateDatabase(t *testing.T) {
	outDir := t.TempDir()
	oa := &openapi.OpenAPI{
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{
				"User": {Type: "object"},
			},
		},
	}

	if err := GenerateDatabase(oa, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "database", "database.go")); os.IsNotExist(err) {
		t.Errorf("expected database.go to exist")
	}
}

func TestGenerateDatabase_Errors(t *testing.T) {
	oa := &openapi.OpenAPI{Components: &openapi.Components{Schemas: map[string]openapi.Schema{"User": {Type: "object"}}}}
	if err := GenerateDatabase(oa, "/dev/null/invalid"); err == nil {
		t.Errorf("expected error for invalid directory")
	}

	outDir := t.TempDir()
	os.MkdirAll(filepath.Join(outDir, "database"), 0555) // read-only
	if err := GenerateDatabase(oa, outDir); err == nil {
		t.Errorf("expected error for read-only directory")
	}
	os.Chmod(filepath.Join(outDir, "database"), 0755) // restore
}

func TestGenerateSeeder(t *testing.T) {
	outDir := t.TempDir()
	oa := &openapi.OpenAPI{
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{
				"User": {Type: "object"},
			},
		},
	}

	if err := GenerateSeeder(oa, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "seeder", "seeder.go")); os.IsNotExist(err) {
		t.Errorf("expected seeder.go to exist")
	}

	if err := GenerateSeeder(&openapi.OpenAPI{}, outDir); err != nil {
		t.Errorf("unexpected error with empty openapi: %v", err)
	}
}

func TestGenerateSeeder_Errors(t *testing.T) {
	oa := &openapi.OpenAPI{Components: &openapi.Components{Schemas: map[string]openapi.Schema{"User": {Type: "object"}}}}
	if err := GenerateSeeder(oa, "/dev/null/invalid"); err == nil {
		t.Errorf("expected error for invalid directory")
	}

	outDir := t.TempDir()
	os.MkdirAll(filepath.Join(outDir, "seeder"), 0555) // read-only
	if err := GenerateSeeder(oa, outDir); err == nil {
		t.Errorf("expected error for read-only directory")
	}
	os.Chmod(filepath.Join(outDir, "seeder"), 0755) // restore
}

func TestGenerateServerMain(t *testing.T) {
	outDir := t.TempDir()
	oa := &openapi.OpenAPI{}

	if err := GenerateServerMain(oa, outDir, false, false, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "cmd", "server", "main.go")); os.IsNotExist(err) {
		t.Errorf("expected main.go to exist")
	}

	outDir2 := t.TempDir()
	if err := GenerateServerMain(oa, outDir2, true, true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir2, "cmd", "server", "main.go")); os.IsNotExist(err) {
		t.Errorf("expected main.go to exist")
	}
}

func TestGenerateDatabase_EmptySchemas(t *testing.T) {
	outDir := t.TempDir()
	oa := &openapi.OpenAPI{}
	if err := GenerateDatabase(oa, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateServerMain_Errors(t *testing.T) {
	oa := &openapi.OpenAPI{}
	if err := GenerateServerMain(oa, "/dev/null/invalid", true, true, true); err == nil {
		t.Errorf("expected error for invalid directory")
	}

	outDir := t.TempDir()
	os.MkdirAll(filepath.Join(outDir, "cmd", "server"), 0555) // read-only
	if err := GenerateServerMain(oa, outDir, true, true, true); err == nil {
		t.Errorf("expected error for read-only directory")
	}
	os.Chmod(filepath.Join(outDir, "cmd", "server"), 0755) // restore
}
