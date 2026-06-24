package database

import (
	"bytes"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitDatabaseConfig(t *testing.T) {
	decl := EmitDatabaseConfig()
	if decl == nil {
		t.Fatal("expected non-nil decl")
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("database"),
		Decls: []dst.Decl{decl},
	}
	if err := restorer.Fprint(&buf, file); err != nil {
		t.Fatalf("failed to print decl: %v", err)
	}

	code := buf.String()
	if !bytes.Contains([]byte(code), []byte("Config struct")) {
		t.Errorf("missing Config struct in code: %s", code)
	}
	if !bytes.Contains([]byte(code), []byte("DatabaseURL string")) {
		t.Errorf("missing DatabaseURL field in code: %s", code)
	}
	if !bytes.Contains([]byte(code), []byte("Ephemeral")) || !bytes.Contains([]byte(code), []byte("bool")) {
		t.Errorf("missing Ephemeral field in code: %s", code)
	}
}

func TestEmitInitDB(t *testing.T) {
	decl := EmitInitDB()
	if decl == nil {
		t.Fatal("expected non-nil decl")
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("database"),
		Decls: []dst.Decl{decl},
	}
	if err := restorer.Fprint(&buf, file); err != nil {
		t.Fatalf("failed to print decl: %v", err)
	}

	code := buf.String()
	if !bytes.Contains([]byte(code), []byte("InitDB(cfg Config)")) {
		t.Errorf("missing InitDB function in code")
	}
	if !bytes.Contains([]byte(code), []byte("sqlite.Open")) {
		t.Errorf("missing sqlite.Open in code")
	}
	if !bytes.Contains([]byte(code), []byte("postgres.Open")) {
		t.Errorf("missing postgres.Open in code")
	}
}

func TestEmitMigrate(t *testing.T) {
	schemas := map[string]openapi.Schema{
		"User": {Type: "object"},
		"Post": {Type: "object"},
	}

	decl := EmitMigrate(schemas)
	if decl == nil {
		t.Fatal("expected non-nil decl")
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("database"),
		Decls: []dst.Decl{decl},
	}
	if err := restorer.Fprint(&buf, file); err != nil {
		t.Fatalf("failed to print decl: %v", err)
	}

	code := buf.String()
	if !bytes.Contains([]byte(code), []byte("Migrate(db *gorm.DB)")) {
		t.Errorf("missing Migrate function in code")
	}
	if !bytes.Contains([]byte(code), []byte("AutoMigrate(&User{})")) {
		t.Errorf("missing User migration in code")
	}
	if !bytes.Contains([]byte(code), []byte("AutoMigrate(&Post{})")) {
		t.Errorf("missing Post migration in code")
	}
}
