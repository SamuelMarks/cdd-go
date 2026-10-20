// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
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
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any generation error.
func GenerateDatabase(oa *openapi.OpenAPI, outDir string) error {
	dbDir := filepath.Join(outDir, "database")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return err
	}

	configDecl := database.EmitDatabaseConfig()
	initDBDecl := database.EmitInitDB()
	var schemas map[string]openapi.Schema
	if oa.Components != nil && len(oa.Components.Schemas) > 0 {
		schemas = oa.Components.Schemas
	} else {
		schemas = make(map[string]openapi.Schema)
	}
	migrateDecl := database.EmitMigrate(schemas)

	importSpecs := []dst.Spec{
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/gorm"`}},
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/driver/sqlite"`}},
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/driver/postgres"`}},
	}
	if len(schemas) > 0 {
		importSpecs = append(importSpecs, &dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}})
	}

	decls := []dst.Decl{
		&dst.GenDecl{Tok: token.IMPORT, Specs: importSpecs},
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
