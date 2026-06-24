package cdd

import (
	"go/token"
	"os"
	"path/filepath"

	"github.com/SamuelMarks/cdd-go/src/database"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// GenerateDatabase generates the database connection and migration routines.
func GenerateDatabase(oa *openapi.OpenAPI, outDir string) error {
	dbDir := filepath.Join(outDir, "database")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return err
	}

	configDecl := database.EmitDatabaseConfig()
	initDBDecl := database.EmitInitDB()
	var migrateDecl *dst.FuncDecl
	if oa.Components != nil && oa.Components.Schemas != nil {
		migrateDecl = database.EmitMigrate(oa.Components.Schemas)
	} else {
		migrateDecl = database.EmitMigrate(make(map[string]openapi.Schema))
	}

	decls := []dst.Decl{
		&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/gorm"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/driver/sqlite"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/driver/postgres"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}},
		}},
		configDecl,
		initDBDecl,
		migrateDecl,
	}

	file := &dst.File{
		Name:  dst.NewIdent("database"),
		Decls: decls,
	}

	return WriteDstFile(filepath.Join(dbDir, "database.go"), file)
}
