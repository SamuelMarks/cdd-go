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
				"ID":     {Type: "string"}, // hits "id" check
				"name":   {Type: "string"},
				"count":  {Type: "integer"},
				"active": {Type: "boolean"},
				"other":  {Type: "object"}, // hits default fallback
				"":       {Type: "string"}, // hits empty string check
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
	if !bytes.Contains([]byte(code), []byte("FakeUser(pool EntityPool)")) {
		t.Errorf("missing FakeUser func in code")
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
	if !bytes.Contains([]byte(code), []byte("gofakeit.Bool()")) {
		t.Errorf("missing gofakeit.Bool() in code")
	}
}
