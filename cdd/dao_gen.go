// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
package cdd

import (
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/daos"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// GenerateDAOs generates the DAO interfaces, stubs, concrete GORM implementations, and factory.
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any generation error.
func GenerateDAOs(oa *openapi.OpenAPI, outDir string) error {
	daosDir := filepath.Join(outDir, "daos")
	if err := os.MkdirAll(daosDir, 0755); err != nil {
		return err
	}

	var schemas map[string]openapi.Schema
	if oa != nil && oa.Components != nil && oa.Components.Schemas != nil {
		schemas = oa.Components.Schemas
	} else {
		schemas = make(map[string]openapi.Schema)
	}

	if len(schemas) > 0 {
		modelsDir := filepath.Join(outDir, "models")
		if err := os.MkdirAll(modelsDir, 0755); err != nil {
			return err
		}
		if err := WriteDstFile(filepath.Join(modelsDir, "errors.go"), daos.EmitModelsErrorsFile()); err != nil {
			return err
		}
	}

	schemaNames := make([]string, 0, len(schemas))
	for name := range schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	for _, name := range schemaNames {
		s := schemas[name]

		bundle, err := daos.EmitSchemaDAOs(name, &s)
		if err != nil {
			return err
		}

		// DAO Interface
		file := &dst.File{
			Name: dst.NewIdent("daos"),
			Decls: []dst.Decl{
				&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"context"`}},
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}},
				}},
				bundle.InterfaceDecl,
			},
		}
		if err := WriteDstFile(filepath.Join(daosDir, strings.ToLower(name)+"_dao.go"), file); err != nil {
			return err
		}

		// Stub DAO
		stubFile := &dst.File{
			Name: dst.NewIdent("daos"),
			Decls: append([]dst.Decl{
				&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"context"`}},
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}},
				}},
				bundle.StubDecl,
			}, bundle.StubMethods...),
		}
		if err := WriteDstFile(filepath.Join(daosDir, strings.ToLower(name)+"_stub.go"), stubFile); err != nil {
			return err
		}

		// Concrete DAO
		concreteFile := &dst.File{
			Name: dst.NewIdent("daos"),
			Decls: append([]dst.Decl{
				&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"context"`}},
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/gorm"`}},
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}},
				}},
				bundle.ConcreteDecl,
			}, bundle.ConcreteMethods...),
		}
		if err := WriteDstFile(filepath.Join(daosDir, strings.ToLower(name)+"_gorm.go"), concreteFile); err != nil {
			return err
		}
	}

	// Factory Interface
	factoryDecl := daos.EmitFactoryInterface(schemas)
	factoryFile := &dst.File{
		Name: dst.NewIdent("daos"),
		Decls: []dst.Decl{
			factoryDecl,
		},
	}
	if err := WriteDstFile(filepath.Join(daosDir, "factory.go"), factoryFile); err != nil {
		return err
	}

	// Concrete Factory
	concreteFactoryDecl, concreteFactoryMethods := daos.EmitConcreteFactory(schemas)
	concreteFactoryFile := &dst.File{
		Name: dst.NewIdent("daos"),
		Decls: append([]dst.Decl{
			&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/gorm"`}},
			}},
			concreteFactoryDecl,
		}, concreteFactoryMethods...),
	}
	if err := WriteDstFile(filepath.Join(daosDir, "factory_impl.go"), concreteFactoryFile); err != nil {
		return err
	}

	if len(schemas) > 0 {
		// Generate tests
		_, factoryTestMethods := daos.EmitFactoryTests(schemas)
		daoTestMethods := daos.EmitConcreteDAOTests(schemas)

		testFile := &dst.File{
			Name: dst.NewIdent("daos"),
			Decls: append(append([]dst.Decl{
				&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"context"`}},
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"testing"`}},
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/gorm"`}},
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/driver/sqlite"`}},
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}},
				}},
			}, factoryTestMethods...), daoTestMethods...),
		}
		if err := WriteDstFile(filepath.Join(daosDir, "daos_test.go"), testFile); err != nil {
			return err
		}
	}

	return nil
}
