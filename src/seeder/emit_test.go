package seeder

import (
	"bytes"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitSeeder(t *testing.T) {
	schemas := map[string]openapi.Schema{
		"User": {
			Type: "object",
			Properties: map[string]openapi.Schema{
				"id":     {Type: "string"},
				"ID":     {Type: "string"},
				"name":   {Type: "string"},
				"count":  {Type: "integer"},
				"score":  {Type: "number"},
				"active": {Type: "boolean"},
				"tags":   {Type: "array", Items: &openapi.Schema{Type: "string"}},
				"parent": {Ref: "#/components/schemas/User"},
				"other":  {Type: "object"},
				"":       {Type: "string"},
			},
		},
		"AuditLog": {
			Type: "object",
			Properties: map[string]openapi.Schema{
				"message": {Type: "string"},
			},
		},
		"Order": {
			Type: "object",
			Properties: map[string]openapi.Schema{
				"id":          {Type: "string"},
				"user_id":     {Type: "string"},
				"auditLogId":  {Type: "string"},
				"unknown_id":  {Type: "string"},
				"description": {Type: "custom"},
			},
		},
	}

	decls := EmitSeeder(schemas)
	if len(decls) == 0 {
		t.Fatal("expected non-empty decls")
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("seeder"),
		Decls: decls,
	}
	if err := restorer.Fprint(&buf, file); err != nil {
		t.Fatalf("failed to print decls: %v", err)
	}

	code := buf.String()
	if !bytes.Contains([]byte(code), []byte("EntityPool struct")) {
		t.Errorf("missing EntityPool in code")
	}
	if !bytes.Contains([]byte(code), []byte("SeedDatabase(db *gorm.DB)")) {
		t.Errorf("missing SeedDatabase func in code")
	}
	if !bytes.Contains([]byte(code), []byte("FakeUser(pool EntityPool) *models.User")) {
		t.Errorf("missing FakeUser func returning *models.User in code: %s", code)
	}
	if !bytes.Contains([]byte(code), []byte("&models.User{")) {
		t.Errorf("missing &models.User composite literal in code: %s", code)
	}
	if !bytes.Contains([]byte(code), []byte("gofakeit.UUID()")) {
		t.Errorf("missing gofakeit.UUID() in code")
	}
	if !bytes.Contains([]byte(code), []byte("gofakeit.Word()")) {
		t.Errorf("missing gofakeit.Word() in code")
	}
	if !bytes.Contains([]byte(code), []byte("gofakeit.Number(1, 100)")) {
		t.Errorf("missing gofakeit.Number(1, 100) in code")
	}
	if !bytes.Contains([]byte(code), []byte("gofakeit.Float64Range(1.0, 100.0)")) {
		t.Errorf("missing gofakeit.Float64Range(1.0, 100.0) in code")
	}
	if !bytes.Contains([]byte(code), []byte("gofakeit.Bool()")) {
		t.Errorf("missing gofakeit.Bool() in code")
	}
}
