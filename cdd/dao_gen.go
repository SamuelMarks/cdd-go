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
	if oa.Components == nil || oa.Components.Schemas == nil {
		return nil
	}

	daosDir := filepath.Join(outDir, "daos")
	if err := os.MkdirAll(daosDir, 0755); err != nil {
		return err
	}

	schemaNames := make([]string, 0, len(oa.Components.Schemas))
	for name := range oa.Components.Schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	for _, name := range schemaNames {
		s := oa.Components.Schemas[name]

		// DAO Interface
		decl, err := daos.EmitDAOInterface(name, &s)
		if err != nil {
			return err
		}
		file := &dst.File{
			Name: dst.NewIdent("daos"),
			Decls: []dst.Decl{
				&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"context"`}},
					&dst.ImportSpec{Name: dst.NewIdent(name), Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}},
				}},
				decl,
			},
		}
		if err := WriteDstFile(filepath.Join(daosDir, strings.ToLower(name)+"_dao.go"), file); err != nil {
			return err
		}

		// Stub DAO
		stubDecl, stubMethods, _ := daos.EmitStubDAO(name, &s)
		stubFile := &dst.File{
			Name: dst.NewIdent("daos"),
			Decls: append([]dst.Decl{
				&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"context"`}},
					&dst.ImportSpec{Name: dst.NewIdent(name), Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}},
				}},
				stubDecl,
			}, stubMethods...),
		}
		if err := WriteDstFile(filepath.Join(daosDir, strings.ToLower(name)+"_stub.go"), stubFile); err != nil {
			return err
		}

		// Concrete DAO
		concreteDecl, concreteMethods, _ := daos.EmitConcreteDAO(name, &s)
		concreteFile := &dst.File{
			Name: dst.NewIdent("daos"),
			Decls: append([]dst.Decl{
				&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"context"`}},
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"gorm.io/gorm"`}},
					&dst.ImportSpec{Name: dst.NewIdent(name), Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/models"`}},
				}},
				concreteDecl,
			}, concreteMethods...),
		}
		if err := WriteDstFile(filepath.Join(daosDir, strings.ToLower(name)+"_gorm.go"), concreteFile); err != nil {
			return err
		}
	}

	// Factory Interface
	factoryDecl := daos.EmitFactoryInterface(oa.Components.Schemas)
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
	concreteFactoryDecl, concreteFactoryMethods := daos.EmitConcreteFactory(oa.Components.Schemas)
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

	// Generate tests
	_, factoryTestMethods := daos.EmitFactoryTests(oa.Components.Schemas)
	daoTestMethods := daos.EmitConcreteDAOTests(oa.Components.Schemas)

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

	return nil
}
