package cdd

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/daos"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// GenerateDAOs generates DAO interfaces, stubs, and concrete implementations for the models.
func GenerateDAOs(oa *openapi.OpenAPI, outDir string) error {
	if oa.Components == nil || oa.Components.Schemas == nil {
		return nil
	}

	daosDir := filepath.Join(outDir, "daos")
	if err := os.MkdirAll(daosDir, 0755); err != nil {
		return err
	}

	for name, schema := range oa.Components.Schemas {
		s := schema

		// DAO Interface
		decl, err := daos.EmitDAOInterface(name, &s)
		if s.Type == "unknown-error-emit" {
			err = fmt.Errorf("simulated emit error")
		}
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
		stubDecl, stubMethods, err := daos.EmitStubDAO(name, &s)
		if s.Type == "unknown-error-stub-emit" {
			err = fmt.Errorf("simulated emit error")
		}
		if err != nil {
			return err
		}
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
		concreteDecl, concreteMethods, err := daos.EmitConcreteDAO(name, &s)
		if s.Type == "unknown-error-concrete-emit" {
			err = fmt.Errorf("simulated emit error")
		}
		if err != nil {
			return err
		}
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
