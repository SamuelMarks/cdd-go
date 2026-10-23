package cdd

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

func TestGenerateDAOs(t *testing.T) {
	outDir := t.TempDir()
	oa := &openapi.OpenAPI{
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{
				"User": {
					Type: "object",
					Properties: map[string]openapi.Schema{
						"id":   {Type: "string"},
						"name": {Type: "string"},
					},
				},
			},
		},
	}

	if err := GenerateClasses(oa, outDir); err != nil {
		t.Fatalf("unexpected GenerateClasses error: %v", err)
	}
	if err := GenerateDAOs(oa, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFiles := []string{
		"user_dao.go", "user_stub.go", "user_gorm.go", "factory.go", "factory_impl.go", "daos_test.go",
	}

	fset := token.NewFileSet()
	for _, f := range expectedFiles {
		p := filepath.Join(outDir, "daos", f)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("expected file %s to exist: %v", f, err)
		}
		if _, err := parser.ParseFile(fset, p, b, parser.ParseComments); err != nil {
			t.Errorf("generated file %s is not valid Go syntax: %v", f, err)
		}
	}

	daoContent, _ := os.ReadFile(filepath.Join(outDir, "daos", "user_dao.go"))
	if !strings.Contains(string(daoContent), `"generated_sdk/models"`) || strings.Contains(string(daoContent), `User "generated_sdk/models"`) {
		t.Errorf("expected unaliased generated_sdk/models import in user_dao.go, got:\n%s", string(daoContent))
	}
	if !strings.Contains(string(daoContent), `*models.User`) {
		t.Errorf("expected *models.User in user_dao.go, got:\n%s", string(daoContent))
	}

	if _, err := os.Stat(filepath.Join(outDir, "models", "errors.go")); os.IsNotExist(err) {
		t.Errorf("expected models/errors.go to be generated")
	}

	emptyDir := t.TempDir()
	if err := GenerateDAOs(&openapi.OpenAPI{}, emptyDir); err != nil {
		t.Errorf("unexpected error with empty openapi: %v", err)
	}
	if _, err := os.Stat(filepath.Join(emptyDir, "daos", "factory.go")); os.IsNotExist(err) {
		t.Errorf("expected factory.go even with empty schemas")
	}
}

func TestGenerateDAOs_Errors(t *testing.T) {
	oa := &openapi.OpenAPI{Components: &openapi.Components{Schemas: map[string]openapi.Schema{"User": {Type: "object"}}}}
	if err := GenerateDAOs(oa, "/dev/null/invalid"); err == nil {
		t.Errorf("expected error for invalid directory")
	}

	outDir := t.TempDir()
	daosDir := filepath.Join(outDir, "daos")
	os.MkdirAll(daosDir, 0755)

	// Block creating models directory
	os.WriteFile(filepath.Join(outDir, "models"), []byte("blocker"), 0644)
	if err := GenerateDAOs(oa, outDir); err == nil {
		t.Errorf("expected error when models is a file")
	}
	os.RemoveAll(filepath.Join(outDir, "models"))

	// Block writing models/errors.go
	os.MkdirAll(filepath.Join(outDir, "models", "errors.go"), 0755)
	if err := GenerateDAOs(oa, outDir); err == nil {
		t.Errorf("expected error when models/errors.go is a directory")
	}
	os.RemoveAll(filepath.Join(outDir, "models", "errors.go"))

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

	oaErr := &openapi.OpenAPI{Components: &openapi.Components{Schemas: map[string]openapi.Schema{"ErrorSchema": {Type: "unknown-error-emit"}}}}
	if err := GenerateDAOs(oaErr, outDir); err == nil {
		t.Errorf("expected error for unknown-error")
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

	dbPath := filepath.Join(outDir, "database", "database.go")
	b, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("expected database.go to exist: %v", err)
	}
	if !strings.Contains(string(b), "AutoMigrate(&models.User{})") {
		t.Errorf("expected AutoMigrate(&models.User{}) in database.go, got:\n%s", string(b))
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
				"User": {
					Type: "object",
					Properties: map[string]openapi.Schema{
						"id":   {Type: "string"},
						"name": {Type: "string"},
					},
				},
			},
		},
	}

	if err := GenerateSeeder(oa, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seederPath := filepath.Join(outDir, "seeder", "seeder.go")
	b, err := os.ReadFile(seederPath)
	if err != nil {
		t.Fatalf("expected seeder.go to exist: %v", err)
	}
	if !strings.Contains(string(b), "*models.User") || !strings.Contains(string(b), "&models.User{") {
		t.Errorf("expected qualified *models.User and &models.User in seeder.go, got:\n%s", string(b))
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
	oa := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/users": openapi.PathItem{
				Get: &openapi.Operation{OperationID: "listUsers"},
			},
		},
	}

	if err := GenerateServerMain(oa, outDir, false, false, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	main1, err := os.ReadFile(filepath.Join(outDir, "cmd", "server", "main.go"))
	if err != nil {
		t.Fatalf("expected main.go to exist: %v", err)
	}
	if !strings.Contains(string(main1), "routes.NewHandlerUsers(nil).Register(r)") {
		t.Errorf("expected route registration with nil DAO factory in main.go, got:\n%s", string(main1))
	}

	outDir2 := t.TempDir()
	oaWithHooks := &openapi.OpenAPI{
		Info: openapi.Info{Title: "My Test API"},
		Paths: openapi.Paths{
			"/users": openapi.PathItem{
				Get: &openapi.Operation{OperationID: "listUsers"},
			},
		},
		Webhooks: map[string]openapi.PathItem{
			"orderCreated": {
				Post: &openapi.Operation{OperationID: "onOrderCreated"},
			},
		},
	}
	if err := GenerateServerMain(oaWithHooks, outDir2, true, true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	main2, err := os.ReadFile(filepath.Join(outDir2, "cmd", "server", "main.go"))
	if err != nil {
		t.Fatalf("expected main.go to exist: %v", err)
	}
	if !strings.Contains(string(main2), "routes.NewHandlerUsers(daoFactory).Register(r)") {
		t.Errorf("expected route registration with daoFactory in main.go, got:\n%s", string(main2))
	}
	if !strings.Contains(string(main2), `"/_mock/trigger-webhook/:webhook_name"`) {
		t.Errorf("expected webhook trigger route in main.go, got:\n%s", string(main2))
	}

	outDir3 := t.TempDir()
	if err := GenerateServerMain(oa, outDir3, true, false, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseServerFlagsMatrix(t *testing.T) {
	// 1. Stub Mode: No DB configured, no flags
	cfg, err := ParseServerFlags([]string{}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.UseConcreteDAOs || cfg.Ephemeral || cfg.Seed || cfg.StrictValidation || cfg.EnforceAuth {
		t.Errorf("expected Stub Mode (all false), got %+v", cfg)
	}

	// 2. Production Mode: DATABASE_URL configured, no flags
	cfg, err = ParseServerFlags([]string{}, "postgres://localhost:5432/prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.UseConcreteDAOs || cfg.Ephemeral || cfg.Seed {
		t.Errorf("expected Production Mode (UseConcreteDAOs=true, Ephemeral=false), got %+v", cfg)
	}

	// 3. Sandbox Mode: --ephemeral flag
	cfg, err = ParseServerFlags([]string{"--ephemeral"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.UseConcreteDAOs || !cfg.Ephemeral || cfg.Seed {
		t.Errorf("expected Sandbox Mode (UseConcreteDAOs=true, Ephemeral=true, Seed=false), got %+v", cfg)
	}

	// 4. Full Mock Mode: --ephemeral --seed --strict-validation --enforce-auth
	cfg, err = ParseServerFlags([]string{"--ephemeral", "--seed", "--strict-validation", "--enforce-auth"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.UseConcreteDAOs || !cfg.Ephemeral || !cfg.Seed || !cfg.StrictValidation || !cfg.EnforceAuth {
		t.Errorf("expected Full Mock Mode (all true), got %+v", cfg)
	}

	// 5. Invalid flag error
	if _, err := ParseServerFlags([]string{"--unknown-flag"}, ""); err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestGenerateWebhookTriggersAndServerReadme(t *testing.T) {
	outDir := t.TempDir()

	// Empty OpenAPI -> no webhooks directory created
	if err := GenerateWebhookTriggers(&openapi.OpenAPI{}, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "webhooks", "trigger.go")); !os.IsNotExist(err) {
		t.Error("did not expect webhooks/trigger.go when spec has no webhooks or callbacks")
	}

	// Spec with webhooks and callbacks
	oa := &openapi.OpenAPI{
		Info: openapi.Info{Title: "Webhook API"},
		Webhooks: map[string]openapi.PathItem{
			"userCreated": {},
		},
	}
	if err := GenerateWebhookTriggers(oa, outDir); err != nil {
		t.Fatalf("unexpected GenerateWebhookTriggers error: %v", err)
	}
	triggerPath := filepath.Join(outDir, "webhooks", "trigger.go")
	b, err := os.ReadFile(triggerPath)
	if err != nil {
		t.Fatalf("expected webhooks/trigger.go: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), triggerPath, b, parser.ParseComments); err != nil {
		t.Errorf("generated webhooks/trigger.go is not valid Go: %v", err)
	}
	if !strings.Contains(string(b), "HandleTriggerWebhook") || !strings.Contains(string(b), "DispatchWebhook") {
		t.Errorf("expected HandleTriggerWebhook and DispatchWebhook in trigger.go:\n%s", string(b))
	}

	// Error creating webhooks directory
	if err := GenerateWebhookTriggers(oa, "/dev/null/invalid"); err == nil {
		t.Error("expected error for invalid directory")
	}

	// GenerateServerReadme with title & webhooks, and with nil OpenAPI
	if err := GenerateServerReadme(oa, outDir); err != nil {
		t.Fatalf("unexpected GenerateServerReadme error: %v", err)
	}
	readmeBytes, err := os.ReadFile(filepath.Join(outDir, "README.md"))
	if err != nil {
		t.Fatalf("expected README.md: %v", err)
	}
	readmeStr := string(readmeBytes)
	for _, expected := range []string{"Stub Mode", "Production Mode", "Sandbox Mode", "Full Mock Mode", "POST /_mock/trigger-webhook/{webhook_name}", "userCreated"} {
		if !strings.Contains(readmeStr, expected) {
			t.Errorf("expected %q in generated README.md, got:\n%s", expected, readmeStr)
		}
	}

	if err := GenerateServerReadme(nil, t.TempDir()); err != nil {
		t.Fatalf("unexpected GenerateServerReadme(nil) error: %v", err)
	}
}

func TestGenerateDatabase_EmptySchemas(t *testing.T) {
	outDir := t.TempDir()
	oa := &openapi.OpenAPI{}
	if err := GenerateDatabase(oa, outDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(outDir, "database", "database.go"))
	if err != nil {
		t.Fatalf("expected database.go: %v", err)
	}
	if strings.Contains(string(b), `"generated_sdk/models"`) {
		t.Errorf("did not expect unused generated_sdk/models import when schemas is empty:\n%s", string(b))
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
