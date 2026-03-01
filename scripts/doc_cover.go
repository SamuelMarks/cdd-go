//go:build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	var total, doced int

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.IsExported() {
					total++
					if d.Doc != nil {
						doced++
					}
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if ts.Name.IsExported() {
							total++
							if d.Doc != nil || ts.Doc != nil {
								doced++
							}
						}
					}
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							if name.IsExported() {
								total++
								if d.Doc != nil || vs.Doc != nil {
									doced++
								}
							}
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if total == 0 {
		fmt.Printf("100.0\n")
		return
	}

	percentage := float64(doced) / float64(total) * 100
	fmt.Printf("%.1f\n", percentage)
}
