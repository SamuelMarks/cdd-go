package cdd

import (
	"go/token"
	"os"
	"path/filepath"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/seeder"
	"github.com/dave/dst"
)

// GenerateSeeder generates the seeding routines to populate an ephemeral DB with fake data.
func GenerateSeeder(oa *openapi.OpenAPI, outDir string) error {
	if oa.Components == nil || oa.Components.Schemas == nil {
		return nil
	}

	seedDir := filepath.Join(outDir, "seeder")
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		return err
	}

	decls := seeder.EmitSeeder(oa.Components.Schemas)

	file := &dst.File{
		Name: dst.NewIdent("seeder"),
		Decls: append([]dst.Decl{
			&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/gorm"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/brianvoe/gofakeit/v6"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}},
			}},
		}, decls...),
	}

	return WriteDstFile(filepath.Join(seedDir, "seeder.go"), file)
}
